package hack

import (
	"sync/atomic"
	"unsafe"
)

type hchan struct {
	qcount   uint           // total data in the queue
	dataqsiz uint           // size of the circular queue
	buf      unsafe.Pointer // points to an array of dataqsiz elements
	elemsize uint16
	closed   uint32
	elemtype unsafe.Pointer // element type
	sendx    uint           // send index
	recvx    uint           // receive index
	recvq    waitq          // list of recv waiters
	sendq    waitq          // list of send waiters

	cid uint64

	// lock protects all fields in hchan, as well as several
	// fields in sudogs blocked on this channel.
	//
	// Do not change another G's status while holding this lock
	// (in particular, do not ready a G), as this can deadlock
	// with stack shrinking.
	lock mutex
}

type waitq struct {
	first unsafe.Pointer
	last  unsafe.Pointer
}

type lockRankStruct struct {
}

type mutex struct {
	// Empty struct if lock ranking is disabled, otherwise includes the lock rank
	lockRankStruct
	// Futex-based impl treats it as uint32 key,
	// while sema-based impl as M* waitm.
	// Used to be a union, but unions break precise GC.
	key uintptr
}

func GetChID(ch any) uint64 {
	/*
		if reflect.TypeOf(ch).Kind() != reflect.Chan {
			panic("only channels!")
		}*/

	// get interface value pointer, from cgo_export
	// typedef struct { void *t; void *v; } GoInterface;
	// then get channel real pointer
	cptr := *(*uintptr)(unsafe.Pointer(
		unsafe.Pointer(uintptr(unsafe.Pointer(&ch)) + unsafe.Sizeof(uint(0))),
	))

	cptr += unsafe.Offsetof(hchan{}.cid)
	u64p := (*uint64)(unsafe.Pointer(cptr))
	return atomic.LoadUint64(u64p)
}
