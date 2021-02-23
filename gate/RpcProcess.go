package gate

import (
	"fmt"

	"github.com/snowyyj001/loumiao/base"

	"github.com/snowyyj001/loumiao/nodemgr"

	"github.com/snowyyj001/loumiao"

	"github.com/snowyyj001/loumiao/config"
	"github.com/snowyyj001/loumiao/define"
	"github.com/snowyyj001/loumiao/gorpc"
	"github.com/snowyyj001/loumiao/llog"
	"github.com/snowyyj001/loumiao/message"
	"github.com/snowyyj001/loumiao/msg"
	"github.com/snowyyj001/loumiao/network"
	"github.com/snowyyj001/loumiao/util"
)

//client connect
func innerConnect(igo gorpc.IGoRoutine, socketId int, data interface{}) {
	//llog.Debugf("GateServer innerConnect: %d", socketId)
	//对于account来说，socket连接数就是在线人数
	//对于gate来说，只有client发送完LouMiaoLoginGate消息，成功登录网关才算在线数
	//本质上来说差别不大
	//对于world来说，内存中玩家数就是在线数，包括暂时离线缓存的那一部分
	if config.NET_NODE_TYPE == config.ServerType_Account {
		This.OnlineNum++
	}
}

//client disconnect
func innerDisConnect(igo gorpc.IGoRoutine, socketId int, data interface{}) {
	//llog.Debugf("GateServer innerDisConnect: %d", socketId)

	if config.NET_NODE_TYPE == config.ServerType_Account {
		This.OnlineNum--
	}

	token, ok := This.tokens[socketId]
	if ok {
		userid := token.UserId
		worlduid := This.users_u[userid]
		onClientDisConnected(userid, worlduid)

		delete(This.tokens, socketId)
		delete(This.tokens_u, userid)
		delete(This.users_u, userid)
		if This.ServerType == network.SERVER_CONNECT {
			This.rpcGates = util.RemoveSlice(This.rpcGates, socketId)
		}
	}
}

//server connect
func outerConnect(igo gorpc.IGoRoutine, socketId int, data interface{}) {
	//llog.Debugf("GateServer outerConnect: %d", socketId)
}

//server disconnect
func outerDisConnect(igo gorpc.IGoRoutine, socketId int, data interface{}) {
	//	llog.Debugf("GateServer outerDisConnect: %d", socketId)

	//if config.NET_NODE_TYPE == config.ServerType_Gate {
	This.removeRpc(socketId)
	//}
}

//client connected to server
func onClientConnected(uid int, tid int) {
	llog.Debugf("GateServer onClientConnected: uid=%d,tid=%d", uid, tid)

	if This.ServerType == network.CLIENT_CONNECT {
		if config.NET_NODE_TYPE == config.ServerType_Gate { //tell world that a client has connected to this gate
			req := &msg.LouMiaoClientConnect{ClientId: int64(uid), GateId: int64(This.Id), State: define.CLIENT_CONNECT}
			buff, _ := message.Encode(0, "LouMiaoClientConnect", req)
			This.SendServer(tid, buff)
			This.OnlineNum++
		}
		handler, ok := handler_Map["ON_CONNECT"]
		if ok {
			m := gorpc.M{Id: tid, Name: "ON_CONNECT", Data: uid}
			gorpc.MGR.Send(handler, "ServiceHandler", m)
		}
	} else {
		//register rpc if has
		req := &msg.LouMiaoRpcRegister{}
		for key, _ := range This.rpcMap {
			req.FuncName = append(req.FuncName, key)
		}
		if len(req.FuncName) > 0 {
			buff, _ := message.Encode(uid, "LouMiaoRpcRegister", req)
			clientid, _ := This.tokens_u[uid]
			This.pInnerService.SendById(clientid, buff)
		}
	}

}

//client disconnected to gate
func onClientDisConnected(uid int, tid int) {
	llog.Debugf("GateServer onClientDisConnected: uid=%d,tid=%d", uid, tid)

	if config.NET_NODE_TYPE == config.ServerType_Gate {
		req := &msg.LouMiaoClientConnect{ClientId: int64(uid), GateId: int64(This.Id), State: define.CLIENT_DISCONNECT}
		buff, _ := message.Encode(0, "LouMiaoClientConnect", req)
		This.SendServer(tid, buff)
		This.OnlineNum--
	} else {
		handler, ok := handler_Map["ON_DISCONNECT"]
		if ok {
			m := gorpc.M{Id: tid, Name: "ON_DISCONNECT", Data: uid}
			gorpc.MGR.Send(handler, "ServiceHandler", m)
		}
	}
}

//gate租约回调
func leaseCallBack(success bool) {
	if success { //成功续租
		//str := fmt.Sprintf("%s%s/%s", define.ETCD_NODESTATUS, config.SERVER_GROUP, config.NET_GATE_SADDR)
		str := fmt.Sprintf("%s/%s", define.ETCD_NODESTATUS, config.NET_GATE_SADDR)
		This.clientEtcd.Put(str, util.Itoa(This.OnlineNum), true) //写入人数
		//llog.Debugf("leaseCallBack %s", str)
	} else {
		llog.Errorf("leaseCallBack续租失败")
		err := This.clientEtcd.SetLease(int64(config.GAME_LEASE_TIME), true)
		if err != nil {
			llog.Debugf("尝试重新续租失败")
		} else {
			llog.Debugf("尝试重新续租成功")
			if This.ServerType != network.CLIENT_CONNECT {
				err := This.clientEtcd.PutService(This.m_etcdKey, config.NET_GATE_SADDR)
				if err != nil {
					llog.Fatalf("leaseCallBack PutService error: %v", err)
				}
			}
		}
	}
}

//login gate
func innerLouMiaoLoginGate(igo gorpc.IGoRoutine, socketId int, data interface{}) {
	m := data.(*msg.LouMiaoLoginGate)
	userid := int(m.UserId)
	llog.Debugf("innerLouMiaoLoginGate: %v, socketId=%d", m, socketId)

	if This.OnlineNum > config.NET_MAX_NUMBER {
		llog.Warningf("0.innerLouMiaoLoginGate too many connections: max=%d, now=%d", config.NET_MAX_NUMBER, This.OnlineNum)
		This.closeClient(socketId)
		return
	}

	old_socketid, ok := This.tokens_u[userid]
	if ok { //close the old connection
		req := &msg.LouMiaoKickOut{}
		if config.NET_NODE_TYPE == config.ServerType_Gate {
			buff, _ := message.Encode(0, "LouMiaoKickOut", req) //顶号,通知老的客户端退出登录
			This.pService.SendById(old_socketid, buff)
			//关闭老的客户端的socket
			innerDisConnect(igo, old_socketid, nil) //时序异步问题，这里直接关闭，不等socket的DISCONNECT消息
			if config.NET_WEBSOCKET {
				This.pService.(*network.WebSocket).StopClient(old_socketid)
			} else {
				This.pService.(*network.ServerSocket).StopClient(old_socketid)
			}
		} else {
			buff, _ := message.Encode(userid, "LouMiaoKickOut", req)
			This.pInnerService.SendById(old_socketid, buff)
			//关闭老的gate的socket
			innerDisConnect(igo, old_socketid, nil) //时序异步问题，这里直接关闭，不等socket的DISCONNECT消息
			This.pInnerService.(*network.ServerSocket).StopClient(old_socketid)
		}
		delete(This.tokens, old_socketid)
		if This.ServerType == network.SERVER_CONNECT {
			This.rpcGates = util.RemoveSlice(This.rpcGates, old_socketid)
		}
	}
	tokenid := int(m.TokenId)
	worldid := int(m.WorldUid)
	if config.NET_NODE_TYPE == config.ServerType_Gate {
		rpcclient := This.GetRpcClient(worldid)
		if rpcclient == nil { //accout分配的world，在gate这里不存在，有可能是刚好world关闭了，这种情况就让客户端重新登录吧
			m.WorldUid = 0
			buff, _ := message.Encode(0, "LouMiaoLoginGate", m)
			This.pService.SendById(socketId, buff)
			return
		}
		This.users_u[userid] = worldid
	} else if config.NET_NODE_TYPE == config.ServerType_Account {
		This.users_u[userid] = socketId
		worldid = socketId
	}
	This.tokens[socketId] = &Token{TokenId: tokenid, UserId: userid}
	This.tokens_u[userid] = socketId
	if This.ServerType == network.SERVER_CONNECT {
		This.rpcGates = append(This.rpcGates, socketId)
	}
	onClientConnected(userid, worldid)
	if config.NET_NODE_TYPE == config.ServerType_Gate { //tell the client login success
		buff, _ := message.Encode(0, "LouMiaoLoginGate", m)
		This.pService.SendById(socketId, buff)
	}
}

//rpc register
func innerLouMiaoRpcRegister(igo gorpc.IGoRoutine, socketId int, data interface{}) {
	req := data.(*msg.LouMiaoRpcRegister)
	llog.Debugf("innerLouMiaoRpcRegister: %v", req)

	client := This.GetRpcClient(socketId)
	if client == nil {
		llog.Warningf("0.innerLouMiaoRpcRegister server has lost[%d] ", socketId)
		return
	}
	for _, key := range req.FuncName {
		if This.rpcMap[key] == nil {
			This.rpcMap[key] = []int{}
		}
		This.rpcMap[key] = append(This.rpcMap[key], socketId)
		//rpcstr, _ := base64.StdEncoding.DecodeString(key)
		//llog.Debugf("rpc register: funcname=%s, uid=%d", string(rpcstr), socketId)
	}
}

//recv rpc msg
func innerLouMiaoRpcMsg(igo gorpc.IGoRoutine, socketId int, data interface{}) {
	req := data.(*msg.LouMiaoRpcMsg)
	llog.Debugf("innerLouMiaoRpcMsg=%s, socurce=%d, target=%d, ByteBuffer=%d", req.FuncName, req.SourceId, req.TargetId, req.ByteBuffer)
	if config.NET_NODE_TYPE == config.ServerType_Gate { //server -> gate
		target := int(req.TargetId)
		var rpcClient *network.ClientSocket
		if target <= 0 {
			rpcClient = This.getCluserServer(req.FuncName)
		} else {
			rpcClient = This.GetRpcClient(target)
		}
		if rpcClient == nil {
			llog.Warningf("1.innerLouMiaoRpcMsg rpc client error %s %d", req.FuncName, target)
			return
		}
		outdata := &msg.LouMiaoRpcMsg{TargetId: req.TargetId, FuncName: req.FuncName, Buffer: req.Buffer, SourceId: req.SourceId, ByteBuffer: req.ByteBuffer}
		buff, _ := message.Encode(target, "LouMiaoRpcMsg", outdata)
		rpcClient.Send(buff)
	} else { //gate -> server or gate self
		handler, ok := handler_Map[req.FuncName]
		if ok {
			if req.ByteBuffer > 0 { //not pb but bytes
				m := gorpc.M{Id: int(req.SourceId), Name: req.FuncName, Data: req.Buffer}
				if handler == "GateServer" {
					igo.CallNetFunc(&m)
				} else {
					gorpc.MGR.Send(handler, "ServiceHandler", m)
				}
			} else {
				err, _, _, pm := message.Decode(This.Id, req.Buffer, len(req.Buffer))
				//llog.Debugf("innerLouMiaoRpcMsg : pm = %v", pm)
				//spm := reflect.Indirect(reflect.ValueOf(pm))
				//llog.Debugf("innerLouMiaoRpcMsg : spm = %v, typename=%v", spm, reflect.TypeOf(spm.Interface()).Name
				if err != nil {
					llog.Warningf("2.innerLouMiaoRpcMsg decode msg error : func=%s, error=%s ", req.FuncName, err.Error())
					return
				}
				m := gorpc.M{Id: int(req.SourceId), Name: req.FuncName, Data: pm}
				if handler == "GateServer" {
					igo.CallNetFunc(&m)
				} else {
					gorpc.MGR.Send(handler, "ServiceHandler", m)
				}
			}
		} else {
			llog.Warningf("3.InnerLouMiaoRpcMsg no rpc hanlder %s, %d", req.FuncName, socketId)
		}
	}
}

//recv broad cast msg
func innerLouMiaoBroadCastMsg(igo gorpc.IGoRoutine, socketId int, data interface{}) {
	req := data.(*msg.LouMiaoBroadCastMsg)
	llog.Debugf("innerLouMiaoBroadCastMsg=%s, type=%d, ByteBuffer=%d", req.FuncName, req.Type, req.ByteBuffer)
	if config.NET_NODE_TYPE == config.ServerType_Gate { //server -> gate
		serverType := int(req.Type)
		outdata := &msg.LouMiaoBroadCastMsg{Type: req.Type, FuncName: req.FuncName, Buffer: req.Buffer, ByteBuffer: req.ByteBuffer}
		buff, _ := message.Encode(0, "LouMiaoBroadCastMsg", outdata)
		for _, client := range This.clients {
			node := nodemgr.GetNodeByAddr(client.GetSAddr())
			if node.Type == serverType {
				client.Send(buff)
			}
		}
	} else { //gate -> server or gate self
		handler, ok := handler_Map[req.FuncName]
		if ok {
			if req.ByteBuffer > 0 { //not pb but bytes
				m := gorpc.M{Id: int(req.Type), Name: req.FuncName, Data: req.Buffer}
				if handler == "GateServer" {
					igo.CallNetFunc(&m)
				} else {
					gorpc.MGR.Send(handler, "ServiceHandler", m)
				}
			} else {
				err, _, _, pm := message.Decode(This.Id, req.Buffer, len(req.Buffer))
				//llog.Debugf("innerLouMiaoRpcMsg : pm = %v", pm)
				//spm := reflect.Indirect(reflect.ValueOf(pm))
				//llog.Debugf("innerLouMiaoRpcMsg : spm = %v, typename=%v", spm, reflect.TypeOf(spm.Interface()).Name
				if err != nil {
					llog.Warningf("2.innerLouMiaoBroadCastMsg decode msg error : func=%s, error=%s ", req.FuncName, err.Error())
					return
				}
				m := gorpc.M{Id: int(req.Type), Name: req.FuncName, Data: pm}
				if handler == "GateServer" {
					igo.CallNetFunc(&m)
				} else {
					gorpc.MGR.Send(handler, "ServiceHandler", m)
				}
			}
		} else {
			llog.Warningf("3.innerLouMiaoBroadCastMsg no rpc hanlder %s, %d", req.FuncName, socketId)
		}
	}
}

//recv net msg
func innerLouMiaoNetMsg(igo gorpc.IGoRoutine, socketId int, data interface{}) {
	req := data.(*msg.LouMiaoNetMsg) //after decode LouMiaoNetMsg msg, post Buffer to next
	//llog.Debugf("innerLouMiaoNetMsg %v", data)
	clientid := int(req.ClientId)

	if config.NET_NODE_TYPE == config.ServerType_Gate { //server -> gate
		socketId, _ := This.tokens_u[clientid]       // get client's socketid by userid
		This.pService.SendById(socketId, req.Buffer) //send to client
	} else { //gate -> server
		m := gorpc.MI{Id: clientid, Name: len(req.Buffer), Data: req.Buffer}
		recvPackMsg(igo, m)
	}
}

//exit the server
func innerLouMiaoKickOut(igo gorpc.IGoRoutine, socketId int, data interface{}) {
	//llog.Debugf("innerLouMiaoKickOut %v", data)

	if config.NET_NODE_TYPE != config.ServerType_Gate {
		llog.Error("0.innerLouMiaoKickOut wrong msg")
		return
	}
	llog.Noticef("innerLouMiaoKickOut: another gate server has already listened the saddr : %s, now close self[%d]", config.NET_GATE_SADDR, This.Id)
	loumiao.Stop()
}

//client be in connected or disconnect a gate
func innerLouMiaoClientConnect(igo gorpc.IGoRoutine, socketId int, data interface{}) {
	llog.Debugf("innerLouMiaoClientConnect %v", data)
	req := data.(*msg.LouMiaoClientConnect)
	clientid := int(req.ClientId)
	gateId := int(req.GateId)
	if req.State == define.CLIENT_CONNECT {
		This.users_u[clientid] = gateId
		handler, ok := handler_Map["ON_CONNECT"]
		if ok {
			m := gorpc.M{Id: gateId, Name: "ON_CONNECT", Data: req.ClientId}
			gorpc.MGR.Send(handler, "ServiceHandler", m)
		}
	} else {
		delete(This.users_u, clientid)
		handler, ok := handler_Map["ON_DISCONNECT"]
		if ok {
			m := gorpc.M{Id: gateId, Name: "ON_DISCONNECT", Data: req.ClientId}
			gorpc.MGR.Send(handler, "ServiceHandler", m)
		}
	}
}

//info about client with the gate
func innerLouMiaoLouMiaoBindGate(igo gorpc.IGoRoutine, socketId int, data interface{}) {
	llog.Debugf("innerLouMiaoLouMiaoBindGate %v", data)
	req := data.(*msg.LouMiaoBindGate)
	This.users_u[int(req.UserId)] = int(req.Uid)
}

func registerNet(igo gorpc.IGoRoutine, data interface{}) interface{} {
	m := data.(gorpc.M)
	sname, ok := handler_Map[m.Name]
	if ok {
		llog.Errorf("registerNet %s has already been registered: %s", m.Name, sname)
		return nil
	}
	handler_Map[m.Name] = m.Data.(string)
	if m.Id < 0 { //rpc register
		This.rpcMap[m.Name] = []int{}
	}
	return nil
}

func unRegisterNet(igo gorpc.IGoRoutine, data interface{}) interface{} {
	return nil
}

func sendClient(igo gorpc.IGoRoutine, data interface{}) interface{} {
	m := data.(gorpc.M)
	if config.NET_NODE_TYPE == config.ServerType_Gate {
		llog.Error("0.sendClient gate can not send client")
		return nil
	}
	if config.NET_NODE_TYPE == config.ServerType_Account {
		This.pService.SendById(m.Id, m.Data.([]byte)) //set to client
		return nil
	}

	uid, _ := This.users_u[m.Id] // get gate's uid by client's userid
	if uid == 0 {
		llog.Noticef("1.sendClient cannot find which gate the client belong: %d", m.Id)
		return nil
	}
	socketId, _ := This.tokens_u[uid]
	if socketId == 0 {
		llog.Noticef("2.sendClient gate has been shut down, uid = %d", m.Id)
		return nil
	}

	//llog.Debugf("sendClient userid = %d, uid = %d", m.Id, uid)
	msg := &msg.LouMiaoNetMsg{ClientId: int64(m.Id), Buffer: m.Data.([]byte)} //m.Id should be client`s userid
	buff, _ := message.Encode(0, "LouMiaoNetMsg", msg)

	This.pInnerService.SendById(socketId, buff)

	return nil
}

func sendMulClient(igo gorpc.IGoRoutine, data interface{}) interface{} {
	m := data.(gorpc.MS)
	if config.NET_NODE_TYPE == config.ServerType_Gate {
		llog.Error("0.sendMulClient gate can not send client")
		return nil
	}
	for _, v := range m.Ids {
		uid, _ := This.users_u[v] // get gate's socketid by client's userid
		socketId, ok := This.tokens_u[uid]
		if ok {
			llog.Noticef("1.sendMulClient gate has been shut down, uid = %d", v)
			return nil
		}
		msg := &msg.LouMiaoNetMsg{ClientId: int64(v), Buffer: m.Data.([]byte)} //m.Id should be client`s userid
		buff, _ := message.Encode(0, "LouMiaoNetMsg", msg)

		This.pInnerService.SendById(socketId, buff)
	}

	return nil
}

//new rpc added
func newRpc(igo gorpc.IGoRoutine, data interface{}) interface{} {
	m := data.(gorpc.M)
	llog.Debugf("GateServer newRpc: %d", m.Id)
	uid := m.Id

	rpcclient, _ := This.clients[uid]
	client := m.Data.(*network.ClientSocket)
	if rpcclient != nil {
		llog.Warningf("GateServer newRpc: server[%d] has already connected", uid)
		client.SetClientId(0)
		client.Stop()
		return nil
	}
	This.clients[uid] = client

	if config.NET_NODE_TYPE == config.ServerType_Gate { //gate才需要向server登录，accoutn目前没有需求
		req := &msg.LouMiaoLoginGate{TokenId: int64(uid), UserId: int64(This.Id)}
		buff, _ := message.Encode(uid, "LouMiaoLoginGate", req)
		client.Send(buff)
	}
	return nil
}

//send rpc msg to gate server
func sendRpc(igo gorpc.IGoRoutine, data interface{}) interface{} {

	base.Assert(This.ServerType == network.SERVER_CONNECT, "gate or account can not call this func")

	m := data.(gorpc.MM)
	clientid := This.getCluserGateClientId()
	if clientid <= 0 {
		llog.Warningf("0.sendRpc no gate server finded %s", m.Name)
		return nil
	}
	//llog.Debugf("sendRpc: %d", clientid)

	outdata := &msg.LouMiaoRpcMsg{TargetId: int64(m.Id), FuncName: m.Name, Buffer: m.Data.([]byte), SourceId: int64(This.Id), ByteBuffer: int32(m.Param)}
	buff, _ := message.Encode(0, "LouMiaoRpcMsg", outdata)
	This.pInnerService.SendById(clientid, buff)

	return nil
}

//broad cast rpc msg to gate server
func broadCastRpc(igo gorpc.IGoRoutine, data interface{}) interface{} {

	base.Assert(This.ServerType == network.SERVER_CONNECT, "gate or account can not call this func")

	m := data.(gorpc.MM)
	clientid := This.getCluserGateClientId()
	if clientid <= 0 {
		llog.Warningf("0.broadCastRpc no gate server finded %s", m.Name)
		return nil
	}
	//llog.Debugf("broadCastRpc: %d", clientid)

	outdata := &msg.LouMiaoBroadCastMsg{Type: int32(m.Id), FuncName: m.Name, Buffer: m.Data.([]byte), ByteBuffer: int32(m.Param)}
	buff, _ := message.Encode(0, "LouMiaoBroadCastMsg", outdata)
	This.pInnerService.SendById(clientid, buff)

	return nil
}

//send msg to gate
func sendGate(igo gorpc.IGoRoutine, data interface{}) interface{} {
	if This.ServerType == network.CLIENT_CONNECT {
		llog.Error("0.sendGate gate or account can not call this function")
		return nil
	}
	m := data.(gorpc.M)
	uid := m.Id
	var clientid int = 0

	//llog.Debugf("sendGate: %d", clientid)
	if uid == 0 {
		clientid = This.getCluserGateClientId()
	} else {
		clientid, _ = This.tokens_u[uid]
	}
	if clientid <= 0 {
		llog.Warning("1.sendGate no gate server finded")
		return nil
	}

	This.pInnerService.SendById(clientid, m.Data.([]byte))
	return nil
}

func recvPackMsg(igo gorpc.IGoRoutine, data interface{}) interface{} {
	m := data.(gorpc.MI)
	socketid := m.Id

	err, target, name, pm := message.Decode(This.Id, m.Data.([]byte), m.Name)

	if err != nil {
		llog.Errorf("recvPackMsg Decode error: %s", err.Error())
		This.closeClient(socketid)
		return nil
	}
	if target == This.Id || target <= 0 { //send to me
		handler, ok := handler_Map[name]
		if ok {
			if handler == This.Name {
				cb, ok := This.NetHandler[name]
				if ok {
					cb(This, socketid, pm)
				} else {
					llog.Warningf("recvPackMsg[%s] handler is nil: %s", name, This.Name)
				}
			} else {
				nm := gorpc.M{Id: socketid, Name: name, Data: pm}
				gorpc.MGR.Send(handler, "ServiceHandler", nm)
			}

		} else {
			llog.Warningf("recvPackMsg self handler is nil, drop it[%s]", name)
		}
	} else { //send to target server
		rpcClient := This.GetRpcClient(target)
		if rpcClient == nil {
			llog.Debugf("1.recvPackMsg msg, but server has lost[%d][%d] ", target, m.Name)
			node := nodemgr.GetNode(target)
			if node == nil || node.Type == config.ServerType_World { //if the world lost,we should let the client relogin
				This.closeClient(socketid)
			}
			return nil
		}

		token, ok := This.tokens[socketid]
		if ok == false {
			llog.Infof("0.recvPackMsg recv client msg, but client has lost[%d][%d][%s][%d] ", target, socketid, name, m.Name)
			return nil
		}
		msg := &msg.LouMiaoNetMsg{ClientId: int64(token.UserId), Buffer: m.Data.([]byte)}
		buff, newlen := message.Encode(target, "LouMiaoNetMsg", msg)
		rpcClient.Send(buff[0:newlen])
	}

	return nil
}

func reportOnLineNum(igo gorpc.IGoRoutine, data interface{}) interface{} {
	This.OnlineNum = data.(int)
	return nil
}

func closeServer(igo gorpc.IGoRoutine, data interface{}) interface{} {
	uid := data.(int)
	llog.Debugf("closeServer: %d", uid)

	if uid == This.Id {
		//if This.ServerType == network.SERVER_CONNECT { //removed from server discover
		//	This.clientEtcd.DelService(This.m_etcdKey)
		//}
		return nil
	}

	if This.ServerType == network.CLIENT_CONNECT {
		rpcClient := This.GetRpcClient(uid)
		if rpcClient != nil {
			This.removeRpcHanlder(uid)
		}
	} else {
		socketId, _ := This.tokens_u[uid]
		token, _ := This.tokens[socketId]
		if token != nil {
			This.rpcGates = util.RemoveSlice(This.rpcGates, socketId)
		}
	}
	return nil
}

//gate send msg to all clients or
//server send msg to all gates
func broadCastClients(buff []byte) {
	llog.Debugf("GateServer broadCastClients: %d", This.Id)
	if config.NET_NODE_TYPE == config.ServerType_Gate {
		This.pService.BroadCast(buff)
	} else {
		This.pInnerService.BroadCast(buff)
	}
}
