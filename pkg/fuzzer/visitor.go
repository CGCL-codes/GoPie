package fuzzer

import "toolkit/pkg/feedback"

type Visitor struct {
	V_corpus *Corpus
	V_cov    *feedback.Cov
	V_score  *int32
}
