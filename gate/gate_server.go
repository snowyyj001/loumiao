// 网关服务
package gate

import (
	"github.com/snowyyj001/loumiao/config"
	"github.com/snowyyj001/loumiao/gorpc"
	"github.com/snowyyj001/loumiao/log"
	"github.com/snowyyj001/loumiao/message"
	"github.com/snowyyj001/loumiao/network"
)

var (
	This        *GateServer
	handler_Map map[string]string //消息报接受者
)

type GateServer struct {
	gorpc.GoRoutineLogic

	pService network.ISocket
}

func (self *GateServer) DoInit() {
	log.Info("GateServer DoInit")
	This = self
	if config.NET_WEBSOCKET {
		self.pService = new(network.WebSocket)
		self.pService.(*network.WebSocket).SetMaxClients(config.NET_MAX_CONNS)
	} else {
		self.pService = new(network.ServerSocket)
		self.pService.(*network.ServerSocket).SetMaxClients(config.NET_MAX_CONNS)
	}

	self.pService.Init(config.NET_GATE_IP, config.NET_GATE_PORT)
	self.pService.SetConnectType(network.CLIENT_CONNECT)
	self.pService.BindPacketFunc(PacketFunc)

	self.pService.Start()

	handler_Map = make(map[string]string)
}

func (self *GateServer) DoRegsiter() {
	self.Register("RegisterNet", RegisterNet)
	self.Register("UnRegisterNet", UnRegisterNet)
	self.Register("SendClient", SendClient)
}

func (self *GateServer) DoDestory() {
	log.Info("GateServer destory")
}

func PacketFunc(socketid int, buff []byte, nlen int) bool {
	defer func() {
		if err := recover(); err != nil {
			log.Errorf("MsgProcess PacketFunc: %v", err)
		}
	}()
	err, name, pm := message.Decode(buff, nlen)
	if err != nil {
		return false
	}

	handler, ok := handler_Map[name]
	if ok {
		This.Send(handler, "NetRpC", gorpc.SimpleNet(socketid, name, pm))
	} else {
		if name != "CONNECT" && name != "DISCONNECT" {
			log.Noticef("MsgProcess PacketFunc handler is nil, drop it[%s]", name)
		}
	}

	return true
}
