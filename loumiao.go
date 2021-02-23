package loumiao

import (
	"os"
	"os/signal"
	"reflect"
	"runtime"
	"syscall"

	"github.com/snowyyj001/loumiao/config"

	"github.com/snowyyj001/loumiao/util/timer"

	"github.com/snowyyj001/loumiao/gorpc"
	"github.com/snowyyj001/loumiao/llog"
	"github.com/snowyyj001/loumiao/message"
)

var c chan os.Signal

//最开始的初始化
func DoInit() {
	message.DoInit()
}

func init() {
	DoInit()
}

//创建一个服务，稍后开启
//@name: actor名，唯一
//@sync: 是否异步协程
func Prepare(igo gorpc.IGoRoutine, name string, sync bool) {
	igo.SetSync(sync)
	gorpc.MGR.Start(igo, name)
	igo.Register("ServiceHandler", gorpc.ServiceHandler)
}

//创建一个服务,立即开启
func Start(igo gorpc.IGoRoutine, name string, sync bool) {
	Prepare(igo, name, sync)
	gorpc.MGR.DoSingleStart(name)
}

//开启游戏
func Run() {

	gorpc.MGR.DoStart()

	timer.DelayJob(1000, func() {
		igo := gorpc.MGR.GetRoutine("GateServer")
		if igo != nil {
			igo.DoOpen()
			llog.Noticef("loumiao start success: %s", config.SERVER_NAME)
		}
	}, true)

	c = make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, os.Kill, syscall.SIGTERM)
	sig := <-c
	llog.Infof("loumiao closing down (signal: %v)", sig)

	gorpc.MGR.CloseAll()

	llog.Infof("loumiao done !")
}

//关闭游戏
func Stop() {
	llog.Info("loumiao stop the server !")
	c <- os.Kill
}

//注册网络消息
func RegisterNetHandler(igo gorpc.IGoRoutine, name string, call gorpc.HanlderNetFunc) {
	gorpc.MGR.Send("GateServer", "RegisterNet", gorpc.M{Id: 0, Name: name, Data: igo.GetName()})
	igo.RegisterGate(name, call)
}

func UnRegisterNetHandler(igo gorpc.IGoRoutine, name string) {
	gorpc.MGR.Send("GateServer", "UnRegisterNet", gorpc.M{Id: 0, Name: name})
	igo.UnRegisterGate(name)
}

//发送给客户端消息
//@clientid: 客户端userid
//@data: 消息结构体指针
func SendClient(clientid int, data interface{}) {
	buff, _ := message.Encode(0, "", data)
	m := gorpc.M{Id: clientid, Data: buff}
	server := gorpc.MGR.GetRoutine("GateServer")
	job := gorpc.ChannelContext{"SendClient", m, nil, nil}
	server.GetJobChan() <- job
}

//发送给客户端消息
func SendMulClient(igo gorpc.IGoRoutine, clientids []int, data interface{}) {
	buff, _ := message.Encode(0, "", data)
	m := gorpc.MS{Ids: clientids, Data: buff}
	server := gorpc.MGR.GetRoutine("GateServer")
	job := gorpc.ChannelContext{"SendMulClient", m, nil, nil}
	server.GetJobChan() <- job
}

//获得rpc注册名
func RpcFuncName(call gorpc.HanlderNetFunc) string {
	//base64str := base64.StdEncoding.EncodeToString([]byte(funcName))
	funcName := runtime.FuncForPC(reflect.ValueOf(call).Pointer()).Name()
	return funcName
}

//注册rpc消息
func RegisterRpcHandler(igo gorpc.IGoRoutine, call gorpc.HanlderNetFunc) {
	funcName := RpcFuncName(call)
	//base64str := base64.StdEncoding.EncodeToString([]byte(funcName))
	gorpc.MGR.Send("GateServer", "RegisterNet", gorpc.M{Id: -1, Name: funcName, Data: igo.GetName()})
	igo.RegisterGate(funcName, call)
}

func UnRegisterRpcHandler(igo gorpc.IGoRoutine, call gorpc.HanlderNetFunc) {
	funcName := RpcFuncName(call)
	gorpc.MGR.Send("GateServer", "UnRegisterNet", gorpc.M{Id: -1, Name: funcName})
	igo.UnRegisterGate(funcName)
}

//远程rpc调用
//@funcName: rpc函数
//@data: 函数参数,如果data是[]byte类型，则代表使用bitstream或自定义二进制内容，否则data应该是一个messgae注册的pb或json结构体
//@target: 目标server的uid，如果target==0，则随机指定目标地址, 否则gate会把消息转发给指定的target服务
func SendRpc(funcName string, data interface{}, target int) {
	m := gorpc.MM{Id: target, Name: funcName}
	if reflect.TypeOf(data).Kind() == reflect.Slice { //bitstream
		m.Data = data
		m.Param = 1
	} else {
		buff, _ := message.Encode(target, "", data)
		m.Data = buff
	}
	llog.Debugf("SendRpc: %s, %d", funcName, target)
	//base64str := base64.StdEncoding.EncodeToString([]byte(funcName))
	gorpc.MGR.Send("GateServer", "SendRpc", m)
}

//远程rpc消息广播调用-*********还没测试
//@funcName: rpc函数
//@data: 函数参数,如果data是[]byte类型，则代表使用bitstream或自定义二进制内容，否则data应该是一个messgae注册的pb或json结构体
//@target: 目标server的type
func BroadCastRpc(funcName string, data interface{}, target int) {
	m := gorpc.MM{Id: target, Name: funcName}
	if reflect.TypeOf(data).Kind() == reflect.Slice { //bitstream
		m.Data = data
		m.Param = 1
	} else {
		buff, _ := message.Encode(target, "", data)
		m.Data = buff
	}
	llog.Debugf("BroadCastRpc: %s, %d", funcName, target)
	//base64str := base64.StdEncoding.EncodeToString([]byte(funcName))
	gorpc.MGR.Send("GateServer", "BroadCastRpc", m)
}

//发送给gate的网络消息
//@clientid: 目标gate的uid，如果clientid=0，则会随机选择一个gate发送
//@data: 发送消息
func SendGate(clientid int, data interface{}) {
	buff, _ := message.Encode(clientid, "", data)
	m := gorpc.M{Id: clientid, Data: buff}
	gorpc.MGR.Send("GateServer", "SendGate", m)
}

//全局通用发送actor消息
//@actorName: 目标actor的名字
//@actorHandler: 目标actor的处理函数
//@data: 函数参数
func SendAcotr(actorName string, actorHandler string, data interface{}) {
	server := gorpc.MGR.GetRoutine(actorName)
	job := gorpc.ChannelContext{actorHandler, data, nil, nil}
	server.GetJobChan() <- job
}
