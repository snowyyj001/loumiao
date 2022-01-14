package etcf

import (
	"encoding/json"
	"fmt"
	"github.com/snowyyj001/loumiao/config"
	"github.com/snowyyj001/loumiao/define"
	"github.com/snowyyj001/loumiao/llog"
	"github.com/snowyyj001/loumiao/message"
	"github.com/snowyyj001/loumiao/msg"
	"github.com/snowyyj001/loumiao/network"
	"github.com/snowyyj001/loumiao/nodemgr"
	"github.com/snowyyj001/loumiao/timer"
	"sync"
	"time"
)

type hanlderNetFunc func(clientid int, buffer []byte)

var (
	Client        *EtcfClient
	handler_client_Map map[string]hanlderNetFunc
)

type ETKeyValue struct {
	Key string
	Value string
}

const(
	GET_VALUE_TIMEOUT         	= 3         //GetValue超时时间,秒
	WRITE_VALUE_TIMEOUT         = 30         //写操作超时时间,毫秒
	LEASE_TIME         			= 3000        //租约心跳时间,毫秒
	LEASE_ROUND         		= 2         //租约通信周期
)

type hanlderFunc func(string, string, bool)

func NewEtcf() *EtcfClient {
	if Client == nil {
		Client = &EtcfClient{}
	}
	Client.Init()
	Client.Start(false)

	Client.etcdKey = fmt.Sprintf("%s%d/%s", define.ETCD_NODEINFO, config.NET_NODE_ID, config.NET_GATE_SADDR)
	Client.statusKey = fmt.Sprintf("%s%d/%s", define.ETCD_NODESTATUS, config.NET_NODE_ID, config.NET_GATE_SADDR)

	return Client
}

type EtcfClient struct {

	pInnerService *network.ClientSocket

	leaseTime int64
	netStatus bool

	etcdKey   string
	statusKey string

	otherFunc sync.Map //[string]HanlderFunc
	mChanGetValue sync.Map //map[string]chan []KeyValue
	mChanAquireLock sync.Map //map[string]chan bool
}

//初始化
func (self *EtcfClient) Init() {
	llog.Infof("EtcfClient Init")

	handler_client_Map["C_CONNECT"] = innerClientConnect
	handler_client_Map["C_DISCONNECT"] = innerClientDisConnect
	handler_client_Map["LouMiaoAquireLock"] = innerClientLouMiaoAquireLock
	handler_client_Map["LouMiaoNoticeValue"] = innerClientLouMiaoNoticeValue
	handler_client_Map["LouMiaoGetValue"] = innerClientLouMiaoGetValue
	handler_client_Map["LouMiaoLease"] = innerClientLouMiaoLease
}

//连接etcf server
func (self *EtcfClient) Start(reconnect bool) {
	llog.Infof("EtcfClient Start: reconnect = %t", reconnect)

	client := new(network.ClientSocket)
	client.SetClientId(config.SERVER_NODE_UID)
	client.Init(config.Cfg.EtcdAddr[0])		//不支持集群，想要集群使用etcd
	client.SetConnectType(network.CHILD_CONNECT)
	client.BindPacketFunc(packetClientFunc)
	client.Uid = config.SERVER_NODE_UID

	self.pInnerService = client

	if client.Start() {
		llog.Infof("EtcfClient connected successed: %s ", client.GetSAddr())
	} else {
		if reconnect {
			llog.Errorf("EtcfClient connect failed: %s", client.GetSAddr())
		} else {	//首次就失败，不让启动
			llog.Fatalf("EtcfClient connect failed: %s", client.GetSAddr())
		}
	}
	if !reconnect {
		timer.NewTimer(LEASE_TIME, leaseCallBack, true)
	}
	Client.netStatus = client.GetState() == network.SSF_CONNECT
	if Client.netStatus {
		Client.leaseTime = 0
		if reconnect {
			llog.Debugf("etcf lease 尝试重新续租成功")
			PutNode()		//再次写入本服务信息
		}
	}
}

//写服务状态
func putStatus()  {
	val := fmt.Sprintf("%d:%t", nodemgr.OnlineNum, nodemgr.ServerEnabled)
	SetValue(Client.statusKey, val)
}

//写服务信息
func PutNode() {
	obj, _ := json.Marshal(&config.Cfg.NetCfg)
	PutService(Client.etcdKey, string(obj))
}

func leaseCallBack(dt int64) bool {
	Client.leaseTime++
	if Client.leaseTime > 2 {
		llog.Errorf("etcf lease 续租失败: leaseTime = %d", Client.leaseTime)
		Client.Start(true)
		return false
	}
	req := &msg.LouMiaoLease{}
	req.Uid = int32(config.SERVER_NODE_UID)

	buff, err := message.Pack(req)
	if err != nil {
		llog.Errorf("EtcfClient leaseCallBack error")
		return false
	}
	Client.pInnerService.Send(buff)
	return true
}

//设置服务
//@prefix: key值
//@val: 设置的value
func PutService(prefix, val string) {
	if !Client.netStatus {
		llog.Errorf("EtcfClient.PutService: prefix = %s, server not connected", prefix)
		return
	}
	req := &msg.LouMiaoPutValue{}
	req.Prefix = prefix
	req.Value = val
	req.Service = true

	buff, err := message.Pack(req)
	if err != nil {
		llog.Errorf("EtcfClient.SetValue: prefix = %s", prefix)
		return
	}
	Client.pInnerService.Send(buff)
}

//设置值
//@prefix: key值
//@val: 设置的value
func SetValue(prefix, val string) {
	if !Client.netStatus {
		llog.Errorf("EtcfClient.SetValue: prefix = %s, server not connected", prefix)
		return
	}
	req := &msg.LouMiaoPutValue{}
	req.Prefix = prefix
	req.Value = val
	req.Service = false

	buff, err := message.Pack(req)
	if err != nil {
		llog.Errorf("EtcfClient.SetValue: prefix = %s", prefix)
		return
	}
	Client.pInnerService.Send(buff)
}

//获取一个分布式锁,expire毫秒后会超时返回0
//@prefix: 锁key
//@expire: 超时时间,毫秒
//@retutn: 锁剩余时间
func AquireLock(key string, expire int) int {
	if !Client.netStatus {
		llog.Errorf("EtcfClient.AquireLock: key = %s, server not connected", key)
		return 0
	}
	req := &msg.LouMiaoAquireLock{}
	req.Prefix = key
	req.TimeOut = int32(expire)

	buff, err := message.Pack(req)
	if err != nil {
		llog.Errorf("EtcfClient.AquireLock: prefix = %s", key)
		return 0
	}
	Client.pInnerService.Send(buff)

	llog.Infof("EtcfClient.AquireLock: key = %s, expire = %d", key, expire)

	ch, ok := Client.mChanAquireLock.Load(key)
	if !ok {
		ch = make(chan int)
		Client.mChanAquireLock.Store(key, ch)
	}

	select {
	case left := <- ch.(chan int):
		return left
	case <-time.After(time.Duration(expire) * time.Millisecond):
		llog.Errorf("EtcfClient.AquireLock timeout: %s",  key)
		return 0
	}
}

//释放锁
func UnLock(key string) {
	Client.mChanAquireLock.Delete(key)

	if !Client.netStatus {
		llog.Errorf("EtcfClient.UnLock: key = %s, server not connected", key)
		return
	}
	req := &msg.LouMiaoReleaseLock{}
	req.Prefix = key

	buff, err := message.Pack(req)
	if err != nil {
		llog.Errorf("EtcfClient.UnLock: prefix = %s", key)
		return
	}
	Client.pInnerService.Send(buff)

	llog.Infof("EtcfClient.UnLock: key = %s", key)
}

//获取值
//@prefix: 监听key值
//@hanlder: key值变化回调
func GetOne(prefix string) string {
	if !Client.netStatus {
		llog.Errorf("EtcfClient.GetOne: prefix = %s, server not connected", prefix)
		return ""
	}
	req := &msg.LouMiaoGetValue{}
	req.Prefix = prefix

	buff, err := message.Pack(req)
	if err != nil {
		llog.Errorf("EtcfClient.GetOne: prefix = %s", prefix)
		return ""
	}
	Client.pInnerService.Send(buff)

	ch, ok := Client.mChanGetValue.Load(prefix)
	if !ok {
		ch = make(chan []ETKeyValue)
		Client.mChanGetValue.Store(prefix, ch)
	}

	select {
	case kvs := <- ch.(chan []ETKeyValue):
		if len(kvs) > 0 {
			for _, v:= range kvs {
				if v.Key == prefix {
					return v.Value
				}
			}
		}
		return ""
	case <-time.After(GET_VALUE_TIMEOUT * time.Second):
		llog.Errorf("EtcfClient.GetOne timeout: %s",  prefix)
		return ""
	}
}

//获取值
//@prefix: 监听key值
//@hanlder: key值变化回调
func GetAll(prefix string) []string {
	if !Client.netStatus {
		llog.Errorf("EtcfClient.GetAll: prefix = %s, server not connected", prefix)
		return nil
	}

	req := &msg.LouMiaoGetValue{}
	req.Prefix = prefix

	buff, err := message.Pack(req)
	if err != nil {
		llog.Errorf("EtcfClient.GetAll: prefix = %s", prefix)
		return nil
	}
	Client.pInnerService.Send(buff)

	ch, ok := Client.mChanGetValue.Load(prefix)
	if !ok {
		ch = make(chan []ETKeyValue)
		Client.mChanGetValue.Store(prefix, ch)
	}

	select {
	case kvs := <- ch.(chan []ETKeyValue):
		if len(kvs) > 0 {
			arr := make([]string, 0)
			for _, v:= range kvs {
				arr = append(arr, v.Value)
			}
			return arr
		}
		return nil
	case <-time.After(GET_VALUE_TIMEOUT * time.Second):
		llog.Errorf("EtcfClient.GetAll timeout: %s",  prefix)
		return nil
	}

}

//监听key
//@prefix: 监听key值
//@hanlder: key值变化回调
func WatchKey(prefix string, hanlder hanlderFunc) {
	if !Client.netStatus {
		llog.Errorf("EtcfClient.WatchKey: prefix = %s, server not connected", prefix)
		return
	}

	Client.otherFunc.Store(prefix, hanlder)

	req := &msg.LouMiaoWatchKey{}
	req.Prefix = prefix
	req.Opcode = msg.LouMiaoWatchKey_ADD

	buff, err := message.Pack(req)
	if err != nil {
		llog.Errorf("EtcfClient.WatchKey: prefix = %s", prefix)
		return
	}
	Client.pInnerService.Send(buff)
}

//删除key值
//@prefix: 监听key值
func RemoveKey(prefix string) {
	Client.otherFunc.Delete(prefix)

	if !Client.netStatus {
		llog.Errorf("EtcfClient.WatchKey: prefix = %s, server not connected", prefix)
		return
	}
	req := &msg.LouMiaoWatchKey{}
	req.Prefix = prefix
	req.Opcode = msg.LouMiaoWatchKey_DEL

	buff, err := message.Pack(req)
	if err != nil {
		llog.Errorf("EtcfClient.RemoveKey: prefix = %s", prefix)
		return
	}
	Client.pInnerService.Send(buff)
}

//client connect
func innerClientConnect( socketId int, data []byte) {
	llog.Debugf("etcf client innerConnect: %d", socketId)
}

//client disconnect
func innerClientDisConnect( socketId int, data []byte) {
	llog.Debugf("etcf client innerDisConnect: %d", socketId)
	Client.netStatus = false
}

//aquire lock
func innerClientLouMiaoAquireLock( socketId int, data []byte) {
	req := &msg.LouMiaoAquireLock{}
	if message.UnPack(req, data) != nil {
		return
	}
	llog.Debugf("innerClientLouMiaoAquireLock: req = %v", req)

	ch, ok := Client.mChanAquireLock.Load(req.Prefix)
	if !ok {
		llog.Errorf("innerClientLouMiaoAquireLock: prefix = %s", req.Prefix)
		return
	}

	select {
	case ch.(chan int) <- int(req.TimeOut):
	case <-time.After(WRITE_VALUE_TIMEOUT * time.Millisecond):
		llog.Errorf("innerClientLouMiaoAquireLock timeout: %s",  req.Prefix)
	}
}

//value changed
func innerClientLouMiaoNoticeValue(socketId int, data []byte) {
	req := new(msg.LouMiaoNoticeValue)
	if message.UnPack(req, data) != nil {
		return
	}

	llog.Debugf("innerClientLouMiaoNoticeValue: req = %v", req)

	cb, ok := Client.otherFunc.Load(req.Prefix)
	if ok {
		if len(req.Value) > 0 {
			cb.(hanlderFunc)(req.Prefix, req.Value, true)
		} else {
			cb.(hanlderFunc)(req.Prefix, "", false)
		}
	} else {
		llog.Errorf("innerClientLouMiaoNoticeValue prefix no handler: %s", req.Prefix)
	}
}

//get value
func innerClientLouMiaoGetValue(socketId int, data []byte) {
	req := new(msg.LouMiaoGetValue)
	if message.UnPack(req, data) != nil {
		return
	}

	llog.Debugf("innerClientLouMiaoGetValue: req = %v", req)

	ch, ok := Client.mChanGetValue.Load(req.Prefix)
	if !ok {
		llog.Errorf("innerClientLouMiaoGetValue no get owner: %s", req.Prefix)
		return
	}
	chval := ch.(chan []ETKeyValue)
	arr := make([]ETKeyValue, 0)
	for i:=0; i < len(req.Prefixs); i++ {
		kv := ETKeyValue{Key: req.Prefixs[i], Value: req.Values[i]}
		arr = append(arr, kv)
	}

	select {
	case chval <- arr:
	case <-time.After(WRITE_VALUE_TIMEOUT * time.Millisecond):
		llog.Errorf("innerClientLouMiaoGetValue timeout: %s",  req.Prefix)
	}
}

//lease
func innerClientLouMiaoLease(socketId int, data []byte) {
	Client.leaseTime = 0

	putStatus()
}


//goroutine unsafe
//net msg handler,this func belong to socket's goroutine
func packetClientFunc(socketid int, buff []byte, nlen int) bool {
	//llog.Debugf("packetFunc: socketid=%d, bufferlen=%d", socketid, nlen)
	_, name, buffbody, err := message.UnPackHead(buff, nlen)
	//llog.Debugf("packetFunc  %s %v", name, pm)
	if nil != err {
		llog.Errorf("packetClientFunc Decode error: %s", err.Error())
		//This.closeClient(socketid)
	} else {
		handler, ok := handler_client_Map[name]
		if ok {
			go handler(socketid, buffbody)
		} else {
			llog.Errorf("packetClientFunc handler is nil, drop it[%s]", name)
		}

	}
	return true
}
