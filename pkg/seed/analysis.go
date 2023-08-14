package seed

import (
	"toolkit/pkg"
	"toolkit/pkg/feedback"
)

func RandomSeed(m map[uint64][]feedback.OpAndStatus) []*pkg.Pair {
	var prev, next uint64
	for _, l := range m {
		if len(l) > 0 {
			prev = l[0].Opid
			break
		}
	}
	for _, l := range m {
		if len(l) > 0 {
			next = l[0].Opid
			break
		}
	}
	return []*pkg.Pair{&pkg.Pair{
		Prev: pkg.Entry{Opid: prev},
		Next: pkg.Entry{Opid: next},
	}}
}

var MaxSeed = 50

func SODRAnalysis(m map[uint64][]feedback.OpAndStatus) []*pkg.Pair {
	//Same object operated in different routines
	visit := make(map[string]*pkg.Pair, 0)
	res := make([]*pkg.Pair, 0)
	for _, l := range m {
		size := len(l)
		for i := 0; i < size; i++ {
			for j := i + 1; j < size; j++ {
				if l[i].Gid != l[j].Gid {
					// Do not need to do a reverse, mutator can do it.
					pair := &pkg.Pair{
						Prev: pkg.Entry{
							Opid: l[i].Opid,
						},
						Next: pkg.Entry{
							Opid: l[j].Opid,
						},
					}
					visit[pair.ToString()] = pair
				}
				if len(visit) > MaxSeed {
					break
				}
			}
			if len(visit) > MaxSeed {
				break
			}
		}
	}
	for _, v := range visit {
		res = append(res, v)
	}
	return res
}

func SRDOAnalysis(m map[uint64][]feedback.OpAndStatus) []*pkg.Pair {
	//Same routine operate different objects
	visit := make(map[string]*pkg.Pair, 0)

	// switch to routine view
	m2 := make(map[uint64][]feedback.OpAndStatus)
	for _, l := range m {
		size := len(l)
		for i := 0; i < size; i++ {
			gid := l[i].Gid
			if _, ok := m2[gid]; !ok {
				m2[gid] = make([]feedback.OpAndStatus, 0)
			}
			m2[gid] = append(m2[gid], l[i])
		}
	}

	for _, l := range m2 {
		size := len(l)
		for i := 0; i < size; i++ {
			for j := i + 1; j < size; j++ {
				if l[i].Oid != l[j].Oid || l[i].Typ != l[j].Typ {
					// Find different objects operated in same routine, which we easily consider
					// other operations on these objects are possibly related
					oid1 := l[i].Oid
					oid2 := l[j].Oid
					if l1, ok := m[oid1]; ok {
						if l2, ok2 := m[oid2]; ok2 {
							for i, op1 := range l1 {
								for j, op2 := range l2 {
									if op1.Gid != op2.Gid {
										p := &pkg.Pair{
											Prev: pkg.Entry{
												Opid: l1[i].Opid,
											},
											Next: pkg.Entry{
												Opid: l2[j].Opid,
											},
										}
										visit[p.ToString()] = p
										if len(visit) > MaxSeed {
											res := make([]*pkg.Pair, 0)
											for _, v := range visit {
												res = append(res, v)
											}
											return res
										}
									}
								}
							}
						}
					}
				}
			}
		}
	}

	res := make([]*pkg.Pair, 0)
	for _, v := range visit {
		res = append(res, v)
	}
	return res
}
