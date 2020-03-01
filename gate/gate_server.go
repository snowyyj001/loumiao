// 网关服务
package gate

import (
	"container/list"
	"strconv"
	"strings"

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

	pService  network.ISocket
	clients   list.List
	serverMap map[int]int
	beClient  bool
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
	self.serverMap = make(map[int]int)

	handler_Map = make(map[string]string)

}

func (self *GateServer) DoRegsiter() {
	log.Info("GateServer DoRegsiter")
	self.Register("RegisterNet", RegisterNet)
	self.Register("UnRegisterNet", UnRegisterNet)
	self.Register("SendClient", SendClient)
	self.Register("SendMulClient", SendMulClient)
}

func (self *GateServer) DoStart() {
	log.Info("GateServer DoStart")
	self.pService.Start()

	for v := self.clients.Front(); v != nil; v = v.Next() {
		client := v.Value.(*network.ClientSocket)
		if client.Start() {
			log.Infof("GateServer rpc connected%s %d", client.GetIP(), client.GetPort())
			self.ShakeHand(client)
		}
	}
}

func (self *GateServer) DoDestory() {
	log.Info("GateServer DoDestory")
}

func PacketFunc(socketid int, buff []byte, nlen int) bool {
	defer func() {
		if err := recover(); err != nil {
			log.Errorf("PacketFunc self: %v", err)
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
		if !This.IsChild() {
			index := strings.Index(name, "$shakehand&")
			if index == 0 {
				uid, err := strconv.Atoi(name[11:])
				if err == nil {
					This.SetRpcClientId(uid, socketid)
				}
			}
		}
		if name == "DISCONNECT" {
			if This.IsChild() {
				for v := This.clients.Front(); v != nil; v = v.Next() {
					client := v.Value.(*network.ClientSocket)
					if socketid == client.GetClientId() {
						This.clients.Remove(v)
						//remote rpc断开，这里重新连接
						This.BuildRpc(client.GetIP(), client.GetPort(), client.Uid, client.Uuid)
						break
					}
				}
			}
		}

		if name != "CONNECT" && name != "DISCONNECT" {
			log.Noticef("MsgProcess self handler is nil, drop it[%s]", name)
		}

	}

	return true
}

func (self *GateServer) IsChild() bool {
	return self.beClient
}

func (self *GateServer) BuildRpc(ip string, port int, id int, nickname string) {
	if !config.NET_BE_CHILD {
		return
	}
	self.beClient = true
	client := new(network.ClientSocket)
	client.SetClientId(0)
	client.Init(ip, port)
	client.SetConnectType(network.SERVER_CONNECT)
	client.BindPacketFunc(PacketFunc)
	client.Uuid = nickname
	client.Uid = id
	self.clients.PushBack(client)
}

func (self *GateServer) ShakeHand(client *network.ClientSocket) {
	client.SendMsg("$shakehand&"+strconv.Itoa(client.Uid), nil)
}

func (self *GateServer) GetClientByUid(uid int) *network.ClientSocket {
	for v := This.clients.Front(); v != nil; v = v.Next() {
		client := v.Value.(*network.ClientSocket)
		if uid == client.Uid {
			return client
		}
	}
	return nil
}

func (self *GateServer) GetClientByUuid(uuid string) *network.ClientSocket {
	for v := This.clients.Front(); v != nil; v = v.Next() {
		client := v.Value.(*network.ClientSocket)
		if uuid == client.Uuid {
			return client
		}
	}
	return nil
}

func (self *GateServer) SetRpcClientId(uid int, cid int) {
	self.serverMap[uid] = cid
}

func (self *GateServer) GetRpcClientId(uid int) int {
	return self.serverMap[uid]
}
