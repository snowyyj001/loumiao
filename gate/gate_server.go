// 网关服务
package gate

import (
	"github.com/snowyyj001/loumiao/config"
	"github.com/snowyyj001/loumiao/gorpc"
	"github.com/snowyyj001/loumiao/log"
	"github.com/snowyyj001/loumiao/message"
	"github.com/snowyyj001/loumiao/network"
	"github.com/snowyyj001/loumiao/util/timer"
)

var (
	This        *GateServer
	handler_Map map[string]string //消息报接受者
	rpc_Map     map[string]int    //rpc消息报接受者
)

type GateServer struct {
	gorpc.GoRoutineLogic

	pService      network.ISocket
	clients       map[int]*network.ClientSocket
	pInnerService network.ISocket
	serverMap     map[int]int
	ServerType    int
}

func (self *GateServer) DoInit() {
	log.Infof("%s DoInit", self.Name)
	This = self

	if self.ServerType == network.CLIENT_CONNECT { //对外
		if config.NET_WEBSOCKET {
			self.pService = new(network.WebSocket)
			self.pService.(*network.WebSocket).SetMaxClients(config.NET_MAX_CONNS)
		} else {
			self.pService = new(network.ServerSocket)
			self.pService.(*network.ServerSocket).SetMaxClients(config.NET_MAX_CONNS)
		}
		self.pService.Init(config.NET_GATE_IP, config.NET_GATE_PORT)
		if config.NET_BE_CHILD == 1 { //对内作为client
			self.pService.BindPacketFunc(RpcServerPacketFunc)
		} else {
			self.pService.BindPacketFunc(PacketFunc)
		}
		self.pService.SetConnectType(network.CLIENT_CONNECT)
	}

	if config.NET_BE_CHILD == 1 { //对内作为client
		self.clients = make(map[int]*network.ClientSocket)
	} else if config.NET_BE_CHILD == 2 { //对内作为server
		self.pInnerService = new(network.ServerSocket)
		self.pInnerService.(*network.ServerSocket).SetMaxClients(config.NET_MAX_RPC_CONNS)
		self.pInnerService.Init(config.NET_RPC_IP, config.NET_RPC_PORT)
		self.pInnerService.BindPacketFunc(PacketFunc)
		self.pInnerService.SetConnectType(network.SERVER_CONNECT)
	}

	self.serverMap = make(map[int]int)
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
	if self.pService != nil {
		self.pService.Start()
	}
	if self.pInnerService != nil {
		self.pInnerService.Start()
	}
	if config.NET_BE_CHILD == 1 {
		for _, cfg := range config.ServerCfg.ServerNodes {
			client := self.BuildRpc(cfg.Ip, cfg.Port, cfg.Id, cfg.Name)
			if self.EnableRpcClient(client) {
				self.clients[client.Uid] = client
			}
		}
	}

	if config.NET_BE_CHILD == 1 {
		handler_Map["DISCONNECT"] = "GateServer"
		self.RegisterGate("DISCONNECT", InnerDisConnect)
	}
	if config.NET_BE_CHILD == 2 {
		handler_Map["LouMiaoHandShake"] = "GateServer"
		self.RegisterGate("LouMiaoHandShake", InnerHandShake)
	}
	if config.NET_BE_CHILD != 0 {
		handler_Map["LouMiaoRpcMsg"] = "GateServer"
		self.RegisterGate("LouMiaoRpcMsg", InnerRpcMsg)
		handler_Map["LouMiaoHeartBeat"] = "GateServer"
		self.RegisterGate("LouMiaoHeartBeat", InnerHeartBeat)
	}
}

func (self *GateServer) EnableRpcClient(client *network.ClientSocket) bool {
	log.Info("GateServer EnableRpcClient")
	if client.Start() {
		log.Infof("GateServer rpc connected %s:%d", client.GetIP(), client.GetPort())

		req := LouMiaoHeartBeat{Uid: client.Uid}
		client.SendMsg("LouMiaoHandShake", req)

		timer.NewTimer(3000, func(dt int64) { //heart beat
			client.SendTimes++
			if client.SendTimes > 3 { //time out?
				client.Close()
			} else {
				req := &LouMiaoHeartBeat{Uid: client.Uid}
				client.SendMsg("LouMiaoHeartBeat", req)
			}
		}, true)
		return true
	} else {
		return false
	}
}

func (self *GateServer) DoDestory() {
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
		m := gorpc.M{Id: socketid, Name: name, Data: pm}
		if handler == "GateServer" {
			This.CallNetFunc(m)
		} else {
			This.Send(handler, "NetRpC", m)
		}
	} else {
		if name != "CONNECT" && name != "DISCONNECT" {
			log.Noticef("MsgProcess self handler is nil, drop it[%s]", name)
		}
	}
	return true
}

func RpcClientPacketFunc(socketid int, buff []byte, nlen int) bool {
	defer func() {
		if err := recover(); err != nil {
			log.Errorf("RpcClientPacketFunc self: %v", err)
		}
	}()

	err, name, pm := message.Decode(buff, nlen)
	log.Debugf("RpcClientPacketFunc %s %v", name, pm)
	if err != nil {
		return false
	}

	uid, ok := rpc_Map[name]
	if ok { //msg to client
		resp := pm.(*LouMiaoRpcMsg)
		This.pService.SendById(resp.ClientId, resp.Buffer)
	} else { //msg to local service
		handler, ok := handler_Map[name]
		if ok {
			m := gorpc.M{Id: uid, Name: name, Data: pm}
			if handler == "GateServer" { //msg to gate server
				This.CallNetFunc(m)
			} else { //to local rpc
				This.Send(handler, "NetRpC", m)
			}
		} else {
			log.Warningf("rpc client hanlder[%d] is nil, drop msg[%s]", uid, name)
		}
	}
	return true
}

func RpcServerPacketFunc(socketid int, buff []byte, nlen int) bool {
	defer func() {
		if err := recover(); err != nil {
			log.Errorf("RpcServerPacketFunc self: %v", err)
		}
	}()

	err, name, pm := message.Decode(buff, nlen)
	log.Debugf("RpcServerPacketFunc %s %v", name, pm)
	if err != nil {
		return false
	}

	uid, ok := rpc_Map[name]
	if ok { //msg to other rpc node
		client := This.GetRpcClient(uid)
		if client != nil {
			req := &LouMiaoRpcMsg{ClientId: socketid, Buffer: buff}
			client.SendMsg("LouMiaoRpcMsg", req)
		} else {
			log.Warningf("RpcServerPacketFunc[%d] lost, drop msg[%s]", uid, name)
		}
	} else { //msg to local service
		handler, ok1 := handler_Map[name]
		if ok1 {
			m := gorpc.M{Id: socketid, Name: name, Data: pm}
			if handler == "GateServer" {
				This.CallNetFunc(m)
			} else {
				This.Send(handler, "NetRpC", m)
			}
		} else {
			if name != "CONNECT" && name != "DISCONNECT" {
				log.Noticef("RpcServerPacketFunc msg handler is nil, drop it[%s]", name)
			}
		}
	}

	return true
}

func (self *GateServer) BuildRpc(ip string, port int, id int, nickname string) *network.ClientSocket {
	client := new(network.ClientSocket)
	client.SetClientId(0)
	client.Init(ip, port)
	client.SetConnectType(network.CHILD_CONNECT)
	client.BindPacketFunc(RpcClientPacketFunc)
	client.Uuid = nickname
	client.Uid = id

	return client
}

func (self *GateServer) GetRpcClient(uid int) *network.ClientSocket {
	return self.clients[uid]
}

func (self *GateServer) RegisterRpc(name string, uid int) {
	rpc_Map[name] = uid
}

func init() {
	handler_Map = make(map[string]string)
	rpc_Map = make(map[string]int)
}
