package gate

import (
	"github.com/snowyyj001/loumiao/config"
	"github.com/snowyyj001/loumiao/gorpc"
	"github.com/snowyyj001/loumiao/log"
	"github.com/snowyyj001/loumiao/message"
	"github.com/snowyyj001/loumiao/network"
)

//client connect
func InnerConnect(igo gorpc.IGoRoutine, socketId int, data interface{}) interface{} {
	return nil
}

//client disconnect
func InnerDisConnect(igo gorpc.IGoRoutine, socketId int, data interface{}) interface{} {
	token, ok := This.tokens[socketId]
	if ok == nil {
		delete(This.tokens, socketId)
		delete(This.tokens_u, token.UserId)
		
		if This.OnClientDisConnected != nil {
			This.OnClientDisConnected(token.UserId)
		}
	}
	return nil
}

//login gate
func InnerLouMiaoLoginGate(igo gorpc.IGoRoutine, socketId int, data interface{}) interface{} {
	m := data.(*LouMiaoLoginGate)
	
	old_socketid, ok := This.tokens_u[m.UserId]
	if ok { //close the old connection
		req := &LouMiaoKickOut{}
		if config.NET_NODE_TYPE == config.ServerType_Gate {
			buff, _ := message.Encode(-1, 0, "LouMiaoKickOut", req)
			This.pService.SendById(old_socketid, buff)
			if config.NET_WEBSOCKET {
				This.pService.(*network.WebSocket).StopClient(old_socketid)
			} else {
				This.pService.(*network.ServerSocket).StopClient(old_socketid)
			}
		} else {
			buff, _ := message.Encode(m.UserId, This.Id, "LouMiaoKickOut", req)
			This.pInnerService.SendById(old_socketid, buff)
			This.pInnerService.(*network.ServerSocket).StopClient(old_socketid)
		}
	}
	This.tokens[socketId] = &Token{TokenId: m.TokenId, UserId: m.UserId}
	This.tokens_u[m.UserId] = socketId

	if This.OnClientConnected != nil {
		This.OnClientConnected(m.UserId)
	}
	return nil
}

//recv rpc msg
func InnerLouMiaoRpcMsg(igo gorpc.IGoRoutine, socketId int, data interface{}) interface{} {
	req := data.(*LouMiaoRpcMsg)
	handler, ok := handler_Map[req.FuncName]
	if ok {
		This.Send(handler, "ServiceHandler", m)
	} else {
		log.Warningf("0.InnerLouMiaoRpcMsg no rpc hanlder %s, %d", req.FuncName, socketId)
	}

	return nil
}

//recv net msg
func InnerLouMiaoNetMsg(igo gorpc.IGoRoutine, socketId int, data interface{}) interface{} {
	req := data.(*LouMiaoNetMsg) //after decode LouMiaoNetMsg msg, post Buffer to next

	if config.NET_NODE_TYPE == config.ServerType_Gate { //server -> gate
		socketId, _ := This.tokens_u[req.ClientId] // get client's socketid by userid
		This.pService.SendById(socketId, buff)     //set to client
	} else { //gate -> server
		This.users_u[req.ClientId] = socketId
		PacketFunc(req.ClientId, req.Buffer, len(req.Buffer)) //post msg to server service
	}
	return nil
}

func RegisterNet(igo gorpc.IGoRoutine, data interface{}) interface{} {
	m := data.(gorpc.M)
	handler_Map[m.Name] = m.Data.(string)
	if m.Id < 0 { //rpc register
		key := fmt.Sprintf("%s%s/%d", define.ETCD_RPCADDR, m.Name, This.Id)
		This.serverReg.Put(key, This.Id)
	}
	return nil
}

func UnRegisterNet(igo gorpc.IGoRoutine, data interface{}) interface{} {
	m := data.(gorpc.M)
	delete(handler_Map, m.Name)
	if m.Id < 0 { //rpc unregister
		This.serverReg.Delete(m.Name)
	}
	return nil
}

func SendClient(igo gorpc.IGoRoutine, data interface{}) interface{} {
	m := data.(gorpc.M)
	var buff []byte = m.Data.([]byte)
	if config.NET_NODE_TYPE == config.ServerType_Gate {
		log.Error("0.SendClient gate can not send client")
		return
	}
	socketId, _ := This.users_u[m.Id] // get gate's socketid by client's userid
	token, ok := This.tokens[socketId]
	if ok != nil {
		log.Noticef("1.SendClient gate has been shut down, uid = %d", m.Id)
		return
	}
	msg := LouMiaoNetMsg{ClientId: m.Id, Buffer: m.Data} //m.Id should be client`s userid
	buff, _ := message.Encode(token.UserId, This.Id, "LouMiaoNetMsg", msg)

	This.pInnerService.SendById(socketId, buff)

	return nil
}

func SendMulClient(igo gorpc.IGoRoutine, data interface{}) interface{} {
	m := data.(gorpc.MS)
	var buff []byte = m.Data.([]byte)
	if config.NET_NODE_TYPE == config.ServerType_Gate {
		log.Error("0.SendMulClient gate can not send client")
		return
	}
	for k, v range(m.Ids) {
		socketId, _ := This.users_u[v] // get gate's socketid by client's userid
		token, ok := This.tokens[socketId]
		if ok != nil {
			log.Noticef("1.SendMulClient gate has been shut down, uid = %d", v)
			return
		}
		msg := LouMiaoNetMsg{ClientId: v, Buffer: m.Data} //m.Id should be client`s userid
		buff, _ := message.Encode(token.UserId, This.Id, "LouMiaoNetMsg", msg)
	
		This.pInnerService.SendById(socketId, buff)
	}

	return nil
}

func SendRpc(igo gorpc.IGoRoutine, data interface{}) interface{} {
	m := data.(gorpc.M)

	arr := This.rpcMap[m.Name]
	sz := len(arr)
	if sz == 0 {
		log.Warningf("0.SendRpc no rpc server hanlder finded %s", m.Name)
		return nil
	}
	index := util.Random(0, sz) //随机一个server进行rpc调用
	target := arr[index]

	rpcClient := This.GetRpcClient(target)
	if rpcClient == nil {
		return false
	}
	indata := message.Encode(target, This.Id, "", m.Data)
	outdata := LouMiaoRpcMsg{FuncName: m.Name, Buffer: indata}

	buff, _ := message.Encode(target, 0, "LouMiaoRpcMsg", outdata)
	rpcClient.Send(buff[0:nlen])

	return nil
}
