package timer

import (
	"github.com/snowyyj001/loumiao/lutil"
	"sync"
	"sync/atomic"
	"time"

	"github.com/snowyyj001/loumiao/llog"
)

/*
关于时间轮算法：
5万个timer的情况下，该算法与系统自带定时器性能差别不大
定时器delat准确与否的关键在于TimeNode.C读取端是否及时
func sendTime(current *TimeNode)函数是在定时器协程内调用的
读取不及时会影响delat精度
既然时间轮算法并没有质的提升，还是用系统的小顶堆吧

*/

const (
	TIME_NEAR_SHIFT  = 8
	TIME_NEAR        = 1 << TIME_NEAR_SHIFT
	TIME_LEVEL_SHIFT = 6
	TIME_LEVEL       = 1 << TIME_LEVEL_SHIFT
	TIME_NEAR_MASK   = TIME_NEAR - 1
	TIME_LEVEL_MASK  = TIME_LEVEL - 1
)

type TimeNode struct {
	next        *TimeNode
	expire      uint32
	cd          uint32
	lasttrigger int
	Session     int64
	C           chan int
	triggerChan chan int
}

type link_list struct {
	head TimeNode
	tail *TimeNode
}

type wl_timer struct {
	near [TIME_NEAR]link_list
	t    [4][TIME_LEVEL]link_list

	time          uint32
	starttime     int
	current_point int
	current       int
}

var (
	ti           *wl_timer
	addTimerChan chan *TimeNode
	timers       sync.Map
	session      int64
)

func link_clear(list *link_list) (ret *TimeNode) {
	ret = list.head.next
	list.head.next = nil
	list.tail = &list.head

	return
}

func link(list *link_list, node *TimeNode) {
	list.tail.next = node
	list.tail = node
	node.next = nil
}

func add_node(T *wl_timer, node *TimeNode) {
	time := int(node.expire)
	current_time := int(T.time)

	if (time | TIME_NEAR_MASK) == (current_time | TIME_NEAR_MASK) {
		link(&T.near[time&TIME_NEAR_MASK], node)
	} else {
		var i int
		mask := int(TIME_NEAR << TIME_LEVEL_SHIFT)
		for i = 0; i < 3; i++ {
			if (time | (mask - 1)) == (current_time | (mask - 1)) {
				break
			}
			mask <<= TIME_LEVEL_SHIFT
		}

		link(&T.t[i][((time>>(TIME_NEAR_SHIFT+i*TIME_LEVEL_SHIFT))&TIME_LEVEL_MASK)], node)
	}
}

func timer_add(T *wl_timer, node *TimeNode) {
	node.cd = uint32(node.cd)
	node.expire = node.cd + T.time
	node.lasttrigger = int(time.Now().UnixNano() / int64(time.Millisecond))
	add_node(T, node)
}

func move_list(T *wl_timer, level int, idx int) {
	current := link_clear(&T.t[level][idx])
	for current != nil {
		temp := current.next
		add_node(T, current)
		current = temp
	}
}

func timer_shift(T *wl_timer) {
	mask := TIME_NEAR
	T.time++
	ct := int(T.time)
	if ct == 0 {
		move_list(T, 3, 0)
	} else {
		time := ct >> TIME_NEAR_SHIFT
		i := 0

		for (ct & (mask - 1)) == 0 {
			idx := time & TIME_LEVEL_MASK
			if idx != 0 {
				move_list(T, i, idx)
				break
			}
			mask <<= TIME_LEVEL_SHIFT
			time >>= TIME_LEVEL_SHIFT
			i++
		}
	}
}

func sendTime(current *TimeNode) {
	// Non-blocking send of time on c.
	// Used in NewTimer, it cannot block anyway (buffer).
	// Used in NewTicker, dropping sends on the floor is
	// the desired behavior when the reader gets behind,
	// because the sends are periodic.
	nt := int(time.Now().UnixNano() / int64(time.Millisecond))
	select {
	case current.C <- int(nt - current.lasttrigger):
		current.lasttrigger = nt
	default:
	}
}

func dispatch_list(current *TimeNode) {
	for current != nil {
		if current.cd > 0 { //has been closed, do not trigger
			sendTime(current)
		}
		temp := current
		current = current.next
		if temp.cd > 0 { //has been closed, do not add again
			temp.expire = temp.cd + ti.time
			add_node(ti, temp)
		}
	}
}

func timer_execute(T *wl_timer) {
	idx := T.time & TIME_NEAR_MASK

	for T.near[idx].head.next != nil {
		current := link_clear(&T.near[idx])
		//SPIN_UNLOCK(T);
		// dispatch_list don't need lock T
		dispatch_list(current)
		//SPIN_LOCK(T);
	}
}

func timer_update(T *wl_timer) {
	// try to dispatch timeout 0 (rare condition)
	timer_execute(T)

	// shift time first, and then dispatch timer message
	timer_shift(T)

	timer_execute(T)
}

func gettime() int {
	return int(time.Now().UnixNano() / 10000000)
}

func loumiao_updatetime() {

	cp := gettime()
	if cp < ti.current_point {
		llog.Errorf("loumiao_updatetime error: change from %d to %d", cp, ti.current_point)
		ti.current_point = cp
	} else if cp != ti.current_point {
		diff := cp - ti.current_point
		ti.current_point = cp
		ti.current += diff
		for i := 0; i < diff; i++ {
			timer_update(ti)
		}
	}
}

func timer_create_timer() *wl_timer {
	r := new(wl_timer)

	for i := 0; i < TIME_NEAR; i++ {
		link_clear(&r.near[i])
	}

	for i := 0; i < 4; i++ {
		for j := 0; j < TIME_LEVEL; j++ {
			link_clear(&r.t[i][j])
		}
	}

	r.current = 0

	return r
}

func LoumiaoStartTimer() {
	ti = timer_create_timer()
	ti.starttime = gettime()
	ti.current = ti.starttime
	ti.current_point = ti.starttime

	addTimerChan = make(chan *TimeNode, 2000)

	go func() {
		defer lutil.Recover()
		for {
			//llog.Debugf("woker run: %s", self.Name)
			select {
			case ev := <-addTimerChan:
				timer_add(ti, ev)
			default:
				loumiao_updatetime()
				time.Sleep(time.Microsecond * 2500) //this effect timer of accuracy
			}
		}
	}()

}

func StartTime() int {
	return ti.starttime * 10
}

func NowTime() int {
	return ti.current * 10
}

func NewWlTimer(dt int) *TimeNode {
	ev := new(TimeNode)
	ev.cd = uint32(dt)
	ev.C = make(chan int, 1)
	ev.triggerChan = make(chan int, 1)
	atomic.AddInt64(&session, 1)
	ev.Session = session
	timers.Store(session, ev)

	addTimerChan <- ev
	return ev
}

func (this *TimeNode) Stop() {
	this.cd = 0
	timers.Delete(this.Session)
}
