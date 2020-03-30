package gate

import (
	"fmt"

	"github.com/snowyyj001/loumiao/config"
	"github.com/snowyyj001/loumiao/gorpc"
	"github.com/snowyyj001/loumiao/log"
	"github.com/snowyyj001/loumiao/message"
)

func InnerRpcMsg(igo gorpc.IGoRoutine, socketId int, data interface{}) interface{} {
	fmt.Println("InnerRpcMsg", socketId, data)
	resp := data.(*LouMiaoRpcMsg)

	err, name, pm := message.Decode(resp.Buffer, len(resp.Buffer))
	if err != nil {
		log.Warningf("InnerRpcMsg hanlder[%d] decode error", socketId)
		return nil
	}
	handler, ok := handler_Map[name]
	if ok {
		m := gorpc.M{Id: socketId, Name: name, Data: pm}
		if handler == "GateServer" { //msg to gate server
			This.CallNetFunc(m)
		} else { //to local rpc
			This.Send(handler, "NetRpC", m)
		}
	} else {
		log.Warningf("InnerRpcMsg hanlder[%d] is nil, drop msg[%s]", socketId, name)
	}

	return nil
}

func InnerDisConnect(igo gorpc.IGoRoutine, socketId int, data interface{}) interface{} {
	for _, client := range This.clients {
		if client.GetClientId() == socketId {
			delete(This.clients, client.Uid)
			//remote rpc断开，这里自动重新连接
			This.BuildRpc(client.GetIP(), client.GetPort(), client.Uid, client.Uuid)
			break
		}
	}
	return nil
}

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
	This.pService.SendById(m.Id, m.Data.([]byte))
	return nil
}

func SendMulClient(igo gorpc.IGoRoutine, data interface{}) interface{} {
	m := data.(gorpc.MS)
	for _, clientid := range m.Ids {
		This.pService.SendById(clientid, m.Data.([]byte))
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
