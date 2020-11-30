package gate

import (
	"fmt"

	"github.com/snowyyj001/loumiao/base"

	"github.com/snowyyj001/loumiao/nodemgr"

	"github.com/snowyyj001/loumiao"

	"github.com/snowyyj001/loumiao/config"
	"github.com/snowyyj001/loumiao/define"
	"github.com/snowyyj001/loumiao/gorpc"
	"github.com/snowyyj001/loumiao/log"
	"github.com/snowyyj001/loumiao/message"
	"github.com/snowyyj001/loumiao/msg"
	"github.com/snowyyj001/loumiao/network"
	"github.com/snowyyj001/loumiao/util"
)

//client connect
func innerConnect(igo gorpc.IGoRoutine, socketId int, data interface{}) {
	log.Debugf("GateServer innerConnect: %d", socketId)
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
	log.Debugf("GateServer innerDisConnect: %d", socketId)

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

	}
}

//server connect
func outerConnect(igo gorpc.IGoRoutine, socketId int, data interface{}) {
	log.Debugf("GateServer outerConnect: %d", socketId)
}

//server disconnect
func outerDisConnect(igo gorpc.IGoRoutine, socketId int, data interface{}) {
	log.Debugf("GateServer outerDisConnect: %d", socketId)

	//if config.NET_NODE_TYPE == config.ServerType_Gate {
	This.removeRpc(socketId)
	//}
}

//client connected to server
func onClientConnected(uid int, tid int) {
	log.Debugf("GateServer onClientConnected: uid=%d,tid=%d", uid, tid)

	if This.ServerType == network.CLIENT_CONNECT {
		if config.NET_NODE_TYPE == config.ServerType_Gate { //tell world that a client has connected to this gate
			req := &msg.LouMiaoClientConnect{ClientId: int64(uid), GateId: int64(This.Id), State: define.CLIENT_CONNECT}
			buff, _ := message.Encode(0, 0, "LouMiaoClientConnect", req)
			This.SendServer(tid, buff)
			This.OnlineNum++
		}
	} else {
		//register rpc if has
		req := &msg.LouMiaoRpcRegister{}
		for key, _ := range This.rpcMap {
			req.FuncName = append(req.FuncName, key)
		}
		if len(req.FuncName) > 0 {
			buff, _ := message.Encode(uid, 0, "LouMiaoRpcRegister", req)
			clientid, _ := This.tokens_u[uid]
			This.pInnerService.SendById(clientid, buff)
		}
	}

}

//client disconnected to gate
func onClientDisConnected(uid int, tid int) {
	log.Debugf("GateServer onClientDisConnected: uid=%d,tid=%d", uid, tid)

	if config.NET_NODE_TYPE == config.ServerType_Gate {
		req := &msg.LouMiaoClientConnect{ClientId: int64(uid), GateId: int64(This.Id), State: define.CLIENT_DISCONNECT}
		buff, _ := message.Encode(0, 0, "LouMiaoClientConnect", req)
		This.SendServer(tid, buff)
		This.OnlineNum--
	}
}

//gate租约回调
func leaseCallBack(success bool) {
	if success { //成功续租
		str := fmt.Sprintf("%s%s/%s", define.ETCD_NODESTATUS, config.SERVER_GROUP, config.NET_GATE_SADDR)
		This.clientEtcd.Put(str, util.Itoa(This.OnlineNum), true) //写入人数
		//log.Debugf("leaseCallBack %s", str)
	} else {
		log.Errorf("leaseCallBack续租失败")
		err := This.clientEtcd.SetLease(int64(config.GAME_LEASE_TIME), true)
		if err != nil {
			log.Debugf("尝试重新续租失败")
		} else {
			log.Debugf("尝试重新续租成功")
			if This.ServerType != network.CLIENT_CONNECT {
				err := This.clientEtcd.PutService(fmt.Sprintf("%s%/%d", define.ETCD_SADDR, config.SERVER_GROUP, This.Id), config.NET_GATE_SADDR)
				if err != nil {
					log.Fatalf("leaseCallBack PutService error: %v", err)
				}
			}
		}
	}
}

//login gate
func innerLouMiaoLoginGate(igo gorpc.IGoRoutine, socketId int, data interface{}) {
	m := data.(*msg.LouMiaoLoginGate)
	userid := int(m.UserId)
	log.Debugf("innerLouMiaoLoginGate: %v, socketId=%d", m, socketId)

	if This.OnlineNum > config.NET_MAX_NUMBER {
		log.Warningf("0.innerLouMiaoLoginGate too many connections: max=%d, now=%d", config.NET_MAX_NUMBER, This.OnlineNum)
		This.closeClient(socketId)
		return
	}

	old_socketid, ok := This.tokens_u[userid]
	if ok { //close the old connection
		req := &msg.LouMiaoKickOut{}
		if config.NET_NODE_TYPE == config.ServerType_Gate {
			buff, _ := message.Encode(0, 0, "LouMiaoKickOut", req) //顶号,通知老的客户端退出登录
			This.pService.SendById(old_socketid, buff)
			//关闭老的客户端的socket
			innerDisConnect(igo, old_socketid, nil) //时序异步问题，这里直接关闭，不等socket的DISCONNECT消息
			if config.NET_WEBSOCKET {
				This.pService.(*network.WebSocket).StopClient(old_socketid)
			} else {
				This.pService.(*network.ServerSocket).StopClient(old_socketid)
			}
		} else {
			buff, _ := message.Encode(userid, 0, "LouMiaoKickOut", req)
			This.pInnerService.SendById(old_socketid, buff)
			//关闭老的gate的socket
			innerDisConnect(igo, old_socketid, nil) //时序异步问题，这里直接关闭，不等socket的DISCONNECT消息
			This.pInnerService.(*network.ServerSocket).StopClient(old_socketid)
		}
		delete(This.tokens, old_socketid)
	}
	tokenid := int(m.TokenId)
	worldid := int(m.WorldUid)
	if config.NET_NODE_TYPE == config.ServerType_Gate {
		rpcclient := This.GetRpcClient(worldid)
		if rpcclient == nil { //accout分配的world，在gate这里不存在，有可能是刚好world关闭了，这种情况就让客户端重新登录吧
			m.WorldUid = 0
			buff, _ := message.Encode(0, 0, "LouMiaoLoginGate", m)
			This.pService.SendById(socketId, buff)
			return
		}
		This.users_u[userid] = worldid
	}
	This.tokens[socketId] = &Token{TokenId: tokenid, UserId: userid}
	This.tokens_u[userid] = socketId
	onClientConnected(userid, worldid)
	if config.NET_NODE_TYPE == config.ServerType_Gate { //tell the client login success
		buff, _ := message.Encode(0, 0, "LouMiaoLoginGate", m)
		This.pService.SendById(socketId, buff)
	}
}

//rpc register
func innerLouMiaoRpcRegister(igo gorpc.IGoRoutine, socketId int, data interface{}) {
	req := data.(*msg.LouMiaoRpcRegister)
	log.Debugf("innerLouMiaoRpcRegister: %v", req)

	client := This.GetRpcClient(socketId)
	if client == nil {
		log.Warningf("0.innerLouMiaoRpcRegister server has lost[%d] ", socketId)
		return
	}
	for _, key := range req.FuncName {
		if This.rpcMap[key] == nil {
			This.rpcMap[key] = []int{}
		}
		This.rpcMap[key] = append(This.rpcMap[key], socketId)
		//rpcstr, _ := base64.StdEncoding.DecodeString(key)
		//log.Debugf("rpc register: funcname=%s, uid=%d", string(rpcstr), socketId)
	}
}

//recv rpc msg
func innerLouMiaoRpcMsg(igo gorpc.IGoRoutine, socketId int, data interface{}) {
	req := data.(*msg.LouMiaoRpcMsg)
	log.Debugf("innerLouMiaoRpcMsg=%s, socurce=%d, target=%d, ByteBuffer=%d", req.FuncName, req.SourceId, req.TargetId, req.ByteBuffer)
	if config.NET_NODE_TYPE == config.ServerType_Gate { //server -> gate
		target := int(req.TargetId)
		var rpcClient *network.ClientSocket
		if target <= 0 {
			rpcClient = This.getCluserServer(req.FuncName)
		} else {
			rpcClient = This.GetRpcClient(target)
		}
		if rpcClient == nil {
			log.Warningf("1.innerLouMiaoRpcMsg rpc client error %s %d", req.FuncName, target)
			return
		}
		outdata := &msg.LouMiaoRpcMsg{TargetId: req.TargetId, FuncName: req.FuncName, Buffer: req.Buffer, SourceId: req.SourceId, ByteBuffer: req.ByteBuffer}
		buff, _ := message.Encode(target, 0, "LouMiaoRpcMsg", outdata)
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
				//log.Debugf("innerLouMiaoRpcMsg : pm = %v", pm)
				//spm := reflect.Indirect(reflect.ValueOf(pm))
				//log.Debugf("innerLouMiaoRpcMsg : spm = %v, typename=%v", spm, reflect.TypeOf(spm.Interface()).Name
				if err != nil {
					log.Warningf("2.innerLouMiaoRpcMsg decode msg error : func=%s, error=%s ", req.FuncName, err.Error())
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
			log.Warningf("3.InnerLouMiaoRpcMsg no rpc hanlder %s, %d", req.FuncName, socketId)
		}
	}
}

//recv net msg
func innerLouMiaoNetMsg(igo gorpc.IGoRoutine, socketId int, data interface{}) {
	req := data.(*msg.LouMiaoNetMsg) //after decode LouMiaoNetMsg msg, post Buffer to next
	//log.Debugf("innerLouMiaoNetMsg %v", data)
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
	//log.Debugf("innerLouMiaoKickOut %v", data)

	if config.NET_NODE_TYPE != config.ServerType_Gate {
		log.Error("0.innerLouMiaoKickOut wrong msg")
		return
	}
	log.Noticef("innerLouMiaoKickOut: another gate server has already listened the saddr : %s, now close self[%d]", config.NET_GATE_SADDR, This.Id)
	loumiao.Stop()
}

//client be in connected or disconnect a gate
func innerLouMiaoClientConnect(igo gorpc.IGoRoutine, socketId int, data interface{}) {
	log.Debugf("innerLouMiaoClientConnect %v", data)
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
	log.Debugf("innerLouMiaoLouMiaoBindGate %v", data)
	req := data.(*msg.LouMiaoBindGate)
	This.users_u[int(req.UserId)] = int(req.Uid)
}

func registerNet(igo gorpc.IGoRoutine, data interface{}) interface{} {
	m := data.(gorpc.M)
	sname, ok := handler_Map[m.Name]
	if ok {
		log.Errorf("registerNet %s has already been registered: %s", m.Name, sname)
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
		log.Error("0.sendClient gate can not send client")
		return nil
	}
	if config.NET_NODE_TYPE == config.ServerType_Account {
		This.pService.SendById(m.Id, m.Data.([]byte)) //set to client
		return nil
	}

	uid, _ := This.users_u[m.Id] // get gate's uid by client's userid
	if uid == 0 {
		log.Noticef("1.sendClient cannot find which gate the client belong: %d", m.Id)
		return nil
	}
	socketId, _ := This.tokens_u[uid]
	if socketId == 0 {
		log.Noticef("2.sendClient gate has been shut down, uid = %d", m.Id)
		return nil
	}

	//log.Debugf("sendClient userid = %d, uid = %d", m.Id, uid)
	msg := &msg.LouMiaoNetMsg{ClientId: int64(m.Id), Buffer: m.Data.([]byte)} //m.Id should be client`s userid
	buff, _ := message.Encode(0, 0, "LouMiaoNetMsg", msg)

	This.pInnerService.SendById(socketId, buff)

	return nil
}

func sendMulClient(igo gorpc.IGoRoutine, data interface{}) interface{} {
	m := data.(gorpc.MS)
	if config.NET_NODE_TYPE == config.ServerType_Gate {
		log.Error("0.sendMulClient gate can not send client")
		return nil
	}
	for _, v := range m.Ids {
		uid, _ := This.users_u[v] // get gate's socketid by client's userid
		socketId, ok := This.tokens_u[uid]
		if ok {
			log.Noticef("1.sendMulClient gate has been shut down, uid = %d", v)
			return nil
		}
		msg := &msg.LouMiaoNetMsg{ClientId: int64(v), Buffer: m.Data.([]byte)} //m.Id should be client`s userid
		buff, _ := message.Encode(0, 0, "LouMiaoNetMsg", msg)

		This.pInnerService.SendById(socketId, buff)
	}

	return nil
}

//new rpc added
func newRpc(igo gorpc.IGoRoutine, data interface{}) interface{} {
	m := data.(gorpc.M)
	log.Debugf("GateServer newRpc: %d", m.Id)
	uid := m.Id

	rpcclient, _ := This.clients[uid]
	client := m.Data.(*network.ClientSocket)
	if rpcclient != nil {
		log.Warningf("GateServer newRpc: server[%d] has already connected", uid)
		client.SetClientId(0)
		client.Stop()
		return nil
	}
	This.clients[uid] = client

	if config.NET_NODE_TYPE == config.ServerType_Gate { //gate才需要向server登录，accoutn目前没有需求
		req := &msg.LouMiaoLoginGate{TokenId: int64(uid), UserId: int64(This.Id)}
		buff, _ := message.Encode(uid, 0, "LouMiaoLoginGate", req)
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
		log.Warningf("0.sendRpc no gate server finded %s", m.Name)
		return nil
	}
	//log.Debugf("sendRpc: %d", clientid)

	outdata := &msg.LouMiaoRpcMsg{TargetId: int64(m.Id), FuncName: m.Name, Buffer: m.Data.([]byte), SourceId: int64(This.Id), ByteBuffer: int32(m.Param)}
	buff, _ := message.Encode(0, 0, "LouMiaoRpcMsg", outdata)
	This.pInnerService.SendById(clientid, buff)

	return nil
}

//send msg to gate
func sendGate(igo gorpc.IGoRoutine, data interface{}) interface{} {
	if This.ServerType == network.CLIENT_CONNECT {
		log.Error("0.sendGate gate or account can not call this function")
		return nil
	}
	m := data.(gorpc.M)
	uid := m.Id
	var clientid int = 0

	//log.Debugf("sendGate: %d", clientid)
	if uid == 0 {
		clientid = This.getCluserGateClientId()
	} else {
		clientid, _ = This.tokens_u[uid]
	}
	if clientid <= 0 {
		log.Warning("1.sendGate no gate server finded")
		return nil
	}

	This.pInnerService.SendById(clientid, m.Data.([]byte))
	return nil
}

func recvPackMsg(igo gorpc.IGoRoutine, data interface{}) interface{} {
	m := data.(gorpc.MI)
	socketid := m.Id
	/*
		defer func() {
			if err := recover(); err != nil {
				log.Errorf("recvPackMsg self: %v", err)
				This.closeClient(socketid)
			}
		}()*/
	//log.Debugf("recvPackMsg[%d][%d]: %v", len(m.Data.([]byte)), m.Name, data)

	err, target, name, pm := message.Decode(This.Id, m.Data.([]byte), m.Name)
	//log.Debugf("recvPackMsg decode: %d, %s, %v", target, name, pm)

	if err != nil {
		log.Warningf("recvPackMsg Decode error: %s", err.Error())
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
					log.Warningf("recvPackMsg[%s] handler is nil: %s", name, This.Name)
				}
			} else {
				nm := gorpc.M{Id: socketid, Name: name, Data: pm}
				gorpc.MGR.Send(handler, "ServiceHandler", nm)
			}

		} else {
			if !filterWarning[name] {
				log.Warningf("recvPackMsg self handler is nil, drop it[%s]", name)
			} else {
				log.Debugf("recvPackMsg[%s]: %d", name, socketid)
			}
		}
	} else { //send to target server
		rpcClient := This.GetRpcClient(target)
		if rpcClient == nil {
			log.Debugf("1.recvPackMsg msg, but server has lost[%d][%d] ", target, m.Name)
			node := nodemgr.GetNode(target)
			if node == nil || node.Type == config.ServerType_World { //if the world lost,we should let the client relogin
				This.closeClient(socketid)
			}
			return nil
		}

		token, ok := This.tokens[socketid]
		if ok == false {
			log.Debugf("0.recvPackMsg recv client msg, but client has lost[%d][%d][%s][%d] ", target, socketid, name, m.Name)
			return nil
		}
		msg := &msg.LouMiaoNetMsg{ClientId: int64(token.UserId), Buffer: m.Data.([]byte)}
		buff, newlen := message.Encode(target, 0, "LouMiaoNetMsg", msg)
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
	log.Debugf("closeServer: %d", uid)

	if uid == This.Id {
		if This.ServerType == network.SERVER_CONNECT { //removed from server discover
			This.clientEtcd.DelService(fmt.Sprintf("%s%s/%d", define.ETCD_SADDR, config.SERVER_GROUP, This.Id))
		}
	}

	if This.ServerType == network.CLIENT_CONNECT {
		rpcClient := This.GetRpcClient(uid)
		if rpcClient != nil {
			rpcClient.SetState(network.SSF_SHUT_DOWNING) //mark it to will be closed
		}
	} else {
		socketId, _ := This.tokens_u[uid]
		token, _ := This.tokens[socketId]
		if token != nil {
			token.TokenId = 0 //mark it to will be closed
		}
	}
	return nil
}

//gate send msg to all clients or
//server send msg to all gates
func broadCastClients(buff []byte) {
	log.Debugf("GateServer broadCastClients: %d", This.Id)
	if config.NET_NODE_TYPE == config.ServerType_Gate {
		This.pService.BroadCast(buff)
	} else {
		This.pInnerService.BroadCast(buff)
	}
}

//gate send msg to all servers
func broadCastServers(buff []byte, serverType int) {
	log.Debugf("GateServer broadCastServers: uid=%d, servertype=%d", This.Id, serverType)
	if config.NET_NODE_TYPE != config.ServerType_Gate {
		log.Errorf("only gate can call this func broadCastServers, node type = %d", config.NET_NODE_TYPE)
		return
	}

	for _, client := range This.clients {
		if serverType > 0 {
			node := nodemgr.GetServerNode(client.GetSAddr())
			if node.Type == serverType {
				client.Send(buff)
			}
		} else {
			client.Send(buff)
		}
	}
}
