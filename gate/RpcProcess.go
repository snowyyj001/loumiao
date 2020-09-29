package gate

import (
	"fmt"

	"github.com/snowyyj001/loumiao/config"
	"github.com/snowyyj001/loumiao/define"
	"github.com/snowyyj001/loumiao/etcd"
	"github.com/snowyyj001/loumiao/gorpc"
	"github.com/snowyyj001/loumiao/log"
	"github.com/snowyyj001/loumiao/message"
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
			buff, _ := message.Encode(m.UserId, 0, "LouMiaoKickOut", req)
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
		if config.NET_NODE_TYPE == config.ServerType_Gate { //server -> gate
			arr := This.rpcMap[req.FuncName]
			sz := len(arr)
			if sz == 0 {
				log.Warningf("0.InnerLouMiaoRpcMsg no rpc server hanlder finded %s", req.FuncName)
				return nil
			}
			index := util.Random(sz) //随机一个server进行rpc调用
			target := arr[index]

			rpcClient := This.GetRpcClient(target)
			if rpcClient == nil {
				log.Warningf("1.InnerLouMiaoRpcMsg rpc client error %s %d", req.FuncName, target)
				return nil
			}
			outdata := LouMiaoRpcMsg{FuncName: m.Name, Buffer: req.Buffer}
			buff, _ := message.Encode(target, 0, "LouMiaoRpcMsg", outdata)
			rpcClient.Send(buff)
		} else { //gate -> server
			err, target, name, pm := message.Decode(This.Id, req.Buffer, len(req.Buffer))
			if err != nil {
				log.Warningf("2.InnerLouMiaoRpcMsg decode msg error : func=%s, error=%s ", req.FuncName, err.Error())
				return nil
			}
			m := gorpc.M{Id: 0, Name: name, Data: pm}
			This.Send(handler, "ServiceHandler", m)
		}
	} else {
		log.Warningf("0.InnerLouMiaoRpcMsg no rpc hanlder %s, %d", req.FuncName, socketId)
	}

	return nil
}

//recv net msg
func InnerLouMiaoNetMsg(igo gorpc.IGoRoutine, socketId int, data interface{}) interface{} {
	req := data.(*LouMiaoNetMsg) //after decode LouMiaoNetMsg msg, post Buffer to next

	if config.NET_NODE_TYPE == config.ServerType_Gate { //server -> gate
		socketId, _ := This.tokens_u[req.ClientId]   // get client's socketid by userid
		This.pService.SendById(socketId, req.Buffer) //set to client
	} else { //gate -> server
		This.users_u[req.ClientId] = socketId
		PacketFunc(req.ClientId, req.Buffer, len(req.Buffer)) //post msg to server service
	}
	return nil
}

func RegisterNet(igo gorpc.IGoRoutine, data interface{}) interface{} {
	m := data.(gorpc.M)
	v, ok := handler_Map[m.Name]
	if ok {
		log.Errorf("RegisterNet %s has already been registered", m.Name)
		return nil
	}
	handler_Map[m.Name] = m.Data.(string)
	if m.Id < 0 { //rpc register
		key := fmt.Sprintf("%s%s/%d", define.ETCD_RPCADDR, m.Name, This.Id)
		etcd.Put(This.serverReg.GetClient(), key, util.Itoa(This.Id))
	}
	return nil
}

func UnRegisterNet(igo gorpc.IGoRoutine, data interface{}) interface{} {
	m := data.(gorpc.M)
	delete(handler_Map, m.Name)
	if m.Id < 0 { //rpc unregister
		key := fmt.Sprintf("%s%s/%d", define.ETCD_RPCADDR, m.Name, This.Id)
		etcd.Delete(This.serverReg.GetClient(), key)
	}
	return nil
}

func SendClient(igo gorpc.IGoRoutine, data interface{}) interface{} {
	m := data.(gorpc.M)
	if config.NET_NODE_TYPE == config.ServerType_Gate {
		log.Error("0.SendClient gate can not send client")
		return nil
	}
	socketId, _ := This.users_u[m.Id] // get gate's socketid by client's userid
	token, ok := This.tokens[socketId]
	if ok {
		log.Noticef("1.SendClient gate has been shut down, uid = %d", m.Id)
		return nil
	}
	msg := LouMiaoNetMsg{ClientId: m.Id, Buffer: m.Data.([]byte)} //m.Id should be client`s userid
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
		msg := LouMiaoNetMsg{ClientId: v, Buffer: m.Data.([]byte)} //m.Id should be client`s userid
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

	indata, _ := message.Encode(target, 0, "", m.Data)
	outdata := LouMiaoRpcMsg{FuncName: m.Name, Buffer: indata}

	buff, _ := message.Encode(target, 0, "LouMiaoRpcMsg", outdata)
	rpcClient.Send(buff)

	return nil
}
