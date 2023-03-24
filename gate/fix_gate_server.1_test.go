package gate

import (
	"github.com/snowyyj001/loumiao/config"
	"github.com/snowyyj001/loumiao/gorpc"
	"github.com/snowyyj001/loumiao/llog"
	"github.com/snowyyj001/loumiao/message"
	"github.com/snowyyj001/loumiao/network"
	"github.com/snowyyj001/loumiao/nodemgr"
	"github.com/snowyyj001/loumiao/pbmsg"
)

func innerConnect_hotfix(igo gorpc.IGoRoutine, socketId int, data []byte) {
	llog.Debugf("GateServer innerConnect_hotfix: %d", socketId)

	//对于account/gate来说，socket连接数就是在线人数
	//本质上来说差别不大
	//对于world来说，内存中玩家数就是在线数，包括暂时离线缓存的那一部分
	//对于zone来说，就是参与战斗的玩家数
	if This.ServerType == network.CLIENT_CONNECT { //对外(login,gate)
		nodemgr.OnlineNum++
	}
}

func sendRpc_hotfix(igo gorpc.IGoRoutine, data interface{}) interface{} {
	llog.Debugf("GateServer sendRpc_hotfix: %v", data)

	m := data.(*gorpc.M)
	clientuid := This.getClusterRpcGateUid()
	if clientuid <= 0 {
		llog.Errorf("0.sendRpc no rpc gate server founded %s", m.Name)
		return nil
	}
	//llog.Debugf("sendRpc: %d", clientuid)
	client := This.GetRpcClient(clientuid)
	outdata := &pbmsg.LouMiaoRpcMsg{TargetId: int32(m.Id), FuncName: m.Name, Buffer: m.Data.([]byte), SourceId: int32(config.SERVER_NODE_UID), Flag: int32(m.Param)}
	buff, _ := message.EncodeProBuff(0, "LouMiaoRpcMsg", outdata)
	client.Send(buff)

	return nil
}

// 开方给hotfix的调用接口
func Fix_bug_1() {
	igo := gorpc.MGR.GetRoutine("GateServer")

	//覆盖gate_server的innerConnect
	igo.RegisterGate("innerConnect", innerConnect_hotfix)

	//覆盖gate_server的SendRpc
	igo.Register("SendRpc", sendRpc_hotfix)
}
