package gate

import (
	"github.com/snowyyj001/loumiao/gorpc"
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
