package fuzzer

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"time"
)

type Executor struct {
}

type Output struct {
	Err   error
	O     string
	Trace string
	Time  time.Duration
}

type Input struct {
	c              *Chain
	ht             *Chain
	cmd            string
	args           []string
	timeout        int
	recovertimeout int
}

func (e *Executor) Run(in Input) Output {
	command := exec.Command(in.cmd, in.args...)
	var instr, htstr string

	if in.c == nil {
		instr = "Input="
	} else {
		instr = "Input=" + in.c.ToString()
	}

	command.Env = append(os.Environ(), instr, htstr)
	if in.timeout != 0 {
		command.Env = append(command.Env, fmt.Sprintf("TIMEOUT=%v", in.timeout))
	}
	if in.recovertimeout != 0 {
		command.Env = append(command.Env, fmt.Sprintf("RECOVER_TIMEOUT=%v", in.recovertimeout))
	}

	command.Stdout = &bytes.Buffer{}
	command.Stderr = &bytes.Buffer{}
	//执行命令，直到命令结束
	start := time.Now()
	err := command.Run()
	//打印命令行的标准输出
	return Output{
		err,
		command.Stdout.(*bytes.Buffer).String(),
		command.Stderr.(*bytes.Buffer).String(),
		time.Since(start),
	}
}
