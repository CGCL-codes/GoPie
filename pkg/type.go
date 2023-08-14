package pkg

import (
	"fmt"
)

type Pair struct {
	Prev Entry
	Next Entry
}

func (e *Pair) ToString() string {
	return fmt.Sprintf("(%v, %v)", e.Prev, e.Next)
}

func NewPair(op1, op2 uint64) *Pair {
	return &Pair{
		Prev: Entry{op1},
		Next: Entry{op2},
	}
}

type Entry struct {
	Opid uint64
}
