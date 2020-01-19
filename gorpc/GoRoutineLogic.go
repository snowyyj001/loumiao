// mongo_logic
package gorpc

import (
	"github.com/snowyyj001/loumiao/log"
	"runtime"
	"sync"
	"time"
)

const (
	CHAN_BUFFER_LEN = 20000 //channel缓冲数量
)

type Cmdtype map[string]HanlderFunc

type IGoRoutine interface {
	DoInit()     //初始化数据
	DoRegsiter() //注册消息
	DoDestory()  //销毁数据
	Init(string) //GoRoutineLogic初始化
	Run()        //开始执行
	Stop()       //停止执行
	GetName() string
	Send(target string, hanld_name string, sdata interface{})
	SendBack(target string, hanld_name string, sdata interface{}, Cb HanlderFunc)
	Call(target string, hanld_name string, sdata interface{}) interface{}
	RegisterGate(name string, call HanlderNetFunc)
	UnRegisterGate(name string)
	Register(name string, fun HanlderFunc)
	UnRegister(name string)
	GetJobChan() chan *ChannelContext
	CallNetFunc(data interface{})
	SetSync(sync bool)
}

type GoRoutineLogic struct {
	Name       string                    //队列名字
	Cmd        Cmdtype                   //处理函数集合
	JobChan    chan *ChannelContext      //投递任务chan
	ReadChan   chan *ChannelContext      //读取chan
	chanNum    int                       //协程数量
	Level      int                       //队列协程等级
	ChanIndex  int                       //当前使用协程
	NetHandler map[string]HanlderNetFunc //net hanlder
	GoFun      bool                      //true:使用go执行Cmd,GoRoutineLogic非协程安全;false:协程安全
	wg         sync.WaitGroup
}

func (self *GoRoutineLogic) DoInit() {
}
func (self *GoRoutineLogic) DoRegsiter() {
}
func (self *GoRoutineLogic) DoDestory() {
}
func (self *GoRoutineLogic) GetName() string {
	return self.Name
}

func (self *GoRoutineLogic) GetJobChan() chan *ChannelContext {
	return self.JobChan
}
func (self *GoRoutineLogic) SetSync(sync bool) {
	self.GoFun = sync
}

func (self *GoRoutineLogic) CallNetFunc(data interface{}) {
	m := data.(M)
	cid, name, pd := SimpleGNet(m)
	self.NetHandler[name](self, int(cid), pd)
}

//工作队列
func (self *GoRoutineLogic) woker(jc chan *ChannelContext) {
	defer self.wg.Done()

	for ct := range jc {
		if ct.Cb != nil && ct.ReadChan == nil {
			self.CallFunc(ct.Cb, ct.Data)
		} else {
			var hd = self.Cmd[ct.Handler]
			if hd == nil {
				log.Noticef("GoRoutineLogic[%s] handler is nil: %s", self.Name, ct.Handler)
			} else {
				if self.GoFun {
					go self.CallGoFunc(hd, ct)
				} else {
					ct.Data = self.CallFunc(hd, ct.Data)
					if ct.Data != nil && ct.ReadChan != nil {
						ch := ct.ReadChan
						ct.ReadChan = nil
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
	go self.woker(self.JobChan)
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
		log.Noticef("GoRoutineLogic call %s target is nil", hanld_name)
	}

	job := ChannelContext{hanld_name, sdata, nil, nil}
	server.GetJobChan() <- &job
}

//投递任务,拥有回调
func (self *GoRoutineLogic) SendBack(target string, hanld_name string, sdata interface{}, Cb HanlderFunc) {
	server := GetGoRoutineMgr().GetRoutine(target)
	if server == nil {
		log.Noticef("GoRoutineLogic call %s target is nil", hanld_name)
	}

	job := ChannelContext{hanld_name, sdata, self.JobChan, Cb}
	server.GetJobChan() <- &job
}

//阻塞读取数据式投递任务，一直等待（超过三秒属于异常）
func (self *GoRoutineLogic) Call(target string, hanld_name string, sdata interface{}) interface{} {
	server := GetGoRoutineMgr().GetRoutine(target)
	if server == nil {
		log.Noticef("GoRoutineLogic call %s target is nil", hanld_name)
	}

	job := ChannelContext{hanld_name, sdata, self.ReadChan, nil}
	select {
	case server.GetJobChan() <- &job:
		rdata := <-self.ReadChan
		return rdata.Data
	case <-time.After(3 * time.Second):
		log.Noticef("GoRoutineLogic[%s] send timeout: %s", self.Name, hanld_name)
		return nil
	}
}

func (self *GoRoutineLogic) CallFunc(cb HanlderFunc, data interface{}) interface{} {
	defer func() {
		if r := recover(); r != nil {
			buf := make([]byte, 4096)
			l := runtime.Stack(buf, false)
			log.Errorf("CallFunc %v: %s", r, buf[:l])
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
		ch <- ct
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
	self.JobChan = make(chan *ChannelContext, CHAN_BUFFER_LEN)
	self.ReadChan = make(chan *ChannelContext)
	self.Cmd = make(Cmdtype)
	self.NetHandler = make(map[string]HanlderNetFunc)
}

func NetRpC(igo IGoRoutine, data interface{}) interface{} {
	m := data.(M)
	igo.CallNetFunc(m)
	return nil
}
