package gate

import (
	"github.com/snowyyj001/loumiao/config"
	"github.com/snowyyj001/loumiao/gorpc"
	"github.com/snowyyj001/loumiao/log"
	"github.com/snowyyj001/loumiao/message"
	"github.com/snowyyj001/loumiao/network"
)

func InnerRpcMsg(igo gorpc.IGoRoutine, socketId int, data interface{}) interface{} {
	resp := data.(*LouMiaoRpcMsg)
	if This.ServerType == network.SERVER_CONNECT {
		err, name, pm := message.Decode(resp.Buffer, len(resp.Buffer))
		if err != nil {
			log.Warningf("InnerRpcMsg hanlder[%d] decode error", socketId)
			return nil
		}
		handler, ok := handler_Map[name]
		if ok {
			This.Send(handler, "NetRpC", pm)
		} else {
			log.Warningf("InnerRpcMsg hanlder[%d] is nil, drop msg[%s]", socketId, name)
		}
	} else {
		This.pService.SendById(resp.ClientId, resp.Buffer)
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
	return nil
}

//server node heart beat
func InnerHeartBeat(igo gorpc.IGoRoutine, socketId int, data interface{}) interface{} {
	req := data.(*LouMiaoHeartBeat)
	uid := req.Uid
	if This.ServerType == network.SERVER_CONNECT {
		req := &LouMiaoHeartBeat{Uid: uid}
		buff, _ := message.Encode("LouMiaoHeartBeat", req)
		This.pService.SendById(socketId, buff)
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
	buff, _ := message.Encode(m.Name, m.Data)
	This.pService.SendById(m.Id, buff)
	return nil
}

func SendMulClient(igo gorpc.IGoRoutine, data interface{}) interface{} {
	m := data.(gorpc.MS)
	buff, _ := message.Encode(m.Name, m.Data)
	for _, clientid := range m.Ids {
		This.pService.SendById(clientid, buff)
	}
	return nil
}

func SendRpc(igo gorpc.IGoRoutine, data interface{}) interface{} {
	m := data.(*gorpc.M)
	if config.NET_BE_CHILD {
		client := This.GetRpcClient(m.Id)
		if client != nil {
			client.SendMsg(m.Name, m.Data)
		} else {
			log.Warningf("0.SendRpcClient dest id not exist %d %s", m.Id, m.Name)
		}
	} else {
		clientid := This.serverMap[m.Id]
		if clientid > 0 {
			buff, _ := message.Encode(m.Name, m.Data)
			This.pService.SendById(clientid, buff)
		} else {
			log.Warningf("1.SendRpcClient dest id not exist %d %s", m.Id, m.Name)
		}
	}
	return nil
}
