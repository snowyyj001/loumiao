package gorpc

import (
	"sync"

	"github.com/snowyyj001/loumiao/log"
)

type GoRoutineMgr struct {
	go_name_Map map[string]IGoRoutine
	go_name_Tmp map[string]IGoRoutine
	is_starting bool
}

var (
	inst *GoRoutineMgr
	once sync.Once
)

func GetGoRoutineMgr() *GoRoutineMgr {
	once.Do(func() {
		inst = &GoRoutineMgr{}
		inst.go_name_Map = make(map[string]IGoRoutine)
		inst.go_name_Tmp = make(map[string]IGoRoutine)
	})
	return inst
}

func (self *GoRoutineMgr) AddRoutine(rou IGoRoutine, name string) {
	if self.go_name_Map[name] != nil {
		log.Fatalf("AddRoutine error: %s has already been added")
		return
	}
	if self.is_starting {
		self.go_name_Tmp[name] = rou
	} else {
		self.go_name_Map[name] = rou
	}
}

func (self *GoRoutineMgr) GetRoutine(name string) IGoRoutine {
	igo, ok := self.go_name_Map[name]
	if ok {
		return igo
	}
	return nil
}

//关闭所有服务
func (self *GoRoutineMgr) CloseAll() {
	for _, igo := range self.go_name_Map {
		igo.DoDestory()
	}
	for _, igo := range self.go_name_Map {
		igo.Stop()
	}
}

//创建初始化服务
func (self *GoRoutineMgr) Start(igo IGoRoutine, name string) {

	//GoRoutineLogic
	igo.Init(name)

	//init some data
	igo.DoInit()

	//register handler msg
	igo.DoRegsiter()

	self.AddRoutine(igo, name)
}

//开启服务
//开启所有服务
func (self *GoRoutineMgr) DoStart() {
	self.is_starting = true
	for _, igo := range self.go_name_Map {
		if igo.IsRunning() == false {
			igo.Run()
			igo.DoStart()
		}

	}
	for name, igo := range self.go_name_Tmp {
		if igo.IsRunning() == false {
			igo.Run()
			igo.DoStart()
		}
		self.go_name_Map[name] = igo
	}
	self.go_name_Tmp = nil
	self.is_starting = false
}

//开启服务
//启动单个服务
func (self *GoRoutineMgr) DoSingleStart(name string) {
	igo, has := self.go_name_Map[name]
	if has {
		igo.Run()
		igo.DoStart()
	}
}
