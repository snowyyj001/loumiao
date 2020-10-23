package loumiao

import (
	"os"
	"os/signal"
	"reflect"
	"runtime"
	"syscall"

	"github.com/snowyyj001/loumiao/util/timer"

	_ "github.com/snowyyj001/loumiao/config"
	"github.com/snowyyj001/loumiao/gorpc"
	"github.com/snowyyj001/loumiao/log"
	"github.com/snowyyj001/loumiao/message"
	_ "github.com/snowyyj001/loumiao/nsq"
)

//最开始的初始化
func DoInit() {
	message.DoInit()
	//nsq.Init(config.Cfg.NsqAddr)
}

func init() {
	DoInit()
}

//创建一个服务，稍后开启
//@name: actor名，唯一
//@sync: 是否异步协程
func Prepare(igo gorpc.IGoRoutine, name string, sync bool) {
	igo.SetSync(sync)
	gorpc.GetGoRoutineMgr().Start(igo, name)
	igo.Register("ServiceHandler", gorpc.ServiceHandler)
}

//创建一个服务,立即开启
func Start(igo gorpc.IGoRoutine, name string, sync bool) {
	Prepare(igo, name, sync)
	gorpc.GetGoRoutineMgr().DoSingleStart(name)
}

//开启游戏
func Run() {

	gorpc.GetGoRoutineMgr().DoStart()

	timer.DelayJob(1000, func() {
		igo := gorpc.GetGoRoutineMgr().GetRoutine("GateServer")
		if igo != nil {
			igo.DoOpen()
		}
	}, true)

	log.Notice("loumiao start success!")

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, os.Kill, syscall.SIGTERM)
	sig := <-c
	log.Infof("loumiao closing down (signal: %v)", sig)

	gorpc.GetGoRoutineMgr().CloseAll()

	log.Infof("loumiao done !")
}

//注册网络消息
func RegisterNetHandler(igo gorpc.IGoRoutine, name string, call gorpc.HanlderNetFunc) {
	igo.Send("GateServer", "RegisterNet", gorpc.M{Id: 0, Name: name, Data: igo.GetName()})
	igo.RegisterGate(name, call)
}

func UnRegisterNetHandler(igo gorpc.IGoRoutine, name string) {
	igo.Send("GateServer", "UnRegisterNet", gorpc.M{Id: 0, Name: name})
	igo.UnRegisterGate(name)
}

//发送给客户端消息
func SendClient(clientid int, data interface{}) {
	buff, _ := message.Encode(0, 0, "", data)
	m := gorpc.M{Id: clientid, Data: buff}
	server := gorpc.GetGoRoutineMgr().GetRoutine("GateServer")
	job := gorpc.ChannelContext{"SendClient", m, nil, nil}
	server.GetJobChan() <- job
}

//发送给客户端消息
func SendMulClient(igo gorpc.IGoRoutine, clientids []int, data interface{}) {
	buff, _ := message.Encode(0, 0, "", data)
	m := gorpc.MS{Ids: clientids, Data: buff}
	server := gorpc.GetGoRoutineMgr().GetRoutine("GateServer")
	job := gorpc.ChannelContext{"SendMulClient", m, nil, nil}
	server.GetJobChan() <- job
}

//注册rpc消息
func RegisterRpcHandler(igo gorpc.IGoRoutine, call gorpc.HanlderNetFunc) {
	funcName := runtime.FuncForPC(reflect.ValueOf(call).Pointer()).Name()
	//base64str := base64.StdEncoding.EncodeToString([]byte(funcName))
	igo.Send("GateServer", "RegisterNet", gorpc.M{Id: -1, Name: funcName, Data: igo.GetName()})

	igo.RegisterGate(funcName, call)
}

func UnRegisterRpcHandler(igo gorpc.IGoRoutine, call gorpc.HanlderNetFunc) {
	funcName := runtime.FuncForPC(reflect.ValueOf(call).Pointer()).Name()
	//base64str := base64.StdEncoding.EncodeToString([]byte(funcName))
	igo.Send("GateServer", "UnRegisterNet", gorpc.M{Id: -1, Name: funcName})

	igo.UnRegisterGate(funcName)
}

//rpc调用
//@funcName: rpc函数
//@data: 函数参数
//@target: 目标server的uid，如果不指定，则随机指定目标地址
func SendRpc(igo gorpc.IGoRoutine, funcName string, data interface{}, target int) {
	buff, _ := message.Encode(target, 0, "", data)
	//base64str := base64.StdEncoding.EncodeToString([]byte(funcName))
	m := gorpc.M{Id: target, Name: funcName, Data: buff}
	igo.Send("GateServer", "SendRpc", m)
}
