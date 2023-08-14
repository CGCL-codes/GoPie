package main

import (
	"fmt"
	"github.com/jessevdk/go-flags"
	"os"
	"strconv"
	"strings"
	"toolkit/cmd"
)

var opts struct {
	T         string `long:"timeout" description:"Instrument single go source file"`
	RT        string `long:"recovertimeout" description:"Output instrumented golang source file to the given file. Only allow when instrumenting single golang source file"`
	PATH      string `long:"path" description:"path"`
	TASK      string `long:"task" description:"task"`
	LL        string `long:"llevel" description:"log level [info, debug, normal]"`
	MaxWoker  string `long:"max" description:"max workers"`
	Fn        string `long:"func" description:"function"`
	Feature   string `long:"feature" description:"[full, fb (without feedback), mu (without mutation)]"`
	LeakCheck string `long:"check" description:"the position of leakcheck [inside, outside]"`
}

func ParseFlags() {
	if _, err := flags.Parse(&opts); err != nil {
		switch flagsErr := err.(type) {
		case flags.ErrorType:
			if flagsErr == flags.ErrHelp {
				os.Exit(0)
			}
			os.Exit(1)
		default:
			os.Exit(1)
		}
	}
}

func main() {
	ParseFlags()
	switch opts.TASK {
	case "lite":
		var timeout, rtimeout int64
		var maxworker int
		if opts.RT != "" {
			rtimeout, _ = strconv.ParseInt(opts.RT, 10, 32)
		}
		if opts.T != "" {
			timeout, _ = strconv.ParseInt(opts.T, 10, 32)
		}
		if opts.MaxWoker != "" {
			max, _ := strconv.ParseInt(opts.MaxWoker, 10, 32)
			maxworker = int(max)
		}
		Lite(opts.PATH, opts.Fn, opts.LL, int(timeout), int(rtimeout), maxworker)
	case "full":
		var maxworker int
		if opts.MaxWoker != "" {
			max, _ := strconv.ParseInt(opts.MaxWoker, 10, 32)
			maxworker = int(max)
		}
		Full(opts.PATH, opts.LL, opts.Feature, maxworker)
	case "inst":
		paths := cmd.ListFiles(opts.PATH, func(s string) bool {
			return strings.HasSuffix(s, ".go")
		})
		pos := "outside"
		if opts.LeakCheck != "" {
			pos = opts.LeakCheck
		}
		Inst(paths, pos)
	case "bins":
		paths := cmd.ListFiles(opts.PATH, func(s string) bool {
			return strings.HasSuffix(s, ".go")
		})
		Bins(paths)
	default:
		fmt.Println("error argument" + " " + opts.TASK)
	}
}
