package kcpgate

import (
	"github.com/snowyyj001/loumiao/gorpc"
	"github.com/snowyyj001/loumiao/llog"
	"github.com/snowyyj001/loumiao/nodemgr"
)

//client connect
func innerConnect(igo gorpc.IGoRoutine, socketId int, data []byte) {
	llog.Debugf("KcpGateServer innerConnect: %d", socketId)
	nodemgr.OnlineNum++

	igoTarget := gorpc.MGR.GetRoutine("GameServer")		//告诉GameServer，client断开了，让GameServer决定如何处理，所以GameServer要注册这个actor
	if igoTarget != nil {
		igoTarget.SendActor("ON_CONNECT", socketId)
	}

}

//client disconnect
func innerDisConnect(igo gorpc.IGoRoutine, socketId int, data []byte) {
	llog.Debugf("KcpGateServer innerDisConnect: %d", socketId)
	nodemgr.OnlineNum--

	igoTarget := gorpc.MGR.GetRoutine("GameServer")		//告诉GameServer，client断开了，让GameServer决定如何处理，所以GameServer要注册这个actor
	if igoTarget != nil {
		igoTarget.SendActor("ON_DISCONNECT", socketId)
	}


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

//client connected to gate
func onClientConnected(uid int, tid int) {
	llog.Debugf("GateServer onClientConnected: uid=%d,tid=%d", uid, tid)

}