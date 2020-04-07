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
)

type Token struct {
	UserId  int
	TokenId int
}

type GateServer struct {
	gorpc.GoRoutineLogic

	Id                int
	pService          network.ISocket
	clients           map[int]*network.ClientSocket
	pInnerService     network.ISocket
	serverMap         map[int]int
	ServerType        int
	tokens            map[int]*Token
	tokens_u          map[int]int
	OnServerConnected func(int)
	OnClientConnected func(int)
}

func (self *GateServer) DoInit() {
	log.Infof("%s DoInit", self.Name)
	This = self

	if config.NET_NODE_TYPE == config.ServerType_World {
		if config.NET_NODE_ID != 1 {
			log.Errorf("GateServer world node id error [%d]", config.NET_NODE_ID)
			return
		}
	}

	if config.NET_NODE_TYPE < config.ServerType_World || config.NET_NODE_TYPE > config.ServerType_Zoon {
		log.Errorf("GateServer node type error [%d]", config.NET_NODE_TYPE)
		return
	}

	self.Id = config.NET_NODE_ID
	if self.ServerType == network.CLIENT_CONNECT { //对外
		if config.NET_WEBSOCKET {
			self.pService = new(network.WebSocket)
			self.pService.(*network.WebSocket).SetMaxClients(config.NET_MAX_CONNS)
		} else {
			self.pService = new(network.ServerSocket)
			self.pService.(*network.ServerSocket).SetMaxClients(config.NET_MAX_CONNS)
		}
		self.pService.Init(config.NET_GATE_IP, config.NET_GATE_PORT)
		if config.NET_NODE_TYPE == config.ServerType_Gate {
			self.pService.BindPacketFunc(RpcServerPacketFunc)
		} else {
			self.pService.BindPacketFunc(PacketFunc)
		}
		self.pService.SetConnectType(network.CLIENT_CONNECT)
	}

	if config.NET_NODE_TYPE == config.ServerType_Gate {
		self.clients = make(map[int]*network.ClientSocket)
	} else {
		self.pInnerService = new(network.ServerSocket)
		self.pInnerService.(*network.ServerSocket).SetMaxClients(config.NET_MAX_RPC_CONNS)
		self.pInnerService.Init(config.NET_RPC_IP, config.NET_RPC_PORT)
		self.pInnerService.BindPacketFunc(RpcInnerPacketFunc)
		self.pInnerService.SetConnectType(network.SERVER_CONNECT)
	}

	self.serverMap = make(map[int]int)
	self.tokens = make(map[int]*Token)
	self.tokens_u = make(map[int]int)
}

func (self *GateServer) DoRegsiter() {
	log.Info("GateServer DoRegsiter")

	self.Register("RegisterNet", RegisterNet)
	self.Register("UnRegisterNet", UnRegisterNet)
	self.Register("SendClient", SendClient)
	self.Register("SendMulClient", SendMulClient)
	self.Register("SendRpc", SendRpc)

	handler_Map["CONNECT"] = "GateServer"
	self.RegisterGate("CONNECT", InnerConnect)
	handler_Map["DISCONNECT"] = "GateServer"
	self.RegisterGate("DISCONNECT", InnerDisConnect)
	handler_Map["LouMiaoHandShake"] = "GateServer"
	self.RegisterGate("LouMiaoHandShake", InnerHandShake)
	handler_Map["LouMiaoHeartBeat"] = "GateServer"
	self.RegisterGate("LouMiaoHeartBeat", InnerHeartBeat)
	handler_Map["LouMiaoLoginGate"] = "GateServer"
	self.RegisterGate("LouMiaoLoginGate", InnerLoginGate)
}

func (self *GateServer) DoStart() {
	log.Info("GateServer DoStart")

	if self.pService != nil {
		self.pService.Start()
	}
	if self.pInnerService != nil {
		self.pInnerService.Start()
	}
	if config.NET_NODE_TYPE == config.ServerType_Gate {
		for _, cfg := range config.ServerCfg.ServerNodes {
			client := self.BuildRpc(cfg.Ip, cfg.Port, cfg.Id, cfg.Name)
			client.Type = cfg.Type
			if self.EnableRpcClient(client) {
				self.clients[client.Uid] = client
				if self.OnServerConnected != nil {
					self.OnServerConnected(client.Uid)
				}
			}
		}
	}
}

func (self *GateServer) EnableRpcClient(client *network.ClientSocket) bool {
	log.Info("GateServer EnableRpcClient")
	if client.Start() {
		log.Infof("GateServer rpc connected %s:%d", client.GetIP(), client.GetPort())
		req := &LouMiaoHandShake{Uid: client.Uid}
		client.SendMsg("LouMiaoHandShake", req)
		timer.NewTimer(5000, func(dt int64) bool { //heart beat
			client.SendTimes++
			if client.SendTimes > 2 { //time out?
				log.Warningf("UID[%d] to rpc server heartbeat timeout", client.Uid)
				client.Close()
				return false
			} else {
				req := &LouMiaoHeartBeat{Uid: client.Uid}
				client.SendMsg("LouMiaoHeartBeat", req)
			}
			return true
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

func RpcInnerPacketFunc(socketid int, buff []byte, nlen int) bool {
	defer func() {
		if err := recover(); err != nil {
			log.Errorf("RpcClientPacketFunc self: %v", err)
		}
	}()

	err, name, pm := message.Decode(buff, nlen)
	if err != nil {
		return false
	}

	if name == "LouMiaoRpcMsg" {
		resp := pm.(*LouMiaoRpcMsg)
		err, name, pm = message.Decode(resp.Buffer, len(resp.Buffer))
		if err != nil {
			return false
		}
		if This.tokens[resp.ClientId] == nil {
			This.tokens[resp.ClientId] = &Token{TokenId: socketid, UserId: resp.ClientId}
		}
		socketid = resp.ClientId
	}

	{
		handler, ok := handler_Map[name]
		if ok {
			m := gorpc.M{Id: socketid, Name: name, Data: pm}
			if handler == "GateServer" { //msg to gate server
				This.CallNetFunc(m)
			} else { //to local rpc
				This.Send(handler, "NetRpC", m)
			}
		} else {
			log.Warningf("rpc client hanlder[%d] is nil, drop msg[%s]", socketid, name)
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
	if err != nil {
		return false
	}
	//fmt.Println("RpcClientPacketFunc ", name, pm)

	if name == "LouMiaoRpcMsg" {
		resp := pm.(*LouMiaoRpcMsg)
		clientid, ok := This.tokens_u[resp.ClientId]
		if ok {
			This.pService.SendById(clientid, resp.Buffer)
		}
	} else { //msg to local service
		handler, ok := handler_Map[name]
		if ok {
			m := gorpc.M{Id: socketid, Name: name, Data: pm}
			if handler == "GateServer" { //msg to gate server
				This.CallNetFunc(m)
			} else { //to local rpc
				This.Send(handler, "NetRpC", m)
			}
		} else {
			log.Warningf("rpc client hanlder[%d] is nil, drop msg[%s]", socketid, name)
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
	//log.Debugf("RpcServerPacketFunc %s %v", name, pm)
	if err != nil {
		return true
	}

	handler, ok := handler_Map[name]
	if ok { //msg to other rpc node
		m := gorpc.M{Id: socketid, Name: name, Data: pm}
		if handler == "GateServer" {
			This.CallNetFunc(m)
		} else {
			This.Send(handler, "NetRpC", m)
		}
	} else {
		client := This.GetRpcClient(1)
		if client != nil {
			token, ok1 := This.tokens[socketid]
			if !ok1 {
				log.Warningf("RpcServerPacketFunc[%d] tokens nil, drop msg[%s]", socketid, name)
				if config.NET_WEBSOCKET {
					This.pService.(*network.WebSocket).StopClient(socketid)
				} else {
					This.pService.(*network.ServerSocket).StopClient(socketid)
				}
				return true
			}
			req := &LouMiaoRpcMsg{ClientId: token.UserId, Buffer: buff}
			client.SendMsg("LouMiaoRpcMsg", req)
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

func init() {
	handler_Map = make(map[string]string)
}
