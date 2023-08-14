package testdata

import (
	sched "sched"
	"sync"
	"testing"
)

type Connection struct {
	closeChan chan bool
}

type idleAwareFramer struct {
	resetChan chan bool
	writeLock sync.Mutex
	conn      *Connection
}

func (i *idleAwareFramer) monitor() {
	var resetChan = i.resetChan
Loop:
	for {
		select {
		case <-i.conn.closeChan:
			sched.InstChAF(1005022347269, i.conn.closeChan)
			sched.InstChAF(712964571141, i.conn.closeChan)
			sched.InstMutexBF(712964571142, &i.writeLock)
			sched.InstMutexBF(1005022347270, &i.writeLock)
			i.writeLock.Lock()
			sched.InstMutexAF(1005022347270, &i.writeLock)
			sched.InstMutexAF(712964571142, &i.writeLock)
			sched.InstChBF(712964571138, resetChan)
			sched.InstChBF(1005022347266, resetChan)
			close(resetChan)
			sched.InstChAF(1005022347266, resetChan)
			sched.InstChAF(712964571138, resetChan)
			i.resetChan = nil
			sched.InstMutexBF(712964571143, &i.writeLock)
			sched.InstMutexBF(1005022347271, &i.writeLock)
			i.writeLock.Unlock()
			sched.InstMutexAF(1005022347271, &i.writeLock)
			sched.InstMutexAF(712964571143, &i.writeLock)
			break Loop
		}
	}
}

func (i *idleAwareFramer) WriteFrame() {
	sched.InstMutexBF(712964571144, &i.writeLock)
	sched.InstMutexBF(1005022347272, &i.writeLock)
	i.writeLock.Lock()
	sched.InstMutexAF(1005022347272, &i.writeLock)
	sched.InstMutexAF(712964571144, &i.writeLock)
	defer func() {
		sched.InstMutexBF(712964571145, &i.writeLock)
		i.writeLock.Unlock()
		sched.InstMutexAF(712964571145, &i.writeLock)
	}()
	if i.resetChan == nil {
		return
	}
	sched.InstChBF(712964571139, i.resetChan)
	sched.InstChBF(1005022347267, i.resetChan)
	i.resetChan <- true
	sched.InstChAF(1005022347267, i.resetChan)
	sched.InstChAF(712964571139, i.resetChan)
}

func NewIdleAwareFramer() *idleAwareFramer {
	return &idleAwareFramer{
		resetChan: make(chan bool),
		conn: &Connection{
			closeChan: make(chan bool),
		},
	}
}

// /
// / G1                                         G2                                      helper goroutine
// / i.monitor()
// / <-i.conn.closeChan
// /                                                    i.WriteFrame()
// /                                                    i.writeLock.Lock()
// /                                                    i.resetChan <-
// /                                                                                            i.conn.closeChan<-
// /    i.writeLock.Lock()
// /    ----------------------G1,G2 deadlock------------------------
// /
func TestKubernetes6632(t *testing.T) {
	i := NewIdleAwareFramer()

	go func() {
		sched. // helper goroutine
			InstChBF(712964571140, i.conn.closeChan)
		sched.InstChBF(1005022347268, i.conn.closeChan)
		i.conn.closeChan <- true
		sched.InstChAF(1005022347268, i.conn.closeChan)
		sched.InstChAF(712964571140, i.conn.closeChan)
	}()
	go i.monitor()    // G1
	go i.WriteFrame() // G2
}
func TestKubernetes6632_1(t *testing.T) {
	sched.ParseInput()
	defer sched.Leakcheck(t)
	i := NewIdleAwareFramer()
	ch := make(chan int)
	go func() {
		sched.
			InstChBF(712964571140, i.conn.closeChan)
		sched.InstChBF(1005022347268, i.conn.closeChan)
		i.conn.closeChan <- true
		sched.InstChAF(1005022347268, i.conn.closeChan)
		sched.InstChAF(712964571140, i.conn.closeChan)
	}()
	<-ch
	go i.monitor()
	go i.WriteFrame()
}
