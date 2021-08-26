package loumiao

import (
	"os"
	"os/signal"
	"reflect"
	"runtime"
	"syscall"

	"github.com/snowyyj001/loumiao/base"
	"github.com/snowyyj001/loumiao/config"

	"github.com/snowyyj001/loumiao/gorpc"
	"github.com/snowyyj001/loumiao/llog"
	"github.com/snowyyj001/loumiao/message"
	"github.com/snowyyj001/loumiao/msg"
	"github.com/snowyyj001/loumiao/timer"
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
//@sync: 是否异步协程，无状态服务可以是异步协程，有状态服务应该使用同步协程，可以保证协程安全
func Prepare(igo gorpc.IGoRoutine, name string, sync bool) {
	igo.SetSync(sync)
	gorpc.MGR.Start(igo, name)
	igo.Register("ServiceHandler", gorpc.ServiceHandler)
}

//创建一个服务,立即开启
//@name: actor名，唯一
//@sync: 是否异步协程，无状态服务可以是异步协程，有状态服务应该使用同步协程，可以保证协程安全
func Start(igo gorpc.IGoRoutine, name string, sync bool) {
	Prepare(igo, name, sync)
	gorpc.MGR.DoSingleStart(name)
}

//开启游戏
func Run() {
	defer func() {
		if r := recover(); r != nil {
			buf := make([]byte, 2048)
			l := runtime.Stack(buf, false)
			llog.Errorf("loumiao run error: name=%s, error=%v, stack=%s", config.SERVER_NAME, r, buf[:l])
		}
	}()

	gorpc.MGR.DoStart()

	timer.DelayJob(1000, func() {
		igo := gorpc.MGR.GetRoutine("GateServer")
		if igo != nil {
			igo.DoOpen()
		}
		llog.Infof("loumiao start success: %s", config.SERVER_NAME)
	}, true)

	c = make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, os.Kill, syscall.SIGTERM, syscall.SIGINT)
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

//注册网络消息,对于内部server节点来说HanlderNetFunc的第二个参数clientid就是userid
func RegisterNetHandler(igo gorpc.IGoRoutine, name string, call gorpc.HanlderNetFunc) {
	gorpc.MGR.Send("GateServer", "RegisterNet", &gorpc.M{Id: 0, Name: name, Data: igo.GetName()})
	igo.RegisterGate(name, call)
}

//暂时没发现需要撤销注册的情况
func UnRegisterNetHandler(igo gorpc.IGoRoutine, name string) {
	//	gorpc.MGR.Send("GateServer", "UnRegisterNet", &gorpc.M{Id: 0, Name: name})
	//	igo.UnRegisterGate(name)
}

//注册网络消息kcp server,对于内部server节点来说HanlderNetFunc的第二个参数clientid就是真实的socketid
func RegisterKcpNetHandler(igo gorpc.IGoRoutine, name string, call gorpc.HanlderNetFunc) {
	gorpc.MGR.Send("KcpGateServer", "RegisterNet", &gorpc.M{Id: 0, Name: name, Data: igo.GetName()})
	igo.RegisterGate(name, call)
}

//暂时没发现需要撤销注册的情况
func UnRegisterKcpNetHandler(igo gorpc.IGoRoutine, name string) {
	//gorpc.MGR.Send("GateServer", "UnRegisterNet", &gorpc.M{Id: 0, Name: name})
	//igo.UnRegisterGate(name)
}

//发送给客户端消息
//@clientid: 客户端userid
//@data: 消息结构体指针
func SendClient(clientid int, data interface{}) {
	buff, n := message.Encode(0, "", data)
	if n == 0 {
		return
	}
	m := &gorpc.M{Id: clientid, Data: buff}
	gorpc.MGR.Send("GateServer", "SendClient", m)
}

//发送给客户端消息
//@clientids: 客户端userid, 空代表全服发送
//@data: 消息结构体指针
func SendMulClient(clientids []int, data interface{}) {
	buff, n := message.Encode(0, "", data)
	if n == 0 {
		return
	}
	ms := &gorpc.MS{Ids: clientids, Data: buff}
	m := &gorpc.M{Data: ms}
	gorpc.MGR.Send("GateServer", "SendMulClient", m)
}

//广播消息
//@data: 消息结构体指针
func BroadCastMsg(data interface{}) {
	SendMulClient(nil, data)
}

//获得rpc注册名
func RpcFuncName(call interface{}) string {
	//base64str := base64.StdEncoding.EncodeToString([]byte(funcName))
	funcName := runtime.FuncForPC(reflect.ValueOf(call).Pointer()).Name()
	return funcName
}

//注册rpc消息
func RegisterRpcHandler(igo gorpc.IGoRoutine, call gorpc.HanlderNetFunc) {
	funcName := RpcFuncName(call)
	//base64str := base64.StdEncoding.EncodeToString([]byte(funcName))
	gorpc.MGR.Send("GateServer", "RegisterNet", &gorpc.M{Id: -1, Name: funcName, Data: igo.GetName()})
	igo.RegisterGate(funcName, call)
}

func UnRegisterRpcHandler(igo gorpc.IGoRoutine, call gorpc.HanlderNetFunc) {
	funcName := RpcFuncName(call)
	gorpc.MGR.Send("GateServer", "UnRegisterNet", &gorpc.M{Id: -1, Name: funcName})
	igo.UnRegisterGate(funcName)
}

//远程rpc调用
//@funcName: rpc函数
//@data: 函数参数,如果data是[]byte类型，则代表使用bitstream或自定义二进制内容，否则data应该是一个messgae注册的pb或json结构体
//@target: 目标server的uid，如果target==0，则随机指定目标地址, 否则gate会把消息转发给指定的target服务
func SendRpc(funcName string, data interface{}, target int) {
	m := &gorpc.M{Id: target, Name: funcName}
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

//发送给gate的消息
//@uid: 目标gate的uid，如果uid=0，则会随机选择一个gate发送
//@data: 发送消息
func SendGate(uid int, data interface{}) {
	buff, _ := message.Encode(uid, "", data)
	m := &gorpc.M{Id: uid, Data: buff}
	gorpc.MGR.Send("GateServer", "SendGate", m)
}

//全局通用发送actor消息
//@actorName: 目标actor的名字
//@actorHandler: 目标actor的处理函数
//@data: 函数参数
func SendAcotr(actorName string, actorHandler string, data interface{}) {
	m := &gorpc.M{Data: data, Flag: true}
	gorpc.MGR.Send(actorName, actorHandler, m)
}

//主动绑定关于client的gate信息，例如，world通知zone绑定client和gate的信息
//world通知其他server关于client的gate信息,其他server只有知道了client属于哪个gate才能发送消息给client
//@userid: client的userid
//@gateuid: client所属的gate uid
//@targetuid: 目标服务器uid
func BindGate(userid int64, gateuid int, targetuid int) {
	req := &msg.LouMiaoBindGate{Uid: int32(gateuid), UserId: userid}
	SendRpc("LouMiaoBindGate", req, targetuid)
}

//sub/pub系统，只在本节点服务内生效
//发布
//@key: 发布的key
//@value: 发布的值
func Publish(key, value string) { //发布
	igo := gorpc.MGR.GetRoutine("GateServer")
	if igo == nil {
		llog.Errorf("loumiao.Publish error, no gate actor: key=%s,value=%s", key, value)
		return
	}
	bitstream := base.NewBitStream_1(base.BitStrLen(key) + base.BitStrLen(value) + 1)
	bitstream.WriteString(key)
	bitstream.WriteString(value)
	SendAcotr("GateServer", "Publish", bitstream.GetBuffer())
}

//订阅
//@name: 订阅者的igo name
//@key: 订阅的key
//@hanlder: 订阅的actor处理函数,为""即为取消订阅
func Subscribe(igo gorpc.IGoRoutine, key string, call gorpc.HanlderFunc) { //订阅
	gateigo := gorpc.MGR.GetRoutine("GateServer")
	if gateigo == nil {
		llog.Errorf("loumiao.Subscribe error, no gate actor: key=%s", key)
		return
	}
	name := igo.GetName()
	var hanlder string
	if call != nil {
		hanlder = RpcFuncName(call)
		igo.Register(hanlder, call)
	} else {
		hanlder = ""
	}

	bitstream := base.NewBitStream_1(base.BitStrLen(name) + base.BitStrLen(key) + base.BitStrLen(hanlder) + 1)
	bitstream.WriteString(name)
	bitstream.WriteString(key)
	bitstream.WriteString(hanlder)
	SendAcotr("GateServer", "Subscribe", bitstream.GetBuffer())
}
