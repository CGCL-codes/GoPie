package fuzzer

import "toolkit/pkg/bug"

type Config struct {
	Bin string
	Fn  string

	MaxWorker    int
	MaxExecution int
	SingleCrash  bool

	LogLevel string
	LogCh    chan string

	UseFeedBack     bool
	UseCoveredSched bool
	UseStates       bool
	UseAnalysis     bool
	UseMutate       bool
	UseGuide        bool

	TimeOut        int
	RecoverTimeOut int
	InitTurnCnt    int
	MaxQuit        int

	BugSet *bug.BugSet
}

func DefaultConfig() *Config {
	c := &Config{
		MaxWorker:       5,
		MaxExecution:    10000000000,
		SingleCrash:     false,
		LogLevel:        "normal",
		UseFeedBack:     true,
		UseStates:       true,
		UseCoveredSched: true,
		UseAnalysis:     true,
		UseMutate:       true,
		UseGuide:        true,
		TimeOut:         30,
		RecoverTimeOut:  100,
		InitTurnCnt:     100,
		MaxQuit:         500,
	}
	return c
}

func GokerConfig() *Config {
	c := &Config{
		MaxWorker:       5,
		MaxExecution:    100000,
		SingleCrash:     true,
		LogLevel:        "normal",
		UseFeedBack:     true,
		UseStates:       true,
		UseCoveredSched: true,
		UseAnalysis:     true,
		UseMutate:       true,
		UseGuide:        true,
		TimeOut:         30,
		RecoverTimeOut:  100,
		InitTurnCnt:     0,
		MaxQuit:         10000,
	}
	return c
}

func NewConfig(bin, fn string, logCh chan string, bugset *bug.BugSet, typ string) *Config {
	var c *Config
	switch typ {
	case "goker":
		c = GokerConfig()
	default:
		c = DefaultConfig()
	}
	c.Bin = bin
	c.Fn = fn
	c.LogCh = logCh
	c.BugSet = bugset
	return c
}
