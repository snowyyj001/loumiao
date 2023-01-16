package gorpc

import "sync"

var (
	PoolMgr map[string]*GoRoutinePool
	mLocker sync.Mutex

	MGR *GoRoutineMgr
)

func init() {
	PoolMgr = make(map[string]*GoRoutinePool)

	MGR = &GoRoutineMgr{}
	MGR.go_name_Map = make(map[string]IGoRoutine)
	MGR.go_name_Tmp = make(map[string]IGoRoutine)
}
