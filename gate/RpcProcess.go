package gate

import (
	"github.com/snowyyj001/loumiao"
	"github.com/snowyyj001/loumiao/base"
	"github.com/snowyyj001/loumiao/base/maps"
	"github.com/snowyyj001/loumiao/config"
	"github.com/snowyyj001/loumiao/define"
	"github.com/snowyyj001/loumiao/gorpc"
	"github.com/snowyyj001/loumiao/llog"
	"github.com/snowyyj001/loumiao/message"
	"github.com/snowyyj001/loumiao/msg"
	"github.com/snowyyj001/loumiao/network"
	"github.com/snowyyj001/loumiao/nodemgr"
	"github.com/snowyyj001/loumiao/util"
)

/*
对于onconnect， 不论是client -> gate/account，还是gate -> server
都是在login成功之后才算onconnect
对于ondisconnect，在socket断开时就算
所以bind信息时，在login成功之后，unbind时在socket的断开时
*/

//client connect
func innerConnect(igo gorpc.IGoRoutine, socketId int, data []byte) {
	llog.Debugf("GateServer innerConnect: %d", socketId)
	//对于account/gate来说，socket连接数就是在线人数
	//本质上来说差别不大
	//对于world来说，内存中玩家数就是在线数，包括暂时离线缓存的那一部分
	//对于zone来说，就是参与战斗的玩家数
	if This.ServerType == network.CLIENT_CONNECT { //对外(login,gate)
		nodemgr.OnlineNum++
	}
}

//client disconnect
func innerDisConnect(igo gorpc.IGoRoutine, socketId int, data []byte) {
	llog.Debugf("GateServer innerDisConnect: %d", socketId)
	if This.ServerType == network.CLIENT_CONNECT { //对外(login,gate)
		nodemgr.OnlineNum--
	}
	if This.ServerType == network.SERVER_CONNECT { //the server lost connect with gate
		This.UnBindGate(socketId)
	} else { //the client lost connect with gate or account
		This.UnBindClient(socketId)
	}
}

//server connect
func outerConnect(igo gorpc.IGoRoutine, socketId int, data []byte) {
	//llog.Debugf("GateServer outerConnect: %d", socketId)
}

//server disconnect
func outerDisConnect(igo gorpc.IGoRoutine, socketId int, data []byte) {
	//	llog.Debugf("GateServer outerDisConnect: %d", socketId)

	//ClientSocket的socketId就是他连接的server的uid
	This.removeClient(socketId)
}

//client connected to server
func onClientConnected(uid int, tid int) {
	llog.Debugf("GateServer onClientConnected: uid=%d,tid=%d", uid, tid)
	if This.ServerType != network.CLIENT_CONNECT {

	}
	if config.NET_NODE_TYPE == config.ServerType_Gate { //tell world that a client has login to this gate
		req := &msg.LouMiaoClientConnect{ClientId: int64(uid), GateId: int32(config.SERVER_NODE_UID), State: define.CLIENT_CONNECT}
		buff, _ := message.Encode(0, "LouMiaoClientConnect", req)
		This.SendServer(tid, buff)
	}
}

//client/gate disconnected to gate/server
func onClientDisConnected(uid int, tid int) {
	llog.Debugf("GateServer onClientDisConnected: uid=%d,tid=%d", uid, tid)
	if config.NET_NODE_TYPE == config.ServerType_Gate { ////tell world that a client has disconnect with this gate
		req := &msg.LouMiaoClientConnect{ClientId: int64(uid), GateId: int32(config.SERVER_NODE_UID), State: define.CLIENT_DISCONNECT}
		buff, _ := message.Encode(0, "LouMiaoClientConnect", req)
		This.SendServer(tid, buff)
	} else {
		igoTarget := gorpc.MGR.GetRoutine("GameServer")		//告诉GameServer，client断开了，让GameServer决定如何处理，所以GameServer要注册这个actor
		if igoTarget != nil {
			mm := gorpc.MA{Id: tid, Param: uid}
			igoTarget.SendActor("ON_DISCONNECT", mm)
		}
	}
}

//login gate, gate -> server
func innerLouMiaoLoginGate(igo gorpc.IGoRoutine, socketId int, data []byte) {
	req := &msg.LouMiaoLoginGate{}
	if message.UnPack(req, data) != nil {
		return
	}
	userid := int(req.UserId)
	llog.Debugf("innerLouMiaoLoginGate: %v, socketId=%d", req, socketId)

	if This.ServerType != network.SERVER_CONNECT {
		llog.Errorf("0.innerLouMiaoLoginGate only gate can send LouMiaoLoginGate to server for a login， gateuid=%d", userid)
		return
	}
	old_socketid, ok := This.tokens_u[userid]
	if ok { //close the old connection
		req := &msg.LouMiaoKickOut{}
		buff, _ := message.Encode(userid, "LouMiaoKickOut", req)
		This.pInnerService.SendById(old_socketid, buff)
		//关闭老的gate的socket
		innerDisConnect(igo, old_socketid, nil) //时序异步问题，这里直接关闭，不等socket的DISCONNECT消息
		This.pInnerService.(*network.ServerSocket).StopClient(old_socketid)
	}
	This.BindServerGate(socketId, userid, int(req.TokenId))
}

//recv rpc msg
func innerLouMiaoRpcMsg(igo gorpc.IGoRoutine, socketId int, data []byte) {
	req := &msg.LouMiaoRpcMsg{}
	if message.UnPack(req, data) != nil {
		return
	}
	llog.Debugf("innerLouMiaoRpcMsg=%s, socurce=%d, target=%d, Flag=%d", req.FuncName, req.SourceId, req.TargetId, req.Flag)
	if util.HasBit(int(req.Flag), define.RPCMSG_FLAG_CALL) {
		if util.HasBit(int(req.Flag), define.RPCMSG_FLAG_RESP) {
			gorpc.MGR.SendActor("CallRpcServer", "RespRpcCall", req)
		} else {
			handler, ok := handler_Map[req.FuncName]
			if !ok {
				llog.Errorf("2.InnerLouMiaoRpcMsg no rpc hanlder %s, %d", req.FuncName, socketId)
				return
			}
			mm := &gorpc.MM{}
			mm.Id = handler
			mm.Data = req
			gorpc.MGR.SendActor("CallRpcServer", "ReqRpcCall", mm)
		}
		return
	}

	handler, ok := handler_Map[req.FuncName]
	if !ok {
		llog.Errorf("0.InnerLouMiaoRpcMsg no rpc hanlder %s, %d", req.FuncName, socketId)
		return
	}
	m := &gorpc.M{Id: int(req.SourceId), Name: req.FuncName, Data: req.Buffer}
	if handler == "GateServer" {
		igo.CallNetFunc(m)
	} else {
		gorpc.MGR.Send(handler, "ServiceHandler", m)
	}
}

//recv broad cast msg
func innerLouMiaoBroadCastMsg(igo gorpc.IGoRoutine, socketId int, data []byte) {
	req := &msg.LouMiaoBroadCastMsg{}
	if message.UnPack(req, data) != nil {
		return
	}
	llog.Debugf("innerLouMiaoBroadCastMsg=%s, type=%d", req.FuncName, req.Type)
	if config.NET_NODE_TYPE == config.ServerType_Gate { //server -> gate
		serverType := int(req.Type)
		outdata := &msg.LouMiaoBroadCastMsg{Type: req.Type, FuncName: req.FuncName, Buffer: req.Buffer}
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
			m := &gorpc.M{Id: int(req.Type), Name: req.FuncName, Data: req.Buffer}
			if handler == "GateServer" {
				igo.CallNetFunc(m)
			} else {
				gorpc.MGR.Send(handler, "ServiceHandler", m)
			}
		} else {
			llog.Warningf("3.innerLouMiaoBroadCastMsg no rpc hanlder %s, %d", req.FuncName, socketId)
		}
	}
}

//recv net msg
func innerLouMiaoNetMsg(igo gorpc.IGoRoutine, socketId int, data []byte) {
	req := &msg.LouMiaoNetMsg{} //after decode LouMiaoNetMsg msg, post Buffer to next
	if message.UnPack(req, data) != nil {
		return
	}
	//llog.Debugf("innerLouMiaoNetMsg %v", req)
	userid := int(req.ClientId)

	if config.NET_NODE_TYPE == config.ServerType_Gate { //server -> gate, for msg to client
		socketId := This.GetClientId(userid)         // get client's socketid by userid
		This.pService.SendById(socketId, req.Buffer) //send to client
	} else { // gate -> server
		target, name, buffbody, err := message.UnPackHead(req.Buffer, len(req.Buffer))
		if err != nil {
			llog.Errorf("innerLouMiaoNetMsg Decode error: err=%s, uid=%d,target=%d", err.Error(), config.SERVER_NODE_UID, target)
		} else {
			//不校验target了，因为gate没有把server type 替换为server uid，内部转发，没必要校验
			/*if target != config.SERVER_NODE_UID && target > 0 {
				llog.Errorf("innerLouMiaoNetMsg target may be error: targetuid=%d, myuid=%d, name=%s", target, config.SERVER_NODE_UID, name)
				return
			}*/
			handler, ok := handler_Map[name]
			if ok {
				if name == This.Name {
					cb, ok := This.NetHandler[name]
					if ok {
						cb(This, userid, buffbody)
					} else {
						llog.Errorf("innerLouMiaoNetMsg[%s] handler is nil: %s", name, This.Name)
					}
				} else {
					nm := &gorpc.M{Id: userid, Name: name, Data: buffbody}
					gorpc.MGR.Send(handler, "ServiceHandler", nm)
				}
			} else {
				llog.Warningf("innerLouMiaoNetMsg handler is nil, drop it[%s][%d][%d]", name, target, config.SERVER_NODE_UID)
			}
		}
	}
}

//exit the server
func innerLouMiaoKickOut(igo gorpc.IGoRoutine, socketId int, data []byte) {
	//llog.Debugf("innerLouMiaoKickOut %v", data)

	if config.NET_NODE_TYPE != config.ServerType_Gate {
		llog.Error("0.innerLouMiaoKickOut wrong msg")
		return
	}
	llog.Infof("innerLouMiaoKickOut: another gate server has already listened the saddr : %s, now close self[%d]", config.NET_GATE_SADDR, config.SERVER_NODE_UID)
	loumiao.Stop()
}

//client be in connected or disconnect a gate
func innerLouMiaoClientConnect(igo gorpc.IGoRoutine, socketId int, data []byte) {
	req := &msg.LouMiaoClientConnect{}
	if message.UnPack(req, data) != nil {
		return
	}
	llog.Debugf("innerLouMiaoClientConnect %v", req)

	clientid := int(req.ClientId)
	gateId := int(req.GateId)
	if req.State == define.CLIENT_CONNECT {
		This.BindGate(clientid, gateId) //绑定client所属的gate
		igoTarget := gorpc.MGR.GetRoutine("GameServer")		//告诉GameServer，client断开了，让GameServer决定如何处理，所以GameServer要注册这个actor
		if igoTarget != nil {
			mm := gorpc.MA{Id: gateId, Param: clientid}
			igoTarget.SendActor("ON_CONNECT", mm)
		}
	} else {
		This.BindGate(clientid, 0) //解除绑定client所属的gate
		igoTarget := gorpc.MGR.GetRoutine("GameServer")		//告诉GameServer，client断开了，让GameServer决定如何处理，所以GameServer要注册这个actor
		if igoTarget != nil {
			mm := gorpc.MA{Id: gateId, Param: clientid}
			igoTarget.SendActor("ON_DISCONNECT", mm)
		}
	}
}

//info about client with the gate
func innerLouMiaoLouMiaoBindGate(igo gorpc.IGoRoutine, socketId int, data []byte) {
	llog.Debugf("innerLouMiaoLouMiaoBindGate %v", data)
	req := &msg.LouMiaoBindGate{}
	if message.UnPack(req, data) != nil {
		return
	}
	This.BindGate(int(req.UserId), int(req.Uid))
}

func registerNet(igo gorpc.IGoRoutine, data interface{}) interface{} {
	m := data.(*gorpc.M)
	sname, ok := handler_Map[m.Name]
	if ok {
		llog.Errorf("registerNet %s has already been registered: %s", m.Name, sname)
		return nil
	}
	handler_Map[m.Name] = m.Data.(string)
	if m.Id < 0 { //rpc register
		This.rpcMap[m.Name] = m.Data.(string)
	}
	return nil
}

func unRegisterNet(igo gorpc.IGoRoutine, data interface{}) interface{} {
	return nil
}

func sendClient(igo gorpc.IGoRoutine, data interface{}) interface{} {
	m := data.(*gorpc.M)
	if config.NET_NODE_TYPE == config.ServerType_Account || config.NET_NODE_TYPE == config.ServerType_Gate {
		This.pService.SendById(m.Id, m.Data.([]byte)) //set to client
		return nil
	}

	socketId := This.GetGateClientId(m.Id)
	if socketId == 0 {
		llog.Warningf("0.sendClient gate has been lost, userid = %d", m.Id)
		return nil
	}
	//llog.Debugf("sendClient userid = %d, uid = %d", m.Id, uid)
	msg := &msg.LouMiaoNetMsg{ClientId: int64(m.Id), Buffer: m.Data.([]byte)} //m.Id should be client`s userid
	buff, _ := message.Encode(0, "LouMiaoNetMsg", msg)

	This.pInnerService.SendById(socketId, buff)

	return nil
}

func sendMulClient(igo gorpc.IGoRoutine, data interface{}) interface{} {
	m := data.(*gorpc.M)
	if config.NET_NODE_TYPE == config.ServerType_Gate {
		llog.Error("0.sendMulClient gate can not send client")
		return nil
	}
	ms := m.Data.(*gorpc.MS)
	if len(ms.Ids) == 0 {
		broadCastClients(m.Data.([]byte))
	} else {
		for _, v := range ms.Ids {
			socketId := This.GetGateClientId(v) // get gate's socketid by client's userid
			if socketId == 0 {
				llog.Infof("1.sendMulClient gate has been lost, userid = %d", v)
				return nil
			}
			msg := &msg.LouMiaoNetMsg{ClientId: int64(v), Buffer: m.Data.([]byte)} //m.Id should be client`s userid
			buff, _ := message.Encode(0, "LouMiaoNetMsg", msg)
			This.pInnerService.SendById(socketId, buff)
		}
	}

	return nil
}

//new client added
func newClient(igo gorpc.IGoRoutine, data interface{}) interface{} {
	m := data.(*gorpc.M)
	llog.Debugf("GateServer newRpc: uid=%d,param=%d", m.Id, m.Param)
	uid := m.Id

	rpcclient, _ := This.clients[uid]
	client := m.Data.(*network.ClientSocket)
	if rpcclient != nil {
		llog.Errorf("GateServer newRpc: server[%d] has already connected", uid)
		client.SetClientId(0)
		client.Stop()
		return nil
	}
	This.clients[uid] = client
	if m.Param == 1 { //login server
		req := &msg.LouMiaoLoginGate{TokenId: int64(uid), UserId: int64(config.SERVER_NODE_UID)}
		buff, _ := message.Encode(uid, "LouMiaoLoginGate", req)
		client.Send(buff)
	} else { //register rpc
		This.rpcUids.Store(uid, true)
		This.rpcGates = append(This.rpcGates, uid)
		//register rpc if has
		req := &msg.LouMiaoRpcRegister{Uid: int32(config.SERVER_NODE_UID)}
		for key, _ := range This.rpcMap {
			req.FuncName = append(req.FuncName, key)
		}
		if len(req.FuncName) > 0 {
			buff, _ := message.Encode(uid, "LouMiaoRpcRegister", req)
			client.Send(buff)
		}
	}
	return nil
}

//send rpc msg
func sendRpc(igo gorpc.IGoRoutine, data interface{}) interface{} {
	m := data.(*gorpc.M)
	clientuid := This.getCluserRpcGateUid()
	if clientuid <= 0 {
		llog.Errorf("0.sendRpc no rpc gate server founded %s", m.Name)
		return nil
	}
	//llog.Debugf("sendRpc: %d", clientuid)
	client := This.GetRpcClient(clientuid)
	outdata := &msg.LouMiaoRpcMsg{TargetId: int32(m.Id), FuncName: m.Name, Buffer: m.Data.([]byte), SourceId: int32(config.SERVER_NODE_UID), Flag: int32(m.Param)}
	buff, _ := message.Encode(0, "LouMiaoRpcMsg", outdata)
	client.Send(buff)

	return nil
}

//server send msg to gate server
func sendGate(igo gorpc.IGoRoutine, data interface{}) interface{} {
	if This.ServerType == network.CLIENT_CONNECT {
		llog.Error("0.sendGate gate or account can not call this function")
		return nil
	}
	m := data.(*gorpc.M)
	uid := m.Id
	var clientid = This.GetClientId(uid)
	//llog.Debugf("sendGate: %d", clientid)
	if clientid <= 0 {
		llog.Warningf("1.sendGate no gate server founded: uid = %d", uid)
		return nil
	}
	This.pInnerService.SendById(clientid, m.Data.([]byte))
	return nil
}

func recvPackMsgClient(igo gorpc.IGoRoutine, data interface{}) interface{} {
	m := data.(*gorpc.M)
	socketid := m.Id
	target := m.Param
	rebuff := m.Data.([]byte)

	//llog.Debugf("recvPackMsgClient: len = %d", len(rebuff))

	token := This.GetClientToken(socketid)
	if token == nil {
		llog.Warningf("1.recvPackMsgClient msg, client lost, socketid=%d, target=%d ", socketid, target)
		This.closeClient(socketid)
		return nil
	}

	//客户端不接受传uid，只接受server type
	//所以这里做一次转化，根据server type找到对应的uid
	targetUid := 0
	if target == config.ServerType_World {
		targetUid = token.WouldId
	} else if target == config.ServerType_Zone {
		targetUid = token.ZoneUid
	} else if target == config.ServerType_LOGINQUEUE {
		targetUid = This.QueueServerUid
	}

	rpcClient := This.GetRpcClient(targetUid)
	if rpcClient == nil {
		llog.Warningf("2.recvPackMsgClient msg, but server has lost, target=%d,targetuid=%d ", target, targetUid)
		node := nodemgr.GetNode(targetUid)
		if node == nil || node.Type == config.ServerType_World { //if the world lost,we should let the client relogin
			This.closeClient(socketid)
		}
		return nil
	}
	//message.ReplacePakcetTarget(int32(targetUid), rebuff)不替换了，只要目标服务器不校验就行
	msg := &msg.LouMiaoNetMsg{ClientId: int64(token.UserId), Buffer: rebuff}
	buff, newlen := message.Encode(targetUid, "LouMiaoNetMsg", msg)
	rpcClient.Send(buff[0:newlen])

	return nil
}

func closeServer(igo gorpc.IGoRoutine, data interface{}) interface{} {
	uid := data.(int)
	llog.Infof("closeServer: %d", uid)
	nodemgr.DisableNode(uid)
	This.rpcGates = util.RemoveSlice(This.rpcGates, uid) //删除rpc负载，如果是rpcserver被关闭的话
	This.rpcUids.Delete(uid)
	return nil
}

func bindGate(igo gorpc.IGoRoutine, data interface{}) interface{} {
	req := data.(*msg.LouMiaoBindGate)
	This.users_u[int(req.UserId)] = int(req.Uid)
	return nil
}

//gate send msg to all clients or
//server send msg to all gates
func broadCastClients(buff []byte) {
	llog.Debugf("GateServer broadCastClients: %d", config.SERVER_NODE_UID)
	if config.NET_NODE_TYPE == config.ServerType_Gate {
		This.pService.BroadCast(buff)
	} else {
		This.pInnerService.BroadCast(buff)
	}
}

func publish(igo gorpc.IGoRoutine, data interface{}) interface{} {
	mm := data.(*gorpc.MM)
	vec, ok := This.msgQueueMap[mm.Id]
	if !ok {
		return nil
	}
	it := vec.Iterator()
	for it.Next() {
		loumiao.SendAcotr(it.Key().(string), it.Value().(string), mm.Data)
	}
	return nil
}

func subscribe(igo gorpc.IGoRoutine, data interface{}) interface{} {
	buffer := data.([]byte)
	bitstream := base.NewBitStream(buffer, len(buffer))
	name := bitstream.ReadString()
	key := bitstream.ReadString()
	hanlder := bitstream.ReadString()
	if len(hanlder) == 0 {
		if vec, ok := This.msgQueueMap[key]; ok {
			vec.Remove(name)
		}
	} else {
		vec, ok := This.msgQueueMap[key]
		if !ok {
			vec = maps.NewWithStringComparator()
			This.msgQueueMap[key] = vec
		}
		vec.Put(name, hanlder)
	}
	return nil
}
