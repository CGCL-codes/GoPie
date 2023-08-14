package main

import (
	"fmt"
	"time"
	"toolkit/cmd"
	"toolkit/pkg/bug"
	"toolkit/pkg/feedback"
	"toolkit/pkg/fuzzer"
)

func Full(path string, llevel string, feature string, maxworker int) {
	resCh := make(chan string, 100000)
	logCh := make(chan string, 100000)
	// control
	max := 24
	if maxworker != 0 {
		max = maxworker
	}
	limit := make(chan struct{}, max*2)
	for i := 0; i < max; i++ {
		limit <- struct{}{}
	}

	bin2tests := make(map[string][]string)

	bugset := bug.NewBugSet()

	bins := cmd.ListFiles(path, func(s string) bool {
		return true
	})

	// bind tests and visitor to bins
	total := 0
	for _, bin := range bins {
		tests := cmd.ListTests(bin)
		bin2tests[bin] = tests
		total += len(tests)
	}

	go func() {
		for bin, tests := range bin2tests {
			for _, test := range tests {
				cfg := fuzzer.DefaultConfig()
				// shared bugset
				cfg.BugSet = bugset
				cfg.Bin = bin
				cfg.Fn = test
				cfg.MaxWorker = 4
				cfg.TimeOut = 30
				cfg.RecoverTimeOut = 200
				cfg.LogCh = logCh
				cfg.MaxQuit = 64
				cfg.MaxExecution = 10000
				cfg.LogLevel = llevel
				if feature == "mu" {
					cfg.UseMutate = false
				}
				if feature == "fb" {
					cfg.UseFeedBack = false
				}

				cov := feedback.NewCov()
				corpus := fuzzer.NewCorpus()
				v := &fuzzer.Visitor{
					V_cov:    cov,
					V_corpus: corpus,
				}
				<-limit
				go func(v *fuzzer.Visitor, cfg *fuzzer.Config) {
					defer func() {
						limit <- struct{}{}
					}()
					m := &fuzzer.Monitor{}
					ok, detail := m.Start(cfg, v, limit)
					var res string
					if ok {
						res = fmt.Sprintf("%s\tFAIL\t%s\n", cfg.Fn, detail[1])
					} else {
						res = fmt.Sprintf("%s\tPASS\n", cfg.Fn)
					}
					resCh <- res
				}(v, cfg)
			}
		}
	}()

	defer fmt.Printf("%v [Fuzzer] Finish\n", time.Now().String())
	cnt := 0
	for {
		select {
		case v := <-resCh:
			fmt.Printf("%v [%v/%v]\t%s", time.Now().String(), cnt+1, total, v)
			cnt += 1
			if cnt == total {
				return
			}
		case v := <-logCh:
			fmt.Printf("%v [WORKER] %s\n", time.Now().String(), v)
		default:
		}
	}
}
