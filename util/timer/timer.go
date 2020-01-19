package Timer

import (
	"time"
)

type Timer struct {
	t    *time.Ticker
	done chan bool
}

func NewTicker(dt int, cb func(dt int64)) t *Timer{
	
	t = new(Timer)
	ticker := time.NewTicker(time.Duration(dt) * time.Millisecond)
	
	t.done = make(chan, bool)
	t.t = ticker

	go func(timer *Timer) {
		defer timer.t.Stop(
		utm := time.Now().Unix()
		for {
			select {
			case <-timer.t.C:
				utmPre := time.Now().Unix()
				cb(utmPre - utm)
				utm = utmPre
			case <-timer.done:
				return
			}
		}
	} (t)
}

func (self* Timer)Stop() {
	self.done <- true
}
