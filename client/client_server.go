// 客户端服务
package client

import (
	"github.com/snowyyj001/loumiao/config"
	"github.com/snowyyj001/loumiao/gorpc"
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

	pService network.ISocket
}

func (self *ClientServer) DoInit() bool {
	llog.Info("ClientServer DoInit")
	This = self

	self.pService = new(network.ClientSocket)
	//self.client.SetClientId(self.m_ClientId)
	self.pService.Init(config.NET_CLIENT_IP, config.NET_CLIENT_PORT)
	self.pService.SetConnectType(network.SERVER_CONNECT)
	self.pService.BindPacketFunc(PacketFunc)

	handler_Map = make(map[string]string)

	return true
}

func (self *ClientServer) DoRegsiter() {
	llog.Info("ClientServer DoRegsiter")
	self.Register("ServerHanlder", ServerHanlder)
}

func (self *ClientServer) DoStart() {
	llog.Info("ClientServer DoStart")
	self.pService.Start()
}

func (self *ClientServer) DoDestory() {
	llog.Info("ClientServer DoDestory")
}

func PacketFunc(socketid int, buff []byte, nlen int) bool {
	defer func() {
		if err := recover(); err != nil {
			llog.Errorf("MsgProcess PacketFunc: %v", err)
		}
	}()
	err, name, pm := message.Decode(buff, nlen)
	if err != nil {
		return false
	}

	handler, ok := handler_Map[name]
	if ok {
		m := gorpc.M{id: socketid, name: name, data: pm}
		This.Send(handler, "ServiceHandler", m)
	} else {
		if name != "CONNECT" && name != "DISCONNECT" {
			llog.Noticef("MsgProcess PacketFunc handler is nil, drop it[%s]", name)
		}
	}

	return true
}
