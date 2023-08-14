package main

import (
	"bytes"
	"fmt"
	"os/exec"
	"runtime"
)

const (
	localPath = "C:\\Users\\Msk\\GolandProjects\\toolkit\\bin\\go_build_toolkit_cmd_inst.exe"
	linuxPath = "./bin/inst"
)

func Inst(paths []string, check_pos string) {
	resCh := make(chan string, 100)
	var toolpath = localPath
	if runtime.GOOS == "linux" {
		toolpath = linuxPath
	}
	dowork := func(path string) {
		command := exec.Command(toolpath, "--file", path, "--checkpos", check_pos)
		var out, out2 bytes.Buffer
		command.Stdout = &out
		command.Stderr = &out2
		err := command.Run()
		if err == nil {
			resCh <- fmt.Sprintf("Handle\t%s OK", path)
		} else {
			resCh <- fmt.Sprintf("Handle\t%s FAIL", path)
		}
	}

	all := len(paths)
	for _, p := range paths {
		go dowork(p)
	}

	for {
		select {
		case v := <-resCh:
			fmt.Printf("[%v/%v]\t%s\n", len(paths)-all+1, len(paths), v)
			all -= 1
			if all == 0 {
				return
			}
		default:
		}
	}
}
