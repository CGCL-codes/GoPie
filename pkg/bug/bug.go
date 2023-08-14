package bug

import (
	"strings"
	"sync"
	"sync/atomic"
)

func TopF(s string) []string {
	if strings.Contains(s, "panic") {
		return []string{"panic"}
	}
	ss := strings.Split(s, "\n")
	topfunc := make([]string, 0)
	for _, line := range ss {
		if strings.Contains(line, "on top of the stack:") {
			idx := strings.LastIndex(line, " on top of the stack:")
			if idx != -1 {
				idx2 := strings.Index(line, "with ")
				if idx2 != -1 {
					f := line[idx2+5 : idx]
					if !strings.Contains(s, "github.com") {
						topfunc = append(topfunc, f)
					} else {
						topfunc = append(topfunc, "ignore")
					}
				}
			}
		}
	}
	return topfunc
}

type BugSet struct {
	m   sync.Map
	cnt uint32
	mu  sync.Mutex
}

func NewBugSet() *BugSet {
	return &BugSet{}
}

func (bs *BugSet) Exist(fs []string, fn string) bool {
	bs.mu.Lock()
	defer bs.mu.Unlock()
	for _, f := range fs {
		_, exist := bs.m.Load(fn + f)
		if exist {
			return true
		}
	}
	atomic.AddUint32(&bs.cnt, 1)
	for _, f := range fs {
		bs.m.Store(fn+f, struct{}{})
	}
	return false
}

func (bs *BugSet) Size() uint32 {
	return atomic.LoadUint32(&bs.cnt)
}
