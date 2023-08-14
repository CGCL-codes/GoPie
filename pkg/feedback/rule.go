package feedback

type StAndOp struct {
	s  uint64
	op uint64
}

var RuleMap map[uint64]uint64

func init() {
	RuleMap = make(map[uint64]uint64)
	RuleMap[ChanFull] = Chansend
	RuleMap[ChanEmpty] = Chanrecv
	RuleMap[ChanClosed] = Chansend | Chanclose
	RuleMap[MutexLocked] = Lock
	RuleMap[MutexUnlocked] = Unlock
}

func UnderRule(stn, opn uint64) bool {
	for k, v := range RuleMap {
		if k&stn != 0 && v&opn != 0 {
			return true
		}
	}
	return false
}
