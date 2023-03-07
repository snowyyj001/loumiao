package gorpc

import (
	"github.com/snowyyj001/loumiao/timer"
	"sync"

	"github.com/snowyyj001/loumiao/llog"
	"github.com/snowyyj001/loumiao/util"
)

const (
	Pool_Statistics_Time = 30 * 1000 //每30s统计一次actor的未读数量
)

func Close() {

	defer mLocker.Unlock()
	mLocker.Lock()
	for _, v := range PoolMgr {
		v.CloseAll()
	}

	MGR.CloseAll()
}

// 协程池，主要用来管理临时开启关闭的协程
type GoRoutinePool struct {
	go_name_Tmp map[int64]IGoRoutine
	mPoolName   string
	mMinitor    sync.Map

	actorLock *sync.RWMutex
}

func (self *GoRoutinePool) Init(name string) {
	defer mLocker.Unlock()
	mLocker.Lock()
	if _, ok := PoolMgr[name]; ok {
		llog.Fatalf("GoRoutinePool pool has been created: %s", name)
	}
	self.go_name_Tmp = make(map[int64]IGoRoutine)
	self.actorLock = &sync.RWMutex{}
	self.mPoolName = name
	PoolMgr[name] = self
	self.Statistics()
}

// 添加一个actor
func (self *GoRoutinePool) AddRoutine(rou IGoRoutine, name int64) bool {
	self.actorLock.Lock()
	if self.go_name_Tmp[name] != nil {
		self.actorLock.Unlock()
		llog.Errorf("GoRoutinePool AddRoutine error: %d has already been added", name)
		return false
	}
	rou.SetName(util.Itoa64(name))
	self.go_name_Tmp[name] = rou
	self.actorLock.Unlock()
	self.mMinitor.Store(name, rou)
	return true
}

// 删除一个actor
func (self *GoRoutinePool) DelRoutine(name int64) {
	self.actorLock.Lock()
	delete(self.go_name_Tmp, name)
	self.actorLock.Unlock()
}

// 删除一个actor，并返回该actor
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

// 获得一个runing的actor
func (self *GoRoutinePool) GetRunRoutine(name int64) IGoRoutine {
	self.actorLock.RLock()
	igo, ok := self.go_name_Tmp[name]
	self.actorLock.RUnlock()
	if ok && igo.IsRunning() {
		return igo
	}
	return nil
}

// 获得一个actor
func (self *GoRoutinePool) GetRoutine(name int64) IGoRoutine {
	self.actorLock.RLock()
	igo, _ := self.go_name_Tmp[name]
	self.actorLock.RUnlock()
	return igo
}

// 关闭单个actor，立刻关闭
func (self *GoRoutinePool) CloseRoutine(name int64) {
	igo := self.GetAndDelRoutine(name)
	if igo != nil {
		igo.DoDestroy()
		igo.Close() //直接关闭
	}
}

// 关闭单个actor，要等待工作队列清空才会关闭
func (self *GoRoutinePool) CloseRoutineCleanly(name int64) {
	igo := self.GetAndDelRoutine(name)
	if igo != nil {
		igo.DoDestroy()
		igo.CloseCleanly() //协程池的关闭要优雅
	}
}

// 关闭所有actor
func (self *GoRoutinePool) CloseAll() {
	self.actorLock.RLock()
	for _, igo := range self.go_name_Tmp {
		igo.DoDestroy()
	}
	for _, igo := range self.go_name_Tmp {
		igo.Close()
	}
	self.actorLock.RUnlock()
}

// 遍历所有actor，效率堪忧
func (self *GoRoutinePool) RangeRoutine(handler func(igo IGoRoutine)) {
	self.actorLock.RLock()
	for _, igo := range self.go_name_Tmp {
		handler(igo)
	}
	self.actorLock.RUnlock()
}

// 创建初始化服务
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
	igo.DoRegister()

	igo.Register("ServiceHandler", ServiceHandler)
	return true
}

// 开启服务
// 开启所有服务
func (self *GoRoutinePool) DoStart() {
	self.actorLock.RLock()
	for _, igo := range self.go_name_Tmp {
		if igo.IsRunning() == false && igo.IsInited() == true {
			igo.Run()
			igo.SetRunning()
			igo.DoStart()
			igo.SetRunning()
		}
	}
	self.actorLock.RUnlock()
}

// 开启服务
// 启动单个服务
// @igo: actor本身
// @name: actor唯一标识
// @doAdd: 是否添加进actor列表
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
	igo.SetRunning()
	igo.DoStart()
	igo.SetRunning()

	if igo.IsRunning() == false || igo.IsInited() == false {
		self.DelRoutine(name)
	}
}

// 获取pool大小
func (self *GoRoutinePool) GetPoolSize() (sz int) {
	self.actorLock.RLock()
	sz = len(self.go_name_Tmp)
	self.actorLock.RUnlock()
	return
}

// 统计当前actor数量
func (self *GoRoutinePool) Statistics() {
	timer.NewTimer(Pool_Statistics_Time, func(dt int64) bool {
		dele := make([]interface{}, 0)
		cnt := 0
		self.mMinitor.Range(func(key, value interface{}) bool {
			igo := value.(IGoRoutine)
			if igo.IsCoExited() {
				dele = append(dele, key)
			} else {
				cnt++
			}
			if igo.IsSync() {
				llog.Debugf("GoRoutinePool.Statistics: poolname = %s, name=%s, jobing=%d,jobleft=%d, jobmax=%d", self.mPoolName, igo.GetName(), igo.RunningJobNumber(), igo.LeftJobNumber(), igo.GetChanLen())
			} else {
				llog.Debugf("GoRoutinePool.Statistics: poolname = %s, name=%s, ,jobleft=%d, jobmax=%d", self.mPoolName, igo.GetName(), igo.LeftJobNumber(), igo.GetChanLen())
			}
			return true
		})
		for _, v := range dele {
			self.mMinitor.Delete(v)
		}
		llog.Debugf("GoRoutinePool.Statistics: name=%s, poolsize=%d, minitortsize=%d", self.mPoolName, self.GetPoolSize(), cnt)
		return true
	}, true)

}

func init() {
}
