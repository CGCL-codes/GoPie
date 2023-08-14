package hack

import (
	"fmt"
	"testing"
)

func TestGetChID(t *testing.T) {
	ch := make(chan int, 5)
	fmt.Printf("chan id := %v\n", GetChID(ch))
	close(ch)
}
