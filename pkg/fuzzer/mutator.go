package fuzzer

import (
	"math/rand"
	"toolkit/pkg"
	"toolkit/pkg/feedback"
)

const (
	BOUND       = 3
	MUTATEBOUND = 128
)

type Mutator struct {
	Cov *feedback.Cov
}

func (m *Mutator) mutate(chain *Chain, energy int) ([]*Chain, map[uint64]map[uint64]struct{}) {
	// TODO : energy
	gs := m.mutateg(chain, energy)
	ht := make(map[uint64]map[uint64]struct{})

	/*
		for _, g := range gs {
			m.mutatet(g, ht)
		}*/
	return gs, ht
}

func (m *Mutator) random(chain *Chain, energy int) ([]*Chain, map[uint64]map[uint64]struct{}) {
	// TODO : energy
	gs := m.randomg(chain, energy)
	ht := make(map[uint64]map[uint64]struct{})

	/*
		for _, g := range gs {
			m.mutatet(g, ht)
		}*/
	return gs, ht
}

func (m *Mutator) mutateg(chain *Chain, energy int) []*Chain {
	// TODO
	if energy > 100 {
		energy = 100
	}
	if chain == nil {
		return []*Chain{}
	}
	res := make([]*Chain, 0)
	set := make(map[string]*Chain, 0)

	if chain.Len() == 0 {
		op1, op2 := m.Cov.One()
		chain = &Chain{[]*pkg.Pair{pkg.NewPair(op1, op2)}}
	}

	set[chain.ToString()] = chain
	if chain.Len() == 1 {
		nc := &Chain{[]*pkg.Pair{&pkg.Pair{chain.T().Next, chain.T().Prev}}}
		set[nc.ToString()] = nc
	}
	for {
		for _, chain := range set {
			tset := make(map[string]*Chain, 0)
			// reduce the length
			if chain.Len() >= 2 {
				nc := chain.Copy()
				nc.pop()
				tset[nc.ToString()] = nc
				nc2 := chain.Copy()
				nc2.item = nc2.item[1:len(nc2.item)]
				tset[nc2.ToString()] = nc2
			}

			//replace
			if chain.Len() <= BOUND {
				if rand.Int()%2 == 1 {
					lastopid := chain.T().Next.Opid
					rels := m.Cov.NextO2(lastopid)
					for _, rel := range rels {
						nc := chain.Copy()
						nc.add(pkg.NewPair(lastopid, rel))
						tset[nc.ToString()] = nc
					}
				}
			}

			// increase the length
			if chain.Len() <= BOUND {
				if rand.Int()%2 == 1 {
					lastopid := chain.T().Next.Opid
					rels := m.Cov.NextR(lastopid)
					if len(rels) == 0 {
						rels = m.Cov.Next(lastopid)
					}
					for _, rel := range rels {
						nc := chain.Copy()
						nc.add(pkg.NewPair(lastopid, rel))
						tset[nc.ToString()] = nc
					}
				}
			} else {
				// nc := chain.Copy()
				// nc.add(GetGlobalCorpus().GetC())
				// tset[nc.ToString()] = nc
			}
			for k, v := range tset {
				set[k] = v
			}
		}
		if len(set) > MUTATEBOUND {
			break
		}
		// up to 50%
		if (rand.Int() % 200) < energy {
			break
		}
	}

	// merge two chain
	// TODO
	for _, v := range set {
		if m.filter(v) {
			res = append(res, v)
		}
	}

	return res
}

func (m *Mutator) randomg(chain *Chain, energy int) []*Chain {
	// TODO
	if energy > 100 {
		energy = 100
	}
	if chain == nil {
		return []*Chain{}
	}
	res := make([]*Chain, 0)
	set := make(map[string]*Chain, 0)

	if chain.Len() == 0 {
		op1 := m.Cov.OneRandom()
		op2 := m.Cov.OneRandom()
		chain = &Chain{[]*pkg.Pair{pkg.NewPair(op1, op2)}}
	}

	set[chain.ToString()] = chain
	if chain.Len() == 1 {
		nc := &Chain{[]*pkg.Pair{&pkg.Pair{chain.T().Next, chain.T().Prev}}}
		set[nc.ToString()] = nc
	}
	for {
		for _, chain := range set {
			tset := make(map[string]*Chain, 0)
			// reduce the length
			if chain.Len() >= 2 {
				nc := chain.Copy()
				nc.pop()
				tset[nc.ToString()] = nc
				nc2 := chain.Copy()
				nc2.item = nc2.item[1:len(nc2.item)]
				tset[nc2.ToString()] = nc2
			}

			// increase the length
			if chain.Len() <= BOUND {
				if rand.Int()%2 == 1 {
					lastopid := chain.T().Next.Opid
					rel := m.Cov.OneRandom()
					nc := chain.Copy()
					nc.add(pkg.NewPair(lastopid, rel))
					tset[nc.ToString()] = nc
				}
			}
			for k, v := range tset {
				set[k] = v
			}
		}
		if len(set) > MUTATEBOUND {
			break
		}
		// up to 50%
		if (rand.Int() % 200) < energy {
			break
		}
	}

	// merge two chain
	// TODO
	for _, v := range set {
		if m.filter(v) {
			res = append(res, v)
		}
	}

	return res
}

// mutatet find out possible attack pairs for each pair in the EC
// func (m *Mutator) mutatet(c *Chain, ht map[uint64]map[uint64]struct{}) {
// 	for _, p := range c.item {
// 		nop := p.Next
// 		if _, ok := ht[nop.Opid]; !ok {
// 			ht[nop.Opid] = make(map[uint64]struct{}, 0)
// 		}
//
// 		// hang attack
// 		ht[nop.Opid][nop.Opid] = struct{}{}
//
// 		st, ok := m.Cov.GetStatus(feedback.OpID(nop.Opid))
// 		if !ok {
// 			return
// 		}
// 		stn := st.ToU64()
//
// 		typefilter := func(stbit, tbit uint64) {
// 			if stn&stbit != 0 {
// 				next, ok := m.Cov.NextTyp(nop.Opid, tbit, nil)
// 				if ok {
// 					ht[nop.Opid][next] = struct{}{}
// 				}
// 			}
// 		}
//
// 		// find status by typ
// 		statusfilter := func(stbit, sbit uint64) {
// 			if stn&stbit != 0 {
// 				next, ok2 := m.Cov.NextStatus(nop.Opid, stbit, nil)
// 				if ok2 {
// 					ht[nop.Opid][next] = struct{}{}
// 				}
// 			}
// 		}
// 		rule := feedback.RuleMap
// 		for st2, op := range rule {
// 			statusfilter(st2, op)
// 			typefilter(st2, op)
// 		}
// 	}
// }

// filter the output of mutator by rules
func (m *Mutator) filter(chain *Chain) bool {
	return true
}
