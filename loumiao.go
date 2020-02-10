package loumiao

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/snowyyj001/loumiao/gorpc"
	"github.com/snowyyj001/loumiao/log"
	"github.com/snowyyj001/loumiao/message"
)

//开启一个任务
func Prepare(igo gorpc.IGoRoutine, name string, sync bool) {
	igo.SetSync(sync)
	gorpc.GetGoRoutineMgr().Start(igo, name)
}

func Run() {

	message.DoInit()

	gorpc.GetGoRoutineMgr().DoStart()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, os.Kill, syscall.SIGTERM)
	sig := <-c
	log.Infof("loumiao closing down (signal: %v)", sig)

	gorpc.GetGoRoutineMgr().CloseAll()

	log.Infof("loumiao done !")
}

//向网关注册网络消息
func RegisterNetHandler(igo gorpc.IGoRoutine, name string, call gorpc.HanlderNetFunc) {
	igo.Register("NetRpC", gorpc.NetRpC)
	igo.Send("GateServer", "RegisterNet", gorpc.M{"name": name, "receiver": igo.GetName()})
	igo.RegisterGate(name, call)
}

func UnRegisterNetHandler(igo gorpc.IGoRoutine, name string) {
	igo.Send("GateServer", "UnRegisterNet", gorpc.M{"name": name})
	igo.UnRegisterGate(name)
}

//发送给客户端消息
func SendClient(igo gorpc.IGoRoutine, clientid int, data interface{}) {
	igo.Send("GateServer", "SendClient", gorpc.SimpleNet(clientid, "", data))
}

func WriteSocket(name string, clientid int, data interface{}) {
	igo := gorpc.GetGoRoutineMgr().GetRoutine(name)
	SendClient(igo, clientid, data)
}
