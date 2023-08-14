package main

import (
	"fmt"
	"toolkit/cmd"
	"toolkit/pkg/bug"
	"toolkit/pkg/fuzzer"
)

func Lite(bin, fn string, llevel string, timeout, recovertimeout int, maxworker int) {
	resCh := make(chan string, 100)
	logCh := make(chan string, 100)
	bugset := bug.NewBugSet()
	nolimit := make(chan struct{})
	close(nolimit)
	dowork := func(bin string, fn string) {
		m := fuzzer.Monitor{}

		cfg := fuzzer.NewConfig(bin, fn, logCh, bugset, "default")
		cfg.LogLevel = llevel
		cfg.TimeOut = timeout
		cfg.RecoverTimeOut = recovertimeout
		cfg.MaxWorker = maxworker

		ok, detail := m.Start(cfg, &fuzzer.Visitor{}, nolimit)
		var res string
		if ok {
			res = fmt.Sprintf("%s\tFAIL\t%s\n", fn, detail[1])
		} else {
			res = fmt.Sprintf("%s\tPASS\n", fn)
		}
		resCh <- res
	}
	fmt.Printf("[FUZZER] Start %s\n", bin)
	var cnt, total int
	if fn != "" {
		for i := 0; i < maxworker; i++ {
			go dowork(bin, fn)
		}
		total = maxworker
	} else {
		tests := cmd.ListTests(bin)
		for _, test := range tests {
			fmt.Printf("[WORKER] Start %s\n", test)
			go dowork(bin, test)
		}
		total = len(tests)
	}
	for {
		select {
		case v := <-resCh:
			fmt.Printf("[%v/%v]\t%s", cnt+1, total, v)
			cnt += 1
			if cnt == total {
				return
			}
		case v := <-logCh:
			fmt.Printf("[WORKER] %s\n", v)
		default:
		}
	}
}
