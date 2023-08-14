package feedback

import "fmt"

const (
	Chanmake = iota
	Chansend
	Chanrecv
	Chanclose
	Lock
	Unlock
)

const (
	ChanFull = 1 << iota
	ChanEmpty
	ChanClosed
	MutexLocked
	MutexUnlocked
)

type ObjectStatus struct {
	isCh bool
	// for channel
	dataqsize int
	qcount    int
	closed    bool
	// for mutex
	locked bool
}

func (s *ObjectStatus) IsCritical() uint64 {
	if s.isCh {
		res := uint64(0)
		if s.qcount == 0 {
			res |= ChanEmpty
		}
		if s.qcount == s.dataqsize {
			res |= ChanFull
		}
		if s.closed {
			res = ChanClosed
		}
		return res
	}

	// mutex
	if s.locked {
		return MutexLocked
	}
	return MutexUnlocked
}

type OpAndStatus struct {
	Opid   uint64
	Oid    uint64
	Gid    uint64
	Typ    uint64
	status ObjectStatus
}

func (ops *OpAndStatus) ToString() string {
	var s string
	switch ops.Typ {
	case Chanmake:
		s = fmt.Sprintf("make(chan, %v)", ops.status.dataqsize)
	case Chansend:
		s = fmt.Sprintf("%v: v -> chan, (%v/%v)", ops.Gid, ops.status.qcount, ops.status.dataqsize)
	case Chanrecv:
		s = fmt.Sprintf("%v: <- chan, (%v/%v)", ops.Gid, ops.status.qcount, ops.status.dataqsize)
	case Chanclose:
		s = fmt.Sprintf("%v: close(chan)", ops.Gid)
	case Lock:
		s = fmt.Sprintf("%v: mu.lock", ops.Gid)
	case Unlock:
		s = fmt.Sprintf("%v: mu.unlock", ops.Gid)
	default:
	}
	return s
}
