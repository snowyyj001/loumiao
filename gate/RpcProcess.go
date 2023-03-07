package gate

import (
	"github.com/snowyyj001/loumiao/base"
	"github.com/snowyyj001/loumiao/base/maps"
	"github.com/snowyyj001/loumiao/config"
	"github.com/snowyyj001/loumiao/define"
	"github.com/snowyyj001/loumiao/gorpc"
	"github.com/snowyyj001/loumiao/llog"
	"github.com/snowyyj001/loumiao/message"
	"github.com/snowyyj001/loumiao/network"
	"github.com/snowyyj001/loumiao/nodemgr"
	"github.com/snowyyj001/loumiao/pbmsg"
	"github.com/snowyyj001/loumiao/util"
)

/*
对于onconnect， 不论是client -> gate/account，还是gate -> server
都是在login成功之后才算onconnect
对于ondisconnect，在socket断开时就算
所以bind信息时，在login成功之后，unbind时在socket的断开时
*/

// client connect
func innerConnect(igo gorpc.IGoRoutine, socketId int, data []byte) {
	llog.Debugf("GateServer innerConnect: %d", socketId)
}

// client disconnect
func innerDisConnect(igo gorpc.IGoRoutine, socketId int, data []byte) {
	llog.Debugf("GateServer innerDisConnect: %d", socketId)
	This.onSocketDisconnected(socketId)
}

// server connect
func outerConnect(igo gorpc.IGoRoutine, socketId int, data []byte) {
	//llog.Debugf("GateServer outerConnect: %d", socketId)
}

// server disconnect
func outerDisConnect(igo gorpc.IGoRoutine, socketId int, data []byte) {
	//	llog.Debugf("GateServer outerDisConnect: %d", socketId)

	//ClientSocket的socketId就是他连接的server的uid
	This.removeClient(socketId)
}

// client connected to server
func onClientConnected(userId, worldUid int) {
	llog.Debugf("GateServer onClientConnected: user id=%d, world uid=%d", userId, worldUid)
	//tell world that a client has login to this gate
	req := &pbmsg.LouMiaoClientConnect{ClientId: int64(userId), GateId: int32(config.SERVER_NODE_UID), State: define.CLIENT_CONNECT}
	buff, _ := message.EncodeProBuff(0, "LouMiaoClientConnect", req)
	This.SendServer(worldUid, buff)
}

// client disconnected to server
func onClientDisConnected(userId, worldUid int) {
	llog.Debugf("GateServer onClientDisConnected: uid=%d,tid=%d", userId, worldUid)
	//tell world that a client has disconnected with this gate
	req := &pbmsg.LouMiaoClientConnect{ClientId: int64(userId), GateId: int32(config.SERVER_NODE_UID), State: define.CLIENT_DISCONNECT}
	buff, _ := message.EncodeProBuff(0, "LouMiaoClientConnect", req)
	This.SendServer(worldUid, buff)
}

// login gate, gate -> server
func innerLouMiaoLoginGate(igo gorpc.IGoRoutine, socketId int, data []byte) {
	req := &pbmsg.LouMiaoLoginGate{}
	if message.UnPackProto(req, data) != nil {
		return
	}
	userid := int(req.UserId)
	llog.Debugf("innerLouMiaoLoginGate: %v, socketId=%d", req, socketId)

	if This.ServerType != network.SERVER_CONNECT {
		llog.Errorf("0.innerLouMiaoLoginGate only gate can send LouMiaoLoginGate to server for a login， gateuid=%d", userid)
		return
	}
	old_socketid := This.GetGateSocketId(userid)
	if old_socketid > 0 { //close the old connection，这种情况是，相同uid的gate启动了，但占用不同端口，算是人为错误，那么就关闭老的gate
		pack := &pbmsg.LouMiaoKickOut{}
		buff, _ := message.EncodeProBuff(userid, "LouMiaoKickOut", pack)
		This.pInnerService.SendById(old_socketid, buff)
		//关闭老的gate的socket
		innerDisConnect(igo, old_socketid, nil) //时序异步问题，这里直接关闭，不等socket的DISCONNECT消息
		This.pInnerService.(*network.ServerSocket).StopClient(old_socketid)
	}
	BindServerGate(socketId, userid, req.TokenId)
}

// recv rpc msg
func innerLouMiaoRpcMsg(igo gorpc.IGoRoutine, socketId int, data []byte) {
	req := &pbmsg.LouMiaoRpcMsg{}
	if message.UnPackProto(req, data) != nil {
		return
	}
	llog.Debugf("innerLouMiaoRpcMsg=%s, socurce=%d, target=%d, Flag=%d", req.FuncName, req.SourceId, req.TargetId, req.Flag)
	if util.HasBit(int(req.Flag), define.RPCMSG_FLAG_CALL) {
		if util.HasBit(int(req.Flag), define.RPCMSG_FLAG_RESP) {
			gorpc.MGR.SendActor("CallRpcServer", "RespRpcCall", req)
		} else {
			handler, ok := handlerMap[req.FuncName]
			if !ok {
				llog.Errorf("2.InnerLouMiaoRpcMsg no rpc handler %s, %d", req.FuncName, socketId)
				return
			}
			mm := &gorpc.MM{}
			mm.Id = handler
			mm.Data = req
			gorpc.MGR.SendActor("CallRpcServer", "ReqRpcCall", mm)
		}
		return
	}

	handler, ok := handlerMap[req.FuncName]
	if !ok {
		llog.Errorf("0.InnerLouMiaoRpcMsg no rpc handler %s, %d", req.FuncName, socketId)
		return
	}
	m := &gorpc.M{Id: int(req.SourceId), Name: req.FuncName, Data: req.Buffer}
	if handler == "GateServer" {
		igo.CallNetFunc(m)
	} else {
		gorpc.MGR.Send(handler, "ServiceHandler", m)
	}
}

// recv net msg
func innerLouMiaoNetMsg(igo gorpc.IGoRoutine, socketId int, data []byte) {
	req := &pbmsg.LouMiaoNetMsg{} //after decode LouMiaoNetMsg msg, post Buffer to next
	if message.UnPackProto(req, data) != nil {
		return
	}
	//llog.Debugf("innerLouMiaoNetMsg %v", req)
	userid := int(req.ClientId)

	if config.NET_NODE_TYPE == config.ServerType_Gate { //server -> gate, for msg to client
		if userid == -1 { //broad msg
			network.BoardSend(req.Buffer)
		} else {
			socketId := This.GetUserSocketId(userid)     // get client's socket id by userid
			This.pService.SendById(socketId, req.Buffer) //send to client
		}
	} else { // gate -> server
		target, name, buffbody, err := message.UnPackHead(req.Buffer, len(req.Buffer))
		if err != nil {
			llog.Errorf("innerLouMiaoNetMsg Decode error: err=%s, uid=%d,target=%d", err.Error(), config.SERVER_NODE_UID, target)
		} else {
			handler, ok := handlerMap[name]
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
				llog.Warningf("innerLouMiaoNetMsg handler is nil, drop it,userid = %d, msgname = %s, target = %d", userid, name, target)
			}
		}
	}
}

// exit the server
func innerLouMiaoKickOut(igo gorpc.IGoRoutine, socketId int, data []byte) {
	//llog.Debugf("innerLouMiaoKickOut %v", data)

	if config.NET_NODE_TYPE != config.ServerType_Gate {
		llog.Error("0.innerLouMiaoKickOut wrong msg")
		return
	}
	llog.Fatalf("innerLouMiaoKickOut: another gate server has already listened the saddr : %s, now close self[%d]，we have the same uid", config.NET_GATE_SADDR, config.SERVER_NODE_UID)
}

// client be in connected or disconnect a gate
func innerLouMiaoClientConnect(igo gorpc.IGoRoutine, socketId int, data []byte) {
	req := &pbmsg.LouMiaoClientConnect{}
	if message.UnPackProto(req, data) != nil {
		return
	}
	llog.Debugf("innerLouMiaoClientConnect %v", req)

	clientid := int(req.ClientId)
	gateId := int(req.GateId)
	if req.State == define.CLIENT_CONNECT {
		BindGate(clientid, gateId) //绑定client所属的gate
	} else {
		UnBindGate(clientid, gateId)                    //解除绑定client所属的gate
		igoTarget := gorpc.MGR.GetRoutine("GameServer") //告诉GameServer，client断开了，让GameServer决定如何处理，所以GameServer要注册这个actor
		if igoTarget != nil {
			igoTarget.SendActor("ON_DISCONNECT", clientid)
		} else {
			llog.Warningf("no GameServer actor to handler ON_DISCONNECT msg")
		}
	}
}

// new client added
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
		req := &pbmsg.LouMiaoLoginGate{TokenId: util.Itoa(uid), UserId: int64(config.SERVER_NODE_UID)}
		buff, _ := message.EncodeProBuff(uid, "LouMiaoLoginGate", req)
		client.Send(buff)
	} else { //register rpc
		This.rpcUids.Store(uid, true)
		This.rpcGates = append(This.rpcGates, uid)
		//register rpc if has
		req := &pbmsg.LouMiaoRpcRegister{Uid: int32(config.SERVER_NODE_UID)}
		for key, _ := range This.rpcMap {
			req.FuncName = append(req.FuncName, key)
		}
		if len(req.FuncName) > 0 {
			buff, _ := message.EncodeProBuff(uid, "LouMiaoRpcRegister", req)
			client.Send(buff)
		}

	}
	return nil
}

// send rpc msg
func sendRpc(igo gorpc.IGoRoutine, data interface{}) interface{} {
	m := data.(*gorpc.M)
	clientuid := This.getCluserRpcGateUid()
	if clientuid <= 0 {
		llog.Errorf("0.sendRpc no rpc gate server founded %s", m.Name)
		return nil
	}
	//llog.Debugf("sendRpc: %d %s %d", clientuid, m.Name, m.Param)
	client := This.GetRpcClient(clientuid)
	outdata := &pbmsg.LouMiaoRpcMsg{TargetId: int32(m.Id), FuncName: m.Name, Buffer: m.Data.([]byte), SourceId: int32(config.SERVER_NODE_UID), Flag: int32(m.Param)}
	buff, _ := message.EncodeProBuff(0, "LouMiaoRpcMsg", outdata)
	client.Send(buff)

	return nil
}

func recvPackMsgClient(igo gorpc.IGoRoutine, data interface{}) interface{} {
	m := data.(*gorpc.M)
	socketid := m.Id
	target := m.Param
	rebuff := m.Data.([]byte)

	//llog.Debugf("recvPackMsgClient: len = %d", len(rebuff))

	userAgent := This.GetClientAgent(socketid)
	if userAgent == nil {
		llog.Warningf("1.recvPackMsgClient msg, client lost, socketid=%d, target=%d ", socketid, target)
		This.closeClient(socketid)
		return nil
	}

	//客户端不接受传uid，只接受server type
	//所以这里做一次转化，根据server type找到对应的uid
	//publicserver是单节点的，这里就不对client开放了，有确实需要可以通过lobby来进行rpc转发
	targetUid := 0
	if target == config.ServerType_World {
		targetUid = int(userAgent.LobbyUid)
	} else if target == config.ServerType_Zone {
		targetUid = int(userAgent.ZoneUid)
	} else if target == config.ServerType_LOGINQUEUE {
		targetUid = This.QueueServerUid
	}

	rpcClient := This.GetRpcClient(targetUid)
	if rpcClient == nil {
		llog.Warningf("2.recvPackMsgClient msg, but server has lost, target type = %d,target uid = %d ", target, targetUid)
		if target == config.ServerType_World { //if the world lost,we should let the client relogin
			This.closeClient(socketid)
		}
		return nil
	}
	//message.ReplacePakcetTarget(int32(targetUid), rebuff)不替换了，只要目标服务器不校验就行
	msg := &pbmsg.LouMiaoNetMsg{ClientId: int64(userAgent.UserId), Buffer: rebuff}
	buff, newlen := message.EncodeProBuff(targetUid, "LouMiaoNetMsg", msg)
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

func publish(igo gorpc.IGoRoutine, data interface{}) interface{} {
	mm := data.(*gorpc.MM)
	vec, ok := This.msgQueueMap[mm.Id]
	if !ok {
		return nil
	}
	it := vec.Iterator()
	for it.Next() {
		//loumiao.SendActor(it.Key().(string), it.Value().(string), mm.Data)
		m := &gorpc.M{Data: mm.Data, Flag: true}
		gorpc.MGR.Send(it.Key().(string), it.Value().(string), m)
	}
	return nil
}

func subscribe(igo gorpc.IGoRoutine, data interface{}) interface{} {
	buffer := data.([]byte)
	bitstream := base.NewBitStream(buffer, len(buffer))
	name := bitstream.ReadString()
	key := bitstream.ReadString()
	handler := bitstream.ReadString()
	if len(handler) == 0 {
		if vec, ok := This.msgQueueMap[key]; ok {
			vec.Remove(name)
		}
	} else {
		vec, ok := This.msgQueueMap[key]
		if !ok {
			vec = maps.NewWithStringComparator()
			This.msgQueueMap[key] = vec
		}
		vec.Put(name, handler)
	}
	return nil
}
