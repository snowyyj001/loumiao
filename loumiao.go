package loumiao

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/snowyyj001/loumiao/gorpc"
	"github.com/snowyyj001/loumiao/log"
	"github.com/snowyyj001/loumiao/message"
)

//创建一个服务，稍后开启
func Prepare(igo gorpc.IGoRoutine, name string, sync bool) {
	igo.SetSync(sync)
	gorpc.GetGoRoutineMgr().Start(igo, name)
}

//创建一个服务,立即开启
func Start(igo gorpc.IGoRoutine, name string, sync bool) {
	Prepare(igo, name, sync)
	gorpc.GetGoRoutineMgr().DoSingleStart(name)
}

//最开始的初始化
func DoInit() {
	message.DoInit()
}

//开启游戏
func Run() {

	DoInit()

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
	igo.Send("GateServer", "RegisterNet", gorpc.M{Name: name, Data: igo.GetName()})
	igo.RegisterGate(name, call)
}

func UnRegisterNetHandler(igo gorpc.IGoRoutine, name string) {
	igo.Send("GateServer", "UnRegisterNet", gorpc.M{Name: name})
	igo.UnRegisterGate(name)
}

//发送给客户端消息
func SendClient(clientid int, data interface{}) {
	server := gorpc.GetGoRoutineMgr().GetRoutine("GateServer")
	buff, _ := message.Encode("", data)
	m := gorpc.M{Id: clientid, Data: buff}
	job := gorpc.ChannelContext{"SendClient", m, nil, nil}
	server.GetJobChan() <- job
}

//发送给客户端消息
func SendMulClient(clientids []int, data interface{}) {
	server := gorpc.GetGoRoutineMgr().GetRoutine("GateServer")
	buff, _ := message.Encode("", data)
	ms := gorpc.MS{Ids: clientids, Data: buff}
	job := gorpc.ChannelContext{"SendClient", ms, nil, nil}
	server.GetJobChan() <- job
}

//发送给remote消息
func SendRpc(uid int, data interface{}) {
	server := gorpc.GetGoRoutineMgr().GetRoutine("GateServer")
	buff, _ := message.Encode("", data)
	m := gorpc.M{Id: uid, Data: buff}
	job := gorpc.ChannelContext{"SendRpc", m, nil, nil}
	server.GetJobChan() <- job
}
