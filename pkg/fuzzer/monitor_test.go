package fuzzer

import "testing"

const (
	usefeedback = true
)

func TestMonitor(t *testing.T) {
	m := Monitor{}
	m.Start("C:\\Users\\Msk\\GolandProjects\\toolkit\\testdata\\project\\bins\\etcd7443.exe", "_1", usefeedback)
}
