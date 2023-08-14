package compiler

import (
	"bytes"
	"log"
	"os/exec"
)

type C struct {
}

func (c *C) compile(path string, deps []string, output_name string) bool {
	cmd := "C:\\Users\\Msk\\go\\go1.19\\bin\\go"
	args := make([]string, 0)
	args = append(args, "test", "-c", path, "-o", output_name)
	args = append(args, deps...)
	var out, out2 bytes.Buffer
	command := exec.Command(cmd, args...)
	command.Stdout = &out
	command.Stderr = &out2
	err := command.Run()
	if err != nil {
		log.Fatalf("Compile %s failed, %v", path, err)
		log.Fatalf("Output:\n%v", out2)
		return false
	} else {
		log.Printf("Compile %s success", path)
		return true
	}
}
