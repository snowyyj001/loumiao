// 客户端服务
package client

import (
	"fmt"
	"github.com/snowyyj001/loumiao/gorpc"
	"github.com/snowyyj001/loumiao/lconfig"
	"github.com/snowyyj001/loumiao/llog"
	"github.com/snowyyj001/loumiao/message"
	"github.com/snowyyj001/loumiao/network"
)

var (
	This        *ClientServer
	handler_Map map[string]string //消息报接受者
)

type ClientServer struct {
	gorpc.GoRoutineLogic

	pService *network.ClientSocket
}

func (self *ClientServer) DoInit() bool {
	llog.Info("ClientServer DoInit")
	This = self

	self.pService = new(network.ClientSocket)
	//self.client.SetClientId(self.m_ClientId)
	self.pService.Init(lconfig.NET_GATE_SADDR)
	self.pService.SetConnectType(network.SERVER_CONNECT)
	self.pService.BindPacketFunc(PacketFunc)

	handler_Map = make(map[string]string)

	return true
}

func (self *ClientServer) DoRegister() {
	llog.Info("ClientServer DoRegister")
	self.Register("Serverhandler", gorpc.ServiceHandler)
}

func (self *ClientServer) DoStart() {
	llog.Info("ClientServer DoStart")
	self.pService.Start()
}

func (self *ClientServer) DoDestroy() {
	llog.Info("ClientServer DoDestroy")
}

func PacketFunc(socketid int, buff []byte, nlen int) error {
	_, name, buffbody, err := message.UnPackHead(buff, nlen)

	if err != nil {
		return fmt.Errorf("KcpGateServer recvPackMsg Decode error: %s", err.Error())
	}

	handler, ok := handler_Map[name]
	if ok {
		if handler == This.Name {
			cb, ok := This.NetHandler[name]
			if ok {
				cb(This, socketid, buffbody)
			} else {
				llog.Errorf("ClientServer packetFunc[%s] handler is nil: %s", name, This.Name)
			}
		} else {
			nm := &gorpc.M{Id: socketid, Name: name, Data: buffbody}
			gorpc.MGR.Send(handler, "ServiceHandler", nm)
		}
	} else {
		llog.Errorf("ClientServer recvPackMsg self handler is nil, drop it[%s]", name)
	}
	return nil
}
