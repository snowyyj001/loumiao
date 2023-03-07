package gate

import (
	"github.com/snowyyj001/loumiao/config"
	"github.com/snowyyj001/loumiao/gorpc"
	"github.com/snowyyj001/loumiao/llog"
	"github.com/snowyyj001/loumiao/message"
	"github.com/snowyyj001/loumiao/network"
	"github.com/snowyyj001/loumiao/pbmsg"
)

// HasRpcGate 是否有rpc网关可用
func HasRpcGate() bool {
	return len(This.rpcGates) > 0
}

// BroadCastClients server 广播客户端消息，server使用
func BroadCastClients(data []byte) interface{} {
	if config.NET_NODE_TYPE == config.ServerType_Account || config.NET_NODE_TYPE == config.ServerType_Gate {
		llog.Error("gate can not call this function BroadCastClients")
		return nil
	}
	msg := &pbmsg.LouMiaoNetMsg{ClientId: -1, Buffer: data} // -1 means all client
	buff, _ := message.EncodeProBuff(0, "LouMiaoNetMsg", msg)
	This.pInnerService.BroadCast(buff)
	return nil
}

// BindServerGate server绑定gate，server使用
func BindServerGate(socketId, gateUid int, tokenId string) {
	This.gatesUid.Store(socketId, gateUid)
	This.gatesSocketId.Store(gateUid, socketId)
	This.gatesSocketId.Store(gateUid, socketId)
}

// BindGate 绑定client的gate, server使用
func BindGate(userId, gateUid int) {
	This.userGates.Store(userId, gateUid)
	if sid, ok := This.gatesSocketId.Load(gateUid); ok {
		This.userGateSockets.Store(userId, sid)
	}
}

// UnBindGate 接触绑定client所属的gate, server使用
func UnBindGate(userId, gateUid int) {
	This.userGates.Delete(userId)
	This.gatesSocketId.Delete(gateUid)
	This.userGateSockets.Delete(userId)
}

// GetSocketNum 获取socket的连接数
func GetSocketNum() int {
	if config.NET_WEBSOCKET {
		return This.pService.(*network.WebSocket).GetClientNumber()
	} else {
		return This.pService.(*network.ServerSocket).GetClientNumber()
	}
}

// GetGateId 获取玩家gate
func GetGateId(userId int) int {
	if gateId, ok := This.userGates.Load(userId); ok {
		return gateId.(int)
	}
	return 0
}

// RegisterNet 注册网络消息
func RegisterNet(igo gorpc.IGoRoutine, funcName string, rpc bool) interface{} {
	sname, ok := handlerMap[funcName]
	if ok {
		llog.Errorf("registerNet %s has already been registered: has = %s, now = %s", sname, funcName)
		return nil
	}
	//llog.Debugf("registerNet: name = %s, id = %d", m.Name, m.Id)
	handlerMap[funcName] = igo.GetName()
	if rpc { //rpc register
		This.rpcMap[funcName] = igo.GetName()
	}
	return nil
}

// SendClient 发送消息给客户端
func SendClient(userId int, data []byte) {
	if config.NET_NODE_TYPE == config.ServerType_Account || config.NET_NODE_TYPE == config.ServerType_Gate {
		This.pService.SendById(userId, data)
		return
	}
	socketId := This.GetGateSocketId(userId)
	if socketId == 0 {
		llog.Warningf("SendClient gate has been lost, userid = %d", userId)
		return
	}

	//llog.Debugf("sendClient userid = %d, uid = %d", m.Id, uid)
	msg := &pbmsg.LouMiaoNetMsg{ClientId: int64(userId), Buffer: data} //m.Id should be client`s userid
	buff, _ := message.EncodeProBuff(0, "LouMiaoNetMsg", msg)
	This.pInnerService.SendById(socketId, buff)
}
