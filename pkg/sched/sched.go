package sched

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"runtime"
	"sched/goleak"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

var event sync.Map
var timeout, recovertimeout time.Duration

var config *Config
var cancel chan struct{}

var once sync.Once

const (
	debugSched = true
)

func init() {
	config = NewConfig()
	cancel = make(chan struct{})
	timeout = time.Second * 20
	recovertimeout = time.Second * 1
	if s := os.Getenv("TIMEOUT"); s != "" {
		t, err := strconv.ParseInt(s, 10, 32)
		if err == nil {
			timeout = time.Duration(t) * time.Second
		}
	}
	if s := os.Getenv("RECOVER_TIMEOUT"); s != "" {
		t, err := strconv.ParseInt(s, 10, 32)
		if err == nil {
			recovertimeout = time.Duration(t) * time.Millisecond
		}
	}
}

// find sender with current wait ID
func (c *Config) findPrev(i uint64) uint64 {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if int(c.top) < len(c.wait_queue) && i == c.wait_queue[c.top][1] {
		return c.wait_queue[c.top][0]
	}
	return 0
}

// find waiter with current send ID
func (c *Config) findNext(i uint64) uint64 {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if int(c.top) < len(c.wait_queue) && i == c.wait_queue[c.top][0] {
		return c.wait_queue[c.top][1]
	}
	return 0
}

// 1. add the pairs to wait_queue
// 2. add to the active
// 3. add the next IDs to waitmap with a counter
func ParsePair(s string) {
	config.mu.Lock()
	defer config.mu.Unlock()
	var prev, next uint64
	for {
		left := strings.Index(s, "(")
		right := strings.Index(s, ")")
		if left < right {
			_, err := fmt.Sscanf(s[left:right+1], "({%v}, {%v})", &prev, &next)
			if err == nil {
				if _, ok := config.waitmap[next]; !ok {
					config.waitmap[next] = 0
				}
				config.waitmap[next] += 1
				config.wait_queue = append(config.wait_queue, []uint64{prev, next})
				config.active[prev] = struct{}{}
				config.active[next] = struct{}{}
			}
		}
		if right+1 < len(s) {
			s = s[right+1 : len(s)]
		} else {
			break
		}
	}
}

func ParseAttackPair(s string) {
	config.mu.Lock()
	defer config.mu.Unlock()
	var prev, next uint64
	for {
		left := strings.Index(s, "(")
		right := strings.Index(s, ")")
		if left < right {
			_, err := fmt.Sscanf(s[left:right+1], "({%v}, {%v})", &prev, &next)
			if err == nil {
				config.attackmap[next] = prev
				config.attack_queue = append(config.wait_queue, []uint64{prev, next})
				config.active[prev] = struct{}{}
				config.active[next] = struct{}{}
			}
		}
		if right+1 < len(s) {
			s = s[right+1 : len(s)]
		} else {
			break
		}
	}
}

func ParseInput() {
	input_pairs := os.Getenv("Input")
	if input_pairs != "" {
		ParsePair(input_pairs)
	}
}

func SetTimeout(s int) {
	timeout = time.Second * time.Duration(s)
}

func (c *Config) doWait(id uint64) (wait bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if _, ok := c.active[id]; !ok {
		return false
	}
	if v, ok := c.waitmap[id]; ok {
		if v <= 0 {
			return false
		} else {
			return true
		}
	}
	return false
}

func (c *Config) waitDec(id uint64) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if v, ok := c.waitmap[id]; ok {
		if v <= 1 {
			delete(config.waitmap, id)
		}
		if v > 1 {
			c.waitmap[id] -= 1
		}
	}
}

func InstChBF[T any | chan T | <-chan T | chan<- T](id uint64, o T) {
	var wait bool
	if wait = config.doWait(id); !wait {
		return
	}
	pid := config.findPrev(id)
	if pid == 0 {
		return
	}
	timer := time.After(timeout / 5)
	for {
		if _, ok := event.LoadAndDelete(pid); ok {
			atomic.AddInt32(&config.top, 1)
			config.waitDec(id)
			fmt.Printf("[COVERED] {%v, %v}\n", pid, id)
			return
		}
		select {
		case <-cancel:
			return
		case <-timer:
			return
		default:
		}
	}
}

func InstChAF[T any | chan T | <-chan T | chan<- T](id uint64, o T) {
	if debugSched {
		print("[FB] chan: obj=", o, "; id=", id, ";\n")
	}
	event.Store(id, struct{}{})
}

func InstMutexBF(id uint64, o any) {
	var wait bool
	if wait = config.doWait(id); !wait {
		return
	}
	for {
		pid := config.findPrev(id)
		if pid == 0 {
			time.Sleep(recovertimeout)
		}
		if _, ok := event.LoadAndDelete(pid); ok {
			atomic.AddInt32(&config.top, 1)
			config.waitDec(id)
			fmt.Printf("[COVERED] {%v, %v}\n", pid, id)
			return
		}
		select {
		case <-cancel:
			return
		default:
		}
	}
	return
}

func InstMutexAF(id uint64, o any) {
	if debugSched {
		var islocked int
		var locked bool
		var mid uint64
		switch mu := o.(type) {
		case *sync.Mutex:
			locked = mu.IsLocked()
			mid = mu.ID()
		case *sync.RWMutex:
			locked = mu.IsLocked()
			mid = mu.ID()
		}
		if locked {
			islocked = 1
		} else {
			islocked = 0
		}
		print("[FB] mutex: obj=", mid, "; id=", id, "; locked=", islocked, "; gid=", runtime.Goid(), "\n")
	}
	event.Store(id, struct{}{})
}

func GetDone() chan struct{} {
	return make(chan struct{})
}

func GetTimeout() <-chan time.Time {
	return time.After(timeout)
}

func Done(ch chan struct{}) {
	close(ch)
}

func Leakcheck(t *testing.T) {
	once.Do(func() {
		close(cancel)
		baseCheck(t)
	})
}

func readlines(filename string) []string {
	res := make([]string, 0)
	if _, err := os.Stat(filename); err != nil {
		return res
	}
	f, err := os.Open(filename)
	if err != nil {
		log.Fatalf("open file error: %v\n", err.Error())
		return []string{}
	}
	// remember to close the file at the end of the program
	defer f.Close()
	// read the file line by line using scanner
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		// do something with a line
		res = append(res, scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		return res
	}
	return res
}

func baseCheck(t *testing.T) {
	opts := []goleak.Option{
		goleak.IgnoreTopFunction("time.Sleep"),
		goleak.IgnoreTopFunction("testing.(*F).Fuzz.func1"),
		goleak.IgnoreTopFunction("testing.runFuzzTests"),
		goleak.IgnoreTopFunction("testing.runFuzzing"),
		goleak.IgnoreTopFunction("os/signal.NotifyContext.func1"),
		goleak.IgnoreTopFunction("testing.tRunner.func1"),
		goleak.IgnoreTopFunction("github.com/ethereum/go-ethereum/metrics.(*meterArbiter).tick"),
		goleak.IgnoreTopFunction("github.com/ethereum/go-ethereum/core.(*txSenderCacher).cache"),
		goleak.IgnoreTopFunction("github.com/ethereum/go-ethereum/consensus/ethash.(*remoteSealer).loop"),
		goleak.MaxRetryAttempts(24),
		goleak.MaxSleepInterval(10 * time.Second),
	}

	goleak.VerifyNone(t, opts...)
}
