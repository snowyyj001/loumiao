package gate

import (
	"github.com/snowyyj001/loumiao/gorpc"
	"github.com/snowyyj001/loumiao/log"
	"github.com/snowyyj001/loumiao/message"
)

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

func SendRpcClient(igo gorpc.IGoRoutine, data interface{}) interface{} {
	m := data.(gorpc.M)
	if This.IsChild() {
		client := This.GetClientByUid(m.Id)
		if client != nil {
			client.SendMsg(m.Name, m.Data)
		} else {
			log.Warningf("0.SendRpcClient dest id not exist %d %s", m.Id, m.Name)
		}
	} else {
		clientid := This.GetRpcClientId(m.Id)
		if clientid > 0 {
			buff, _ := message.Encode(m.Name, m.Data)
			This.pService.SendById(clientid, buff)
		} else {
			log.Warningf("1.SendRpcClient dest id not exist %d %s", m.Id, m.Name)
		}
	}
	return nil
}
