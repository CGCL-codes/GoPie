package fuzzer

import (
	"fmt"
	"hash/fnv"
	"toolkit/pkg"
	"toolkit/pkg/feedback"
)

func ColorCovered(covlog string, c *Chain) (string, *Chain) {
	if c == nil || c.Len() == 0 {
		return "empty chain", &Chain{item: []*pkg.Pair{}}
	}
	idx := 0
	schedres := ""
	covered := &Chain{item: []*pkg.Pair{}}
	schedcov := feedback.ParseCovered(covlog)
	for _, p := range c.item {
		if idx < len(schedcov) {
			cov := schedcov[idx]
			if p.Prev.Opid == cov[0] && p.Next.Opid == cov[1] {
				schedres += fmt.Sprintf("\033[1;31;40m%s\033[0m", p.ToString())
				covered.item = append(covered.item, p)
				idx += 1
			}
		} else {
			schedres += p.ToString()
		}
	}
	return schedres, covered
}

func Hash32(s string) uint32 {
	h := fnv.New32a()
	h.Write([]byte(s))
	return h.Sum32()
}
