package fuzzer

import (
	"fmt"
	"strconv"
	"strings"
	"sync/atomic"
	"time"
	"toolkit/pkg/bug"
	"toolkit/pkg/feedback"
	"toolkit/pkg/seed"
)

var (
	debug  = false
	info   = false
	normal = true
)

type Monitor struct {
	etimes int32
	max    int32
	doinit uint32
}

type RunContext struct {
	In      Input
	Out     Output
	timeout bool
}

var workerID uint32

func (m *Monitor) Start(cfg *Config, visitor *Visitor, ticket chan struct{}) (bool, []string) {
	if m.max == int32(0) {
		m.max = int32(cfg.MaxExecution)
	}
	m.doinit = uint32(1)
	switch cfg.LogLevel {
	case "debug":
		debug = true
		info = true
	case "info":
		info = true
	default:
	}
	var fncov *feedback.Cov
	var corpus *Corpus
	var maxscore *int32

	if visitor.V_cov == nil {
		fncov = feedback.NewCov()
	} else {
		fncov = visitor.V_cov
	}

	if visitor.V_corpus == nil {
		corpus = NewCorpus()
	} else {
		corpus = visitor.V_corpus
	}

	if visitor.V_score == nil {
		score := int32(10)
		maxscore = &score
	} else {
		maxscore = visitor.V_score
	}

	wid := atomic.AddUint32(&workerID, 1)
	ch := make(chan RunContext)
	cancel := make(chan struct{})
	quit := cfg.MaxQuit

	dowork := func() {
		for {
			var c, ht *Chain
			c, ht = corpus.Get()
			if !cfg.UseMutate || atomic.LoadUint32(&m.doinit) == uint32(1) { // if no feedback, no seed and mutation
				c = nil
				ht = nil
			}
			e := Executor{}
			in := Input{
				c:    c,
				ht:   ht,
				cmd:  cfg.Bin,
				args: []string{"-test.v", "-test.run", cfg.Fn},
				// args:           []string{"-test.v", "-test.run", cfg.Fn, "-test.timeout", "30s"},
				timeout:        cfg.TimeOut,
				recovertimeout: cfg.RecoverTimeOut,
			}
			// atomic.AddInt32(&m.etimes, 1)
			timeout := time.After(1 * time.Minute)
			var istimeout bool
			done := make(chan int)
			var o *Output
			go func() {
				t := e.Run(in)
				o = &t
				close(done)
			}()
			select {
			case <-done:
			case <-timeout:
				istimeout = true
			}
			// if ok {
			//	ticket <- struct{}{}
			//}
			if o == nil {
				continue
			}
			if debug {
				cfg.LogCh <- fmt.Sprintf("%s\t[EXECUTOR] Finish, USE %s", time.Now().String(), o.Time.String())
			}
			ch <- RunContext{In: in, Out: *o, timeout: istimeout}
			select {
			case <-cancel:
				break
			default:
			}
		}
	}

	for i := 0; i < cfg.MaxWorker; i++ {
		go dowork()
	}

	for {
		if m.etimes > m.max {
			close(cancel)
			return false, []string{}
		}
		ctx := <-ch
		atomic.AddInt32(&m.etimes, 1)
		var inputc string
		if ctx.In.c != nil {
			inputc = ctx.In.c.ToString()
		} else {
			inputc = "empty chain"
		}
		if debug {
			cfg.LogCh <- fmt.Sprintf("%s\t[WORKER %v] Input: %s", time.Now().String(), wid, inputc)
		}
		// global corpus is not thread safe now
		if ctx.Out.Err != nil {
			// ignore normal test fail
			if ctx.Out.Time < time.Duration(cfg.TimeOut)*time.Second &&
				(strings.Contains(ctx.Out.O, "panic") || strings.Contains(ctx.Out.O, "found unexpected goroutines") || strings.Contains(ctx.Out.Trace, "all goroutines are asleep - deadlock!")) {
				tfs := bug.TopF(ctx.Out.O)
				exist := cfg.BugSet.Exist(tfs, cfg.Fn)
				if !exist {
					detail := []string{inputc, strconv.FormatInt(int64(atomic.LoadInt32(&m.etimes)), 10), ctx.Out.O}
					if normal {
						if strings.Contains(ctx.Out.Trace, "all goroutines are asleep - deadlock!") {
							cfg.LogCh <- fmt.Sprintf("%s\t[WORKER %v] CRASH [%v] \n %s", time.Now().String(), cfg.BugSet.Size(), inputc, "all goroutines are asleep - deadlock!")
						} else {
							cfg.LogCh <- fmt.Sprintf("%s\t[WORKER %v] CRASH [%v] \n %s", time.Now().String(), cfg.BugSet.Size(), inputc, ctx.Out.O)
						}
					}
					if debug {
						topfs := ""
						for _, f := range tfs {
							topfs += f + "\n"
						}
						cfg.LogCh <- fmt.Sprintf("%s\t[BUG] [%s] TopF : \n%s", time.Now().String(), cfg.Fn, topfs)
					}
					if cfg.SingleCrash {
						close(cancel)
						return true, detail
					}
				}
			}
		}
		op_st, all := feedback.ParseLog(ctx.Out.Trace)
		schedcov := feedback.ParseCovered(ctx.Out.O)
		schedres, coveredinput := ColorCovered(ctx.Out.O, ctx.In.c)

		cov := feedback.Log2Cov(op_st, all)
		score := cov.Score(cfg.UseStates)
		// if len(schedcov) != 0 && cfg.UseCoveredSched {
		//	score += (len(schedcov) / (ctx.In.c.Len())) * len(schedcov) * 10
		// }
		curmax := atomic.LoadInt32(maxscore)
		if int32(score) > curmax {
			atomic.StoreInt32(maxscore, int32(score))
		}
		energy := int(float64(score+1) / float64(curmax) * 100)
		if debug {
			cfg.LogCh <- fmt.Sprintf("%s\t[WORKER %v] score : %v\tenergy %v", time.Now().String(), wid, score, energy)
		}

		init := atomic.LoadUint32(&m.doinit) == 1
		if init && atomic.LoadInt32(&m.etimes) > int32(cfg.InitTurnCnt) {
			atomic.StoreUint32(&m.doinit, 0)
			if cfg.UseMutate {
				fmt.Printf("[MUTATE] SWITCH TO MUTATION MODE, CURRENT INITCNT %v\n", cfg.InitTurnCnt)
			}
		}
		if init {
			if cfg.UseFeedBack {
				go func() { // static analysis at a single routine
					seeds := seed.SRDOAnalysis(op_st)
					seeds = append(seeds, seed.SODRAnalysis(op_st)...)
					if debug {
						if len(seeds) != 0 {
							cfg.LogCh <- fmt.Sprintf("%s\t[WORKER %v] %v SEEDS %s ...", time.Now().String(), wid, len(seeds), seeds[0].ToString())
						}
					}
					corpus.GUpdateSeed(seeds)
				}()
			}
			seeds := seed.RandomSeed(op_st)
			// if debug {
			// 	if len(seeds) != 0 {
			//		cfg.LogCh <- fmt.Sprintf("%s\t[WORKER %v] %v SEEDS %s ...", time.Now().String(), wid, len(seeds), seeds[0].ToString())
			//		}
			// }
			corpus.GUpdateSeed(seeds)
		}
		ok := fncov.Merge(cov)
		if (init && ok) || !cfg.UseGuide || (inputc != "empty chain" && coveredinput.Len() != 0 && !ctx.timeout) {
			corpus.IncSchedCnt(schedres)
			if info {
				cfg.LogCh <- fmt.Sprintf("%s\t[WORKER %v] NEW score: [%v/%v] Input:%s", time.Now().String(), wid, score, curmax, schedres)
			}
			fncov.UpdateR(schedcov)
			quit = cfg.MaxQuit
			if init && ok { // init can get more coverage, do init instead of mutation
				cfg.InitTurnCnt = cfg.InitTurnCnt * 2
				if cfg.InitTurnCnt > 100 {
					cfg.InitTurnCnt = 100
				}
			}
			go func() { // do mutation in a single routine, check the concurrency safety of corpus
				var mu Mutator
				mu = Mutator{Cov: fncov}
				var ncs []*Chain
				var hts map[uint64]map[uint64]struct{}
				if cfg.UseFeedBack {
					ncs, hts = mu.mutate(coveredinput, energy)
				} else {
					ncs, hts = mu.random(ctx.In.c, 100)
				}
				if debug {
					cfg.LogCh <- fmt.Sprintf("%s\t[WORKER %v] MUTATE %s", time.Now().String(), wid, coveredinput.ToString())
				}
				corpus.Update(ncs, hts) // concurrency safe
			}()
		} else {
			quit -= 1
			if quit <= 0 {
				if info {
					cfg.LogCh <- fmt.Sprintf("%s\t[WORKER %v] Fuzzing seems useless, QUIT", time.Now().String(), wid)
				}
				close(cancel)
				return false, []string{}
			}
		}
	}
}
