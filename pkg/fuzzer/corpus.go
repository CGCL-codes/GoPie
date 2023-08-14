package fuzzer

import (
	"fmt"
	"math/rand"
	"sync"
	"sync/atomic"
	"toolkit/pkg"
)

var (
	allowDup = 100
	AttackP  = 40
)

const BANMAX = 100

type Chain struct {
	item []*pkg.Pair
}

func (c *Chain) Copy() *Chain {
	res := Chain{item: []*pkg.Pair{}}
	if c == nil {
		return &res
	}
	for _, p := range c.item {
		res.add(p)
	}
	return &res
}

func (c *Chain) ToString() string {
	res := ""
	if c == nil {
		return res
	}
	if c.item == nil {
		return res
	}
	for _, e := range c.item {
		if e == nil {
			continue
		}
		res += e.ToString()
	}
	return res
}

func (c *Chain) Len() int {
	if c == nil {
		return 0
	}
	return len(c.item)
}

func (c *Chain) G() *Chain {
	if c.Len() <= 1 {
		return nil
	}
	return &Chain{c.item[0 : len(c.item)-1]}
}

// tail
func (c *Chain) T() *pkg.Pair {
	if c.Len() == 0 {
		return nil
	}
	return c.item[len(c.item)-1]
}

func (c *Chain) pop() *pkg.Pair {
	if c.Len() == 0 {
		return nil
	}
	last := c.item[len(c.item)-1]
	c.item = c.item[0 : len(c.item)-1]
	return last
}

func (c *Chain) add(e *pkg.Pair) {
	if e != nil {
		c.item = append(c.item, e)
	}
}

func (c *Chain) merge(cc *Chain) {
	p1 := 0
	p2 := 0
	for {
		if p1 < c.Len() && p2 < cc.Len() {
			if rand.Int()%2 == 0 {
				p1 += 1
			} else {
				p2 += 1
			}
			if p1 < c.Len() && p2 < cc.Len() && rand.Int()%2 == 0 {
				temp := c.item[p1]
				c.item[p1] = c.item[p2]
				c.item[p2] = temp
			}
		}
	}
}

type Corpus struct {
	gm map[uint32]*Chain
	// map prev to nexts
	tm              map[uint64]map[uint64]struct{}
	covered         map[uint64]uint64
	preban          map[uint32]uint64
	ban             map[uint32]struct{}
	allow           map[uint32]struct{}
	schedCovered    sync.Map
	schedCoveredCnt uint64
	fetchCnt        uint64
	hash            sync.Map
	gmu             sync.RWMutex
	tmu             sync.RWMutex
	bmu             sync.RWMutex
	cmu             sync.RWMutex
}

var once sync.Once
var GlobalCorpus Corpus

func init() {
	GlobalCorpus = Corpus{}
	GlobalCorpus.Init()
}

func (cp *Corpus) Init() {
	once.Do(func() {
		GlobalCorpus.gm = make(map[uint32]*Chain)
		GlobalCorpus.tm = make(map[uint64]map[uint64]struct{})
		GlobalCorpus.ban = make(map[uint32]struct{})
		GlobalCorpus.preban = make(map[uint32]uint64)
		GlobalCorpus.allow = make(map[uint32]struct{})
		GlobalCorpus.covered = make(map[uint64]uint64)
	})
}

func NewCorpus() *Corpus {
	corpus := &Corpus{}
	corpus.gm = make(map[uint32]*Chain)
	corpus.tm = make(map[uint64]map[uint64]struct{})
	corpus.ban = make(map[uint32]struct{})
	corpus.preban = make(map[uint32]uint64)
	corpus.allow = make(map[uint32]struct{})
	corpus.covered = make(map[uint64]uint64)
	return corpus
}

func (cp *Corpus) IncSchedCnt(sc string) {
	if _, exist := cp.schedCovered.LoadOrStore(Hash32(sc), struct{}{}); !exist {
		atomic.AddUint64(&cp.schedCoveredCnt, 1)
	}
}

func (cp *Corpus) SchedCnt() uint64 {
	return atomic.LoadUint64(&cp.schedCoveredCnt)
}

func (cp *Corpus) FetchCnt() uint64 {
	return atomic.LoadUint64(&cp.fetchCnt)
}

func (cp *Corpus) IncFetchCnt() {
	atomic.AddUint64(&cp.fetchCnt, 1)
}

func (cp *Corpus) Get() (*Chain, *Chain) {
	// TODO
	cp.IncFetchCnt()
	cp.gmu.Lock()
	defer cp.gmu.Unlock()
	htc := &Chain{}
	for k, v := range cp.gm {
		if _, ok := cp.hash.LoadOrStore(Hash32(v.ToString()), struct{}{}); ok {
			if rand.Int()%100 > allowDup {
				continue
			}
		} /*
			for _, p := range v.item {
				if rand.Int()%100 < AttackP {
					ht := cp.HTGet(p.Next.Opid)
					if ht != 0 {
						htc.add(pkg.NewPair(p.Next.Opid, ht))
					}
				}
			}*/
		delete(cp.gm, k) // reduce the size of corpus
		return v, htc
	}
	return nil, nil
}

func (cp *Corpus) GGet() *Chain {
	// TODO
	cp.gmu.RLock()
	defer cp.gmu.RUnlock()
	for _, v := range cp.gm {
		if _, ok := cp.hash.LoadOrStore(Hash32(v.ToString()), struct{}{}); !ok {
			return v
		}
		if rand.Int()%100 < allowDup {
			return v
		}
	}
	return nil
}

func (cp *Corpus) HTGet(id uint64) uint64 {
	cp.tmu.RLock()
	defer cp.tmu.RUnlock()
	ts := cp.tm[id]
	for v, _ := range ts {
		return v
	}
	return 0
}

func (cp *Corpus) Ban(ps [][]uint64) {
	cp.bmu.Lock()
	defer cp.bmu.Unlock()
	for _, p := range ps {
		s := Hash32(fmt.Sprintf("{%v, %v}", p[0], p[1]))
		if _, ok := cp.ban[s]; ok {
			continue
		}
		if _, ok := cp.allow[s]; ok {
			continue
		}
		if _, ok := cp.preban[s]; !ok {
			cp.preban[s] = 0
		}
		cp.preban[s] += 1
		if cp.preban[s] >= BANMAX {
			cp.ban[s] = struct{}{}
			delete(cp.preban, s)
		}
	}
}

func (cp *Corpus) Allow(ps [][]uint64) {
	cp.bmu.Lock()
	defer cp.bmu.Unlock()
	for _, p := range ps {
		s := Hash32(fmt.Sprintf("{%v, %v}", p[0], p[1]))
		cp.allow[s] = struct{}{}
	}
}

func (cp *Corpus) GExist(chain *Chain) bool {
	if chain == nil {
		return false
	}
	k := Hash32(chain.ToString())
	cp.gmu.RLock()
	defer cp.gmu.RUnlock()
	if _, ok := cp.gm[k]; ok {
		return true
	}
	return false
}

func (cp *Corpus) GUpdate(chain *Chain) bool {
	if cp.GExist(chain) {
		return false
	}
	k := Hash32(chain.ToString())
	cp.gmu.Lock()
	defer cp.gmu.Unlock()
	cp.gm[k] = chain
	return true
}

func (cp *Corpus) GSUpdate(chains []*Chain) bool {
	var ok bool
	for _, chain := range chains {
		t := cp.GUpdate(chain)
		if t {
			ok = true
		}
	}
	return ok
}

func (cp *Corpus) TUpdate(m map[uint64]map[uint64]struct{}) bool {
	cp.tmu.Lock()
	defer cp.tmu.Unlock()
	var update bool
	for k, v := range m {
		if vv, ok := cp.tm[k]; !ok {
			cp.tm[k] = v
			update = true
		} else {
			for k2, v2 := range v {
				vv[k2] = v2
				update = true
			}
		}
	}
	return update
}

func (cp *Corpus) UpdateSeed(seeds []*pkg.Pair) {
	chs := make([]*Chain, 0)
	hts := make(map[uint64]map[uint64]struct{}, 0)
	for _, seed := range seeds {
		chs = append(chs, &Chain{
			item: []*pkg.Pair{seed},
		})
		hts[seed.Prev.Opid][seed.Next.Opid] = struct{}{}
	}
	cp.GSUpdate(chs)
	cp.TUpdate(hts)
}

func (cp *Corpus) GUpdateSeed(seeds []*pkg.Pair) {
	chs := make([]*Chain, 0)
	for _, seed := range seeds {
		chs = append(chs, &Chain{
			item: []*pkg.Pair{seed},
		})
	}
	cp.GSUpdate(chs)
}

func (cp *Corpus) CUpdate(cs [][]uint64) {
	cp.cmu.Lock()
	defer cp.cmu.Unlock()
	for _, v := range cs {
		cp.covered[v[0]] = v[1]
	}
}

func (cp *Corpus) GetC() *pkg.Pair {
	cp.cmu.RLock()
	defer cp.cmu.RUnlock()
	for k, v := range cp.covered {
		return pkg.NewPair(k, v)
	}
	return nil
}

func (cp *Corpus) Update(ncs []*Chain, hts map[uint64]map[uint64]struct{}) {
	cp.GSUpdate(ncs)
	cp.TUpdate(hts)
}

func GetGlobalCorpus() *Corpus {
	return &GlobalCorpus
}

func (cp *Corpus) GSize() int {
	cp.gmu.RLock()
	defer cp.gmu.RUnlock()
	return len(cp.gm)
}
