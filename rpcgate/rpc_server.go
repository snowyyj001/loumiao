// rpc网关服务
package rpcgate

import (
	"fmt"
	"github.com/snowyyj001/loumiao/lnats"
	"sync"

	"github.com/snowyyj001/loumiao/message"

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
	This        *RpcGateServer
	handler_Map map[string]string
)

/*
rpc转发专用的gate server，仅做rpc转发
作为sever监听，等待rpc cleint来连接
*/

type RpcGateServer struct {
	gorpc.GoRoutineLogic

	pInnerService network.ISocket

	rpcMap    map[string][]int
	rpcUids   sync.Map
	users_u   map[int]int //socketid -> uid
	clients_u map[int]int //uid -> socketid

	lock sync.Mutex
}

func (self *RpcGateServer) DoInit() bool {
	llog.Info("RpcGateServer DoInit")
	This = self

	self.pInnerService = new(network.ServerSocket)
	self.pInnerService.(*network.ServerSocket).SetMaxClients(config.NET_MAX_RPC_CONNS)
	self.pInnerService.Init(config.NET_LISTEN_SADDR)
	self.pInnerService.BindPacketFunc(packetFunc_rpc)
	self.pInnerService.SetConnectType(network.SERVER_CONNECT)

	self.rpcMap = make(map[string][]int) //base64(funcname) -> [socketid,socketid,...]
	handler_Map = make(map[string]string)
	self.users_u = make(map[int]int)
	self.clients_u = make(map[int]int)

	return true
}

func (self *RpcGateServer) DoRegister() {
	llog.Info("RpcGateServer DoRegister")

	self.Register("SendRpcMsgToServer", sendRpcMsgToServer)
	self.Register("CloseServer", closeServer)

	//equal to RegisterSelfNet
	handler_Map["CONNECT"] = "GateServer" //server connect with rpc server
	self.RegisterGate("CONNECT", innerConnect)

	handler_Map["DISCONNECT"] = "GateServer"
	self.RegisterGate("DISCONNECT", innerDisConnect)

	handler_Map["LouMiaoRpcMsg"] = "GateServer"
	self.RegisterGate("LouMiaoRpcMsg", innerLouMiaoRpcMsg)

	handler_Map["LouMiaoRpcRegister"] = "GateServer"
	self.RegisterGate("LouMiaoRpcRegister", innerLouMiaoRpcRegister)
}

// begin communicate with other nodes
func (self *RpcGateServer) DoStart() {
	llog.Info("RpcGateServer DoStart")

	//etcd client
	err := etcd.NewClient()
	if util.CheckErr(err) {
		llog.Fatalf("etcd connect failed: %v", config.Cfg.EtcdAddr)
	}

	nodemgr.ServerEnabled = true
	//etcd.Client.PutStatus() //服务如果异常关闭，是没有撤销租约的，在三秒内重启会保留上次状态，这里强制刷新一下

	llog.Debug("ServerSocket.Start i")
	util.Assert(self.pInnerService.Start(), fmt.Sprintf("GateServer listen failed: saddr=%s", self.pInnerService.GetSAddr()))

	llog.Infof("RpcGateServer DoStart success: name=%s,saddr=%s,uid=%d", self.Name, config.NET_GATE_SADDR, config.SERVER_NODE_UID)
}

// begin start socket servie
func (self *RpcGateServer) DoOpen() {

	//server discover
	//watch status, for balance
	err := etcd.Client.WatchCommon(fmt.Sprintf("%s%d", define.ETCD_NODESTATUS, config.NET_NODE_ID), self.serverStatusUpdate)
	if err != nil {
		llog.Fatalf("etcd watch ETCD_NODESTATUS error : %s", err.Error())
	}
	//watch all node, just for account, to gate balance
	err = etcd.Client.WatchCommon(fmt.Sprintf("%s%d", define.ETCD_NODEINFO, config.NET_NODE_ID), self.newServerDiscover)
	if err != nil {
		llog.Fatalf("etcd watch NET_GATE_SADDR error : %s", err.Error())
	}

	//register to etcd when the socket is ok
	if err := etcd.Client.PutNode(); err != nil {
		llog.Fatalf("etcd PutNode error %v", err)
	}
	nodemgr.ServerEnabled = true
	llog.Infof("RpcGateServer DoOpen success: name=%s, saddr=%s, uid=%d", self.Name, config.NET_GATE_SADDR, config.SERVER_NODE_UID)

	lnats.ReportMail(define.MAIL_TYPE_START, "服务器完成启动")
}

// goroutine unsafe
func (self *RpcGateServer) serverStatusUpdate(key, val string, dis bool) {
	node := nodemgr.NodeStatusUpdate(key, val, dis)
	if node == nil {
		return
	}
	if node.Uid > 0 && node.SocketActive == false {
		This.rpcUids.Delete(node.Uid)
	}
}

// goroutine unsafe
func (self *RpcGateServer) newServerDiscover(key, val string, dis bool) {
	node := nodemgr.NodeDiscover(key, val, dis)
	if node == nil {
		return
	}
}

func (self *RpcGateServer) DoDestroy() {
	llog.Info("RpcGateServer DoDestroy")
	nodemgr.ServerEnabled = false
}

// goroutine unsafe
// net msg handler,this func belong to socket's goroutine
func packetFunc_rpc(socketid int, buff []byte, nlen int) error {
	//llog.Debugf("packetFunc_rpc: socketid=%d, bufferlen=%d", socketid, nlen)
	target, name, buffbody, err := message.UnPackHead(buff, nlen)
	//llog.Debugf("packetFunc_rpc  %s %v", name, pm)
	if nil != err {
		return fmt.Errorf("packetFunc_rpc Decode error: %s", err.Error())
		//This.closeClient(socketid)
	} else {
		if target == config.SERVER_NODE_UID || target <= 0 { //server使用的是server uid
			handler, ok := handler_Map[name]
			if ok {
				nm := &gorpc.M{Id: socketid, Name: name, Data: buffbody}
				gorpc.MGR.Send(handler, "ServiceHandler", nm)
			} else {
				llog.Errorf("packetFunc_rpc handler is nil, drop it[%s]", name)
			}
		} else {
			llog.Errorf("packetFunc_rpc target may be error: targetuid=%d, myuid=%d, name=%s", target, config.SERVER_NODE_UID, name)
		}
	}
	return nil
}

func (self *RpcGateServer) removeRpc(socketId int) {

	uid := self.getUserId(socketId)
	if uid == 0 { //已经关闭了
		return
	}

	This.rpcUids.Delete(uid)

	delete(self.clients_u, uid)

	delete(self.users_u, socketId)

	//remove rpc handler
	self.removeRpchandler(socketId)

	//reset node ststus
	nodemgr.RemoveNodeById(uid)
}

func (self *RpcGateServer) removeRpchandler(socketid int) {
	//remove rpc handler
	for key, arr := range self.rpcMap {
		for i, val := range arr {
			if val == socketid {
				self.rpcMap[key] = append(arr[:i], arr[i+1:]...)
				break
			}
		}
	}
}

// rpc调用的目标server选择
func (self *RpcGateServer) getClusterServerSocketId(funcName string) int {
	arr := self.rpcMap[funcName]
	sz := len(arr)
	if sz == 0 {
		llog.Warningf("0.getCluserServerSocketId no rpc server handler finded %s", funcName)
		return 0
	}
	index := util.Random(sz) //choose a server by random
	sid := arr[index]
	return sid
}

func (self *RpcGateServer) getUserId(socketId int) int {
	uid, _ := self.users_u[socketId]
	return uid
}

func (self *RpcGateServer) getClientId(userId int) int {
	sid, _ := self.clients_u[userId]
	return sid
}
