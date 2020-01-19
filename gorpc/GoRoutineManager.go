package gorpc

import (
	"sync"
)

type GoRoutineMgr struct {
	go_name_Map map[string]IGoRoutine
}

var (
	inst *GoRoutineMgr
	once sync.Once
)

func GetGoRoutineMgr() *GoRoutineMgr {
	once.Do(func() {
		inst = &GoRoutineMgr{}
		inst.go_name_Map = make(map[string]IGoRoutine)

	})
	return inst
}

func (self *GoRoutineMgr) AddRoutine(rou IGoRoutine, name string) {
	self.go_name_Map[name] = rou
}

func (self *GoRoutineMgr) GetRoutine(name string) IGoRoutine {
	igo, ok := self.go_name_Map[name]
	if ok {
		return igo
	}
	return nil
}

func (self *GoRoutineMgr) CloseAll() {
	for _, igo := range self.go_name_Map {
		igo.DoDestory()
	}
	for _, igo := range self.go_name_Map {
		igo.Stop()
	}
}

//创建任务
func (self *GoRoutineMgr) Start(igo IGoRoutine, name string) {

	//GoRoutineLogic
	igo.Init(name)

	//init some data
	igo.DoInit()

	//register handler msg
	igo.DoRegsiter()

	self.AddRoutine(igo, name)
}

//开启任务
func (self *GoRoutineMgr) DoStart() {
	for _, igo := range self.go_name_Map {
		igo.Run()
	}
}
