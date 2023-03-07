package loumiao

import (
	"fmt"
	"github.com/golang/protobuf/proto"
	"github.com/snowyyj001/loumiao/gate"
	"github.com/snowyyj001/loumiao/kcpgate"
	"github.com/snowyyj001/loumiao/udpgate"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"reflect"
	"syscall"

	"github.com/snowyyj001/loumiao/callrpc"
	"github.com/snowyyj001/loumiao/define"
	"github.com/snowyyj001/loumiao/nodemgr"

	"github.com/snowyyj001/loumiao/base"
	"github.com/snowyyj001/loumiao/config"
	"github.com/snowyyj001/loumiao/gorpc"
	"github.com/snowyyj001/loumiao/llog"
	"github.com/snowyyj001/loumiao/message"
	"github.com/snowyyj001/loumiao/util"
)

var c chan os.Signal

// 最开始的初始化
func DoInit() {
	message.DoInit()
	Start(new(callrpc.CallRpcServer), "CallRpcServer", true)
}

func init() {
	DoInit()
}

// 创建一个服务，稍后开启
// @name: actor名，唯一
// @sync: 是否异步协程，无状态服务可以是异步协程，有状态服务应该使用同步协程，可以保证协程安全
func Prepare(igo gorpc.IGoRoutine, name string, sync bool) {
	igo.SetSync(sync)
	gorpc.MGR.Start(igo, name)
	igo.Register("ServiceHandler", gorpc.ServiceHandler)
}

// 创建一个服务,立即开启
// @name: actor名，唯一
// @sync: 是否异步协程，无状态服务可以是异步协程，有状态服务应该使用同步协程，可以保证协程安全
func Start(igo gorpc.IGoRoutine, name string, sync bool) {
	Prepare(igo, name, sync)
	gorpc.MGR.DoSingleStart(name)
}

// 开启游戏
func Run() {
	defer util.Recover()

	util.DumpPid()

	if config.SERVER_DEBUGPORT > 0 {
		go func() {
			http.ListenAndServe(fmt.Sprintf("0.0.0.0:%d", config.SERVER_DEBUGPORT), nil)
		}()
	}

	gorpc.MGR.DoStart()

	//timer.DelayJob(1000, func() {
	igo := gorpc.MGR.GetRoutine("GateServer")
	if igo != nil {
		igo.DoOpen()
	}
	llog.Infof("loumiao start success: %s", config.SERVER_NAME)
	//}, true)

	c = make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, os.Kill, syscall.SIGTERM, syscall.SIGINT)
	sig := <-c
	llog.Infof("loumiao closing down (signal: %v)", sig)

	gorpc.Close()

	llog.Infof("loumiao done !")
}

// 关闭游戏
func Stop() {
	llog.Info("loumiao stop the server !")
	c <- os.Kill
}

// RegisterNetHandler 注册网络消息,对于内部server节点来说HandlerNetFunc的第二个参数clientid就是userid
func RegisterNetHandler(igo gorpc.IGoRoutine, name string, call gorpc.HandlerNetFunc) {
	//gorpc.MGR.Send("GateServer", "RegisterNet", &gorpc.M{Id: 0, Name: name, Data: igo.GetName()})
	gate.RegisterNet(igo, name, false)
	igo.RegisterGate(name, call)
}

// 暂时没发现需要撤销注册的情况
func UnRegisterNetHandler(igo gorpc.IGoRoutine, name string) {
	//	gorpc.MGR.Send("GateServer", "UnRegisterNet", &gorpc.M{Id: 0, Name: name})
	//	igo.UnRegisterGate(name)
}

// RegisterKcpNetHandler 注册网络消息kcp server,对于内部server节点来说HandlerNetFunc的第二个参数clientid就是真实的socketid
func RegisterKcpNetHandler(igo gorpc.IGoRoutine, name string, call gorpc.HandlerNetFunc) {
	//gorpc.MGR.Send("KcpGateServer", "RegisterNet", &gorpc.M{Id: 0, Name: name, Data: igo.GetName()})
	kcpgate.RegisterNet(igo, &gorpc.M{Id: 0, Name: name, Data: igo.GetName()})
	igo.RegisterGate(name, call)
}

// 暂时没发现需要撤销注册的情况
func UnRegisterKcpNetHandler(igo gorpc.IGoRoutine, name string) {
	//gorpc.MGR.Send("GateServer", "UnRegisterNet", &gorpc.M{Id: 0, Name: name})
	//igo.UnRegisterGate(name)
}

// RegisterUdpNetHandler 注册网络消息udp server,
func RegisterUdpNetHandler(igo gorpc.IGoRoutine, msgId int, userid int, call gorpc.HandlerNetFunc) {
	udpgate.RegisterHandler(userid, igo, call)
}

// UnRegisterUdpNetHandler 取消注册网络消息udp server,
func UnRegisterUdpNetHandler(userId int) {
	udpgate.UnRegisterHandler(userId)
}

// SendClient 发送给客户端消息
// @clientid: 客户端userid
// @data: 消息结构体指针
func SendClient(clientid int, data interface{}) {
	buff, n := message.Encode(0, "", data)
	if n == 0 {
		return
	}
	gate.SendClient(clientid, buff)
}

// SendClients 广播给客户端消息
// @data: 消息结构体指针
func SendClients(ids []int, data interface{}) {
	buff, n := message.Encode(0, "", data)
	if n == 0 {
		return
	}
	for _, clientid := range ids {
		gate.SendClient(clientid, buff)
	}
}

// BroadCastClients 广播给客户端消息
// @data: 消息结构体指针
func BroadCastClients(data interface{}) {
	buff, n := message.Encode(0, "", data)
	if n == 0 {
		return
	}
	gate.BroadCastClients(buff)
}

// RegisterRpcHandler 注册rpc消息
func RegisterRpcHandler(igo gorpc.IGoRoutine, call gorpc.HandlerRpcFunc) {
	util.Assert(nodemgr.ServerEnabled == false, "RegisterRpcHandler can not register after server started")
	funcName := util.RpcFuncName(call)
	//llog.Debugf("funcName = %s", funcName)
	//rpc send
	igo.Register(funcName, func(igo gorpc.IGoRoutine, data interface{}) interface{} {
		return call(data.([]byte))
	})
	//rpc call
	igo.RegisterGate(funcName, func(igo gorpc.IGoRoutine, clientId int, buffer []byte) {
		call(buffer)
	})
	gate.RegisterNet(igo, funcName, true)
}

// SendRpc 远程rpc调用
// @funcName: rpc函数
// @data: 函数参数,一个二进制buff或pb结构体
// @target: 目标server的uid，如果target==0，则随机指定目标地址, 否则gate会把消息转发给指定的target服务
func SendRpc(funcName string, data interface{}, target int) {
	m := &gorpc.M{Id: target, Name: funcName}
	if reflect.TypeOf(data).Kind() == reflect.Slice { //bitstream
		m.Data = data
	} else {
		buff, err := message.PackProto(data.(proto.Message))
		if err != nil {
			llog.Errorf("SendRpc: %s", err.Error())
			return
		}
		m.Data = buff
	}
	llog.Debugf("SendRpc: %s, %d", funcName, target)
	//base64str := base64.StdEncoding.EncodeToString([]byte(funcName))
	gorpc.MGR.Send("GateServer", "SendRpc", m)
}

// CallRpc 远程rpc调用
// igo: 调用者的actor，会阻塞此actor
// @funcName: rpc函数
// @data: 一个二进制buff或pb结构体
// @target: 目标server的uid，如果target==0，则随机指定目标地址, 否则gate会把消息转发给指定的target服务
// return: 返回的[]byte结果或nil
func CallRpc(igo gorpc.IGoRoutine, funcName string, data interface{}, target int) ([]byte, bool) {
	m := &gorpc.M{Id: target, Name: funcName}
	m.Param = define.RPCMSG_FLAG_CALL
	session := igo.GetName()
	if data == nil {
		m.Data = []byte{}
	} else if reflect.TypeOf(data).Kind() == reflect.Slice { //bitstream
		orgbuff := data.([]byte)
		bitstream := base.NewBitStreamS(len(orgbuff) + base.BitStrLen(session))
		bitstream.WriteString(session)
		bitstream.WriteBytes(orgbuff)
		m.Data = bitstream.GetBuffer()
	} else {
		orgbuff, err := message.PackProto(data.(proto.Message))
		if err != nil {
			llog.Errorf("CallRpc: %s", err.Error())
			return nil, false
		}
		bitstream := base.NewBitStreamS(len(orgbuff) + base.BitStrLen(session))
		bitstream.WriteString(session)
		bitstream.WriteBytes(orgbuff)
		m.Data = bitstream.GetBuffer()
	}
	//llog.Debugf("CallRpc: session=%s, funcName=%s, target=%d", session, funcName, target)
	//base64str := base64.StdEncoding.EncodeToString([]byte(funcName))
	gorpc.MGR.Send("GateServer", "SendRpc", m)

	resp, ok := igo.CallActor("CallRpcServer", "CallRpc", session)
	if resp == nil || !ok { //既然调用call了就是为了返回数据，nil是不能接受的
		llog.Errorf("loumiao.CallRpc: src = %s, func = %s, target = %d", session, funcName, target)
		return nil, false
	}

	return resp.([]byte), ok
}

// RpcBoolResult 构造一个rpc result返回结果
func RpcBoolResult(ret bool) []byte {
	if ret {
		return base.IntToBytes(1)
	} else {
		return base.IntToBytes(0)
	}
}

// RpcBoolResultOK rpc的返回结果
func RpcBoolResultOK(data []byte, ok bool) bool {
	if ok && base.BytesToInt(data) == 1 {
		return true
	}
	return false
}

// SendActor 全局通用发送actor消息
// @actorName: 目标actor的名字
// @actorHandler: 目标actor的处理函数
// @data: 函数参数
func SendActor(actorName string, actorHandler string, data interface{}) {
	m := &gorpc.M{Data: data, Flag: true}
	gorpc.MGR.Send(actorName, actorHandler, m)
}

// BindGate 绑定网关信息
func BindGate(userid, gateid int) {
	gate.BindGate(userid, gateid)
}

// Publish sub/pub系统，只在本节点服务内生效
// 发布
// @key: 发布的key
// @value: 发布的值
func Publish(key string, data interface{}) { //发布
	igo := gorpc.MGR.GetRoutine("GateServer")
	if igo == nil {
		llog.Errorf("loumiao.Publish error, no gate actor: key=%s,value=%v", key, data)
		return
	}
	mm := &gorpc.MM{}
	mm.Id = key
	mm.Data = data
	SendActor("GateServer", "Publish", mm)
}

// Subscribe 订阅
// @name: 订阅者的igo
// @key: 订阅的key
// @handler: 订阅的actor处理函数,为""即为取消订阅
func Subscribe(igo gorpc.IGoRoutine, key string, call gorpc.HandlerFunc) { //订阅
	gateigo := gorpc.MGR.GetRoutine("GateServer")
	if gateigo == nil {
		llog.Errorf("loumiao.Subscribe error, no gate actor: key=%s", key)
		return
	}
	name := igo.GetName()
	var handler string
	if call != nil {
		handler = util.RpcFuncName(call)
		igo.Register(handler, call)
	} else {
		handler = ""
	}

	bitstream := base.NewBitStreamS(base.BitStrLen(name) + base.BitStrLen(key) + base.BitStrLen(handler) + 1)
	bitstream.WriteString(name)
	bitstream.WriteString(key)
	bitstream.WriteString(handler)
	SendActor("GateServer", "Subscribe", bitstream.GetBuffer())
}

// GetSocketNum 获取socket的连接数
func GetSocketNum() int {
	return gate.GetSocketNum()
}

// GetUserGate 获取玩家所属gate id
func GetUserGate(userId int) int {
	return gate.GetGateId(userId)
}

// GetServerUid 本节点服务uid
func GetServerUid() int {
	return config.SERVER_NODE_UID
}

// GetAreaId 服id
func GetAreaId() int {
	return config.NET_NODE_ID
}
