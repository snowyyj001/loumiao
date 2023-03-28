package lgate

import (
	"github.com/snowyyj001/loumiao/gorpc"
	"github.com/snowyyj001/loumiao/lconfig"
	"github.com/snowyyj001/loumiao/llog"
	"github.com/snowyyj001/loumiao/message"
	"github.com/snowyyj001/loumiao/network"
	"github.com/snowyyj001/loumiao/nodemgr"
	"github.com/snowyyj001/loumiao/pbmsg"
)

// BroadCastClients server 广播客户端消息，server使用
func BroadCastClients(data []byte) interface{} {
	if lconfig.NET_NODE_TYPE == lconfig.ServerType_Account || lconfig.NET_NODE_TYPE == lconfig.ServerType_Gate {
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
}

// BindClientGate 绑定client的gate, server使用
func BindClientGate(userId, gateUid int) {
	This.userGates.Store(userId, gateUid)
	if sid, ok := This.gatesSocketId.Load(gateUid); ok {
		This.userGateSockets.Store(userId, sid)
	}
}

// UnBindClientGate 接触绑定client所属的gate, server使用
func UnBindClientGate(userId, gateUid int) {
	This.userGates.Delete(userId)
	This.userGateSockets.Delete(userId)
}

// GetSocketNum 获取socket的连接数
func GetSocketNum() int {
	if lconfig.NET_WEBSOCKET {
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

func CloseServer(uid int) {
	llog.Infof("CloseServer: %d", uid)
	nodemgr.DisableNode(uid)
	This.rpcUids.Delete(uid)
}

// SendClient 发送消息给客户端
func SendClient(userId int, data []byte) {
	if lconfig.NET_NODE_TYPE == lconfig.ServerType_Account || lconfig.NET_NODE_TYPE == lconfig.ServerType_Gate {
		This.pService.SendById(userId, data)
		return
	}
	socketId := This.GetGateSocketIdByUserId(userId)
	if socketId == 0 {
		llog.Warningf("SendClient gate has been lost, userid = %d", userId)
		return
	}

	//llog.Debugf("sendClient userid = %d, uid = %d", m.Id, uid)
	msg := &pbmsg.LouMiaoNetMsg{ClientId: int64(userId), Buffer: data} //m.Id should be client`s userid
	buff, _ := message.EncodeProBuff(0, "LouMiaoNetMsg", msg)
	This.pInnerService.SendById(socketId, buff)
}

// SendRpc 发送rpc消息
func SendRpc(targetId int, rpcFunc string, buffer []byte, flag int) {
	clientuid := This.getClusterRpcGateUid()
	if clientuid <= 0 {
		llog.Errorf("0.sendRpc no rpc gate server founded %s", rpcFunc)
		return
	}
	//llog.Debugf("sendRpc: %d %s %d", clientuid, m.Name, m.Param)
	client := This.GetRpcClient(clientuid)
	outdata := &pbmsg.LouMiaoRpcMsg{TargetId: int32(targetId), FuncName: rpcFunc, Buffer: buffer, SourceId: int32(lconfig.SERVER_NODE_UID), Flag: int32(flag)}
	buff, _ := message.EncodeProBuff(0, "LouMiaoRpcMsg", outdata)
	client.Send(buff)
}
