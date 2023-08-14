package fuzzer

import (
	"testing"
	"toolkit/pkg/feedback"
)

/*
g1           g2

11: lock       21: lock
12: chan <- 1  22: <- chan
13: unlock     23: unlock
*/

func init() {
}

func TestFilter(t *testing.T) {
	// TODO
}

func TestMutate(t *testing.T) {
	// TODO
}

func TestMutateG(t *testing.T) {
	feedback.SetGlobalCov()
}
