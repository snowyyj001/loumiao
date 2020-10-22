package gate

import (
	"encoding/base64"
	"fmt"

	"github.com/snowyyj001/loumiao/etcd"

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
func InnerConnect(igo gorpc.IGoRoutine, socketId int, data interface{}) interface{} {
	return nil
}

//client disconnect
func InnerDisConnect(igo gorpc.IGoRoutine, socketId int, data interface{}) interface{} {
	token, ok := This.tokens[socketId]
	if ok {
		delete(This.tokens, socketId)
		delete(This.tokens_u, token.UserId)

		if This.OnClientDisConnected != nil {
			This.OnClientDisConnected(token.UserId)
		}
	}
	return nil
}

//gate connected to server
func onClientConnected(uid int) {
	log.Debugf("onClientConnected: %d", uid)

	//register rpc if has
	req := &msg.LouMiaoRpcRegister{}
	for key, _ := range This.rpcMap {
		req.FuncName = append(req.FuncName, key)
	}
	if len(req.FuncName) > 0 {
		buff, _ := message.Encode(uid, 0, "LouMiaoRpcRegister", req)
		This.pInnerService.SendById(This.tokens_u[uid], buff)
	}
}

//gate租约回调
func leaseCallBack(success bool) {
	if success { //成功续租
		var ietcd etcd.IEtcdBase
		if This.ServerType == network.CLIENT_CONNECT { //gate watch server
			ietcd = This.clientReq
		} else {
			ietcd = This.serverReg
		}

		str := fmt.Sprintf("%s%s", define.ETCD_NODESTATUS, config.NET_GATE_SADDR)
		ietcd.Put(str, util.Itoa(len(This.users_u)), true) //写入人数
		//log.Debugf("leaseCallBack %s", str)
	}
}

//login gate
func InnerLouMiaoLoginGate(igo gorpc.IGoRoutine, socketId int, data interface{}) interface{} {
	m := data.(*msg.LouMiaoLoginGate)
	userid := int(m.UserId)

	log.Debugf("InnerLouMiaoLoginGate: %v", m)

	old_socketid, ok := This.tokens_u[userid]
	if ok { //close the old connection
		req := &msg.LouMiaoKickOut{}
		if config.NET_NODE_TYPE == config.ServerType_Gate {
			buff, _ := message.Encode(0, 0, "LouMiaoKickOut", req)
			This.pService.SendById(old_socketid, buff)
			if config.NET_WEBSOCKET {
				This.pService.(*network.WebSocket).StopClient(old_socketid)
			} else {
				This.pService.(*network.ServerSocket).StopClient(old_socketid)
			}
		} else {
			buff, _ := message.Encode(userid, 0, "LouMiaoKickOut", req)
			This.pInnerService.SendById(old_socketid, buff)
			This.pInnerService.(*network.ServerSocket).StopClient(old_socketid)
		}
	}

	This.tokens[socketId] = &Token{TokenId: int(m.TokenId), UserId: userid}
	This.tokens_u[userid] = socketId

	if This.OnClientConnected != nil {
		This.OnClientConnected(userid)
	}
	return nil
}

//rpc register
func InnerLouMiaoRpcRegister(igo gorpc.IGoRoutine, socketId int, data interface{}) interface{} {
	req := data.(*msg.LouMiaoRpcRegister)
	log.Debugf("InnerLouMiaoRpcRegister: %v", req)

	client := This.GetRpcClient(socketId)
	if client == nil {
		log.Warningf("0.InnerLouMiaoRpcRegister server has lost[%d] ", socketId)
		return nil
	}
	for _, key := range req.FuncName {
		if This.rpcMap[key] == nil {
			This.rpcMap[key] = make([]int, len(req.FuncName))
		}
		This.rpcMap[key] = append(This.rpcMap[key], socketId)
		rpcstr, _ := base64.StdEncoding.DecodeString(key)
		log.Debugf("rpc register: funcname=%s, uid=%d", string(rpcstr), socketId)
	}
	return nil
}

//recv rpc msg
func InnerLouMiaoRpcMsg(igo gorpc.IGoRoutine, socketId int, data interface{}) interface{} {
	req := data.(*msg.LouMiaoRpcMsg)
	handler, ok := handler_Map[req.FuncName]
	if ok {
		if config.NET_NODE_TYPE == config.ServerType_Gate { //server -> gate
			arr := This.rpcMap[req.FuncName]
			sz := len(arr)
			if sz == 0 {
				log.Warningf("0.InnerLouMiaoRpcMsg no rpc server hanlder finded %s", req.FuncName)
				return nil
			}
			target := int(req.TargetId)
			if target <= 0 {
				index := util.Random(sz) //cluser server rpc call
				target = arr[index]
			}
			rpcClient := This.GetRpcClient(target)
			if rpcClient == nil {
				log.Warningf("1.InnerLouMiaoRpcMsg rpc client error %s %d", req.FuncName, target)
				return nil
			}
			outdata := &msg.LouMiaoRpcMsg{TargetId: int32(target), FuncName: req.FuncName, Buffer: req.Buffer, SourceId: req.SourceId}
			buff, _ := message.Encode(target, 0, "LouMiaoRpcMsg", outdata)
			rpcClient.Send(buff)
		} else { //gate -> server
			err, _, name, pm := message.Decode(This.Id, req.Buffer, len(req.Buffer))
			if err != nil {
				log.Warningf("2.InnerLouMiaoRpcMsg decode msg error : func=%s, error=%s ", req.FuncName, err.Error())
				return nil
			}
			m := gorpc.M{Id: int(req.SourceId), Name: name, Data: pm}
			This.Send(handler, "ServiceHandler", m)
		}
	} else {
		log.Warningf("0.InnerLouMiaoRpcMsg no rpc hanlder %s, %d", req.FuncName, socketId)
	}

	return nil
}

//recv net msg
func InnerLouMiaoNetMsg(igo gorpc.IGoRoutine, socketId int, data interface{}) interface{} {
	req := data.(*msg.LouMiaoNetMsg) //after decode LouMiaoNetMsg msg, post Buffer to next
	log.Debugf("InnerLouMiaoNetMsg %v", data)
	clientid := int(req.ClientId)

	if config.NET_NODE_TYPE == config.ServerType_Gate { //server -> gate
		socketId, _ := This.tokens_u[clientid]       // get client's socketid by userid
		This.pService.SendById(socketId, req.Buffer) //send to client
	} else { //gate -> server
		This.users_u[clientid] = socketId
		PacketFunc(clientid, req.Buffer, len(req.Buffer)) //post msg to server service
	}
	return nil
}

func RegisterNet(igo gorpc.IGoRoutine, data interface{}) interface{} {
	m := data.(gorpc.M)
	_, ok := handler_Map[m.Name]
	if ok {
		log.Errorf("RegisterNet %s has already been registered", m.Name)
		return nil
	}
	handler_Map[m.Name] = m.Data.(string)
	if m.Id < 0 { //rpc register
		This.rpcMap[m.Name] = []int{}
	}
	return nil
}

func UnRegisterNet(igo gorpc.IGoRoutine, data interface{}) interface{} {
	return nil
}

func SendClient(igo gorpc.IGoRoutine, data interface{}) interface{} {
	m := data.(gorpc.M)
	if config.NET_NODE_TYPE == config.ServerType_Gate {
		log.Error("0.SendClient gate can not send client")
		return nil
	}
	if config.NET_NODE_TYPE == config.ServerType_Account {
		This.pService.SendById(m.Id, m.Data.([]byte)) //set to client
		return nil
	}

	socketId, _ := This.users_u[m.Id] // get gate's socketid by client's userid
	token, ok := This.tokens[socketId]
	if ok == false {
		log.Noticef("1.SendClient gate has been shut down, uid = %d", m.Id)
		return nil
	}
	msg := &msg.LouMiaoNetMsg{ClientId: int32(m.Id), Buffer: m.Data.([]byte)} //m.Id should be client`s userid
	buff, _ := message.Encode(token.UserId, 0, "LouMiaoNetMsg", msg)

	This.pInnerService.SendById(socketId, buff)

	return nil
}

func SendMulClient(igo gorpc.IGoRoutine, data interface{}) interface{} {
	m := data.(gorpc.MS)
	if config.NET_NODE_TYPE == config.ServerType_Gate {
		log.Error("0.SendMulClient gate can not send client")
		return nil
	}
	for _, v := range m.Ids {
		socketId, _ := This.users_u[v] // get gate's socketid by client's userid
		token, ok := This.tokens[socketId]
		if ok {
			log.Noticef("1.SendMulClient gate has been shut down, uid = %d", v)
			return nil
		}
		msg := &msg.LouMiaoNetMsg{ClientId: int32(v), Buffer: m.Data.([]byte)} //m.Id should be client`s userid
		buff, _ := message.Encode(token.UserId, 0, "LouMiaoNetMsg", msg)

		This.pInnerService.SendById(socketId, buff)
	}

	return nil
}

func SendRpc(igo gorpc.IGoRoutine, data interface{}) interface{} {
	m := data.(gorpc.M)
	clientid := This.GetCluserGate()
	if clientid <= 0 {
		log.Warningf("0.SendRpc no gate server finded %s", m.Name)
		return nil
	}

	//indata, _ := message.Encode(m.Id, 0, "", m.Data)
	outdata := &msg.LouMiaoRpcMsg{TargetId: int32(m.Id), FuncName: m.Name, Buffer: m.Data.([]byte), SourceId: int32(This.Id)}

	buff, _ := message.Encode(clientid, 0, "LouMiaoRpcMsg", outdata)
	This.pInnerService.SendById(clientid, buff)

	return nil
}
