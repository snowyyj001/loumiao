// 网关服务
package gate

import (
	"fmt"

	"github.com/snowyyj001/loumiao/config"
	"github.com/snowyyj001/loumiao/define"
	"github.com/snowyyj001/loumiao/etcd"
	"github.com/snowyyj001/loumiao/gorpc"
	"github.com/snowyyj001/loumiao/log"
	"github.com/snowyyj001/loumiao/message"
	"github.com/snowyyj001/loumiao/network"
	"github.com/snowyyj001/loumiao/util"
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

	Id            int
	pService      network.ISocket
	clients       map[int]*network.ClientSocket
	pInnerService network.ISocket
	serverMap     map[int]int
	ServerType    int
	tokens        map[int]*Token
	tokens_u      map[int]int
	serverReg     *etcd.ServiceReg
	clientReq     *etcd.ClientDis

	OnServerConnected func(int)
	OnClientConnected func(int)
}

func (self *GateServer) DoInit() {
	self.Id = int(util.UUID())
	log.Infof("%s DoInit： uid = %d, %d", self.Name, self.Id)
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

	if self.ServerType == network.CLIENT_CONNECT { //对外
		if config.NET_WEBSOCKET {
			self.pService = new(network.WebSocket)
			self.pService.(*network.WebSocket).SetMaxClients(config.NET_MAX_CONNS)
		} else {
			self.pService = new(network.ServerSocket)
			self.pService.(*network.ServerSocket).SetMaxClients(config.NET_MAX_CONNS)
		}
		self.pService.Init(config.NET_GATE_SADDR)
		self.pService.BindPacketFunc(PacketFunc)
		self.pService.SetConnectType(network.CLIENT_CONNECT)
	}

	if config.NET_NODE_TYPE == config.ServerType_Gate {
		self.clients = make(map[int]*network.ClientSocket)
	} else {
		self.pInnerService = new(network.ServerSocket)
		self.pInnerService.(*network.ServerSocket).SetMaxClients(config.NET_MAX_RPC_CONNS)
		self.pInnerService.Init(config.NET_RPC_ADDR)
		self.pInnerService.BindPacketFunc(PacketFunc)
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
		client, err := etcd.NewClientDis(config.Cfg.EtcdAddr)
		if err == nil {
			self.clientReq = client
			//watch addr
			_, err = self.clientReq.Watch(define.ETCD_SADDR, self.NewClientDiscover)
			if err != nil {
				log.Errorf("etcd watch error ", err)
			}
		} else {
			log.Errorf("GateServer NewClientDis ", err)
		}

	} else {
		_, err := etcd.NewServiceReg(config.Cfg.EtcdAddr, 3)
		if err == nil {
			//register my addr
			err = self.serverReg.PutService(fmt.Sprintf("%s%d", define.ETCD_SADDR, self.Id), config.NET_GATE_SADDR)
			if err != nil {
				log.Errorf("etcd PutService error ", err)
			}
		} else {
			log.Errorf("GateServer NewClientDis", err)
		}
	}
}

func (self *GateServer) NewClientDiscover(key string, val string, dis bool) {
	var uid int
	fmt.Sscanf(key, define.ETCD_SADDR+"%d", &uid)

	if dis == true {
		client := self.BuildRpc(uid, val)
		if self.EnableRpcClient(client) {
			self.clients[uid] = client
			if self.OnServerConnected != nil {
				self.OnServerConnected(uid)
			}
		}
	} else {
		self.RemoveRpc(uid)
	}
}

func (self *GateServer) EnableRpcClient(client *network.ClientSocket) bool {
	log.Info("GateServer EnableRpcClient")
	if client.Start() {
		log.Infof("GateServer rpc connected %s", client.GetSAddr())
		return true
	} else {
		return false
	}
}

func (self *GateServer) DoDestory() {
	if self.serverReg != nil {
		self.serverReg.RevokeLease()
	}
}

func PacketFunc(socketid int, buff []byte, nlen int) bool {
	defer func() {
		if err := recover(); err != nil {
			log.Errorf("PacketFunc self: %v", err)
		}
	}()

	err, target, name, pm := message.Decode(This.Id, buff, nlen)
	if err != nil {
		return false
	}
	if target == This.Id { //send to me
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
	} else { //send to other server
		rpcClient := This.GetRpcClient(target)
		if rpcClient == nil {
			return false
		}
		rpcClient.Send(buff[0:nlen])
	}
	return true
}

func (self *GateServer) BuildRpc(uid int, addr string) *network.ClientSocket {
	client := new(network.ClientSocket)
	client.SetClientId(0)
	client.Init(addr)
	client.SetConnectType(network.CHILD_CONNECT)
	client.BindPacketFunc(PacketFunc)
	client.Uid = uid

	return client
}

func (self *GateServer) RemoveRpc(uid int) {
	client, ok := self.clients[uid]
	if ok {
		client.Stop()
	}
	delete(self.clients, uid)
}

func (self *GateServer) GetRpcClient(uid int) *network.ClientSocket {
	return self.clients[uid]
}

func init() {
	handler_Map = make(map[string]string)
}
