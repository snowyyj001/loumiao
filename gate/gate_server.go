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
)

var (
	This        *GateServer
	handler_Map map[string]string //消息报接受者
)

type Token struct {
	UserId  int
	TokenId int
}

/*tokens, tokens_u, users_u说明
当loumiao是gate时，tokens的key是客户端的socketid，Token.UserId是客户端的userid，
Token.TokenId是account生成的秘钥，用来校验client的合法性；
tokens_u的key是客户端的userid,value是socketid
users_u无效
当loumiao是server时，tokens的key是gate的socketid，Token.UserId是gate的uid，
Token.TokenId是server的uid
tokens_u的key是gate的uid,value是socketid
users_u的key是client的userid，value是gate的socketid
*/

type GateServer struct {
	gorpc.GoRoutineLogic

	Id            int
	pService      network.ISocket
	clients       map[int]*network.ClientSocket
	pInnerService network.ISocket //pService  pInnerService just for account server
	ServerType    int
	tokens        map[int]*Token
	tokens_u      map[int]int
	users_u       map[int]int
	serverReg     *etcd.ServiceReg
	clientReq     *etcd.ClientDis
	rpcMap        map[string][]int

	OnClientConnected    func(int)
	OnClientDisConnected func(int)
}

func (self *GateServer) DoInit() {
	log.Infof("%s DoInit", self.Name)
	This = self

	if config.NET_NODE_TYPE == config.ServerType_World {
		if config.NET_NODE_ID != 1 {
			log.Fatalf("GateServer world node id error [%d]", config.NET_NODE_ID)
			return
		}
	}

	if config.NET_NODE_TYPE < config.ServerType_World || config.NET_NODE_TYPE > config.ServerType_IM {
		log.Fatalf("GateServer node type error [%d]", config.NET_NODE_TYPE)
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

	self.tokens = make(map[int]*Token)
	self.tokens_u = make(map[int]int)
	self.users_u = make(map[int]int)
	self.rpcMap = make(map[string][]int) //md5(funcname) -> [uid,uid,...]

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

	handler_Map["LouMiaoLoginGate"] = "GateServer"
	self.RegisterGate("LouMiaoLoginGate", InnerLouMiaoLoginGate)

	handler_Map["LouMiaoRpcMsg"] = "GateServer"
	self.RegisterGate("LouMiaoRpcMsg", InnerLouMiaoRpcMsg)

	handler_Map["LouMiaoNetMsg"] = "GateServer"
	self.RegisterGate("LouMiaoNetMsg", InnerLouMiaoNetMsg)
}

func (self *GateServer) DoStart() {
	log.Info("GateServer DoStart")

	if self.pService != nil {
		self.pService.Start()
	}
	if self.pInnerService != nil {
		self.pInnerService.Start()
	}

	//server discover
	if config.NET_NODE_TYPE == config.ServerType_Gate {
		client, err := etcd.NewClientDis(config.Cfg.EtcdAddr)
		if err == nil {
			self.clientReq = client
			self.Id = etcd.GetServerUid(client.GetClient(), config.NET_GATE_SADDR)
			//watch addr
			_, err = self.clientReq.Watch(define.ETCD_SADDR, self.NewServerDiscover)
			if err != nil {
				log.Fatalf("etcd watch error ", err)
			}
		} else {
			log.Fatalf("GateServer NewClientDis ", err)
		}

	} else {
		client, err := etcd.NewServiceReg(config.Cfg.EtcdAddr, 3)
		if err == nil {
			self.serverReg = client
			self.Id = etcd.GetServerUid(client.GetClient(), config.NET_GATE_SADDR)

			//register my addr
			err = self.serverReg.PutService(fmt.Sprintf("%s%d", define.ETCD_SADDR, self.Id), config.NET_GATE_SADDR)
			if err != nil {
				log.Fatalf("etcd PutService error ", err)
			}
		} else {
			log.Fatalf("GateServer NewClientDis", err)
		}
	}

	log.Infof("GateServer DoStart success: %s,%s,%d", self.Name, config.NET_GATE_SADDR, self.Id)
}

func (self *GateServer) NewServerDiscover(key string, val string, dis bool) {
	var uid int
	fmt.Sscanf(key, define.ETCD_SADDR+"%d", &uid)

	if dis == true {
		client := self.BuildRpc(uid, val)
		if self.EnableRpcClient(client) {
			self.clients[uid] = client
			self.OnServerConnected(uid)
		}
	} else {
		self.RemoveRpc(uid)
	}
}

func (self *GateServer) NewRpcRegister(key string, val string, dis bool) {
	var uid int
	var funcName string
	var prefix string

	fmt.Sscanf(key, "%s/%s/%d", &prefix, &funcName, &uid)

	log.Debugf("NewRpcRegister: %s,%d,%t", funcName, uid, dis)

	if dis == true {
		arr, ok := self.rpcMap[funcName]
		if ok {
			for _, val := range arr {
				if val == uid {
					log.Warningf("NewRpcRegister func has already been registered %s, %d", funcName, uid)
					return
				}
			}
			arr = append(arr, uid)
		} else {
			self.rpcMap[funcName] = make([]int, 100)
			self.rpcMap[funcName] = append(self.rpcMap[funcName], uid)
		}
	} else {
		arr, ok := self.rpcMap[funcName]
		if ok {
			for i, val := range arr {
				if val == uid {
					arr = append(arr[:i], arr[i+1:]...)
					return
				}
			}
		}
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

//net msg handler
func PacketFunc(socketid int, buff []byte, nlen int) bool {
	defer func() {
		if err := recover(); err != nil {
			log.Errorf("PacketFunc self: %v", err)
		}
	}()

	err, target, name, pm := message.Decode(This.Id, buff, nlen)
	if err != nil {
		log.Warningf("PacketFunc Decode error: %s", err.Error())
		return false
	}
	if target == This.Id { //send to me
		handler, ok := handler_Map[name]
		if ok {
			if handler == "GateServer" {
				This.CallNetFunc(pm)
			} else {
				m := gorpc.M{Id: socketid, Name: name, Data: pm}
				This.Send(handler, "ServiceHandler", m)
			}
		} else {
			if name != "CONNECT" && name != "DISCONNECT" {
				log.Noticef("MsgProcess self handler is nil, drop it[%s]", name)
			}
		}
	} else { //send to target server
		if config.NET_NODE_TYPE != config.ServerType_Gate {
			log.Errorf("PacketFunc target error, [%s]drop it, target=%d,myid=%d", name, target, This.Id)
			return false
		}
		token, ok := This.tokens[socketid]
		if ok {
			log.Debugf("0.PacketFunc recv client msg, but client has lost[%d] ", socketid)
			return false
		}
		msg := LouMiaoNetMsg{ClientId: token.UserId, Buffer: buff}
		buff, _ := message.Encode(target, 0, "LouMiaoNetMsg", msg)

		rpcClient := This.GetRpcClient(target)
		if rpcClient != nil {
			rpcClient.Send(buff[0:nlen])
		} else {
			log.Debugf("1.PacketFunc recv client msg, but server has lost[%d] ", target)
			return false
		}
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

func (self *GateServer) OnServerConnected(uid int) {
	req := &LouMiaoLoginGate{TokenId: uid, UserId: self.Id}
	buff, _ := message.Encode(uid, 0, "LouMiaoLoginGate", req)
	client, _ := self.clients[uid]
	client.Send(buff)
}

func (self *GateServer) GetCluserGate() int {
	if config.NET_NODE_TYPE == config.ServerType_Gate {
		log.Fatalf("GetCluserGate: gate can not call this func")
		return 0
	} else {
		sz := len(self.tokens)
		index := util.Random(sz) //随机一个gate进行rpc转发
		clientid := self.tokens[index]
		return clientid
	}
}

func init() {
	handler_Map = make(map[string]string)
}
