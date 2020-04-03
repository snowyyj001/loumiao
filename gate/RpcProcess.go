package gate

import (
	"github.com/snowyyj001/loumiao/config"
	"github.com/snowyyj001/loumiao/gorpc"
	"github.com/snowyyj001/loumiao/log"
	"github.com/snowyyj001/loumiao/message"
	"github.com/snowyyj001/loumiao/network"
)

//do bind
func InnerHandShake(igo gorpc.IGoRoutine, socketId int, data interface{}) interface{} {
	req := data.(*LouMiaoHandShake)
	uid := req.Uid
	This.serverMap[uid] = socketId
	if This.OnClientConnected != nil {
		This.OnClientConnected(uid)
	}
	return nil
}

//server node heart beat
func InnerHeartBeat(igo gorpc.IGoRoutine, socketId int, data interface{}) interface{} {
	req := data.(*LouMiaoHeartBeat)
	uid := req.Uid
	if config.NET_BE_CHILD == 2 {
		req := &LouMiaoHeartBeat{Uid: uid}
		buff, _ := message.Encode("LouMiaoHeartBeat", req)
		This.pInnerService.SendById(socketId, buff)
	} else {
		client := This.GetRpcClient(uid)
		client.SendTimes = 0 //reset flag
	}
	return nil
}

//register rpc
func InnerRegisterRpc(igo gorpc.IGoRoutine, socketId int, data interface{}) interface{} {
	m := data.(*LouMiaoRegisterRpc)
	This.RegisterRpc(m.Name, m.Uid)

	return nil
}

//login gate
func InnerLoginGate(igo gorpc.IGoRoutine, socketId int, data interface{}) interface{} {
	m := data.(*LouMiaoLoginGate)
	old_socketid, ok := This.tokens_u[m.UserId]
	if ok {
		req := &LouMiaoKickOut{}
		buff, _ := message.Encode("LouMiaoKickOut", req)
		This.pService.SendById(old_socketid, buff)
		if config.NET_WEBSOCKET {
			This.pService.(*network.WebSocket).StopClient(old_socketid)
		} else {
			This.pService.(*network.ServerSocket).StopClient(old_socketid)
		}
	}
	This.tokens[socketId] = &Token{TokenId: m.TokenId, UserId: m.UserId}
	This.tokens_u[m.UserId] = socketId
	return nil
}

func RegisterNet(igo gorpc.IGoRoutine, data interface{}) interface{} {
	m := data.(gorpc.M)
	handler_Map[m.Name] = m.Data.(string)
	return nil
}

func UnRegisterNet(igo gorpc.IGoRoutine, data interface{}) interface{} {
	m := data.(gorpc.M)
	delete(handler_Map, m.Name)
	return "success"
}

func SendClient(igo gorpc.IGoRoutine, data interface{}) interface{} {
	m := data.(gorpc.M)
	var buff []byte = m.Data.([]byte)
	if This.pService != nil {
		This.pService.SendById(m.Id, buff)
	} else {
		req := &LouMiaoRpcMsg{ClientId: m.Id, Buffer: m.Data.([]byte)}
		buff, _ = message.Encode("LouMiaoRpcMsg", req)
		This.pInnerService.SendById(This.tokens[m.Id].TokenId, buff)
	}
	return nil
}

func SendMulClient(igo gorpc.IGoRoutine, data interface{}) interface{} {
	m := data.(gorpc.MS)
	for _, clientid := range m.Ids {
		if This.pService != nil {
			This.pService.SendById(clientid, m.Data.([]byte))
		} else {
			This.pInnerService.SendById(clientid, m.Data.([]byte))
		}
	}
	return nil
}

func SendRpc(igo gorpc.IGoRoutine, data interface{}) interface{} {
	m := data.(gorpc.M)
	if config.NET_BE_CHILD == 1 {
		client := This.GetRpcClient(m.Id)
		if client != nil {
			client.Send(m.Data.([]byte))
		} else {
			log.Warningf("0.SendRpc dest id not exist %d %s", m.Id, m.Name)
		}
	} else {
		clientid := This.serverMap[m.Id]
		if clientid > 0 {
			This.pInnerService.SendById(clientid, m.Data.([]byte))
		} else {
			log.Warningf("1.SendRpc dest id not exist %d %s", m.Id, m.Name)
		}
	}
	return nil

}
