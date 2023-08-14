package compiler

import "testing"

func TestCompiler(t *testing.T) {
	path := "C:\\Users\\Msk\\GolandProjects\\toolkit\\a_test.go"
	deps := []string{}
	c := C{}
	c.compile(path, deps, "toolkit.test.exe")
}
