// mongo_logic
package gorpc

import (
	"github.com/snowyyj001/loumiao/base"
	"github.com/snowyyj001/loumiao/util"
	"sync"
	"time"
	"unsafe"

	"github.com/snowyyj001/loumiao/llog"
	"github.com/snowyyj001/loumiao/timer"
)

const (
	ACTION_CLOSE = iota //0
)

const (
	CHAN_BUFFER_LEN   = 20000 //channel缓冲数量
	WAIT_TIMEOUT      = 5     //wait超时时间,秒
	CALL_TIMEOUT      = 10    //call超时时间,秒
	CALL_RESP_TIMEOUT = 2     //call回复超时时间,秒
	CHAN_LIMIT_TIMES  = 2     //异步协程上限倍数
	CHAN_OVERLOW      = 30    //chan可能溢出，告警CD，秒
	CHAN_CALL_LONG    = 1000  //call警告，执行超过1秒
	ACTOR_EXEC_LONG   = 200   //actor执行时长过长，毫秒
)

type Cmdtype map[string]HanlderFunc

type IGoRoutine interface {
	DoInit() bool  //初始化数据//非线程安全
	DoRegsiter()   //注册消息//非线程安全
	DoStart()      //启动完成//非线程安全
	DoOpen()       //开始服务//非线程安全
	DoDestory()    //销毁数据//线程是否安全要看调用方是否是自己
	init(string)   //GoRoutineLogic初始化
	Close()        //关闭,线程是否安全要看调用方是否是自己
	CloseCleanly() //关闭,线程是否安全要看调用方是否是自己
	Run()          //开始执行
	stop()         //停止执行
	GetName() string
	SetName(str string)
	GetJobChan() chan ChannelContext
	Send(handler_name string, sdata *M)
	SendActor(handler_name string, sdata interface{})
	SendBack(target IGoRoutine, handler_name string, sdata *M, Cb HanlderFunc)
	Call(target IGoRoutine, handler_name string, sdata *M) (interface{}, bool)
	CallActor(target string, handler_name string, sdata interface{}) (interface{}, bool)
	RegisterGate(name string, call HanlderNetFunc)
	UnRegisterGate(name string)
	Register(name string, fun HanlderFunc)
	UnRegister(name string)
	CallNetFunc(*M)
	SetSync(sync bool)
	IsSync() bool
	IsRunning() bool
	SetRunning()
	IsCoExited() bool
	GetChanLen() int
	GetWarnChanLen() int
	IsInited() bool
	SetInited(bool)
	LeftJobNumber() int
	RunningJobNumber() int
	GetWarningTime() int64
	SetWarningTime(int64)
	WaitResp(name string) interface{}
	SignalResp(name string, data interface{})
	SelfStart(string) //自己管理自己
}

type RoutineTimer struct {
	timer        *timer.Timer
	lastCallTime int64
	timerCall    func(dt int64)
	param        interface{}
}

type GoRoutineLogic struct {
	Name         string                    //队列名字
	Cmd          Cmdtype                   //处理函数集合
	jobChan      chan ChannelContext       //投递任务chan
	actionChan   chan int                  //命令控制chan
	chanNum      int                       //协程数量
	NetHandler   map[string]HanlderNetFunc //net hanlder
	goFun        bool                      //true:使用go执行Cmd,GoRoutineLogic非协程安全;false:协程安全
	cRoLimitChan chan struct{}             //异步协程上限控制
	execTime     int64
	callNum      int //当前call数量

	rpcWaitchan map[string]chan interface{}

	timerChan  chan int //定时器chan
	timerFuncs map[int]*RoutineTimer

	tickerChan   chan int64 //定时器chan
	delayJobs    map[int64]*RoutineTimer
	tickerIndex  int64
	delayJobLock sync.Mutex

	started         int  //是否已启动
	closed          bool //协程是否退出
	inited          bool //是否初始化失败
	ChanSize        int  //job chan size
	chanWarningSize int  //
	lastWarningTime int64
	HasCrashMsg     bool //标记有panic信息
}

func (self *GoRoutineLogic) DoInit() bool {
	//llog.Infof("%s DoInit", self.Name)
	return true
}
func (self *GoRoutineLogic) DoRegsiter() {
	//llog.Infof("%s DoRegsiter", self.Name)
}
func (self *GoRoutineLogic) DoStart() {
	//llog.Infof("%s DoStart", self.Name)
}
func (self *GoRoutineLogic) DoOpen() {
	//llog.Infof("%s DoOpen", self.Name)
}
func (self *GoRoutineLogic) DoDestory() {
	//llog.Infof("%s DoDestory", self.Name)
}
func (self *GoRoutineLogic) GetName() string {
	return self.Name
}
func (self *GoRoutineLogic) SetName(str string) {
	self.Name = str
}
func (self *GoRoutineLogic) GetJobChan() chan ChannelContext {
	return self.jobChan
}
func (self *GoRoutineLogic) SetSync(sync bool) {
	self.goFun = sync
}
func (self *GoRoutineLogic) IsSync() bool {
	return self.goFun
}
func (self *GoRoutineLogic) IsRunning() bool {
	return self.started == 2
}
func (self *GoRoutineLogic) SetRunning() {
	self.started++
}
func (self *GoRoutineLogic) IsCoExited() bool {
	return self.closed
}
func (self *GoRoutineLogic) GetChanLen() int {
	return self.ChanSize
}
func (self *GoRoutineLogic) GetWarnChanLen() int {
	return self.chanWarningSize
}
func (self *GoRoutineLogic) IsInited() bool {
	return self.inited
}
func (self *GoRoutineLogic) SetInited(in bool) {
	self.inited = in
}

func (self *GoRoutineLogic) LeftJobNumber() int {
	return len(self.jobChan)
}

func (self *GoRoutineLogic) RunningJobNumber() int {
	return len(self.cRoLimitChan)
}

func (self *GoRoutineLogic) GetWarningTime() int64 {
	return self.lastWarningTime
}
func (self *GoRoutineLogic) SetWarningTime(cd int64) {
	self.lastWarningTime = cd
}
func (self *GoRoutineLogic) CallNetFunc(m *M) {
	if handler, ok := self.NetHandler[m.Name]; ok {
		handler(self, m.Id, m.Data.([]byte))
	} else {
		llog.Errorf("GoRoutineLogic.CallNetFunc no net handler: %s, %d", m.Name, m.Id)
	}
}

func (self *GoRoutineLogic) SelfStart(name string) {
	//GoRoutineLogic
	self.init(name)

	//init some data
	if self.DoInit() == false {
		return
	}
	self.SetInited(true)

	//register handler msg
	self.DoRegsiter()

	self.Run()

	self.SetRunning()

	self.DoStart()

	self.SetRunning()
}

// RunDelayJob 一次性定时任务，协程安全
// @delta: 延迟时间，单位毫秒
// @cb: 回调函数，在GoRoutineLogic中调用，协程安全
// @param: 回调参数
func (self *GoRoutineLogic) RunDelayJob(delta int, cb func(interface{}), param interface{}) {
	defer self.delayJobLock.Unlock()
	self.delayJobLock.Lock()
	caller := new(RoutineTimer)
	caller.param = param
	self.tickerIndex++
	caller.lastCallTime = self.tickerIndex
	caller.timerCall = func(dt int64) { //try catch errors here, do not effect woker
		cb(caller.param)
	}
	caller.timer = timer.NewTimer(delta, func(dt int64) bool {
		self.tickerChan <- caller.lastCallTime
		return true
	}, false)
	self.delayJobs[caller.lastCallTime] = caller
}

// RunTicker 同步定时任务，不是协程安全的
// @delta: 时间间隔，单位毫秒
// @f: 回调函数，在GoRoutineLogic中调用，协程安全
func (self *GoRoutineLogic) RunTicker(delta int, f func(int64)) {
	if self.IsRunning() {
		llog.Fatalf("can not run a timer when the actor is running: %s", self.Name)
	}
	_, ok := self.timerFuncs[delta]
	if ok {
		llog.Errorf("GoRoutineLogic.RunTimer[%s]: timer duation[%d] already has one", self.Name, delta)
		return
	}
	caller := new(RoutineTimer)
	caller.lastCallTime = base.TimeStamp()
	caller.timerCall = func(dt int64) { //try catch errors here, do not effect woker
		f(dt)
	}
	caller.timer = timer.NewTicker(delta, func(dt int64) bool {
		self.timerChan <- delta
		return true
	})

	self.timerFuncs[delta] = caller
}

func (self *GoRoutineLogic) StopTimerTicker(delta int) {
	if tm, ok := self.timerFuncs[delta]; ok {
		tm.timer.Stop()
		delete(self.timerFuncs, delta)
	}
}

// RunTimer 同步定时任务，不是协程安全的
// @delta: 时间间隔，单位毫秒
// @f: 回调函数，在GoRoutineLogic中调用，协程安全
func (self *GoRoutineLogic) RunTimer(delta int, f func(int64)) {
	if self.IsRunning() {
		llog.Fatalf("can not run a timer when the actor is running: %s", self.Name)
	}
	_, ok := self.timerFuncs[delta]
	if ok {
		llog.Errorf("GoRoutineLogic.RunTimer[%s]: timer duation[%d] already has one", self.Name, delta)
		return
	}
	caller := new(RoutineTimer)
	caller.lastCallTime = base.TimeStamp()
	caller.timerCall = func(dt int64) { //try catch errors here, do not effect woker
		f(dt)
	}
	caller.timer = timer.NewTimer(delta, func(dt int64) bool {
		self.timerChan <- delta
		return true
	}, true)

	self.timerFuncs[delta] = caller
}

// 工作队列
func (self *GoRoutineLogic) woker() {
	defer util.Recover()
	//utm := base.TimeStamp()
	//utmPre := utm

	for {
		//llog.Debugf("woker run: %s", self.Name)
		select {
		case ct := <-self.jobChan:
			//llog.Debugf("jobchan single: %s %s", self.Name, ct.Handler)
			if ct.Cb != nil && ct.ReadChan == nil { //callback, for remote actor return back, remote actor should set ReadChan = nil
				self.CallFunc(ct.Cb, &ct.Data)
			} else { //
				var hd = self.Cmd[ct.Handler]
				if hd == nil {
					llog.Errorf("GoRoutineLogic[%s] handler is nil: %s", self.Name, ct.Handler)
				} else {
					if self.goFun {
						self.cRoLimitChan <- struct{}{}
						go self.CallGoFunc(hd, &ct)
					} else {
						self.execTime = base.TimeStamp()
						ret := self.CallFunc(hd, &ct.Data)
						endTime := base.TimeStamp()
						if endTime-self.execTime > ACTOR_EXEC_LONG {
							llog.Warningf("GoRoutineLogic exec long time: %s, %d", ct.Handler, endTime-self.execTime)
						}
						if ct.ReadChan != nil {
							//llog.Debugf("call back: %s", ct.Handler)
							retctx := ChannelContext{Cb: ct.Cb}
							retctx.Data.Flag = true
							retctx.Data.Data = ret
							//ct.ReadChan <- retctx
							select {
							case ct.ReadChan <- retctx:
							case <-time.After(CALL_RESP_TIMEOUT * time.Second):
								llog.Errorf("GoRoutineLogic[%s] woker resp timeout:  handler = %s", self.Name, ct.Handler)
							}
						}
					}
				}
			}
			//llog.Debugf("jobchan single done: %s %s", self.Name, ct.Handler)
		case action := <-self.actionChan:
			if action == ACTION_CLOSE {
				goto LabelEnd
			}
		case index := <-self.timerChan:
			if self.started == 0 {
				break
			}
			caller, ok := self.timerFuncs[index]
			if ok {
				nt := base.TimeStamp()
				if self.goFun {
					util.Go(func() {
						caller.timerCall(nt - caller.lastCallTime)
					})
				} else {
					caller.timerCall(nt - caller.lastCallTime)
				}
				caller.lastCallTime = nt
			}
		case index := <-self.tickerChan:
			if self.started == 0 {
				break
			}
			self.delayJobLock.Lock()
			caller, ok := self.delayJobs[index]
			if ok {
				delete(self.delayJobs, index)
				self.delayJobLock.Unlock()
				if self.goFun {
					util.Go(func() {
						caller.timerCall(0)
						self.HasCrashMsg = true
					})
				} else {
					caller.timerCall(0)
				}
			} else {
				self.delayJobLock.Unlock()
			}

		}
	}

LabelEnd:
	self.stop()
}

// 处理任务
func (self *GoRoutineLogic) Run() {
	//	self.started = true
	go self.woker()
}

// 关闭任务
func (self *GoRoutineLogic) stop() {
	self.started = 0
	self.closed = true
	for _, caller := range self.timerFuncs {
		caller.timer.Stop()
	}
	defer self.delayJobLock.Unlock()
	self.delayJobLock.Lock()
	for _, caller := range self.delayJobs {
		caller.timer.Stop()
	}
	//当一个通道不再被任何协程所使用后，它将逐渐被垃圾回收掉，无论它是否已经被关闭
	//这里不关闭jobChan和readChan，让gc处理他们
	close(self.actionChan)
}

// 关闭任务
func (self *GoRoutineLogic) Close() {
	self.started = 0 //先标记关闭
	self.actionChan <- ACTION_CLOSE
	llog.Debugf("GoRoutineLogic.Close: %s", self.Name)
}

// 延迟关闭任务，等待工作队列清空
func (self *GoRoutineLogic) CloseCleanly() {
	self.started = 0 //先标记关闭
	//把定时器给关了
	for _, caller := range self.timerFuncs {
		caller.timer.Stop()
	}
	defer self.delayJobLock.Unlock()
	self.delayJobLock.Lock()
	for _, caller := range self.delayJobs {
		caller.timer.Stop()
	}
	timer.NewTicker(1000, func(dt int64) bool {
		if self.LeftJobNumber() == 0 && self.callNum == 0 {
			timer.DelayJob(1000, func() {
				self.actionChan <- ACTION_CLOSE
			}, true)
			return false
		}
		//llog.Debugf("CloseCleanly: %d", self.LeftJobNumber())
		return true
	})
}

// 投递任务，给自己
func (self *GoRoutineLogic) Send(handler_name string, sdata *M) {
	if self.started == 0 {
		llog.Warningf("GoRoutineLogic.Send has not started: %s, %s, %v", self.Name, handler_name, sdata)
		return
	}
	left := self.LeftJobNumber()
	if left > self.GetWarnChanLen() && self.GetWarningTime() < base.TimeStampSec() {
		llog.Errorf("GoRoutineLogic.Send:[src=%s,target=%s (chan may overlow[now=%d, max=%d])] func=%s", self.Name, self.GetName(), self.LeftJobNumber(), self.GetChanLen(), handler_name)
		self.SetWarningTime(base.TimeStampSec() + CHAN_OVERLOW)
	}
	if left > self.GetChanLen() {
		llog.Errorf("GoRoutineLogic.Send:[src=%s,target=%s (chan overlow[now=%d, max=%d])] func=%s", self.Name, self.GetName(), self.LeftJobNumber(), self.GetChanLen(), handler_name)
	}
	job := ChannelContext{Handler: handler_name}
	if sdata != nil {
		job.Data = *sdata
	}
	self.GetJobChan() <- job
}

// 投递任务，给自己
func (self *GoRoutineLogic) SendActor(handler_name string, sdata interface{}) {
	m := M{Data: sdata, Flag: true}
	self.Send(handler_name, &m)
}

// 投递任务,拥有回调
func (self *GoRoutineLogic) SendBack(server IGoRoutine, handler_name string, sdata *M, Cb HanlderFunc) {
	if self.started == 0 {
		llog.Errorf("GoRoutineLogic.SendBack has not started: %s, %s, %v", self.Name, handler_name, sdata)
		return
	}
	if server == nil {
		llog.Errorf("GoRoutineLogic[%s].SendBack target[%s] is nil: %s", self.Name, server.GetName(), handler_name)
		return
	}
	left := server.LeftJobNumber()
	if left > server.GetWarnChanLen() && server.GetWarningTime() < base.TimeStampSec() {
		llog.Noticef("GoRoutineLogic.SendBack:[src=%s,target=%s (chan may overlow[now=%d, max=%d])] func=%s", self.Name, server.GetName(), server.LeftJobNumber(), server.GetChanLen(), handler_name)
		server.SetWarningTime(base.TimeStampSec() + CHAN_OVERLOW)
	}
	if left > server.GetChanLen() {
		llog.Noticef("GoRoutineLogic.SendBack:[src=%s,target=%s (chan overlow[now=%d, max=%d])] func=%s", self.Name, server.GetName(), server.LeftJobNumber(), server.GetChanLen(), handler_name)
	}
	job := ChannelContext{handler_name, *sdata, self.jobChan, Cb}
	server.GetJobChan() <- job
}

// Call 阻塞读取数据式投递任务，一直等待（超过三秒属于异常）
func (self *GoRoutineLogic) Call(server IGoRoutine, handler_name string, sdata *M) (interface{}, bool) {
	if self.started == 0 {
		llog.Errorf("GoRoutineLogic.Call has not started: %s, %s, %v", self.Name, handler_name, sdata)
		return nil, false
	}
	if server == nil {
		llog.Errorf("GoRoutineLogic[%s].Call target[%s] is nil: %s", self.Name, server.GetName(), handler_name)
		return nil, false
	}
	left := server.LeftJobNumber()
	if left > server.GetWarnChanLen() && server.GetWarningTime() > base.TimeStampSec() {
		llog.Noticef("GoRoutineLogic.Call:[src=%s,target=%s (chan may overlow[now=%d, max=%d])] func=%s", self.Name, server.GetName(), server.LeftJobNumber(), server.GetChanLen(), handler_name)
		server.SetWarningTime(base.TimeStampSec() + CHAN_OVERLOW)
	}

	if left > server.GetChanLen() {
		llog.Noticef("GoRoutineLogic.Call:[src=%s,target=%s (chan overlow[now=%d, max=%d])] func=%s", self.Name, server.GetName(), server.LeftJobNumber(), server.GetChanLen(), handler_name)
	}
	readChan := make(chan ChannelContext)
	job := ChannelContext{Handler: handler_name, ReadChan: readChan}
	if sdata != nil {
		job.Data = *sdata
	}
	defer func() {
		self.callNum--
	}()
	self.callNum++
	before := base.TimeStamp()
	server.GetJobChan() <- job
	select {
	case rdata := <-readChan:
		after := base.TimeStamp()
		if after-before > CHAN_CALL_LONG {
			llog.Noticef("GoRoutineLogic.Call:[src=%s,target=%s (use too long[time=%d])] func=%s", self.Name, server.GetName(), after-before, handler_name)
		}
		return rdata.Data.Data, true
	case <-time.After(CALL_TIMEOUT * time.Second):
		llog.Warningf("GoRoutineLogic[%s] to [%s], Call timeout: %s", self.Name, server.GetName(), handler_name)
		return nil, false
	}
}

// CallActor 阻塞读取数据式投递任务，一直等待（超过三秒属于异常）
func (self *GoRoutineLogic) CallActor(target string, handler_name string, sdata interface{}) (interface{}, bool) {
	igo := MGR.GetRoutine(target)
	if igo == nil {
		llog.Errorf("GoRoutineLogic.CallActor target[%s] is nil: %s", target, handler_name)
		return nil, false
	}
	m := &M{Data: sdata, Flag: true}
	return self.Call(igo, handler_name, m)
}

func (self *GoRoutineLogic) CallFunc(cb HanlderFunc, data *M) interface{} {
	defer util.Recover()
	if data.Flag {
		return cb(self, data.Data)
	} else {
		return cb(self, data)
	}
}

func (self *GoRoutineLogic) CallGoFunc(hd HanlderFunc, ct *ChannelContext) {
	self.execTime = base.TimeStamp()
	ret := self.CallFunc(hd, &ct.Data)
	endTime := base.TimeStamp()
	if endTime-self.execTime > ACTOR_EXEC_LONG {
		llog.Warningf("GoRoutineLogic CallGoFunc exec long time: %s, %d", ct.Handler, endTime-self.execTime)
	}
	if ct.ReadChan != nil {
		retctx := ChannelContext{Cb: ct.Cb}
		retctx.Data.Flag = true
		retctx.Data.Data = ret
		select {
		case ct.ReadChan <- retctx:
		case <-time.After(CALL_RESP_TIMEOUT * time.Second):
			llog.Warningf("GoRoutineLogic[%s]CallGoFunc resp timeout: hanlder = %s", self.Name, ct.Handler)
		}
	}
	<-self.cRoLimitChan
}

func (self *GoRoutineLogic) Register(name string, fun HanlderFunc) {
	self.Cmd[name] = fun
}

func (self *GoRoutineLogic) UnRegister(name string) {
	delete(self.Cmd, name)
}

func (self *GoRoutineLogic) RegisterGate(name string, call HanlderNetFunc) {
	self.NetHandler[name] = call
}

func (self *GoRoutineLogic) UnRegisterGate(name string) {
	delete(self.NetHandler, name)
}

func (self *GoRoutineLogic) init(name string) {
	self.Name = name
	self.ChanSize = CHAN_BUFFER_LEN
	self.chanWarningSize = self.ChanSize * 2 / 3
	self.jobChan = make(chan ChannelContext, self.ChanSize)
	//self.readChan = make(chan ChannelContext)
	self.actionChan = make(chan int, 1)
	self.cRoLimitChan = make(chan struct{}, self.ChanSize*CHAN_LIMIT_TIMES)
	self.Cmd = make(Cmdtype)
	self.NetHandler = make(map[string]HanlderNetFunc)
	//self.timer = time.NewTimer(1<<63 - 1) //默认没有定时器
	//self.timerCall = nil
	self.rpcWaitchan = make(map[string]chan interface{})

	self.timerChan = make(chan int, 1)
	self.timerFuncs = make(map[int]*RoutineTimer)
	self.tickerChan = make(chan int64, 1)
	self.delayJobs = make(map[int64]*RoutineTimer)
}

// wait for another gorpc resp
func (self *GoRoutineLogic) WaitResp(name string) interface{} {
	ch, ok := self.rpcWaitchan[name]
	if ok == false {
		ch = make(chan interface{})
		self.rpcWaitchan[name] = ch
	}
	select {
	case r := <-ch:
		return r
	case <-time.After(WAIT_TIMEOUT * time.Second):
		llog.Errorf("GoRoutineLogic.WaitResp: read timeout [%s], data = %v", name)
		return nil
	}
}

// signal the resp is ok
func (self *GoRoutineLogic) SignalResp(name string, data interface{}) {
	ch, ok := self.rpcWaitchan[name]
	if ok == false {
		llog.Errorf("GoRoutineLogic.SignalResp: no read, name = %s, data = %v", name, data)
		return
	}

	select {
	case ch <- data:
	case <-time.After(WAIT_TIMEOUT * time.Second):
		llog.Errorf("GoRoutineLogic.SignalResp: write timeout, name = %s, data = %v", name, data)
	}
}

func ServiceHandler(igo IGoRoutine, data interface{}) interface{} {
	//llog.Debugf("ServiceHandler[%s]: %v", igo.GetName(), data)
	m := data.(*M)
	igo.CallNetFunc(m)
	return nil
}

func CType(igo IGoRoutine) unsafe.Pointer {
	return unsafe.Pointer(igo.(*GoRoutineLogic))
}
