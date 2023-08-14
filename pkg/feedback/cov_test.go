package feedback

import (
	"fmt"
	"testing"
	"toolkit/pkg/testdata"
)

func TestLog2Cov(t *testing.T) {
	op_st, allops := ParseLog(log)
	cov := Log2Cov(op_st, allops)
	print(cov.ToString())
}

func TestCovUpdate(t *testing.T) {
	op_st, allops := ParseLog(testdata.Log)
	c := Log2Cov(op_st, allops)
	fmt.Print(c.ToString())

	c.UpdateC(OpID(841813590024), ToStatus(ChanFull))
	c.UpdateC(OpID(1337), ToStatus(ChanEmpty))

	c.UpdateO1(1, 2)

	fmt.Print(c.ToString())
}
