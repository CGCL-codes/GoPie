package yield

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"math/rand"
	"runtime"
	"sync"
)

var Bound = 5
var MaxP = 0.9
var MaxCh = 10
var ON = true

type context struct {
	// channel operation ids
	op_chain     []int
	op_chain_max int

	//
	op_hash_map sync.Map
	sync.Mutex
}

var YieldCtx context

func (c *context) AddOp(id int) bool {
	c.Lock()
	defer c.Unlock()
	old := c.op_chain
	if len(c.op_chain) == c.op_chain_max {
		c.op_chain = c.op_chain[1:]
	}
	c.op_chain = append(c.op_chain, id)
	s, _ := json.Marshal(c.op_chain)
	h := sha256.Sum256(s)

	if _, ok := c.op_hash_map.Load(h); !ok {
		fmt.Printf("new path : %v\n", c.op_chain)
		c.op_hash_map.Store(h, struct{}{})
		return true
	}
	fmt.Printf("same path : %v\n", c.op_chain)
	c.op_chain = old
	return false
}

func do_probability(fn func(), p float64) bool {
	if rand.Float64() < p {
		fn()
		return true
	}
	return false
}

func (c *context) init() {
	var once sync.Once
	once.Do(func() {
		if c.op_chain_max == 0 && c.op_chain == nil {
			c.op_chain = make([]int, 0)
			c.op_chain_max = MaxCh
		}
	})
}

func Yield(id int) {
	if !ON {
		fmt.Printf("op : %v\n", id)
		return
	}
	YieldCtx.init()
	cnt := Bound
	for !YieldCtx.AddOp(id) {
		if cnt == 0 {
			break
		}
		do_probability(runtime.Gosched, MaxP)
		cnt--
	}
	return
}

func Clean() {
	YieldCtx.Lock()
	defer YieldCtx.Unlock()
	if !ON {
		return
	}
	YieldCtx.op_chain = make([]int, 0)
}
