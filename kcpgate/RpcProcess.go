package kcpgate

import (
	"github.com/snowyyj001/loumiao/gorpc"
	"github.com/snowyyj001/loumiao/llog"
	"github.com/snowyyj001/loumiao/nodemgr"
)

//client connect
func innerConnect(igo gorpc.IGoRoutine, socketId int, data []byte) {
	//llog.Debugf("KcpGateServer innerConnect: %d", socketId)
	nodemgr.OnlineNum++
}

//client disconnect
func innerDisConnect(igo gorpc.IGoRoutine, socketId int, data []byte) {
	//llog.Debugf("KcpGateServer innerDisConnect: %d", socketId)
	nodemgr.OnlineNum--
}

func registerNet(igo gorpc.IGoRoutine, data interface{}) interface{} {
	m := data.(*gorpc.M)
	sname, ok := handler_Map[m.Name]
	if ok {
		llog.Errorf("registerNet %s has already been registered: %s", m.Name, sname)
		return nil
	}
	handler_Map[m.Name] = m.Data.(string)
	return nil
}