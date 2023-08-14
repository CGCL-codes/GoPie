package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"sync"
	"toolkit/cmd"
)

const (
	localGo = "C:\\Users\\Msk\\go\\go1.19\\bin\\go"
	linuxgo = "go"
)

func dirname(s string) string {
	if strings.Contains(s, ".go") {
		idx := strings.LastIndex(s, "/")
		if idx != -1 {
			return s[0:idx]
		}
	}
	return ""
}

func Bins(paths []string) {
	resCh := make(chan string, 100)
	tests := make([]string, 0)
	var mu sync.Mutex
	gopath := localGo
	if runtime.GOOS == "linux" {
		gopath = linuxgo
	}

	limit := make(chan struct{}, 32)
	for i := 0; i < 16; i++ {
		limit <- struct{}{}
	}

	workpath, _ := os.Getwd()
	dowork := func(dir string) {
		<-limit
		defer func() {
			limit <- struct{}{}
		}()
		opath := workpath + "/testbins/" + strings.Replace(dir, "/", "_", -1)
		c := fmt.Sprintf("cd %s && %s test -o %s -c .", dir, gopath, opath)
		command := exec.Command("bash", "-c", c)
		var out, out2 bytes.Buffer
		command.Stdout = &out
		command.Stderr = &out2
		err := command.Run()
		if err == nil {
			resCh <- fmt.Sprintf("Handle\t%s OK", opath)
			t := cmd.ListTests(opath)
			mu.Lock()
			tests = append(tests, t...)
			mu.Unlock()
		} else {
			resCh <- fmt.Sprintf("Handle\t%s FAIL", dir)
		}
	}

	m := make(map[string]struct{})
	dirs := make([]string, 0)
	for _, path := range paths {
		dir := dirname(path)
		if _, ok := m[dir]; !ok && dir != "" {
			dirs = append(dirs, dir)
			m[dir] = struct{}{}
		}
	}
	for _, dir := range dirs {
		go dowork(dir)
	}

	all := len(dirs)
	for {
		select {
		case v := <-resCh:
			fmt.Printf("[%v/%v]\t%s\n", len(dirs)-all+1, len(dirs), v)
			all -= 1
			if all == 0 {
				fmt.Printf("Finish, Find %v Tests:\n", len(tests))
				for _, test := range tests {
					fmt.Println(test)
				}
				return
			}
		default:
		}
	}
}
