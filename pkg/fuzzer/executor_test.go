package fuzzer

import (
	"fmt"
	"testing"
)

func TestExecutor(t *testing.T) {
	e := Executor{}
	o := e.Run(Input{
		c:    nil,
		cmd:  "C:\\Users\\Msk\\GolandProjects\\toolkit\\toolkit.test.exe",
		args: []string{"-test.v"},
	})
	if o.Err == nil {
		fmt.Printf("[Output]\n%s", o.O)
		fmt.Printf("[Output] END\n")
	} else {
		t.Fatalf("%s", o.Err.Error())
	}
}
