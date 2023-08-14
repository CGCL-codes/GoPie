package main

import (
	"log"
	"strings"
	"toolkit/cmd"
	"toolkit/pkg/inst"
	passes "toolkit/pkg/inst/passes"
)

func main() {
	cmd.ParseFlags()
	if cmd.Opts.Dir != "" {
		files := cmd.ListFiles(cmd.Opts.Dir, func(s string) bool {
			return strings.Contains(s, ".go")
		})

		for _, file := range files {
			reg := inst.NewPassRegistry()
			// register passes
			reg.Register("channel", func() inst.InstPass { return &passes.ChRecPass{} })
			reg.Register("select", func() inst.InstPass { return &passes.SelectPass{} })
			reg.Register("lock", func() inst.InstPass { return &passes.LockPass{} })
			// reg.Register("fuzz", func() inst.InstPass { return &passes.FuzzPass{} })
			reg.Register("test", func() inst.InstPass { return &passes.TestPass{Pos: cmd.Opts.Pos} })
			err := cmd.HandleSrcFile(file, reg, reg.ListOfPassNames())
			log.Println("Inst " + file)
			if err != nil {
				log.Printf("error %v", err.Error())
			}
		}
	} else {
		if cmd.Opts.File == "" {
			log.Fatalf("Need source file")
		}
		reg := inst.NewPassRegistry()
		// register passes
		reg.Register("channel", func() inst.InstPass { return &passes.ChRecPass{} })
		reg.Register("select", func() inst.InstPass { return &passes.SelectPass{} })
		reg.Register("lock", func() inst.InstPass { return &passes.LockPass{} })
		// reg.Register("fuzz", func() inst.InstPass { return &passes.FuzzPass{} })
		reg.Register("test", func() inst.InstPass { return &passes.TestPass{Pos: cmd.Opts.Pos} })
		cmd.HandleSrcFile(cmd.Opts.File, reg, reg.ListOfPassNames())
	}
}
