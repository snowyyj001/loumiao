package gorpc

import (
	"github.com/snowyyj001/loumiao/log"
)

//服务启动后，不允许再开新的service
//这样就不用考虑go_name_Map的全局调用问题了，保证线程安全
type GoRoutineMgr struct {
	go_name_Map map[string]IGoRoutine //持久化actor
	go_name_Tmp map[string]IGoRoutine //临时actor
	is_starting bool
	has_started bool
}

var (
	MGR *GoRoutineMgr
)

func init() {
	MGR = &GoRoutineMgr{}
	MGR.go_name_Map = make(map[string]IGoRoutine)
	MGR.go_name_Tmp = make(map[string]IGoRoutine)
}

func (self *GoRoutineMgr) AddRoutine(rou IGoRoutine, name string) {
	if self.go_name_Map[name] != nil || self.go_name_Tmp[name] != nil {
		log.Fatalf("AddRoutine fatal: %s has already been added", name)
		return
	}
	if self.is_starting {
		self.go_name_Tmp[name] = rou

	} else {
		self.go_name_Map[name] = rou
	}
}

//只会获得永久存在的actor
func (self *GoRoutineMgr) GetRoutine(name string) IGoRoutine {
	igo, ok := self.go_name_Map[name]
	if ok {
		return igo
	}
	return nil
}

//关闭单个服务
func (self *GoRoutineMgr) Close(name string) {
	igo := self.GetRoutine(name)
	if igo != nil {
		igo.DoDestory()
		igo.Close()
	}
}

//关闭所有服务
func (self *GoRoutineMgr) CloseAll() {
	for _, igo := range self.go_name_Map {
		igo.DoDestory()
	}
	for _, igo := range self.go_name_Map {
		igo.Close()
	}
}

//创建初始化服务
func (self *GoRoutineMgr) Start(igo IGoRoutine, name string) {

	//GoRoutineLogic
	igo.init(name)

	//init some data
	if igo.DoInit() == false {
		return
	}
	igo.SetInited(true)

	//register handler msg
	igo.DoRegsiter()

	self.AddRoutine(igo, name)
}

//开启服务
//开启所有服务
func (self *GoRoutineMgr) DoStart() {
	self.is_starting = true
	for _, igo := range self.go_name_Map {
		if igo.IsRunning() == false && igo.IsInited() == true {
			igo.Run()
			igo.DoStart()
		}

	}
	for name, igo := range self.go_name_Tmp {
		if igo.IsRunning() == false && igo.IsInited() == true {
			igo.Run()
			igo.DoStart()
		}
		self.go_name_Map[name] = igo
	}
	self.go_name_Tmp = make(map[string]IGoRoutine)

	for key, igo := range self.go_name_Map {
		if igo.IsRunning() == false || igo.IsInited() == false {
			delete(self.go_name_Map, key)
		}
	}
	self.is_starting = false
	self.has_started = true
}

//开启服务
//启动单个服务
func (self *GoRoutineMgr) DoSingleStart(name string) {
	igo, has := self.go_name_Map[name]
	if has {
		if igo.IsRunning() == false && igo.IsInited() == true {
			igo.Run()
			igo.DoStart()
		}
		if igo.IsRunning() == false || igo.IsInited() == false {
			delete(self.go_name_Map, name)
		}
	}
}

//内部rpc调用
//@target: 目标actor
//@funcName: rpc函数
//@data: 函数参数
func (self *GoRoutineMgr) Send(target string, funcName string, data interface{}) {
	igo := self.GetRoutine(target)
	if igo == nil {
		log.Warningf("GoRoutineMgr.Send target[%s] is nil: %s", target, funcName)
		return
	}
	if len(igo.GetJobChan()) > igo.GetChanLen()*2 {
		log.Warningf("GoRoutineMgr.Send:[%s (chan overlow[%d, %d])] %s", target, igo.GetJobChan(), igo.GetChanLen(), funcName)
		return
	}
	job := ChannelContext{funcName, data, nil, nil}
	igo.GetJobChan() <- job
}
