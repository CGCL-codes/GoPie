package seed

import (
	"testing"
	"toolkit/pkg/feedback"
	"toolkit/pkg/testdata"
)

func TestSRDOAnalysis(t *testing.T) {
	res, _ := feedback.ParseLog(testdata.Log)
	ps := SRDOAnalysis(res)
	for i, p := range ps {
		print("[", i, "] ", p.ToString(), "\n")
	}
}

func TestSODRAnalysis(t *testing.T) {
	res, _ := feedback.ParseLog(testdata.Log)
	ps := SODRAnalysis(res)
	for i, p := range ps {
		print("[", i, "] ", p.ToString(), "\n")
	}
}
