package lgate

import (
	"github.com/snowyyj001/loumiao/gorpc"
	"github.com/snowyyj001/loumiao/lbase"
	"github.com/snowyyj001/loumiao/lbase/maps"
	"github.com/snowyyj001/loumiao/lconfig"
	"github.com/snowyyj001/loumiao/ldefine"
	"github.com/snowyyj001/loumiao/llog"
	"github.com/snowyyj001/loumiao/lutil"
	"github.com/snowyyj001/loumiao/message"
	"github.com/snowyyj001/loumiao/network"
	"github.com/snowyyj001/loumiao/pbmsg"
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
	req := &pbmsg.LouMiaoClientConnect{ClientId: int64(userId), GateId: int32(lconfig.SERVER_NODE_UID), State: ldefine.CLIENT_CONNECT}
	buff, _ := message.EncodeProBuff(0, "LouMiaoClientConnect", req)
	This.SendServer(worldUid, buff)
}

// client disconnected to server
func onClientDisConnected(userId, worldUid int) {
	llog.Debugf("GateServer onClientDisConnected: uid=%d,tid=%d", userId, worldUid)
	//tell world that a client has disconnected with this gate
	req := &pbmsg.LouMiaoClientConnect{ClientId: int64(userId), GateId: int32(lconfig.SERVER_NODE_UID), State: ldefine.CLIENT_DISCONNECT}
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
	llog.Infof("innerLouMiaoLoginGate: %v, socketId=%d, uid=%d", req, socketId, userid)

	if This.ServerType != network.SERVER_CONNECT {
		llog.Errorf("0.innerLouMiaoLoginGate only gate can send LouMiaoLoginGate to server for a login， gateuid=%d", userid)
		return
	}
	old_socketid := This.GetGateSocketIdByGateUid(userid)
	if old_socketid > 0 { //close the old connection，这种情况是，相同uid的gate启动了，但占用不同端口，算是人为错误，那么就关闭老的gate
		pack := &pbmsg.LouMiaoKickOut{}
		buff, _ := message.EncodeProBuff(userid, "LouMiaoKickOut", pack)
		This.pInnerService.SendById(old_socketid, buff)
		//关闭老的gate的socket
		innerDisConnect(igo, old_socketid, nil) //时序异步问题，这里直接关闭，不等socket的DISCONNECT消息
		This.pInnerService.(*network.ServerSocket).StopClient(old_socketid)
	}
	BindServerGate(socketId, userid, req.TokenId)

	//register net handler to gate
	pack := &pbmsg.LouMiaoNetRegister{Ty: int32(lconfig.NET_NODE_TYPE)}
	for f, name := range handlerMap {
		if name != "GateServer" && This.rpcMap[f] == "" {
			pack.FuncName = append(pack.FuncName, f)
		}
	}
	buff, _ := message.EncodeProBuff(userid, "LouMiaoNetRegister", pack)
	This.pInnerService.SendById(socketId, buff)
}

// recv rpc msg
func innerLouMiaoRpcMsg(igo gorpc.IGoRoutine, socketId int, data []byte) {
	req := &pbmsg.LouMiaoRpcMsg{}
	if message.UnPackProto(req, data) != nil {
		return
	}
	llog.Debugf("innerLouMiaoRpcMsg=%s, socurce=%d, target=%d, Flag=%d", req.FuncName, req.SourceId, req.TargetId, req.Flag)
	if lutil.HasBit(int(req.Flag), ldefine.RPCMSG_FLAG_CALL) {
		if lutil.HasBit(int(req.Flag), ldefine.RPCMSG_FLAG_RESP) {
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

	if lconfig.NET_NODE_TYPE == lconfig.ServerType_Gate { //server -> gate, for msg to client
		if userid == -1 { //broad msg
			network.BoardSend(req.Buffer)
		} else {
			socketId := This.GetUserSocketId(userid)     // get client's socket id by userid
			This.pService.SendById(socketId, req.Buffer) //send to client
		}
	} else { // gate -> server
		target, name, buffbody, err := message.UnPackHead(req.Buffer, len(req.Buffer))
		if err != nil {
			llog.Errorf("innerLouMiaoNetMsg Decode error: err=%s, uid=%d,target=%d", err.Error(), lconfig.SERVER_NODE_UID, target)
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

	if lconfig.NET_NODE_TYPE != lconfig.ServerType_Gate {
		llog.Error("0.innerLouMiaoKickOut wrong msg")
		return
	}
	llog.Fatalf("innerLouMiaoKickOut: another gate server has already listened the saddr : %s, now close self[%d]，we have the same uid", lconfig.NET_GATE_SADDR, lconfig.SERVER_NODE_UID)
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
	if req.State == ldefine.CLIENT_CONNECT {
		BindClientGate(clientid, gateId) //绑定client所属的gate
	} else {
		UnBindClientGate(clientid, gateId)              //解除绑定client所属的gate
		igoTarget := gorpc.MGR.GetRoutine("GameServer") //告诉GameServer，client断开了，让GameServer决定如何处理，所以GameServer要注册这个actor
		if igoTarget != nil {
			igoTarget.SendActor("ON_DISCONNECT", clientid)
		} else {
			llog.Warningf("no GameServer actor to handler ON_DISCONNECT msg")
		}
	}
}

// net register
func innerLouMiaoNetRegister(igo gorpc.IGoRoutine, socketId int, data []byte) {
	req := &pbmsg.LouMiaoNetRegister{}
	if message.UnPackProto(req, data) != nil {
		return
	}
	llog.Infof("innerLouMiaoNetRegister: %v", req)
	for _, v := range req.FuncName {
		if ve, ok := This.netMap.LoadOrStore(v, int(req.Ty)); ok {
			if ve.(int) != int(req.Ty) {
				llog.Warningf("innerLouMiaoNetRegister: same handler msg:%s, last type:%d, new type:%d", v, ve.(int), req.Ty)
			}
		}
		This.netMap.Store(v, int(req.Ty))
	}
}

// new gate client added
func newGateClient(uid int, client *network.ClientSocket) {
	llog.Infof("GateServer newGateClient: uid=%d, client=%d, %s", uid, client.Uid, client.GetSAddr())

	if rpcClient, ok := This.clients.Load(uid); ok {
		if rpcClient != nil {
			llog.Errorf("GateServer newGateClient: server[%d] has already connected", uid)
			client.SetClientId(0)
			client.Stop()
			return
		}
	}
	This.clients.Store(uid, client)

	req := &pbmsg.LouMiaoLoginGate{TokenId: lutil.Itoa(uid), UserId: int64(lconfig.SERVER_NODE_UID)}
	buff, _ := message.EncodeProBuff(uid, "LouMiaoLoginGate", req)
	client.Send(buff)
}

// new rpc client added
func newRpcClient(uid int, client *network.ClientSocket) {
	llog.Infof("GateServer newRpcClient: uid=%d, client=%d, %s", uid, client.Uid, client.GetSAddr())

	if rpcClient, ok := This.clients.Load(uid); ok {
		if rpcClient != nil {
			llog.Errorf("GateServer newRpcClient: server[%d] has already connected", uid)
			client.SetClientId(0)
			client.Stop()
			return
		}
	}
	This.clients.Store(uid, client)

	//register rpc
	This.rpcUids.Store(uid, true)
	//register rpc if it has
	req := &pbmsg.LouMiaoRpcRegister{Uid: int32(lconfig.SERVER_NODE_UID)}
	for key := range This.rpcMap {
		req.FuncName = append(req.FuncName, key)
	}
	if len(req.FuncName) > 0 {
		buff, _ := message.EncodeProBuff(uid, "LouMiaoRpcRegister", req)
		client.Send(buff)
	}
}

func recvPackMsgClient(socketId, targetType int, buffer []byte) {
	//llog.Debugf("recvPackMsgClient: len = %d", len(rebuff))
	userAgent := This.GetClientAgent(socketId)
	if userAgent == nil {
		llog.Warningf("1.recvPackMsgClient msg, client lost, socketid=%d, target=%d ", socketId, targetType)
		This.closeClient(socketId)
		return
	}

	targetUid := 0
	if targetType == lconfig.ServerType_World {
		targetUid = int(userAgent.LobbyUid)
	} else if targetType == lconfig.ServerType_Zone {
		targetUid = int(userAgent.ZoneUid)
	} else if targetType == lconfig.ServerType_LOGINQUEUE {
		targetUid = This.QueueServerUid
	} else {
		llog.Warningf("3.recvPackMsgClient msg, wrong target server type, socketId=%d, target type = %d ", socketId, targetType)
		return
	}

	rpcClient := This.GetRpcClient(targetUid)
	if rpcClient == nil {
		llog.Warningf("2.recvPackMsgClient msg, but server has lost, target type = %d,target uid = %d ", targetType, targetUid)
		if targetType == lconfig.ServerType_World { //if the world lost,we should let the client relogin
			This.closeClient(socketId)
		}
		return
	}
	//message.ReplacePakcetTarget(int32(targetUid), rebuff)不替换了，只要目标服务器不校验就行
	msg := &pbmsg.LouMiaoNetMsg{ClientId: int64(userAgent.UserId), Buffer: buffer}
	buff, newlen := message.EncodeProBuff(targetUid, "LouMiaoNetMsg", msg)
	rpcClient.Send(buff[0:newlen])
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
	bitstream := lbase.NewBitStream(buffer, len(buffer))
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
