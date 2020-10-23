// 网关服务
package gate

import (
	"fmt"
	"sync"

	"github.com/snowyyj001/loumiao/msg"

	"github.com/snowyyj001/loumiao/config"
	"github.com/snowyyj001/loumiao/define"
	"github.com/snowyyj001/loumiao/etcd"
	"github.com/snowyyj001/loumiao/gorpc"
	"github.com/snowyyj001/loumiao/log"
	"github.com/snowyyj001/loumiao/message"
	"github.com/snowyyj001/loumiao/network"
	"github.com/snowyyj001/loumiao/nodemgr"
	"github.com/snowyyj001/loumiao/util"
)

var (
	This        *GateServer
	handler_Map map[string]string //消息报接受者
)

type Token struct {
	UserId  int //这里应该是int64，所以，程序只能在64位机上运行，否则会有异常
	TokenId int
}

/*tokens, tokens_u, users_u说明
当loumiao是account时:
tokens: 没有用
tokens_u: 没有用
================================================================================
当loumiao是gate时:
tokens：key是client的socketid，Token.UserId是client的userid， Token.TokenId是world的uid
tokens_u：key是client的userid,value是socketid
users_u: 指向tokens_u
================================================================================
当loumiao是server时:
tokens：key是gate的socketid，Token.UserId是gate的uid， Token.TokenId是自己的uid
tokens_u：key是gate的uid,value是socketid
users_u: key是client的userid,value是gate的socketid
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
	lock          sync.Mutex

	OnClientConnected    func(int)
	OnClientDisConnected func(int)
}

func (self *GateServer) DoInit() {
	log.Infof("%s DoInit", self.Name)
	This = self

	if self.ServerType == network.CLIENT_CONNECT { //对外(login,gate)
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

	if self.ServerType == network.CLIENT_CONNECT { //对外(login,gate)
		self.clients = make(map[int]*network.ClientSocket)
	} else {
		self.pInnerService = new(network.ServerSocket)
		self.pInnerService.(*network.ServerSocket).SetMaxClients(config.NET_MAX_RPC_CONNS)
		self.pInnerService.Init(config.NET_GATE_SADDR)
		self.pInnerService.BindPacketFunc(PacketFunc)
		self.pInnerService.SetConnectType(network.SERVER_CONNECT)

		self.OnClientConnected = onClientConnected
	}

	self.tokens = make(map[int]*Token)
	self.tokens_u = make(map[int]int)
	if self.ServerType == network.CLIENT_CONNECT { //对外(login,gate)
		self.users_u = self.tokens_u
	} else {
		self.users_u = make(map[int]int)
	}
	self.rpcMap = make(map[string][]int) //base64(funcname) -> [uid,uid,...]
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

	handler_Map["LouMiaoRpcRegister"] = "GateServer"
	self.RegisterGate("LouMiaoRpcRegister", InnerLouMiaoRpcRegister)

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

	//etcd client创建
	if self.ServerType == network.CLIENT_CONNECT { //gate watch server
		client, err := etcd.NewClientDis(config.Cfg.EtcdAddr)
		if err == nil {
			self.clientReq = client
			self.Id = nodemgr.GetServerUid(client, config.NET_GATE_SADDR)
		} else {
			log.Fatalf("GateServer NewClientDis ", err)
		}

	} else { // server put address
		client, err := etcd.NewServiceReg(config.Cfg.EtcdAddr, int64(config.GAME_LEASE_TIME))
		if err == nil {
			self.serverReg = client
			self.Id = nodemgr.GetServerUid(client, config.NET_GATE_SADDR)
		} else {
			log.Fatalf("GateServer NewClientDis", err)
		}
	}

	log.Infof("GateServer DoStart success: name=%s,saddr=%s,uid=%d", self.Name, config.NET_GATE_SADDR, self.Id)
}

//begin communicate with other nodes
func (self *GateServer) DoOpen() {
	//server discover
	if self.ServerType == network.CLIENT_CONNECT { //account/gate watch server

		self.clientReq.SetLeasefunc(leaseCallBack)
		self.clientReq.SetLease(int64(config.GAME_LEASE_TIME), true)
		if config.NET_NODE_TYPE == config.ServerType_Account {
			//watch all node, just for account, to manager server node, gate balance
			_, err := self.clientReq.WatchNodeList(define.ETCD_LOCKUID, nodemgr.NewNodeDiscover)
			if err != nil {
				log.Fatalf("etcd watch ETCD_LOCKUID error ", err)
			}

			//watch all node, just for account, to gate balance
			err = self.clientReq.WatchStatusList(define.ETCD_NODESTATUS, nodemgr.NodeStatusUpdate)
			if err != nil {
				log.Fatalf("etcd watch ETCD_NODESTATUS error ", err)
			}
		} else {
			//watch addr
			_, err := self.clientReq.WatchServerList(define.ETCD_SADDR, self.NewServerDiscover)
			if err != nil {
				log.Fatalf("etcd watch NET_GATE_SADDR error ", err)
			}
		}
	} else { // server put address
		self.serverReg.SetLeasefunc(leaseCallBack)
		//register my addr
		err := self.serverReg.PutService(fmt.Sprintf("%s%d", define.ETCD_SADDR, self.Id), config.NET_GATE_SADDR)
		if err != nil {
			log.Fatalf("etcd PutService error ", err)
		}
	}
}

func (self *GateServer) NewServerDiscover(key string, val string, dis bool) {
	self.lock.Lock()
	defer self.lock.Unlock()

	var uid int
	fmt.Sscanf(key, define.ETCD_SADDR+"%d", &uid)
	log.Debugf("GateServer NewServerDiscover: key=%s,val=%s,dis=%t", key, val, dis)

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

func (self *GateServer) EnableRpcClient(client *network.ClientSocket) bool {
	if client.Start() {
		log.Infof("GateServer rpc connected %s success", client.GetSAddr())
		return true
	} else {
		log.Warningf("GateServer rpc connect failed %s", client.GetSAddr())
		return false
	}
	return true
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

	if target == This.Id || target == 0 { //send to me
		handler, ok := handler_Map[name]
		if ok {
			if handler == "GateServer" {
				This.NetHandler[name](This, socketid, pm)
			} else {
				m := gorpc.M{Id: socketid, Name: name, Data: pm}
				This.Send(handler, "ServiceHandler", m)
			}
		} else {
			if name != "CONNECT" && name != "DISCONNECT" {
				log.Noticef("MsgProcess self handler is nil, drop it[%s]", name)
			} else {
				log.Debugf("PacketFunc[%s]: %d", name, socketid)
			}
		}
	} else { //send to target server
		if config.NET_NODE_TYPE != config.ServerType_Gate {
			log.Errorf("PacketFunc target error, [%s]drop it, target=%d,myid=%d", name, target, This.Id)
			return false
		}
		token, ok := This.tokens[socketid]
		if ok == false {
			log.Debugf("0.PacketFunc recv client msg, but client has lost[%d] ", socketid)
			return false
		}
		msg := &msg.LouMiaoNetMsg{ClientId: int64(token.UserId), Buffer: buff}
		buff, newlen := message.Encode(target, 0, "LouMiaoNetMsg", msg)
		//log.Debugf("LouMiaoNetMsg send to server %d", target)
		rpcClient := This.GetRpcClient(target)
		if rpcClient != nil {
			rpcClient.Send(buff[0:newlen])
		} else {
			log.Debugf("1.PacketFunc recv client msg, but server has lost[%d] ", target)
			return false
		}
	}
	return true
}

func (self *GateServer) BuildRpc(uid int, addr string) *network.ClientSocket {
	client := new(network.ClientSocket)
	client.SetClientId(uid)
	client.Init(addr)
	client.SetConnectType(network.CHILD_CONNECT)
	client.BindPacketFunc(PacketFunc)
	client.Uid = uid

	return client
}

func (self *GateServer) RemoveRpc(uid int) {
	log.Debugf("GateServer RemoveRpc: %d", uid)
	client, ok := self.clients[uid]
	if ok {
		client.Stop()
	}
	delete(self.clients, uid)

	//remove rpc handler
	for key, arr := range self.rpcMap {
		for i, val := range arr {
			if val == uid {
				self.rpcMap[key] = append(arr[:i], arr[i+1:]...)
				break
			}
		}
	}
}

func (self *GateServer) GetRpcClient(uid int) *network.ClientSocket {
	return self.clients[uid]
}

func (self *GateServer) OnServerConnected(uid int) {
	log.Debugf("GateServer OnServerConnected: %d", uid)
	req := &msg.LouMiaoLoginGate{TokenId: int64(uid), UserId: int64(self.Id)}
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
		num := 0
		for k, _ := range self.tokens {
			if num == index {
				return k
			}
			num++
		}
		return 0
	}
}

func init() {
	handler_Map = make(map[string]string)
}
