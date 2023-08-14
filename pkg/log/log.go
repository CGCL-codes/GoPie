package log

import (
	"log"
	"sync"
	"time"
)

type Logger struct {
	m sync.Map
}

var LOG Logger

type Sequence struct {
	l  []*Entry
	mu sync.Mutex
}

type Entry struct {
	opid uint64
	eid  uint32
	t    time.Time
}

func (l *Logger) LogWithTime(opid uint64, e uint32, addr any) {
	t := time.Now()
	v, _ := l.m.LoadOrStore(addr, &Sequence{make([]*Entry, 0), sync.Mutex{}})
	if seq, ok := v.(*Sequence); !ok {
		log.Fatalf("logger crash")
	} else {
		seq.mu.Lock()
		defer seq.mu.Unlock()
		seq.l = append(seq.l, &Entry{opid, e, t})
	}
}
