package ratelimiter_test

import (
	"fmt"
	"github.com/snowyyj001/loumiao/lutil/ratelimiter"
	"sync"
	"testing"
	"time"
)

// go test -v -run Test_rl RateLimiter_test.go
func Test_rl(t *testing.T) {

	var permitsPerSecond = 100
	var maxPermits = 200
	limter := ratelimiter.Create(permitsPerSecond, maxPermits)

	suc := 0
	a1 := 0
	a2 := 0

	maxv := 600

	wg := &sync.WaitGroup{}

	t1 := time.Now().Unix()

	wg.Add(1)
	go func() {
		for {
			limter.Acquire()
			suc++
			a1++
			if suc >= maxv {
				break
			}
		}
		wg.Done()
	}()

	wg.Add(1)
	go func() {
		for {
			limter.Acquire()
			suc++
			a2++
			if suc >= maxv {
				break
			}
		}
		wg.Done()
	}()

	wg.Add(1)
	go func() {
		for {
			limter.Acquire()
			suc++
			a2++
			if suc >= maxv {
				break
			}
		}
		wg.Done()
	}()

	wg.Wait()

	t2 := time.Now().Unix()

	//ratelimiter起步就有maxPermits可用，然后每秒产生permitsPerSecond个
	//因此maxv个需要t2 - t1秒，忽略小数部分的毫秒
	if int(t2-t1) != (maxv-maxPermits)/(permitsPerSecond) {
		t.FailNow()
	}

	fmt.Println("Test_rl done", a1, a2, suc, t2-t1)
}
