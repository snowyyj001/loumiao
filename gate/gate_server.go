// 网关服务
package gate

import (
	"fmt"
	"reflect"
	"strings"
	"sync"

	"github.com/snowyyj001/loumiao/msg"

	"github.com/snowyyj001/loumiao/message"

	"github.com/snowyyj001/loumiao/base"

	"github.com/snowyyj001/loumiao/config"
	"github.com/snowyyj001/loumiao/define"
	"github.com/snowyyj001/loumiao/etcd"
	"github.com/snowyyj001/loumiao/gorpc"
	"github.com/snowyyj001/loumiao/log"
	"github.com/snowyyj001/loumiao/network"
	"github.com/snowyyj001/loumiao/nodemgr"
	"github.com/snowyyj001/loumiao/util"
)

var (
	This          *GateServer
	handler_Map   map[string]string
	filterWarning map[string]bool
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
tokens：key是client的socketid，Token.UserId是client的userid， Token.TokenId是client的tokenid
tokens_u：key是client的userid,value是socketid
users_u: key是client的userid,value是world的uid
================================================================================
当loumiao是server时:
tokens：key是gate的socketid，Token.UserId是gate的uid， Token.TokenId是自己的uid(0代表该gate将要被关闭)
tokens_u：key是gate的uid,value是socketid
users_u: key是client的userid,value是gate的uid
*/

type GateServer struct {
	gorpc.GoRoutineLogic

	Id            int
	pService      network.ISocket
	clients       map[int]*network.ClientSocket
	pInnerService network.ISocket
	ServerType    int
	//
	tokens     map[int]*Token
	tokens_u   map[int]int
	users_u    map[int]int
	OnlineNum  int
	clientEtcd *etcd.ClientDis
	rpcMap     map[string][]int

	lock sync.Mutex

	InitFunc func() //需要额外处理的函数回调
}

func (self *GateServer) DoInit() bool {
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
		self.pService.BindPacketFunc(packetFunc)
		self.pService.SetConnectType(network.CLIENT_CONNECT)
	}

	if self.ServerType == network.CLIENT_CONNECT { //对外(login,gate)
		self.clients = make(map[int]*network.ClientSocket)
	} else {
		self.pInnerService = new(network.ServerSocket)
		self.pInnerService.(*network.ServerSocket).SetMaxClients(config.NET_MAX_RPC_CONNS)
		self.pInnerService.Init(config.NET_GATE_SADDR)
		self.pInnerService.BindPacketFunc(packetFunc)
		self.pInnerService.SetConnectType(network.SERVER_CONNECT)
	}

	self.tokens = make(map[int]*Token)
	self.tokens_u = make(map[int]int)
	self.users_u = make(map[int]int)
	self.rpcMap = make(map[string][]int) //base64(funcname) -> [uid,uid,...]

	handler_Map = make(map[string]string)
	filterWarning = make(map[string]bool)
	filterWarning["CONNECT"] = true
	filterWarning["DISCONNECT"] = true

	if self.InitFunc != nil {
		self.InitFunc()
	}

	return true
}

func (self *GateServer) DoRegsiter() {
	log.Info("GateServer DoRegsiter")

	self.Register("RegisterNet", registerNet)
	self.Register("UnRegisterNet", unRegisterNet)
	self.Register("SendClient", sendClient)
	self.Register("SendMulClient", sendMulClient)
	self.Register("NewRpc", newRpc)
	self.Register("SendRpc", sendRpc)
	self.Register("SendGate", sendGate)
	self.Register("RecvPackMsg", recvPackMsg)
	self.Register("ReportOnLineNum", reportOnLineNum)
	self.Register("CloseServer", closeServer)

	handler_Map["CONNECT"] = "GateServer"
	self.RegisterGate("CONNECT", innerConnect)

	handler_Map["DISCONNECT"] = "GateServer"
	self.RegisterGate("DISCONNECT", innerDisConnect)

	handler_Map["C_CONNECT"] = "GateServer"
	self.RegisterGate("C_CONNECT", outerConnect)

	handler_Map["C_DISCONNECT"] = "GateServer"
	self.RegisterGate("C_DISCONNECT", outerDisConnect)

	handler_Map["LouMiaoLoginGate"] = "GateServer"
	self.RegisterGate("LouMiaoLoginGate", innerLouMiaoLoginGate)

	handler_Map["LouMiaoRpcRegister"] = "GateServer"
	self.RegisterGate("LouMiaoRpcRegister", innerLouMiaoRpcRegister)

	handler_Map["LouMiaoRpcMsg"] = "GateServer"
	self.RegisterGate("LouMiaoRpcMsg", innerLouMiaoRpcMsg)

	handler_Map["LouMiaoNetMsg"] = "GateServer"
	self.RegisterGate("LouMiaoNetMsg", innerLouMiaoNetMsg)

	handler_Map["LouMiaoKickOut"] = "GateServer"
	self.RegisterGate("LouMiaoKickOut", innerLouMiaoKickOut)

	handler_Map["LouMiaoClientConnect"] = "GateServer"
	self.RegisterGate("LouMiaoClientConnect", innerLouMiaoClientConnect)

	handler_Map["LouMiaoBindGate"] = "GateServer"
	self.RegisterGate("LouMiaoBindGate", innerLouMiaoLouMiaoBindGate)
}

func (self *GateServer) DoStart() {
	log.Info("GateServer DoStart")

	if self.pService != nil {
		if self.pService.Start() == false {
			util.Assert(nil)
			return
		}
	}
	if self.pInnerService != nil {
		if self.pInnerService.Start() == false {
			util.Assert(nil)
			return
		}
	}

	//etcd client创建
	client, err := etcd.NewClientDis(config.Cfg.EtcdAddr)
	if err != nil {
		base.Assert(err == nil, "NewClientDis error")
	}
	self.Id = nodemgr.GetServerUid(client, config.NET_GATE_SADDR)
	self.clientEtcd = client
	log.Infof("GateServer DoStart success: name=%s,saddr=%s,uid=%d", self.Name, config.NET_GATE_SADDR, self.Id)
}

//begin communicate with other nodes
func (self *GateServer) DoOpen() {
	self.clientEtcd.SetLeasefunc(leaseCallBack)                   //租约回调
	self.clientEtcd.SetLease(int64(config.GAME_LEASE_TIME), true) //创建租约
	//register my addr

	err := self.clientEtcd.PutService(fmt.Sprintf("%s%s/%d/%d", define.ETCD_SADDR, config.SERVER_GROUP, self.Id, config.NET_NODE_TYPE), config.NET_GATE_SADDR)
	if err != nil {
		log.Fatalf("etcd PutService error %v", err)
	}

	//server discover
	if self.ServerType == network.CLIENT_CONNECT { //account/gate watch server
		//watch all node, to manager server node, gate balance
		_, err := self.clientEtcd.WatchNodeList(define.ETCD_NODEINFO, nodemgr.NewNodeDiscover)
		if err != nil {
			log.Fatalf("etcd watch ETCD_LOCKUID error ", err)
		}
		//watch all node, just for account, to gate balance
		err = self.clientEtcd.WatchStatusList(define.ETCD_NODESTATUS, nodemgr.NodeStatusUpdate)
		if err != nil {
			log.Fatalf("etcd watch ETCD_NODESTATUS error ", err)
		}
		//watch server addr, this is for server discover
		//if config.NET_NODE_TYPE == config.ServerType_Gate {
		//watch addr
		_, err = self.clientEtcd.WatchServerList(define.ETCD_SADDR, self.newServerDiscover)
		if err != nil {
			log.Fatalf("etcd watch NET_GATE_SADDR error ", err)
		}
		//}
	} else {
		if config.NET_NODE_TYPE == config.ServerType_World {
			_, err = self.clientEtcd.WatchServerList(define.ETCD_SADDR, self.newServerDiscover)
			if err != nil {
				log.Fatalf("etcd watch NET_GATE_SADDR error ", err)
			}
		}
	}
}

//simple register self net hanlder, this func can only be called before igo started
func (self *GateServer) RegisterSelfNet(hanlderName string, hanlderFunc gorpc.HanlderNetFunc) {
	if self.IsRunning() {
		log.Fatal("RegisterSelfNet error, igo has already started")
		return
	}
	handler_Map[hanlderName] = "GateServer"
	self.RegisterGate(hanlderName, hanlderFunc)
}

//goroutine unsafe
func (self *GateServer) newServerDiscover(key string, val string, dis bool) {
	var uid int
	var group string
	arrStr := strings.Split(key, "/")
	group = arrStr[2]
	uid = util.Atoi(arrStr[3])
	atype := util.Atoi(arrStr[4])
	log.Debugf("GateServer newServerDiscover: key=%s,val=%s,dis=%t", key, val, dis)

	node := nodemgr.GetNode(uid)
	if node == nil && len(val) > 0 {
		node = nodemgr.CreateNode(val, uid, atype, group)
	}
	if node == nil {
		return
	}
	node.SocketActive = dis

	if config.NET_NODE_TYPE == config.ServerType_World {
		return
	}

	if atype == config.ServerType_Gate || atype == config.ServerType_Account { //filter gate and account
		return
	}

	if config.NET_NODE_TYPE == config.ServerType_Gate {
		if group != config.SERVER_GROUP { //only the same group can be discovered
			return
		}
	}

	if dis == true {
		if node.Number > 0 {
			log.Infof("newServerDiscover new server, but has already connected: saddr=%s,uid=%d", val, uid)
			return
		}
		if config.NET_NODE_TYPE == config.ServerType_Account {
			if atype != config.ServerType_World { //Account only connect to world
				return
			}
		}
		client := self.buildRpc(uid, val)
		if self.enableRpcClient(client) {
			m := gorpc.M{Id: uid, Data: client}
			gorpc.MGR.Send("GateServer", "NewRpc", m)
		}
	} else {
		//self.removeRpc(uid)
	}
}

func (self *GateServer) enableRpcClient(client *network.ClientSocket) bool {
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
	if self.clientEtcd != nil {
		self.clientEtcd.RevokeLease()
	}
}

//goroutine unsafe
//net msg handler,this func belong to socket's goroutine
func packetFunc(socketid int, buff []byte, nlen int) bool {
	//log.Debugf("packetFunc: socketid=%d, bufferlen=%d", socketid, nlen)
	m := gorpc.MI{Id: socketid, Name: nlen, Data: buff}
	gorpc.MGR.Send("GateServer", "RecvPackMsg", m)

	return true
}

func (self *GateServer) buildRpc(uid int, addr string) *network.ClientSocket {
	client := new(network.ClientSocket)
	client.SetClientId(uid)
	client.Init(addr)
	client.SetConnectType(network.CHILD_CONNECT)
	client.BindPacketFunc(packetFunc)
	client.Uid = uid

	return client
}

func (self *GateServer) removeRpc(uid int) {
	log.Debugf("GateServer removeRpc: %d", uid)
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

	//reset node ststus
	nodemgr.DisableNode(uid)
}

func (self *GateServer) GetRpcClient(uid int) *network.ClientSocket {
	client, ok := self.clients[uid]
	if ok {
		return client
	}
	return nil
}

//rpc调用的gate选择
func (self *GateServer) getCluserGateClientId() int {
	if config.NET_NODE_TYPE == config.ServerType_Gate {
		log.Fatalf("getCluserGate: gate can not call this func")
		return 0
	} else {
		var uids []int = make([]int, len(self.tokens_u))
		var sz int
		for _, v := range self.tokens_u { //过滤出可用的gate
			token, _ := self.tokens[v]
			if token != nil && token.TokenId > 0 {
				uids[sz] = v
				sz++
			}
		}
		if sz == 0 {
			return 0
		}
		index := util.Random(sz) //随机一个gate进行rpc转发
		return uids[index]
	}
	return 0
}

//rpc调用的目标server选择
func (self *GateServer) getCluserServer(funcName string) *network.ClientSocket {
	arr := self.rpcMap[funcName]
	sz := len(arr)
	if sz == 0 {
		log.Warningf("0.getCluserServerUid no rpc server hanlder finded %s", funcName)
		return nil
	}
	var uids []*network.ClientSocket = make([]*network.ClientSocket, len(self.clients))
	var n int
	for _, uid := range arr {
		client, _ := self.clients[uid]
		if client != nil && client.GetState() == network.SSF_CONNECT {
			uids[n] = client
			n++
		}
	}
	if n == 0 {
		return nil
	}
	index := util.Random(n) //随机一个gate进行rpc转发
	return uids[index]
}

func (self *GateServer) StopClient(userId int) {
	sid := This.tokens_u[userId]
	if sid > 0 {
		self.closeClient(sid)
	}
}

func (self *GateServer) closeClient(clientid int) {
	if self.ServerType == network.CLIENT_CONNECT {
		if config.NET_WEBSOCKET {
			self.pService.(*network.WebSocket).StopClient(clientid)
		} else {
			self.pService.(*network.ServerSocket).StopClient(clientid)
		}
	} else {
		self.pInnerService.(*network.ServerSocket).StopClient(clientid)
	}
}

// 必须保证线程安全，即需要在gateserver的igo中调用该函数
func (self *GateServer) SendServer(target int, buff []byte) {
	rpcClient := self.GetRpcClient(target)
	if rpcClient != nil {
		rpcClient.Send(buff)
	} else {
		log.Warningf("GateServer.SendServer target error: target=%d", target)
	}
}

// 必须保证线程安全，即需要在gateserver的igo中调用该函数
func (self *GateServer) SendRpc(funcName string, data interface{}, target int) {
	var buff []byte
	param := int32(1)
	if reflect.TypeOf(data).Kind() == reflect.Slice { //bitstream
		buff = data.([]byte)
	} else {
		buff, _ = message.Encode(target, 0, "", data)
		param = 0
	}

	outdata := &msg.LouMiaoRpcMsg{TargetId: int64(target), FuncName: funcName, Buffer: buff, SourceId: int64(This.Id), ByteBuffer: param}
	innerLouMiaoRpcMsg(self, target, outdata)
}
