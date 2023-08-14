package feedback

import (
	"fmt"
	"strconv"
	"strings"
)

/*
[FBSDK] makechan: chan=0xc00006a090; elemsize=8; dataqsiz=5
[FBSDK] Chansend: chan=0xc00006a090; elemsize=8; dataqsiz=5; qcount=1
[FB] chan: 0xc00006a090; id :841813590017;
[FBSDK] chanclose: chan=0xc00006a090
[FB] chan: 0xc00006a090; id :841813590018;
[FB] mutex: 0xc000089f3c; id :841813590019;
[FB] mutex: 0xc000089f3c; id :841813590020;
*/
func implParse(s string) (bool, []uint64) {
	res := make([]uint64, 0)
	ss := strings.Split(s, ";")
	for _, v := range ss {
		if len(v) == 0 || v == "\n" || v == "\t" || v == " " {
			continue
		}
		if !strings.Contains(v, "=") {
			return false, nil
		}
		value := strings.Split(v, "=")[1]
		var base int
		if len(value) >= 3 && value[0:2] == "0x" {
			base = 16
			value = value[2 : len(value)-1]
		} else {
			base = 10
		}
		if value[len(value)-1] == ';' {
			value = value[0 : len(value)-1]
		}
		i, err := strconv.ParseUint(value, base, 64)
		if err != nil {
			return false, nil
		}
		res = append(res, i)
	}
	return true, res
}

func parseLine(s string) (bool, uint64, OpAndStatus) {
	defer func() {
		if r := recover(); r != nil {
			// fmt.Printf("parseLine error, recovered\n")
		}
	}()
	if !strings.HasPrefix(s, "[FB") {
		return false, 0, OpAndStatus{}
	}
	ss := strings.Split(s, ":")
	if len(ss) != 2 {
		return false, 0, OpAndStatus{}
	}
	head := ss[0]
	ok, values := implParse(ss[1])
	if !ok {
		return false, 0, OpAndStatus{}
	}
	var typ uint64
	switch head {
	case "[FBSDK] chansend", "[FBSDK] chanrecv":
		if head == "[FBSDK] chansend" {
			typ = Chansend
		} else {
			typ = Chanrecv
		}
		return true, values[0], OpAndStatus{
			Opid:   0,
			Oid:    values[0],
			status: ObjectStatus{isCh: true, dataqsize: int(values[2]), qcount: int(values[3])},
			Gid:    values[4],
			Typ:    typ,
		}
	case "[FBSDK] makechan":
		return true, values[0], OpAndStatus{
			Opid:   0,
			Oid:    values[0],
			status: ObjectStatus{isCh: true, dataqsize: int(values[2]), qcount: 0},
			Typ:    Chanmake,
		}
	case "[FBSDK] chanclose":
		return true, values[0], OpAndStatus{
			Opid:   0,
			Oid:    values[0],
			status: ObjectStatus{isCh: true, closed: true},
			Gid:    values[1],
			Typ:    Chanclose,
		}
	case "[FB] chan":
		return true, values[0], OpAndStatus{
			Opid:   values[1],
			Oid:    values[0],
			status: ObjectStatus{isCh: true},
		}
	case "[FB] mutex":
		var islocked bool
		if values[2] == 1 {
			islocked = true
			typ = Lock
		} else {
			typ = Unlock
		}
		return true, values[0], OpAndStatus{
			Opid:   values[1],
			Oid:    values[0],
			status: ObjectStatus{isCh: false, locked: islocked},
			Gid:    values[3],
			Typ:    typ,
		}
	default:
		return false, 0, OpAndStatus{}
	}
}

func ParseCovered(s string) [][]uint64 {
	res := make([][]uint64, 0)
	lines := strings.Split(s, "\n")
	fblines := make([]string, 0)
	for _, line := range lines {
		if strings.HasPrefix(line, "[COVERED]") {
			fblines = append(fblines, line)
		}
	}
	for _, line := range fblines {
		var prev, next uint64
		_, err := fmt.Sscanf(line, "[COVERED] {%v, %v}", &prev, &next)
		if err == nil {
			res = append(res, []uint64{prev, next})
		}
	}
	return res
}

func ParseLog(s string) (map[uint64][]OpAndStatus, []OpAndStatus) {
	lines := strings.Split(s, "\n")
	fblines := make([]string, 0)
	// global orders, which used to generate coverage
	orders := make([]OpAndStatus, 0)
	for _, line := range lines {
		if strings.HasPrefix(line, "[FB") {
			fblines = append(fblines, line)
		}
	}

	m := make(map[uint64][]OpAndStatus, 0)

	add := func(oid uint64, info OpAndStatus) {
		if _, ok := m[oid]; !ok {
			m[oid] = make([]OpAndStatus, 0)
		}
		// keep the order of logs
		m[oid] = append(m[oid], info)
		orders = append(orders, info)
	}

	// filter, FBSDK->FB for chan and [FB] mutex for mutex
	for i, line := range lines {
		if strings.HasPrefix(line, "[FB] mutex") {
			ok, oid, info := parseLine(line)
			if !ok {
				continue
			}
			add(oid, info)
		} else {
			if strings.HasPrefix(line, "[FBSDK]") {
				if i+1 < len(lines)-1 && strings.HasPrefix(lines[i+1], "[FB]") {
					ok, oid, info := parseLine(line)
					ok2, oid2, info2 := parseLine(lines[i+1])
					if !ok || !ok2 || (oid != oid2 && !info.status.isCh) {
						continue
					}
					info.Opid = info2.Opid
					add(oid, info)
				}
			}
		}
	}
	return m, orders
}

// info is object map to execution orders
func Log2Cov(info map[uint64][]OpAndStatus, allops []OpAndStatus) *Cov {
	cov := NewCov()
	for _, ops := range info { // same primitive
		l := len(ops)
		for i := 0; i < l-1; i++ {
			if ops[i].Gid != ops[i+1].Gid {
				if _, ok := cov.rel[ops[i].Opid]; !ok {
					cov.rel[ops[i].Opid] = make(map[uint64]struct{})
				}
				if _, ok := cov.rel[ops[i+1].Opid]; !ok {
					cov.rel[ops[i+1].Opid] = make(map[uint64]struct{})
				}
				if _, ok := cov.o1[ops[i].Opid]; !ok {
					cov.o1[ops[i].Opid] = make(map[uint64]struct{})
				}
				cov.rel[ops[i].Opid][ops[i+1].Opid] = struct{}{} // same object on different goroutines
				cov.rel[ops[i+1].Opid][ops[i].Opid] = struct{}{}
				cov.o1[ops[i].Opid][ops[i+1].Opid] = struct{}{} // concurrnecy orders
			}
		}
		for i := 0; i < l; i++ {
			st := ops[i].status.IsCritical()
			if st != 0 {
				cov.UpdateC(OpID(ops[i].Opid), ToStatus(st))
			}
		}
	}
	l := len(allops)
	for i := 0; i < l; i++ {
		for j := i + 1; j < l; j++ {
			if allops[i].Gid == allops[j].Gid {
				if _, ok := cov.o2[allops[i].Opid]; !ok {
					cov.o2[allops[i].Opid] = make(map[uint64]struct{})
				}
				cov.o2[allops[i].Opid][allops[i+1].Opid] = struct{}{} // concurrnecy orders
				break
			}
		}
	}
	return cov
}

type Fragments struct {
	m map[uint64]uint64
}

func (f *Fragments) Size() int {
	return len(f.m)
}

func (f *Fragments) Root(x uint64) uint64 {
	if x == 0 {
		return 0
	}
	for {
		if v, ok := f.m[x]; !ok {
			return 0
		} else {
			if v != x {
				x = v
			} else {
				break
			}
		}
	}
	return x
}

func (f *Fragments) IsSame(x uint64, y uint64) bool {
	return f.Root(x) != 0 && f.Root(x) == f.Root(y)
}

func (f *Fragments) Exist(x uint64) bool {
	_, ok := f.m[x]
	return ok
}

func (f *Fragments) Uint(x uint64, y uint64) uint64 {
	if !f.Exist(x) {
		f.Add(x)
	}
	if !f.Exist(y) {
		f.Add(y)
	}
	r1 := f.Root(x)
	r2 := f.Root(y)
	f.m[r2] = r1
	return r1
}

func (f *Fragments) Add(x uint64) {
	f.m[x] = x
}

func NewFragments(m map[uint64][]OpAndStatus) Fragments {
	// TODO
	return Fragments{}
}
