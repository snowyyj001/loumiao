package rpcgate

import (
	"github.com/snowyyj001/loumiao/define"
	"github.com/snowyyj001/loumiao/gorpc"
	"github.com/snowyyj001/loumiao/llog"
	"github.com/snowyyj001/loumiao/message"
	"github.com/snowyyj001/loumiao/msg"
	"github.com/snowyyj001/loumiao/nodemgr"
	"github.com/snowyyj001/loumiao/util"
)

//client connect
func innerConnect(igo gorpc.IGoRoutine, socketId int, data []byte) {
	llog.Debugf("RpcGateServer innerConnect: %d", socketId)
	nodemgr.OnlineNum++
}

//client disconnect
func innerDisConnect(igo gorpc.IGoRoutine, socketId int, data []byte) {
	llog.Debugf("RpcGateServer innerDisConnect: %d", socketId)
	nodemgr.OnlineNum--
	This.removeRpc(socketId)

}

//rpc register
func innerLouMiaoRpcRegister(igo gorpc.IGoRoutine, socketId int, data []byte) {
	req := &msg.LouMiaoRpcRegister{}
	if message.UnPackProto(req, data) != nil {
		return
	}
	llog.Debugf("innerLouMiaoRpcRegister: %v", req)

	if sid, ok := This.clients_u[int(req.Uid)]; ok {
		llog.Errorf("innerLouMiaoRpcRegister: rpc has already register: %d %d, ", req.Uid, sid, socketId)
		return
	}

	This.rpcUids.Store(req.Uid, true)
	This.clients_u[int(req.Uid)] = socketId
	This.users_u[socketId] = int(req.Uid)
	for _, key := range req.FuncName {
		if This.rpcMap[key] == nil {
			This.rpcMap[key] = []int{}
		}
		This.rpcMap[key] = append(This.rpcMap[key], socketId)
	}
}

//recv rpc msg
func innerLouMiaoRpcMsg(igo gorpc.IGoRoutine, socketId int, data []byte) {
	req := &msg.LouMiaoRpcMsg{}
	if message.UnPackProto(req, data) != nil {
		return
	}
	llog.Debugf("0.innerLouMiaoRpcMsg=%s, socurce=%d, target=%d, Flag=%d", req.FuncName, req.SourceId, req.TargetId, req.Flag)
	if util.HasBit(int(req.Flag), define.RPCMSG_FLAG_BROAD) {
		if req.TargetId == 0 {
			for _, sid := range This.clients_u {
				//This.pInnerService.SendById(sid, buff)
				This.pInnerService.SendById(sid, data)
			}
		} else {
			nodes := nodemgr.GetTypeServer(int(req.TargetId))
			for _, node := range nodes {
				rpcClientId := This.getClientId(node.Uid)
				This.pInnerService.SendById(rpcClientId, data)
			}
		}
		return
	}
	var rpcClientId int
	if req.TargetId <= 0 {
		rpcClientId = This.getCluserServerSocketId(req.FuncName)
	} else {
		rpcClientId = This.getClientId(int(req.TargetId))
	}
	if rpcClientId == 0 {
		llog.Errorf("1.innerLouMiaoRpcMsg no target error funcName=%s,sourceid=%d,target=%d", req.FuncName, req.SourceId, req.TargetId)
		return
	}
	buff, _ := message.EncodeProBuff(0, "LouMiaoRpcMsg", req)
	//fmt.Println("innerLouMiaoRpcMsg: ", rpcClientId, req.FuncName)
	This.pInnerService.SendById(rpcClientId, buff)
	//fmt.Println("innerLouMiaoRpcMsg: ", rpcClientId, req.FuncName, n)
}

func sendRpcMsgToServer(igo gorpc.IGoRoutine, data interface{}) interface{} {
	m := data.(*gorpc.M)
	target := m.Id
	buff := m.Data.([]byte)
	funcName := m.Name

	var rpcClientId int
	if target <= 0 {
		rpcClientId = This.getCluserServerSocketId(funcName)
	} else {
		rpcClientId = This.getClientId(target)
	}
	if rpcClientId == 0 {
		llog.Warningf("0.sendRpcMsgToServer target error funcName=%s,target=%d", funcName, target)
		return nil
	}
	This.pInnerService.SendById(rpcClientId, buff)
	return nil
}

func closeServer(igo gorpc.IGoRoutine, data interface{}) interface{} {
	uid := data.(int)
	nodemgr.DisableNode(uid)
	llog.Infof("closeServer: %d", uid)
	//仅仅不参与新的rpc负载调用，节点还是照样可用，节点的真实关闭依赖于socket的断开
	socketId := This.getClientId(uid)
	//remove rpc handler
	This.removeRpcHanlder(socketId)
	This.rpcUids.Delete(uid)
	return nil
}
