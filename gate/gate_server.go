// 网关服务
package gate

import (
	"fmt"
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
	This        *GateServer
	handler_Map map[string]string //goroutine unsafe
)

type Token struct {
	UserId  int    //userid
	WouldId int    //world uid
	ZoneUid int    //zone uid
	TokenId string //合法性校验tokenid
}

/*tokens, tokens_u, users_u说明
当loumiao是account时:
tokens: 没有用
tokens_u: 没有用
================================================================================
当loumiao是gate时:
tokens：key是client的socketid，Token.UserId是client的userid，Token.TokenId是client的tokenid，Token.WouldId是client所属的world uid，Token.ZoneUid是client所属的zone uid
tokens_u：key是client的userid,value是client的socketid
users_u: 没有用
================================================================================
当loumiao是server时:
tokens：key是gate的socketid，Token.UserId是gate的uid，Token.TokenId是自己的uid(0代表该gate将要被关闭)
tokens_u：key是gate的uid,value是socketid
users_u: key是client的userid,value是gate的uid
*/

type GateServer struct {
	gorpc.GoRoutineLogic

	pService      network.ISocket
	clients       map[int]*network.ClientSocket
	pInnerService network.ISocket
	ServerType    int
	//
	tokens   map[int]*Token
	tokens_u map[int]int
	users_u  map[int]int

	//rpc相关
	rpcMap   map[string]string //rpc functon -> actor name
	rpcGates []int             //space for time, all rpcserver's uid
	rpcUids  sync.Map          //rpc uid -> true

	//单个actor内的消息pub/sub
	msgQueueMap map[string]*maps.Map

	//排队服uid
	QueueServerUid int

	lock sync.Mutex

	InitFunc func() //需要额外处理的函数回调
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

	self.clients = make(map[int]*network.ClientSocket)

	self.tokens = make(map[int]*Token)
	self.tokens_u = make(map[int]int)
	self.users_u = make(map[int]int)
	self.rpcMap = make(map[string]string)
	self.rpcGates = make([]int, 0)
	self.msgQueueMap = make(map[string]*maps.Map) //key -> [actorname,actorname,...]

	handler_Map = make(map[string]string)

	if self.InitFunc != nil {
		self.InitFunc()
	}

	return true
}

func (self *GateServer) DoRegsiter() {
	llog.Infof("%s DoRegsiter", self.Name)

	self.Register("RegisterNet", registerNet)
	self.Register("UnRegisterNet", unRegisterNet)
	self.Register("SendClient", sendClient)
	self.Register("SendMulClient", sendMulClient)
	self.Register("NewClient", newClient)
	self.Register("SendRpc", sendRpc)
	self.Register("SendGate", sendGate)
	self.Register("RecvPackMsgClient", recvPackMsgClient)
	self.Register("CloseServer", closeServer)
	self.Register("Publish", publish)
	self.Register("Subscribe", subscribe)
	self.Register("BindGate", bindGate)

	//equal to RegisterSelfNet
	handler_Map["CONNECT"] = "GateServer" //in gate, client connect with gate, in server, gate(as client) connect with server
	self.RegisterGate("CONNECT", innerConnect)

	handler_Map["DISCONNECT"] = "GateServer"
	self.RegisterGate("DISCONNECT", innerDisConnect)

	handler_Map["C_CONNECT"] = "GateServer" //in gate, gate connect with server, no in server
	self.RegisterGate("C_CONNECT", outerConnect)

	handler_Map["C_DISCONNECT"] = "GateServer"
	self.RegisterGate("C_DISCONNECT", outerDisConnect)

	handler_Map["LouMiaoLoginGate"] = "GateServer"
	self.RegisterGate("LouMiaoLoginGate", innerLouMiaoLoginGate)

	handler_Map["LouMiaoRpcMsg"] = "GateServer"
	self.RegisterGate("LouMiaoRpcMsg", innerLouMiaoRpcMsg)

	handler_Map["LouMiaoBroadCastMsg"] = "GateServer"
	self.RegisterGate("LouMiaoBroadCastMsg", innerLouMiaoBroadCastMsg)

	handler_Map["LouMiaoNetMsg"] = "GateServer"
	self.RegisterGate("LouMiaoNetMsg", innerLouMiaoNetMsg)

	handler_Map["LouMiaoKickOut"] = "GateServer"
	self.RegisterGate("LouMiaoKickOut", innerLouMiaoKickOut)

	handler_Map["LouMiaoClientConnect"] = "GateServer"
	self.RegisterGate("LouMiaoClientConnect", innerLouMiaoClientConnect)

	self.RegisterSelfCallRpc(bindServer)
}

//begin communicate with other nodes
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
			llog.Debug("ServerSocket.Start000")
			util.Assert(self.pService.Start(), fmt.Sprintf("GateServer listen failed: saddr=%s", self.pService.GetSAddr()))
		}
		if self.pInnerService != nil {
			util.Assert(self.pInnerService.Start(), fmt.Sprintf("GateServer listen failed: saddr=%s", self.pInnerService.GetSAddr()))
		}
	}

	need := config.NET_NODE_TYPE == config.ServerType_Gate              //挑选world给client需要
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
	err = etcd.Client.WatchCommon(key, self.newServerDiscover)
	if err != nil {
		llog.Fatalf("etcd watch NET_GATE_SADDR error: %s", err.Error())
	}

	llog.Infof("GateServer DoStart success: name=%s,saddr=%s,uid=%d", self.Name, config.NET_GATE_SADDR, config.SERVER_NODE_UID)
}

//begin start socket servie
func (self *GateServer) DoOpen() {
	llog.Info("GateServer.DoOpen")

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
// simple register self net hanlder, this func can only be called before igo started
func (self *GateServer) RegisterSelfNet(hanlderName string, hanlderFunc gorpc.HanlderNetFunc) {
	if self.IsRunning() {
		llog.Fatal("RegisterSelfNet error, igo has already started")
		return
	}
	handler_Map[hanlderName] = "GateServer"
	self.RegisterGate(hanlderName, hanlderFunc)
}

// RegisterSelfRpc
// simple register self rpc hanlder, this func can only be called before igo started
func (self *GateServer) RegisterSelfRpc(hanlderFunc gorpc.HanlderNetFunc) {
	if self.IsRunning() {
		llog.Fatal("RegisterSelfNet error, igo has already started")
		return
	}
	funcName := util.RpcFuncName(hanlderFunc)
	handler_Map[funcName] = "GateServer"
	This.rpcMap[funcName] = "GateServer"
	self.RegisterGate(funcName, hanlderFunc)
}

// RegisterSelfCallRpc
// simple register self rpc hanlder, this func can only be called before igo started
func (self *GateServer) RegisterSelfCallRpc(hanlderFunc gorpc.HanlderFunc) {
	if self.IsRunning() {
		llog.Fatal("RegisterSelfNet error, igo has already started")
		return
	}
	funcName := util.RpcFuncName(hanlderFunc)
	//fmt.Println("funcName", funcName)
	self.Register(funcName, hanlderFunc)
	handler_Map[funcName] = "GateServer"
	This.rpcMap[funcName] = "GateServer"
}

//
func (self *GateServer) getServerService() network.ISocket {
	if self.ServerType == network.CLIENT_CONNECT { //对外(login,gate)
		return self.pService
	} else {
		return self.pInnerService
	}
}

//goroutine unsafe
func (self *GateServer) serverStatusUpdate(key, val string, dis bool) {
	node := nodemgr.NodeStatusUpdate(key, val, dis)
	if node == nil || !dis {
		return
	}

	if node.Uid > 0 && node.SocketActive == false && node.Type == config.ServerType_RPCGate {
		_, ok := This.rpcUids.Load(node.Uid)
		if ok {
			gorpc.MGR.SendActor("GateServer", "CloseServer", node.Uid)
		}
	}
}

//goroutine unsafe
func (self *GateServer) newServerDiscover(key, val string, dis bool) {
	node := nodemgr.NodeDiscover(key, val, dis)
	if node == nil || !dis {
		return
	}

	if node.Type == config.ServerType_WEB_LOGIN ||
		node.Type == config.ServerType_WEB_GM { //filter weblogin/webserver,他们不参与rpc调用，不参与集群组网
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
				m := &gorpc.M{Id: node.Uid, Data: client, Param: 0} //rpc client
				gorpc.MGR.Send("GateServer", "NewClient", m)
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
			if rpcClient == nil { //this conditation can be etcd reconnect
				client := self.buildClient(node.Uid, node.SAddr)
				if self.enableClient(client) {
					m := &gorpc.M{Id: node.Uid, Data: client, Param: 1} //client client
					gorpc.MGR.Send("GateServer", "NewClient", m)
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

func (self *GateServer) DoDestory() {
	llog.Info("GateServer DoDestory")
	nodemgr.ServerEnabled = false
	etcd.Client.RevokeLease()
}

//goroutine unsafe
//net msg handler,this func belong to socket's goroutine
func packetFunc_client(socketid int, buff []byte, nlen int) error {
	//llog.Debugf("packetFunc: socketid=%d, bufferlen=%d", socketid, nlen)
	target, name, buffbody, err := message.UnPackHead(buff, nlen)
	//llog.Debugf("packetFunc_client %d %s %d", target, name, nlen)
	if nil != err {
		return fmt.Errorf("packetFunc_client Decode error: %s", err.Error())
		//This.closeClient(socketid)
	} else {
		if target == config.NET_NODE_TYPE || target <= 0 { //msg to me，client使用的是server type
			handler, ok := handler_Map[name]
			if ok {
				nm := &gorpc.M{Id: socketid, Name: name, Data: buffbody}
				gorpc.MGR.Send(handler, "ServiceHandler", nm)
			} else {
				llog.Errorf("packetFunc_client handler is nil, drop it[%s]", name)
			}
		} else { //msg to other server
			if config.NET_NODE_TYPE != config.ServerType_Gate { //only gate can forward msg
				return fmt.Errorf("packetFunc_client target may be error: target=%d, mytype=%d, name=%s", target, config.NET_NODE_TYPE, name)
			}
			m := &gorpc.M{Id: socketid, Param: target, Data: buff}
			gorpc.MGR.Send("GateServer", "RecvPackMsgClient", m)
		}
	}
	return nil
}

//goroutine unsafe
//net msg handler,this func belong to socket's goroutine
func packetFunc_server(socketid int, buff []byte, nlen int) error {
	//llog.Debugf("packetFunc_server: socketid=%d, bufferlen=%d", socketid, nlen)
	target, name, buffbody, err := message.UnPackHead(buff, nlen)
	//llog.Debugf("packetFunc_server %d %s %v", target, name, err)
	if nil != err {
		return fmt.Errorf("packetFunc_server Decode error: %s", err.Error())
		//This.closeClient(socketid)
	} else {
		if target == config.SERVER_NODE_UID || target <= 0 { //server使用的是server uid
			handler, ok := handler_Map[name] //handler_Map will not changed, so use here is ok
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

//创建一个clientsocket连接server
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
	_, ok := self.clients[uid]
	if !ok {
		return
	}
	llog.Debugf("GateServer removeRpc: %d", uid)

	delete(self.clients, uid)

	self.rpcGates = util.RemoveSlice(self.rpcGates, uid) //将这个gate的rpc调用移除
	self.rpcUids.Delete(uid)
}

func HasRpcGate() bool {
	return len(This.rpcGates) > 0
}

func (self *GateServer) GetRpcClient(uid int) *network.ClientSocket {
	client, ok := self.clients[uid]
	if ok {
		return client
	}
	return nil
}

//rpc调用的gate选择
func (self *GateServer) getCluserRpcGateUid() int {

	sz := len(self.rpcGates)
	if sz == 0 {
		llog.Warning("0.getCluserGateUid no rpc gate server finded ")
		return 0
	} else {
		index := util.Random(sz) //choose a gate by random
		return self.rpcGates[index]
	}

	return 0
}

func (self *GateServer) GetClientToken(socketId int) *Token {
	if token, ok := self.tokens[socketId]; ok {
		return token
	}
	return nil
}

func (self *GateServer) GetClientId(userid int) int {
	socketId, _ := self.tokens_u[userid]
	return socketId
}

func (self *GateServer) GetGateUid(userid int) int {
	uid, _ := self.users_u[userid]
	return uid
}

// get gate's socketid by client's userid, for server
func (self *GateServer) GetGateClientId(userid int) int {
	uid, _ := self.users_u[userid]
	socketId, _ := self.tokens_u[uid]
	return socketId
}

//关闭客户端
//@sync 是否等立即清除连接记录
//@userId 客户端uid
func (self *GateServer) StopClient(sync bool, userId int) {
	sid := This.tokens_u[userId]
	if sid > 0 {
		if sync { //时序异步问题，直接关闭，不等socket的DISCONNECT消息
			innerDisConnect(self, sid, nil)
		}
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

// gate向client发送消息
func (self *GateServer) SendClient(clientid int, buff []byte) {
	//llog.Debugf("GateServer.SendClient: clientid=%d, bufflen=%d, buff=%v", clientid, len(buff), buff)
	self.pService.SendById(clientid, buff)
}

// BindZone 客户端绑定zone，gate使用
func (self *GateServer) BindZone(userid, Zoneuid int) {
	socketid := self.GetClientId(userid)
	token := self.GetClientToken(socketid)
	if token == nil {
		return
	}
	token.ZoneUid = Zoneuid
}

// 客户端绑定gate，gate使用
func (self *GateServer) BindClient(socketId, userid, worlduid int, tokenid string) {
	self.tokens[socketId] = &Token{TokenId: tokenid, UserId: userid, WouldId: worlduid}
	self.tokens_u[userid] = socketId
	onClientConnected(userid, worlduid)
}

// 客户端解绑gate，gate使用
func (self *GateServer) UnBindClient(socketId int) {
	token := self.GetClientToken(socketId)
	if token != nil {
		delete(self.tokens, socketId)
		delete(self.tokens_u, token.UserId)
		onClientDisConnected(token.UserId, token.WouldId)
	}
}

// gate绑定server，server使用
func (self *GateServer) BindServerGate(socketId, userid int, tokenid string) {
	self.tokens[socketId] = &Token{TokenId: tokenid, UserId: userid}
	self.tokens_u[userid] = socketId
}

// 解除/绑定client所属的gate, server使用
func (self *GateServer) BindGate(userid, gateuid int) {
	if gateuid > 0 {
		self.users_u[userid] = gateuid
	} else {
		delete(self.users_u, userid)
	}
}

// 解除gate的绑定，server使用
func (self *GateServer) UnBindGate(socketId int) {
	token := This.GetClientToken(socketId)
	if token != nil {
		gateuid := token.UserId
		delete(This.tokens, socketId) //和gate解绑
		delete(This.tokens_u, gateuid)

		for userid, mygateuid := range This.users_u { //在这个gate上的user都应该掉线
			if mygateuid == gateuid {
				onClientDisConnected(userid, gateuid)
				self.BindGate(userid, 0)
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
