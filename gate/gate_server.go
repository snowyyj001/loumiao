// 网关服务
package gate

import (
	"fmt"
	"github.com/snowyyj001/loumiao/agent"
	"github.com/snowyyj001/loumiao/base/maps"
	"github.com/snowyyj001/loumiao/lnats"
	"github.com/snowyyj001/loumiao/message"
	"sync"

	"github.com/snowyyj001/loumiao/config"
	"github.com/snowyyj001/loumiao/define"
	"github.com/snowyyj001/loumiao/etcd"
	"github.com/snowyyj001/loumiao/gorpc"
	"github.com/snowyyj001/loumiao/llog"
	"github.com/snowyyj001/loumiao/network"
	"github.com/snowyyj001/loumiao/nodemgr"
	"github.com/snowyyj001/loumiao/util"
)

var (
	This       *GateServer
	handlerMap map[string]string //goroutine unsafe,func name -> actor name
)

type GateServer struct {
	gorpc.GoRoutineLogic

	pService      network.ISocket
	clients       sync.Map //uid -> *network.ClientSocket
	pInnerService network.ISocket
	ServerType    int
	//
	tokensLocker sync.RWMutex

	//only for SERVER_CONNECT
	gatesUid        sync.Map //gate socket id -> gate uid
	gatesSocketId   sync.Map //gate uid -> gate socket id
	userGates       sync.Map //user id -> gate uid
	userGateSockets sync.Map //user id -> gate socket id

	//only for CLIENT_CONNECT
	uidAgents sync.Map //user id -> socket id
	sidAgents sync.Map //socket id -> *agent.LouMiaoAgent
	netMap    sync.Map //net func -> server type

	//rpc相关
	rpcMap  map[string]string //rpc function -> actor name
	rpcUids sync.Map          //rpc uid -> true

	//单个actor内的消息pub/sub
	msgQueueMap map[string]*maps.Map

	//排队服uid
	QueueServerUid int

	lock sync.Mutex
}

func (self *GateServer) DoInit() bool {
	llog.Infof("%s DoInit", self.Name)
	This = self

	if self.ServerType == network.CLIENT_CONNECT { //对外(login,gate)
		if config.NET_WEBSOCKET {
			self.pService = new(network.WebSocket)
			self.pService.(*network.WebSocket).SetMaxClients(config.NET_MAX_CONNS)
		} else {
			self.pService = new(network.ServerSocket)
			self.pService.(*network.ServerSocket).SetMaxClients(config.NET_MAX_CONNS)
		}
		self.pService.Init(config.NET_LISTEN_SADDR)
		self.pService.BindPacketFunc(packetFunc_client)
		self.pService.SetConnectType(network.CLIENT_CONNECT)
	} else {
		self.pInnerService = new(network.ServerSocket)
		self.pInnerService.(*network.ServerSocket).SetMaxClients(config.NET_MAX_RPC_CONNS)
		self.pInnerService.Init(config.NET_LISTEN_SADDR)
		self.pInnerService.BindPacketFunc(packetFunc_server)
		self.pInnerService.SetConnectType(network.SERVER_CONNECT)
	}

	//self.clients = make(map[int]*network.ClientSocket)

	self.rpcMap = make(map[string]string)
	self.msgQueueMap = make(map[string]*maps.Map) //key -> [actorname,actorname,...]

	handlerMap = make(map[string]string)

	return true
}

func (self *GateServer) DoRegister() {
	llog.Infof("%s DoRegister", self.Name)
	
	self.Register("Publish", publish)
	self.Register("Subscribe", subscribe)

	//in gate, client connect with gate, in server, gate(as client) connect with server
	self.RegisterSelfNet("CONNECT", innerConnect)
	self.RegisterSelfNet("DISCONNECT", innerDisConnect)

	//in gate, gate connect with server, no in server
	self.RegisterSelfNet("C_CONNECT", outerConnect)
	self.RegisterSelfNet("C_DISCONNECT", outerDisConnect)

	self.RegisterSelfNet("LouMiaoLoginGate", innerLouMiaoLoginGate)

	self.RegisterSelfNet("LouMiaoRpcMsg", innerLouMiaoRpcMsg)

	self.RegisterSelfNet("LouMiaoNetMsg", innerLouMiaoNetMsg)

	self.RegisterSelfNet("LouMiaoKickOut", innerLouMiaoKickOut)

	self.RegisterSelfNet("LouMiaoClientConnect", innerLouMiaoClientConnect)

	self.RegisterSelfNet("LouMiaoNetRegister", innerLouMiaoNetRegister)
}

// begin communicate with other nodes
func (self *GateServer) DoStart() {
	llog.Info("GateServer.DoStart")

	//etcd client
	err := etcd.NewClient()
	if util.CheckErr(err) {
		llog.Fatalf("etcd connect failed: %v", config.Cfg.EtcdAddr)
	}

	//nodemgr.ServerEnabled = true
	//etcd.Client.PutStatus() //服务如果异常关闭，是没有撤销租约的，在三秒内重启会保留上次状态(可能是关闭状态)，这里强制刷新一下，
	//nodemgr.ServerEnabled = false

	lnats.SubscribeAsync(define.TOPIC_SERVER_LOG, llog.Tp_SetLevel)

	if config.NET_NODE_TYPE != config.ServerType_WEB_LOGIN {
		if self.pService != nil {
			util.Assert(self.pService.Start(), fmt.Sprintf("GateServer listen failed: saddr=%s", self.pService.GetSAddr()))
		}
		if self.pInnerService != nil { //内部组网监听，地址自动切换也没关系
			util.Assert(self.pInnerService.Start(), fmt.Sprintf("GateServer listen failed: saddr=%s", self.pInnerService.GetSAddr()))
		}
	}

	llog.Infof("GateServer DoStart success: name=%s,saddr=%s,uid=%d", self.Name, config.NET_GATE_SADDR, config.SERVER_NODE_UID)
}

// DoOpen 监听并注册etcd
func (self *GateServer) DoOpen() {
	llog.Info("GateServer.DoOpen")

	need := config.NET_NODE_TYPE == config.ServerType_Account           //挑选网关和world给client需要
	need = config.NET_NODE_TYPE == config.ServerType_Gate               //挑选world给client需要
	need = need || config.NET_NODE_TYPE == config.ServerType_World      //挑选zone给client需要，其他主逻辑也可能需要
	need = need || config.NET_NODE_TYPE == config.ServerType_LOGINQUEUE //统计所有world在线人数需要
	need = need || config.NET_NODE_TYPE == config.ServerType_WEB_LOGIN  //挑选网关给client需要
	if need {                                                           //只让必要的节点参与状态监视，减少消息量
		//watch status
		var key string
		if config.NET_NODE_ID == 0 { //跨服
			key = define.ETCD_NODESTATUS
		} else { //只关心本服的节点信息
			key = fmt.Sprintf("%s%d", define.ETCD_NODESTATUS, config.NET_NODE_ID)
		}
		err := etcd.Client.WatchCommon(key, self.serverStatusUpdate)
		if err != nil {
			llog.Fatalf("etcd watch ETCD_NODESTATUS error: %s", err.Error())
		}
	}

	var key string
	if config.NET_NODE_ID == 0 { //跨服
		key = define.ETCD_NODEINFO
	} else { //只关心本服的节点信息
		key = fmt.Sprintf("%s%d", define.ETCD_NODEINFO, config.NET_NODE_ID)
	}
	//watch all node，所有集群内的节点都需要参与服发现
	err := etcd.Client.WatchCommon(key, self.newServerDiscover)
	if err != nil {
		llog.Fatalf("etcd watch NET_GATE_SADDR error: %s", err.Error())
	}

	//timer.DelayJob(100, func() { //delay 100ms, that all RegisterRpcHandler should be compled
	//应该首先watch，注册rpc消息，再把自己put进去
	//register to etcd when the socket is ok
	if err := etcd.Client.PutNode(); err != nil {
		llog.Fatalf("etcd PutService error %v", err)
	}
	nodemgr.ServerEnabled = true
	llog.Infof("GateServer DoOpen success: name=%s, saddr=%s, uid=%d", self.Name, config.NET_GATE_SADDR, config.SERVER_NODE_UID)
	//	}, false)

	lnats.ReportMail(define.MAIL_TYPE_START, "服务器完成启动")
}

// RegisterSelfNet
// simple register self net handler, this func can only be called before igo started
func (self *GateServer) RegisterSelfNet(handlerName string, HandlerFunc gorpc.HandlerNetFunc) {
	if self.IsRunning() {
		llog.Fatal("RegisterSelfNet error, igo has already started")
		return
	}
	handlerMap[handlerName] = "GateServer"
	self.RegisterGate(handlerName, HandlerFunc)
}

// RegisterSelfRpc
// simple register self rpc handler, this func can only be called before igo started
func (self *GateServer) RegisterSelfRpc(handlerFunc gorpc.HandlerNetFunc) {
	if self.IsRunning() {
		llog.Fatal("RegisterSelfNet error, igo has already started")
		return
	}
	funcName := util.RpcFuncName(handlerFunc)
	handlerMap[funcName] = "GateServer"
	This.rpcMap[funcName] = "GateServer"
	self.RegisterGate(funcName, handlerFunc)
}

// RegisterSelfCallRpc
// simple register self rpc handler, this func can only be called before igo started
func (self *GateServer) RegisterSelfCallRpc(handlerFunc gorpc.HandlerFunc) {
	if self.IsRunning() {
		llog.Fatal("RegisterSelfNet error, igo has already started")
		return
	}
	funcName := util.RpcFuncName(handlerFunc)
	//fmt.Println("funcName", funcName)
	self.Register(funcName, handlerFunc)
	handlerMap[funcName] = "GateServer"
	This.rpcMap[funcName] = "GateServer"
}

func (self *GateServer) GetServerService() network.ISocket {
	if self.ServerType == network.CLIENT_CONNECT { //对外(login,gate)
		return self.pService
	} else {
		return self.pInnerService
	}
}

// goroutine unsafe
func (self *GateServer) serverStatusUpdate(key, val string, dis bool) {
	node := nodemgr.NodeStatusUpdate(key, val, dis)
	if node == nil || !dis {
		return
	}

	if node.Uid > 0 && node.SocketActive == false && node.Type == config.ServerType_RPCGate {
		This.rpcUids.Delete(node.Uid)
	}
}

// goroutine unsafe
func (self *GateServer) newServerDiscover(key, val string, dis bool) {
	node := nodemgr.NodeDiscover(key, val, dis)
	if node == nil || !dis {
		return
	}

	if node.Type == config.ServerType_WEB_LOGIN ||
		node.Type == config.ServerType_WEB_GM ||
		node.Type == config.ServerType_WebKeyPoint { //filter 他们不参与rpc调用，不参与集群组网，其实参与了也没啥影响
		return
	}

	if !node.SocketActive { //已经关闭的节点就不去连接了
		llog.Infof("newServerDiscover: discover a closed node: uid = %d", node.Uid)
		return
	}

	if node.Type == config.ServerType_LOGINQUEUE {
		self.QueueServerUid = node.Uid //这里记录一下，方便转发client的排队网络消息
	}

	if node.Type == config.ServerType_RPCGate { //发现了一个rpc server，咱作为客户端去连上它，参与rpc的狂欢
		rpcClient := self.GetRpcClient(node.Uid)
		if rpcClient == nil { //this conditation can be etcd reconnect
			client := self.buildClient(node.Uid, node.SAddr)
			if self.enableClient(client) {
				newRpcClient(node.Uid, client)
			} else {
				node.SocketActive = false
			}
		}
	} else if config.NET_NODE_TYPE == config.ServerType_Gate { //网关和其他server建立一条专线，用来直接转发client的消息
		gateConn := node.Type == config.ServerType_World                 //世界服
		gateConn = gateConn || node.Type == config.ServerType_Zone       //战斗服
		gateConn = gateConn || node.Type == config.ServerType_LOGINQUEUE //排队服
		if gateConn {                                                    //目前只有这三个需要直接接受来自client的消息(其实这里没必要过滤，多一条socket连接没有任何影响)
			rpcClient := self.GetRpcClient(node.Uid)
			if rpcClient == nil { //this condition can be etcd reconnect
				client := self.buildClient(node.Uid, node.SAddr)
				if self.enableClient(client) {
					newGateClient(node.Uid, client)
				} else {
					node.SocketActive = false
				}
			}
		}
	}
}

func (self *GateServer) enableClient(client *network.ClientSocket) bool {
	if client.Start() {
		llog.Infof("GateServer rpc connected %s success", client.GetSAddr())
		return true
	} else {
		llog.Warningf("GateServer rpc connect failed %s", client.GetSAddr())
		return false
	}
	return true
}

func (self *GateServer) DoDestroy() {
	llog.Info("GateServer DoDestroy")
	nodemgr.ServerEnabled = false
	etcd.Client.RevokeLease()
}

// goroutine unsafe
// net msg handler,this func belong to socket's goroutine
func packetFunc_client(socketid int, buff []byte, nlen int) error {
	//llog.Debugf("packetFunc: socketid=%d, bufferlen=%d", socketid, nlen)
	target, name, buffbody, err := message.UnPackHead(buff, nlen)
	//llog.Debugf("packetFunc_client %d %s %d", target, name, nlen)
	if nil != err {
		return fmt.Errorf("packetFunc_client Decode error: %s", err.Error())
		//This.closeClient(socketid)
	} else {
		ty, ok := This.netMap.Load(name)
		if !ok { //msg to me，client使用的是server type
			handler, ok := handlerMap[name]
			if ok {
				nm := &gorpc.M{Id: socketid, Name: name, Data: buffbody}
				gorpc.MGR.Send(handler, "ServiceHandler", nm)
			} else {
				llog.Errorf("packetFunc_client handler is nil, drop it[%s]", name)
			}
		} else { //msg to other server
			target = ty.(int)
			recvPackMsgClient(socketid, target, buff)
		}
	}
	return nil
}

// goroutine unsafe
// net msg handler,this func belong to socket's goroutine
func packetFunc_server(socketid int, buff []byte, nlen int) error {
	//llog.Debugf("packetFunc_server: socketid=%d, bufferlen=%d", socketid, nlen)
	target, name, buffbody, err := message.UnPackHead(buff, nlen)
	//llog.Debugf("packetFunc_server %d %s %v", target, name, err)
	if nil != err {
		return fmt.Errorf("packetFunc_server Decode error: %s", err.Error())
		//This.closeClient(socketid)
	} else {
		if target == config.SERVER_NODE_UID || target <= 0 { //server使用的是server uid
			handler, ok := handlerMap[name] //handler_Map will not changed, so use here is ok
			if ok {
				nm := &gorpc.M{Id: socketid, Name: name, Data: buffbody}
				gorpc.MGR.Send(handler, "ServiceHandler", nm)
			} else {
				llog.Errorf("packetFunc_server handler is nil, drop it[%s]", name)
			}
		} else { //msg to other server
			return fmt.Errorf("packetFunc_server target may be error: targetuid=%d, myuid=%d, name=%s", target, config.SERVER_NODE_UID, name)
		}
	}
	return nil
}

// 创建一个clientsocket连接server
func (self *GateServer) buildClient(uid int, addr string) *network.ClientSocket {
	client := new(network.ClientSocket)
	client.SetClientId(uid)
	client.Init(addr)
	client.SetConnectType(network.CHILD_CONNECT)
	client.BindPacketFunc(packetFunc_server)
	client.Uid = uid

	return client
}

func (self *GateServer) removeClient(uid int) {
	if _, ok := self.clients.LoadAndDelete(uid); ok {
		llog.Debugf("GateServer remove client: %d", uid)
		self.rpcUids.Delete(uid)
	}
}

func (self *GateServer) GetRpcClient(uid int) *network.ClientSocket {
	client, ok := self.clients.Load(uid)
	if ok {
		return client.(*network.ClientSocket)
	}
	return nil
}

// rpc调用的gate选择
func (self *GateServer) getClusterRpcGateUid() (uid int) {
	self.rpcUids.Range(func(key, value any) bool {
		uid = key.(int)
		return false
	})
	return
}

// StopClientByUserId 关闭客户端
func (self *GateServer) StopClientByUserId(userId int) {
	if socketId, ok := self.uidAgents.Load(userId); ok {
		self.closeClient(socketId.(int))
	}
}

func (self *GateServer) closeClient(clientId int) {
	if self.ServerType == network.CLIENT_CONNECT {
		if config.NET_WEBSOCKET {
			self.pService.(*network.WebSocket).StopClient(clientId)
		} else {
			self.pService.(*network.ServerSocket).StopClient(clientId)
		}
	} else {
		self.pInnerService.(*network.ServerSocket).StopClient(clientId)
	}
}

// SendClient gate向client发送消息
func (self *GateServer) SendClient(clientid int, buff []byte) {
	self.pService.SendById(clientid, buff)
}

// onSocketDisconnected socket连接断开
func (self *GateServer) onSocketDisconnected(socketId int) {
	if self.ServerType == network.SERVER_CONNECT { //the server lost connect with gate
		if value, ok := self.gatesUid.LoadAndDelete(socketId); ok {
			gateUid := value.(int)
			self.gatesSocketId.Delete(gateUid)
			self.userGates.Range(func(key, value any) bool { //在这个gate上的user都应该掉线
				if value.(int) == gateUid {
					igoTarget := gorpc.MGR.GetRoutine("GameServer") //告诉GameServer，client断开了，让GameServer决定如何处理，所以GameServer要注册这个actor
					if igoTarget != nil {
						igoTarget.SendActor("ON_DISCONNECT", key)
					}
					self.userGates.Delete(key)
				}
				return true
			})
		}
	} else { //the client lost connect with gate or account
		if value, ok := self.sidAgents.LoadAndDelete(socketId); ok {
			user := value.(*agent.LouMiaoAgent)
			self.uidAgents.Delete(user.UserId)
			if user.LobbyUid > 0 {
				onClientDisConnected(int(user.UserId), int(user.LobbyUid))
			}

		}

	}
}

// 向内部server直接发送buffer消息,专为gate使用，
// 必须保证线程安全，即需要在gateserver的igo中调用该函数
func (self *GateServer) SendServer(target int, buff []byte) {
	rpcClient := self.GetRpcClient(target)
	if rpcClient != nil {
		rpcClient.Send(buff)
	} else {
		llog.Warningf("GateServer.SendServer target error: target=%d", target)
	}
}

func (self *GateServer) GetClientAgent(socketId int) *agent.LouMiaoAgent {
	if v, ok := self.sidAgents.Load(socketId); ok {
		return v.(*agent.LouMiaoAgent)
	}
	return nil
}

func (self *GateServer) GetUserSocketId(userid int) int {
	if socketId, ok := self.uidAgents.Load(userid); ok {
		return socketId.(int)
	}
	return 0
}

func (self *GateServer) GetGateSocketIdByUserId(userId int) int {
	if socketId, ok := self.userGateSockets.Load(userId); ok {
		return socketId.(int)
	}
	return 0
}

func (self *GateServer) GetGateSocketIdByGateUid(gateUid int) int {
	if socketId, ok := self.gatesSocketId.Load(gateUid); ok {
		return socketId.(int)
	}
	return 0
}

// BindWorld 建立client的profile，并通知server, gate使用
func (self *GateServer) BindWorld(socketId, userId, worldUid int, tokenId string) {
	user := &agent.LouMiaoAgent{
		UserId:     int64(userId),
		LobbyUid:   worldUid,
		ZoneUid:    0,
		MatchUid:   0,
		GateUid:    config.SERVER_NODE_UID,
		SocketId:   socketId,
		WorldToken: tokenId,
	}
	self.sidAgents.Store(socketId, user)
	self.uidAgents.Store(userId, socketId)
	onClientConnected(userId, worldUid)
}

// BindZone 绑定client所属的zone, gate使用
func (self *GateServer) BindZone(userId, zoneUid int) {
	if sid, ok := self.uidAgents.Load(userId); ok {
		if user, ok := self.sidAgents.Load(sid); ok {
			user.(*agent.LouMiaoAgent).ZoneUid = zoneUid
		}
	}
}
