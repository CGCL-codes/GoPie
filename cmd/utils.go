package cmd

import (
	"bytes"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"toolkit/pkg/inst"
	"toolkit/pkg/utils/gofmt"
)

func ListFiles(d string, f func(s string) bool) []string {
	var files []string

	err := filepath.Walk(d, func(path string, info os.FileInfo, err error) error {
		if f(path) {
			files = append(files, path)
		}
		return nil
	})
	if err != nil {
		panic(err)
	}
	return files
}

func ListTests(bin string) []string {
	res := make([]string, 0)
	command := exec.Command(bin, "-test.list", "_1")
	var out bytes.Buffer
	command.Stdout = &out
	err := command.Run()
	if err == nil {
		outstr := out.String()
		ss := strings.Split(outstr, "\n")
		for _, s := range ss {
			if strings.HasPrefix(s, "Test") && strings.HasSuffix(s, "_1") {
				res = append(res, s)
			}
		}
	}
	return res
}

func HandleSrcFile(src string, reg *inst.PassRegistry, passes []string) error {
	iCtx, err := inst.NewInstContext(src)
	if err != nil {
		return err
	}

	err = inst.Run(iCtx, reg, passes)
	if err != nil {
		return err
	}

	var dst string
	if Opts.Out != "" {
		dst = Opts.Out
	} else {
		// dump AST in-place
		dst = iCtx.File

	}
	err = inst.DumpAstFile(iCtx.FS, iCtx.AstFile, dst)
	if err != nil {
		return err
	}

	// check if output is valid, revert if error happened
	if gofmt.HasSyntaxError(dst) {
		// we simply ignored the instrumented result,
		// and revert the file content back to original version.
		err = ioutil.WriteFile(dst, iCtx.OriginalContent, 0777)
		if err != nil {
			log.Panicf("failed to recover file '%s'", dst)
		}
		log.Printf("recovered '%s' from syntax error\n", dst)
	}

	return nil
}
