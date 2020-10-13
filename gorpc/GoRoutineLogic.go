// mongo_logic
package gorpc

import (
	"runtime"
	"sync"
	"time"

	"github.com/snowyyj001/loumiao/log"
	"github.com/snowyyj001/loumiao/util/timer"
)

const (
	CHAN_BUFFER_LEN = 20000  //channel缓冲数量
	CHAN_BUFFER_MAX = 100000 //channel缓冲最大数量
	CALL_TIMEOUT    = 3      //call超时时间,秒

)

type Cmdtype map[string]HanlderFunc

type IGoRoutine interface {
	DoInit()     //初始化数据
	DoRegsiter() //注册消息
	DoStart()    //启动完成
	DoDestory()  //销毁数据
	Init(string) //GoRoutineLogic初始化
	Run()        //开始执行
	Stop()       //停止执行
	GetName() string
	GetJobChan() chan ChannelContext
	Send(target string, hanld_name string, sdata interface{})
	SendBack(target string, hanld_name string, sdata interface{}, Cb HanlderFunc)
	Call(target string, hanld_name string, sdata interface{}) interface{}
	RegisterGate(name string, call HanlderNetFunc)
	UnRegisterGate(name string)
	Register(name string, fun HanlderFunc)
	UnRegister(name string)
	CallNetFunc(data interface{})
	SetSync(sync bool)
}

type GoRoutineLogic struct {
	Name       string                    //队列名字
	Cmd        Cmdtype                   //处理函数集合
	JobChan    chan ChannelContext       //投递任务chan
	ReadChan   chan ChannelContext       //读取chan
	chanNum    int                       //协程数量
	Level      int                       //队列协程等级
	ChanIndex  int                       //当前使用协程
	NetHandler map[string]HanlderNetFunc //net hanlder
	GoFun      bool                      //true:使用go执行Cmd,GoRoutineLogic非协程安全;false:协程安全
	wg         sync.WaitGroup
	Ticker     *timer.Timer
}

func (self *GoRoutineLogic) DoInit() {
}
func (self *GoRoutineLogic) DoRegsiter() {
}
func (self *GoRoutineLogic) DoStart() {
	log.Infof("%s DoStart", self.Name)
}
func (self *GoRoutineLogic) DoDestory() {
}
func (self *GoRoutineLogic) GetName() string {
	return self.Name
}
func (self *GoRoutineLogic) GetJobChan() chan ChannelContext {
	return self.JobChan
}
func (self *GoRoutineLogic) SetSync(sync bool) {
	self.GoFun = sync
}

func (self *GoRoutineLogic) CallNetFunc(data interface{}) {
	m := data.(M)
	self.NetHandler[m.Name](self, m.Id, m.Data)
}

//同步定时任务
func (self *GoRoutineLogic) RunTimer(delat int, f HanlderFunc) {
	// callback
	var cb = func(dt int64) bool {
		job := ChannelContext{"", dt, nil, f}
		self.JobChan <- job
		return true
	}
	self.Ticker = timer.NewTimer(delat, cb, true)
}

//同步定时任务
func (self *GoRoutineLogic) RunTicker(delat int, f HanlderFunc) {
	// callback
	var cb = func(dt int64) bool {
		job := ChannelContext{"", dt, nil, f}
		self.JobChan <- job
		return true
	}
	self.Ticker = timer.NewTicker(delat, cb)
}

//工作队列
func (self *GoRoutineLogic) woker() {
	defer self.wg.Done()

	for ct := range self.JobChan {
		//log.Debugf("jobchan single %s %s", self.Name, ct.Handler)
		if ct.Cb != nil && ct.ReadChan == nil {
			self.CallFunc(ct.Cb, ct.Data)
		} else {
			var hd = self.Cmd[ct.Handler]
			if hd == nil {
				log.Noticef("GoRoutineLogic[%s] handler is nil: %s", self.Name, ct.Handler)
			} else {
				if self.GoFun {
					go self.CallGoFunc(hd, &ct)
				} else {
					ct.Data = self.CallFunc(hd, ct.Data)
					if ct.Data != nil && ct.ReadChan != nil {
						ch := ct.ReadChan
						ct.ReadChan = nil //this is importent, for send callback
						ch <- ct
					}
				}

			}
		}
	}
}

//处理任务
func (self *GoRoutineLogic) Run() {

	self.wg.Add(1)
	go self.woker()
}

//关闭任务
func (self *GoRoutineLogic) Stop() {
	close(self.JobChan)
	self.wg.Wait()
}

//投递任务
func (self *GoRoutineLogic) Send(target string, hanld_name string, sdata interface{}) {
	server := GetGoRoutineMgr().GetRoutine(target)
	if server == nil {
		log.Noticef("GoRoutineLogic.Send:[%s --> %s(is nil)] %s", self.Name, target, hanld_name)
		return
	}
	if len(server.GetJobChan()) > CHAN_BUFFER_MAX {
		log.Warningf("GoRoutineLogic.Send:[%s --> %s(chan overlow)] %s", self.Name, target, hanld_name)
		return
	}
	job := ChannelContext{hanld_name, sdata, nil, nil}
	server.GetJobChan() <- job
}

//投递任务,拥有回调
func (self *GoRoutineLogic) SendBack(target string, hanld_name string, sdata interface{}, Cb HanlderFunc) {
	server := GetGoRoutineMgr().GetRoutine(target)
	if server == nil {
		log.Warningf("GoRoutineLogic.SendBack:[%s --> %s(is nil)] %s", self.Name, target, hanld_name)
		return
	}
	if len(server.GetJobChan()) > CHAN_BUFFER_MAX {
		log.Warningf("GoRoutineLogic.SendBack:[%s --> %s(chan overlow)] %s", self.Name, target, hanld_name)
		return
	}
	job := ChannelContext{hanld_name, sdata, self.JobChan, Cb}
	server.GetJobChan() <- job
}

//阻塞读取数据式投递任务，一直等待（超过三秒属于异常）
func (self *GoRoutineLogic) Call(target string, hanld_name string, sdata interface{}) interface{} {
	server := GetGoRoutineMgr().GetRoutine(target)
	if server == nil {
		log.Noticef("GoRoutineLogic.Call:[%s --> %s(is nil)] %s", self.Name, target, hanld_name)
		return nil
	}
	if len(server.GetJobChan()) > CHAN_BUFFER_MAX {
		log.Warningf("GoRoutineLogic.Call:[%s --> %s(chan overlow)] %s", self.Name, target, hanld_name)
		return
	}
	job := ChannelContext{hanld_name, sdata, self.ReadChan, nil}
	select {
	case server.GetJobChan() <- job:
		rdata := <-self.ReadChan
		return rdata.Data
	case <-time.After(CALL_TIMEOUT * time.Second):
		log.Noticef("GoRoutineLogic[%s] send timeout: %s", self.Name, hanld_name)
		return nil
	}
}

func (self *GoRoutineLogic) CallFunc(cb HanlderFunc, data interface{}) interface{} {
	defer func() {
		if r := recover(); r != nil {
			buf := make([]byte, 4096)
			l := runtime.Stack(buf, false)
			log.Errorf("%s CallFunc %v: %s\n%v", self.Name, r, buf[:l], data)
		}
	}()

	if cb != nil {
		return cb(self, data)
	}

	return nil
}

func (self *GoRoutineLogic) CallGoFunc(hd HanlderFunc, ct *ChannelContext) {
	ct.Data = self.CallFunc(hd, ct.Data)
	if ct.Data != nil && ct.ReadChan != nil {
		ch := ct.ReadChan
		ct.ReadChan = nil
		ch <- *ct
	}
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

func (self *GoRoutineLogic) Init(name string) {
	self.Name = name
	self.JobChan = make(chan ChannelContext, CHAN_BUFFER_LEN)
	self.ReadChan = make(chan ChannelContext)
	self.Cmd = make(Cmdtype)
	self.NetHandler = make(map[string]HanlderNetFunc)
}

func ServiceHandler(igo IGoRoutine, data interface{}) interface{} {
	m := data.(M)
	igo.CallNetFunc(m)
	return nil
}
