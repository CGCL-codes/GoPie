package sched

import "sync"

type Config struct {
	waitmap      map[uint64]uint64
	attackmap    map[uint64]uint64
	active       map[uint64]struct{}
	mu           sync.RWMutex
	wait_queue   [][]uint64
	attack_queue [][]uint64
	top          int32
}

func NewConfig() *Config {
	config := Config{}
	config.waitmap = make(map[uint64]uint64)
	config.active = make(map[uint64]struct{})
	config.wait_queue = make([][]uint64, 0)
	config.attack_queue = make([][]uint64, 0)
	return &config
}
