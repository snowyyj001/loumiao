package kcpgate

import (
	"github.com/snowyyj001/loumiao/gorpc"
	"github.com/snowyyj001/loumiao/llog"
	"github.com/snowyyj001/loumiao/message"
)

//client connect
func innerConnect(igo gorpc.IGoRoutine, socketId int, data interface{}) {
	//llog.Debugf("KcpGateServer innerConnect: %d", socketId)
	This.OnlineNum++
}

//client disconnect
func innerDisConnect(igo gorpc.IGoRoutine, socketId int, data interface{}) {
	//llog.Debugf("KcpGateServer innerDisConnect: %d", socketId)
	This.OnlineNum--
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

func recvPackMsg(igo gorpc.IGoRoutine, data interface{}) interface{} {
	m := data.(*gorpc.M)
	socketid := m.Id

	err, _, name, pm := message.Decode(This.Id, m.Data.([]byte), m.Param)

	if err != nil {
		llog.Errorf("KcpGateServer recvPackMsg Decode error: %s", err.Error())
		This.closeClient(socketid)
		return nil
	}

	handler, ok := handler_Map[name]
	if ok {
		if handler == This.Name {
			cb, ok := This.NetHandler[name]
			if ok {
				cb(This, socketid, pm)
			} else {
				llog.Errorf("KcpGateServer recvPackMsg[%s] handler is nil: %s", name, This.Name)
			}
		} else {
			nm := &gorpc.M{Id: socketid, Name: name, Data: pm}
			gorpc.MGR.Send(handler, "ServiceHandler", nm)
		}
	} else {
		llog.Errorf("KcpGateServer recvPackMsg handler is nil, drop it[%s]", name)
	}

	return nil
}
