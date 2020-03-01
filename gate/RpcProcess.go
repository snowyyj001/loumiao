package gate

import (
	"github.com/snowyyj001/loumiao/gorpc"
	"github.com/snowyyj001/loumiao/log"
	"github.com/snowyyj001/loumiao/message"
)

func RegisterNet(igo gorpc.IGoRoutine, data interface{}) interface{} {
	m := data.(gorpc.M)
	name := gorpc.SimpleGK(m, "name")
	receiver := gorpc.SimpleGK(m, "receiver")
	handler_Map[name.(string)] = receiver.(string)
	return nil
}

func UnRegisterNet(igo gorpc.IGoRoutine, data interface{}) interface{} {
	m := data.(gorpc.M)
	name := gorpc.SimpleGK(m, "name")
	delete(handler_Map, name.(string))
	return "success"
}

func SendClient(igo gorpc.IGoRoutine, data interface{}) interface{} {
	m := data.(gorpc.M)
	cid, name, pd := gorpc.SimpleGNet(m)
	buff, _ := message.Encode(name, pd)
	This.pService.SendById(cid, buff)
	return nil
}

func SendMulClient(igo gorpc.IGoRoutine, data interface{}) interface{} {
	m := data.(gorpc.M)
	cid, name, pd := gorpc.SimpleGMNet(m)
	buff, _ := message.Encode(name, pd)
	for _, clientid := range cid {
		This.pService.SendById(clientid, buff)
	}
	return nil
}

func SendRpcClient(igo gorpc.IGoRoutine, data interface{}) interface{} {
	m := data.(gorpc.M)
	cid, name, pd := gorpc.SimpleGNet(m)
	if This.IsChild() {
		client := This.GetClientByUid(cid)
		if client != nil {
			client.SendMsg(name, pd)
		} else {
			log.Warningf("0.SendRpcClient dest id not exist %d %s", cid, name)
		}
	} else {
		clientid := This.GetRpcClientId(cid)
		if clientid > 0 {
			buff, _ := message.Encode(name, pd)
			This.pService.SendById(clientid, buff)
		} else {
			log.Warningf("1.SendRpcClient dest id not exist %d %s", cid, name)
		}
	}
	return nil
}
