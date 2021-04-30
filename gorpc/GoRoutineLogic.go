// mongo_logic
package gorpc

import (
	"runtime"
	"time"
	"unsafe"

	"github.com/snowyyj001/loumiao/llog"
	"github.com/snowyyj001/loumiao/util"
	"github.com/snowyyj001/loumiao/util/timer"
)

const (
	ACTION_CLOSE = iota //0
)

const (
	CHAN_BUFFER_LEN = 20000 //channel缓冲数量
	CHAN_BUFFER_MAX = 50000 //channel缓冲最大数量
	CALL_TIMEOUT    = 3     //call超时时间,秒
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
	WriteSync(ct *ChannelContext)
	ReadSync() interface{}
	Send(handler_name string, sdata *M)
	SendActor(handler_name string, sdata interface{})
	SendBack(target IGoRoutine, handler_name string, sdata *M, Cb HanlderFunc)
	Call(target IGoRoutine, handler_name string, sdata *M) interface{}
	RegisterGate(name string, call HanlderNetFunc)
	UnRegisterGate(name string)
	Register(name string, fun HanlderFunc)
	UnRegister(name string)
	CallNetFunc(*M)
	SetSync(sync bool)
	IsRunning() bool
	GetChanLen() int
	IsInited() bool
	SetInited(bool)
	LeftJobNumber() int
}

type GoRoutineLogic struct {
	Name         string                    //队列名字
	Cmd          Cmdtype                   //处理函数集合
	jobChan      chan ChannelContext       //投递任务chan
	readChan     chan ChannelContext       //读取chan
	actionChan   chan int                  //命令控制chan
	chanNum      int                       //协程数量
	NetHandler   map[string]HanlderNetFunc //net hanlder
	goFun        bool                      //true:使用go执行Cmd,GoRoutineLogic非协程安全;false:协程安全
	timer        *time.Timer
	timerCall    func(dt int64)
	timerDuation time.Duration
	started      bool //是否已启动
	inited       bool //是否初始化失败
	ChanSize     int  //job chan size
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
func (self *GoRoutineLogic) IsRunning() bool {
	return self.started
}
func (self *GoRoutineLogic) GetChanLen() int {
	return self.ChanSize
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

func (self *GoRoutineLogic) CallNetFunc(m *M) {
	self.NetHandler[m.Name](self, m.Id, m.Data)
}

//同步定时任务
func (self *GoRoutineLogic) RunTimer(delat int64, f func(int64)) {
	self.timerCall = func(dt int64) {
		defer func() {
			if r := recover(); r != nil {
				buf := make([]byte, 2048)
				l := runtime.Stack(buf, false)
				llog.Errorf("GoRoutineLogic.RunTimer[%s] %v: %s", self.Name, r, buf[:l])
			}
		}()
		f(dt)
	}
	self.timerDuation = time.Duration(delat) * time.Millisecond
	self.timer.Reset(self.timerDuation)
	//self.timer = time.NewTimer(self.timerDuation)
}

//工作队列
func (self *GoRoutineLogic) woker() {
	defer func() {
		if r := recover(); r != nil {
			buf := make([]byte, 2048)
			l := runtime.Stack(buf, false)
			llog.Errorf("GoRoutineLogic.woker[%s] %v: %s", self.Name, r, buf[:l])
		}
	}()
	utm := util.TimeStamp()
	utmPre := utm

	for {
		//llog.Debugf("woker run: %s", self.Name)
		select {
		case ct := <-self.jobChan:
			//llog.Debugf("jobchan single: %s %s", self.Name, ct.Handler)
			if ct.Cb != nil && ct.ReadChan == nil { //callback, for remote actor return back, remote actor should set ReadChan = nil, look line:172
				self.CallFunc(ct.Cb, &ct.Data)
			} else { //
				var hd = self.Cmd[ct.Handler]
				if hd == nil {
					llog.Errorf("GoRoutineLogic[%s] handler is nil: %s", self.Name, ct.Handler)
				} else {
					if self.goFun {
						go self.CallGoFunc(hd, &ct)
					} else {
						ret := self.CallFunc(hd, &ct.Data)
						if ct.ReadChan != nil {
							retctx := ChannelContext{Cb: ct.Cb}
							retctx.Data.Flag = true
							retctx.Data.Data = ret
							ct.ReadChan <- retctx
						}
					}
				}
			}
			//llog.Debugf("jobchan single done: %s %s", self.Name, ct.Handler)
		case action := <-self.actionChan:
			if action == ACTION_CLOSE {
				goto LabelEnd
			}

		case <-self.timer.C:
			utmPre = util.TimeStamp()
			if self.timerCall != nil {
				self.timerCall(utmPre - utm)
				self.timer.Reset(self.timerDuation)
			}
			utm = utmPre
		}
	}

LabelEnd:
	self.stop()
}

//处理任务
func (self *GoRoutineLogic) Run() {
	self.started = true
	go self.woker()
}

//关闭任务
func (self *GoRoutineLogic) stop() {
	self.started = false
	if self.timer != nil {
		self.timer.Stop()
		self.timer = nil
	}
	//当一个通道不再被任何协程所使用后，它将逐渐被垃圾回收掉，无论它是否已经被关闭
	//这里不关闭jobChan和readChan，让gc处理他们
	close(self.actionChan)

}

//关闭任务
func (self *GoRoutineLogic) Close() {
	self.started = false //先标记关闭
	self.actionChan <- ACTION_CLOSE
	llog.Debugf("GoRoutineLogic.Close: %s", self.Name)
}

//延迟关闭任务，等待工作队列清空
func (self *GoRoutineLogic) CloseCleanly() {
	self.started = false //先标记关闭
	timer.NewTicker(1000, func(dt int64) bool {
		if self.LeftJobNumber() == 0 {
			timer.DelayJob(1000, func() {
				self.actionChan <- ACTION_CLOSE
			}, true)
			return false
		}
		//llog.Debugf("CloseCleanly: %d", self.LeftJobNumber())
		return true
	})
}

//投递任务，给自己
func (self *GoRoutineLogic) Send(handler_name string, sdata *M) {
	if self.started == false {
		llog.Warningf("GoRoutineLogic.Send has not started: %s, %s, %v", self.Name, handler_name, sdata)
		return
	}
	if len(self.GetJobChan()) > self.GetChanLen()*2 {
		llog.Noticef("GoRoutineLogic[%s].Send too many job chan: %s", self.Name, handler_name)
		return
	}
	job := ChannelContext{Handler: handler_name}
	if sdata != nil {
		job.Data = *sdata
	}
	self.GetJobChan() <- job
}

//投递任务，给自己
func (self *GoRoutineLogic) SendActor(handler_name string, sdata interface{}) {
	if self.started == false {
		llog.Warningf("GoRoutineLogic.SendActor has not started: %s, %s, %v", self.Name, handler_name, sdata)
		return
	}
	if len(self.GetJobChan()) > self.GetChanLen()*2 {
		llog.Noticef("GoRoutineLogic[%s].SendActor too many job chan: %s", self.Name, handler_name)
		return
	}
	m := M{Data: sdata, Flag: true}
	job := ChannelContext{handler_name, m, nil, nil}
	self.GetJobChan() <- job
}

//投递任务,拥有回调
func (self *GoRoutineLogic) SendBack(server IGoRoutine, handler_name string, sdata *M, Cb HanlderFunc) {
	if self.started == false {
		llog.Warningf("GoRoutineLogic.SendBack has not started: %s, %s, %v", self.Name, handler_name, sdata)
		return
	}
	if server == nil {
		llog.Noticef("GoRoutineLogic[%s].SendBack target[%s] is nil: %s", self.Name, server.GetName(), handler_name)
		return
	}
	if len(server.GetJobChan()) > server.GetChanLen()*2 {
		llog.Noticef("GoRoutineLogic[%s].SendBack[%s] too many job chan: %s", self.Name, server.GetName(), handler_name)
		return
	}
	job := ChannelContext{handler_name, *sdata, self.jobChan, Cb}
	server.GetJobChan() <- job
}

//阻塞读取数据式投递任务，一直等待（超过三秒属于异常）
func (self *GoRoutineLogic) Call(server IGoRoutine, handler_name string, sdata *M) interface{} {
	if self.started == false {
		llog.Warningf("GoRoutineLogic.Call has not started: %s, %s, %v", self.Name, handler_name, sdata)
		return nil
	}
	if server == nil {
		llog.Noticef("GoRoutineLogic[%s].Call target[%s] is nil: %s", self.Name, server.GetName(), handler_name)
		return nil
	}
	if len(server.GetJobChan()) > server.GetChanLen()*2 {
		llog.Noticef("GoRoutineLogic[%s].Call target[%s] too many jon chan: %s", self.Name, server.GetName(), handler_name)
		return nil
	}
	job := ChannelContext{Handler: handler_name, ReadChan: self.readChan}
	if sdata != nil {
		job.Data = *sdata
	}
	select {
	case server.GetJobChan() <- job:
		rdata := <-self.readChan
		return rdata.Data.Data
	case <-time.After(CALL_TIMEOUT * time.Second):
		llog.Noticef("GoRoutineLogic[%s] Call timeout: %s", self.Name, handler_name)
		return nil
	}
}

func (self *GoRoutineLogic) CallFunc(cb HanlderFunc, data *M) interface{} {
	defer func() {
		if r := recover(); r != nil {
			buf := make([]byte, 2048)
			l := runtime.Stack(buf, false)
			llog.Errorf("GoRoutineLogic.CallFunc[%s] %v: %s", self.Name, r, buf[:l])
		}
	}()
	if data.Flag {
		return cb(self, data.Data)
	} else {
		return cb(self, data)
	}
}

func (self *GoRoutineLogic) CallGoFunc(hd HanlderFunc, ct *ChannelContext) {
	ret := self.CallFunc(hd, &ct.Data)
	if ct.ReadChan != nil {
		retctx := ChannelContext{Cb: ct.Cb}
		retctx.Data.Flag = true
		retctx.Data.Data = ret
		ct.ReadChan <- retctx
	}
}

func (self *GoRoutineLogic) WriteSync(ct *ChannelContext) {
	if self.started == false {
		return
	}
	self.readChan <- *ct
}

func (self *GoRoutineLogic) ReadSync() interface{} {
	if self.started == false {
		return nil
	}
	ct := <-self.readChan
	return ct.Data
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
	n := self.ChanSize
	if n == 0 {
		n = CHAN_BUFFER_LEN
	}
	self.ChanSize = n
	self.jobChan = make(chan ChannelContext, n)
	self.readChan = make(chan ChannelContext)
	self.actionChan = make(chan int, 1)
	self.Cmd = make(Cmdtype)
	self.NetHandler = make(map[string]HanlderNetFunc)
	self.timer = time.NewTimer(1<<63 - 1) //默认没有定时器
	self.timerCall = nil
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
