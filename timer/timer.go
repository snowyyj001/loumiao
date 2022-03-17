package timer

import (
	"github.com/snowyyj001/loumiao/util"
	"time"
)

const (
	DAY_SECONDS  int64  = 86400        //每日秒数
	SECOND_MILLI int64  = 1000         //1秒钟,毫秒
	MINITE_MILLI int64  = 1000 * 60    //1分钟,毫秒
	RFC3339Day   string = "2006_01_02" //RFC3339格式，精确到天
)

type Timer struct {
	t1   *time.Ticker
	t2   *time.Timer
	done chan bool
	loop bool
	over bool
}

//创建一个定时器
//@dt: 时间间隔，毫秒
//@cb：触发回调
//@repeat：是否循环触发
func NewTimer(dt int, cb func(dt int64) bool, repeat bool) *Timer {
	delat := time.Duration(dt) * time.Millisecond

	t := new(Timer)
	t.t2 = time.NewTimer(delat)
	t.done = make(chan bool, 1)
	t.loop = repeat
	t.over = true

	go func(t *Timer) {
		defer t.t2.Stop()
		utm := time.Now().UnixNano() / int64(time.Millisecond)
		utmPre := utm
		for {
			select {
			case <-t.t2.C:
				if !t.over {
					return
				}
				utmPre = time.Now().UnixNano() / int64(time.Millisecond)
				t.over = cb(utmPre - utm)
				if !t.over {
					return
				}
				utm = utmPre
				if t.loop {
					t.t2.Reset(delat)
				}

			case <-t.done:
				return
			}
		}
	}(t)

	return t
}

//创建一个定时器
//@dt: 时间间隔，毫秒
//@cb：触发回调
func NewTicker(dt int, cb func(dt int64) bool) *Timer {

	t := new(Timer)
	ticker := time.NewTicker(time.Duration(dt) * time.Millisecond)

	t.done = make(chan bool, 1)
	t.t1 = ticker
	t.loop = true
	t.over = true

	go func(timer *Timer) {
		defer timer.t1.Stop()
		utm := time.Now().UnixNano() / int64(time.Millisecond)
		utmPre := utm
		for {
			select {
			case <-timer.t1.C:
				if !t.over {
					return
				}
				utmPre = time.Now().UnixNano() / int64(time.Millisecond)
				t.over = cb(utmPre - utm)
				if !t.over {
					return
				}
				utm = utmPre
			case <-timer.done:
				return
			}
		}
	}(t)

	return t
}

func (self *Timer) Stop() {
	self.over = false
	self.done <- true
}

//延迟dt毫秒，执行一个任务
//@dt:延迟时间，毫秒
//@cb:任务
//@sync:是否同步执行
func DelayJob(dt int64, cb func(), sync bool) {
	if sync {
		<-time.After(time.Duration(dt) * time.Millisecond)
		defer util.Recover()
		cb()
	} else {
		go func() {
			<-time.After(time.Duration(dt) * time.Millisecond)
			defer util.Recover()
			cb()
		}()
	}
}

//获取当天的0点和24点时间
//@st：指定那一天，0默认当天
func GetDayTime(st int64) (int64, int64) {
	var timeStr string
	if st == 0 {
		timeStr = time.Now().Format(RFC3339Day)
	} else {
		timeStr = time.Unix(st, 0).Format(RFC3339Day)
	}
	t, _ := time.ParseInLocation(RFC3339Day, timeStr, time.Local)
	var beginTimeNum int64 = t.Unix()
	var endTimeNum int64 = beginTimeNum + DAY_SECONDS
	return beginTimeNum, endTimeNum
}

//是否是同一天
//请确保stmp2 > stmp1
func IsSameDay(stmp1, stmp2 int64) bool {
	if stmp2-stmp1 >= DAY_SECONDS {
		return false
	}
	_, end := GetDayTime(stmp1)
	if stmp2 <= end {
		return false
	}
	return true
}
