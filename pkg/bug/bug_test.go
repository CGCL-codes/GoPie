package bug

import (
	"log"
	"testing"
)

var testlog = "=== RUN   TestLockVerify_1\n" +
	"=== RUN   TestLockVerify_1/call_lockInsideApply_from_inside_apply\n" +
	"=== RUN   TestLockVerify_1/call_lockInsideApply_from_outside_apply_(without_txPostLockInsideApplyHook)\n" +
	"=== RUN   TestLockVerify_1/call_lockInsideApply_from_outside_apply_(with_txPostLockInsideApplyHook)\n" +
	"    logger.go:130: 2023-03-27T16:45:34.233+0800 PANIC   Called outside of APPLY!\n" +
	"=== RUN   TestLockVerify_1/call_lockOutsideApply_from_inside_apply\n" +
	"    logger.go:130: 2023-03-27T16:45:34.236+0800 PANIC   Called inside of APPLY!\n" +
	"=== RUN   TestLockVerify_1/call_Lock_from_unit_test\n" +
	"=== CONT  TestLockVerify_1\n" +
	"    sched.go:331: found unexpected goroutines:\n" +
	"        [Goroutine 136 in state select, with go.etcd.io/etcd/server/v3/storage/backend.(*backend).run on top of the stack:\n" +
	"        goroutine 136 [select]:\n" +
	"        go.etcd.io/etcd/server/v3/storage/backend.(*backend).run(0xc00047ef70)\n" +
	"                /tool/etcd/server/storage/backend/backend.go:466 +0x109\n" +
	"        created by go.etcd.io/etcd/server/v3/storage/backend.newBackend\n" +
	"                /tool/etcd/server/storage/backend/backend.go:227 +0x5ca\n" +
	"        \n" +
	"         Goroutine 138 in state select, with go.etcd.io/etcd/server/v3/storage/backend.(*backend).run on top of the stack:\n" +
	"        goroutine 138 [select]:\n" +
	"        go.etcd.io/etcd/server/v3/storage/backend.(*backend).run(0xc00047f5f0)\n" +
	"                /tool/etcd/server/storage/backend/backend.go:466 +0x109\n" +
	"        created by go.etcd.io/etcd/server/v3/storage/backend.newBackend\n" +
	"                /tool/etcd/server/storage/backend/backend.go:227 +0x5ca\n" +
	"        \n" +
	"         Goroutine 140 in state select, with go.etcd.io/etcd/server/v3/storage/backend.(*backend).run on top of the stack:\n" +
	"        goroutine 140 [select]:\n" +
	"        go.etcd.io/etcd/server/v3/storage/backend.(*backend).run(0xc00047fba0)\n" +
	"                /tool/etcd/server/storage/backend/backend.go:466 +0x109\n" +
	"        created by go.etcd.io/etcd/server/v3/storage/backend.newBackend\n" +
	"                /tool/etcd/server/storage/backend/backend.go:227 +0x5ca\n" +
	"        \n" +
	"         Goroutine 142 in state select, with go.etcd.io/etcd/server/v3/storage/backend.(*backend).run on top of the stack:\n" +
	"        goroutine 142 [select]:\n" +
	"        go.etcd.io/etcd/server/v3/storage/backend.(*backend).run(0xc000217520)\n" +
	"                /tool/etcd/server/storage/backend/backend.go:466 +0x109\n" +
	"        created by go.etcd.io/etcd/server/v3/storage/backend.newBackend\n" +
	"                /tool/etcd/server/storage/backend/backend.go:227 +0x5ca\n" +
	"        \n" +
	"         Goroutine 144 in state select, with go.etcd.io/etcd/server/v3/storage/backend.(*backend).run on top of the stack:\n" +
	"        goroutine 144 [select]:\n" +
	"        go.etcd.io/etcd/server/v3/storage/backend.(*backend).run(0xc000217d40)\n" +
	"                /tool/etcd/server/storage/backend/backend.go:466 +0x109\n" +
	"        created by go.etcd.io/etcd/server/v3/storage/backend.newBackend\n" +
	"                /tool/etcd/server/storage/backend/backend.go:227 +0x5ca\n" +
	"        \n" +
	"         Goroutine 146 in state select, with go.etcd.io/etcd/server/v3/storage/backend.(*backend).run on top of the stack:\n" +
	"        goroutine 146 [select]:\n" +
	"        go.etcd.io/etcd/server/v3/storage/backend.(*backend).run(0xc000321930)\n" +
	"                /tool/etcd/server/storage/backend/backend.go:466 +0x109\n" +
	"        created by go.etcd.io/etcd/server/v3/storage/backend.newBackend\n" +
	"                /tool/etcd/server/storage/backend/backend.go:227 +0x5ca\n" +
	"        ]\n" +
	"--- FAIL: TestLockVerify_1 (0.23s)\n" +
	"    --- PASS: TestLockVerify_1/call_lockInsideApply_from_inside_apply (0.00s)\n" +
	"    --- PASS: TestLockVerify_1/call_lockInsideApply_from_outside_apply_(without_txPostLockInsideApplyHook) (0.00s)\n" +
	"    --- PASS: TestLockVerify_1/call_lockInsideApply_from_outside_apply_(with_txPostLockInsideApplyHook) (0.00s)\n" +
	"    --- PASS: TestLockVerify_1/call_lockOutsideApply_from_outside_apply (0.00s)\n" +
	"    --- PASS: TestLockVerify_1/call_lockOutsideApply_from_inside_apply (0.00s)\n" +
	"    --- PASS: TestLockVerify_1/call_Lock_from_unit_test (0.00s)\n" +
	"FAIL"

var testlog2 = "sched.go:331: found unexpected goroutines:\n" +
	" [Goroutine 152 in state select, with go.etcd.io/etcd/server/v3/storage/backend.(*backend).run on top of the stack:\n" +
	" goroutine 152 [select]:\n" +
	" go.etcd.io/etcd/server/v3/storage/backend.(*backend).run(0xc00019af70)\n" +
	" /tool/etcd/server/storage/backend/backend.go:466 +0x109\n" +
	" created by go.etcd.io/etcd/server/v3/storage/backend.newBackend\n" +
	" /tool/etcd/server/storage/backend/backend.go:227 +0x5ca\n" +
	" \n" +
	" Goroutine 154 in state select, with go.etcd.io/etcd/server/v3/storage/backend.(*backend).run on top of the stack:\n" +
	" goroutine 154 [select]:\n" +
	" go.etcd.io/etcd/server/v3/storage/backend.(*backend).run(0xc00019b5f0)\n" +
	" /tool/etcd/server/storage/backend/backend.go:466 +0x109\n" +
	" created by go.etcd.io/etcd/server/v3/storage/backend.newBackend\n" +
	" /tool/etcd/server/storage/backend/backend.go:227 +0x5ca\n" +
	" \n" +
	" Goroutine 156 in state select, with go.etcd.io/etcd/server/v3/storage/backend.(*backend).run on top of the stack:\n" +
	" goroutine 156 [select]:\n" +
	" go.etcd.io/etcd/server/v3/storage/backend.(*backend).run(0xc00019bba0)\n" +
	" /tool/etcd/server/storage/backend/backend.go:466 +0x109\n" +
	" created by go.etcd.io/etcd/server/v3/storage/backend.newBackend\n" +
	" /tool/etcd/server/storage/backend/backend.go:227 +0x5ca\n" +
	" \n" +
	" Goroutine 158 in state select, with go.etcd.io/etcd/server/v3/storage/backend.(*backend).run on top of the stack:\n" +
	" goroutine 158 [select]:\n" +
	" go.etcd.io/etcd/server/v3/storage/backend.(*backend).run(0xc0004da5b0)\n" +
	" /tool/etcd/server/storage/backend/backend.go:466 +0x109\n" +
	" created by go.etcd.io/etcd/server/v3/storage/backend.newBackend\n" +
	" /tool/etcd/server/storage/backend/backend.go:227 +0x5ca\n" +
	" \n" +
	" Goroutine 160 in state select, with go.etcd.io/etcd/server/v3/storage/backend.(*backend).run on top of the stack:\n" +
	" goroutine 160 [select]:\n" +
	" go.etcd.io/etcd/server/v3/storage/backend.(*backend).run(0xc0004dab60)\n" +
	" /tool/etcd/server/storage/backend/backend.go:466 +0x109\n" +
	" created by go.etcd.io/etcd/server/v3/storage/backend.newBackend\n" +
	" /tool/etcd/server/storage/backend/backend.go:227 +0x5ca\n" +
	" \n" +
	" Goroutine 162 in state select, with go.etcd.io/etcd/server/v3/storage/backend.(*backend).run on top of the stack:\n" +
	" goroutine 162 [select]:\n" +
	" go.etcd.io/etcd/server/v3/storage/backend.(*backend).run(0xc0004db450)\n" +
	" /tool/etcd/server/storage/backend/backend.go:466 +0x109\n" +
	" created by go.etcd.io/etcd/server/v3/storage/backend.newBackend\n" +
	" /tool/etcd/server/storage/backend/backend.go:227 +0x5ca\n" +
	" ]\n" +
	" --- FAIL: TestLockVerify_1 (0.23s)\n" +
	" --- PASS: TestLockVerify_1/call_lockInsideApply_from_inside_apply (0.00s)\n" +
	" --- PASS: TestLockVerify_1/call_lockInsideApply_from_outside_apply_(without_txPostLockInsideApplyHook) (0.00s)\n" +
	" --- PASS: TestLockVerify_1/call_lockInsideApply_from_outside_apply_(with_txPostLockInsideApplyHook) (0.00s)\n" +
	" --- PASS: TestLockVerify_1/call_lockOutsideApply_from_outside_apply (0.00s)\n" +
	" --- PASS: TestLockVerify_1/call_lockOutsideApply_from_inside_apply (0.00s)\n" +
	" --- PASS: TestLockVerify_1/call_Lock_from_unit_test (0.00s)\n" +
	" FAIL\n"

func TestTopF(t *testing.T) {
	fs := TopF(testlog)
	for _, f := range fs {
		t.Log(f)
	}

	fs = TopF(testlog2)
	for _, f := range fs {
		t.Log(f)
	}
}

func TestBugSet(t *testing.T) {
	bs := NewBugSet()
	log.Printf("%v", bs.Exist(TopF(testlog), "Test1"))
	log.Printf("%v", bs.Exist(TopF(testlog2), "Test1"))
}
