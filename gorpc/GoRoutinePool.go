package gorpc

import (
	"sync"

	"github.com/snowyyj001/loumiao/log"
	"github.com/snowyyj001/loumiao/util"
)

//协程池，主要用来管理临时开启关闭的协程
type GoRoutinePool struct {
	go_name_Tmp map[int64]IGoRoutine

	actorLock *sync.RWMutex
}

func (self *GoRoutinePool) Init() {
	self.go_name_Tmp = make(map[int64]IGoRoutine)
	self.actorLock = &sync.RWMutex{}
}

//添加一个actor
func (self *GoRoutinePool) AddRoutine(rou IGoRoutine, name int64) bool {
	self.actorLock.Lock()
	if self.go_name_Tmp[name] != nil {
		self.actorLock.Unlock()
		log.Errorf("GoRoutinePool AddRoutine error: %d has already been added", name)
		return false
	}
	self.go_name_Tmp[name] = rou
	self.actorLock.Unlock()
	return true
}

//删除一个actor
func (self *GoRoutinePool) DelRoutine(name int64) {
	self.actorLock.Lock()
	delete(self.go_name_Tmp, name)
	self.actorLock.Unlock()
}

//删除一个actor，并返回该actor
func (self *GoRoutinePool) GetAndDelRoutine(name int64) IGoRoutine {
	self.actorLock.Lock()
	igo := self.go_name_Tmp[name]
	if igo != nil {
		delete(self.go_name_Tmp, name)
		self.actorLock.Unlock()
		return igo
	}
	self.actorLock.Unlock()
	return nil
}

//获得一个actor
func (self *GoRoutinePool) GetRoutine(name int64) IGoRoutine {
	self.actorLock.RLock()
	igo, ok := self.go_name_Tmp[name]
	self.actorLock.RUnlock()
	if ok {
		return igo
	}
	return nil
}

//关闭单个actor
func (self *GoRoutinePool) CloseRoutine(name int64) {
	igo := self.GetAndDelRoutine(name)
	if igo != nil {
		igo.DoDestory()
		igo.Close()
	}
}

//关闭单个actor，要等待工作队列清空才会关闭
func (self *GoRoutinePool) CloseRoutineCleanly(name int64) {
	igo := self.GetAndDelRoutine(name)
	if igo != nil {
		igo.CloseCleanly()
	}
}

//关闭所有actor
func (self *GoRoutinePool) CloseAll() {
	self.actorLock.RLock()
	for _, igo := range self.go_name_Tmp {
		igo.DoDestory()
	}
	for _, igo := range self.go_name_Tmp {
		igo.Close()
	}
	self.actorLock.RUnlock()
}

//创建初始化服务
func (self *GoRoutinePool) Start(igo IGoRoutine, name int64) bool {

	//GoRoutineLogic
	igo.init(util.Itoa64(name))

	//init some data
	if igo.DoInit() == true {
		igo.SetInited(true)
	} else {
		self.DelRoutine(name)
		return false
	}

	//register handler msg
	igo.DoRegsiter()

	return true
}

//开启服务
//开启所有服务
func (self *GoRoutinePool) DoStart() {
	self.actorLock.RLock()
	for _, igo := range self.go_name_Tmp {
		if igo.IsRunning() == false && igo.IsInited() == true {
			igo.Run()
			igo.DoStart()
		}
	}
	self.actorLock.RUnlock()
}

//开启服务
//启动单个服务
//@igo: actor本身
//@name: actor唯一标识
//@doAdd: 是否添加进actor列表
func (self *GoRoutinePool) DoSingleStart(igo IGoRoutine, name int64, doAdd bool) {

	if doAdd {
		//add first
		ok := self.AddRoutine(igo, name)
		if !ok {
			return
		}
	}

	if igo.IsRunning() {
		return
	}

	ok := self.Start(igo, name)
	if !ok {
		return
	}
	igo.Run()
	igo.DoStart()

	if igo.IsRunning() == false || igo.IsInited() == false {
		self.DelRoutine(name)
	}
}

//获取pool大小
func (self *GoRoutinePool) GetPoolSize() (sz int) {
	self.actorLock.RLock()
	sz = len(self.go_name_Tmp)
	self.actorLock.RUnlock()
	return
}
