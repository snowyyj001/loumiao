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
	"runtime"
	"strings"
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
	LEASE_TIME         			= 1000        //租约心跳时间,毫秒
	PUT_STATUS_TIME				= 1			//刷新服务状态时间
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
	leaseCallBackTimes int

	etcdKey   string
	statusKey string

	otherFunc sync.Map //[string]HanlderFunc
	mChanGetValue sync.Map //map[string]chan []KeyValue
	mChanAquireLock sync.Map //map[string]chan bool
}

//初始化
func (self *EtcfClient) Init() {
	llog.Infof("EtcfClient Init")

	handler_client_Map = make(map[string]hanlderNetFunc)

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
		if !reconnect { //首次就失败，不让启动
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
			ReWatchKey()		//重新注册要监听的key
		}
	}
}

//重新监听key
func ReWatchKey() {
	Client.otherFunc.Range(func(key, value interface{}) bool {
		req := &msg.LouMiaoWatchKey{}
		req.Prefix = key.(string)
		req.Opcode = msg.LouMiaoWatchKey_ADD

		buff, _ := message.EncodeProBuff(0, "LouMiaoWatchKey", req)
		Client.pInnerService.Send(buff)
		return true
	})
}

//写服务状态
func PutStatus()  {
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
	//llog.Debugf("leaseCallBack: %d", Client.leaseTime)
	if Client.leaseTime > 5 {
		llog.Errorf("etcf lease 续租失败: leaseTime = %d", Client.leaseTime)
		Client.Start(true)
	}
	if !Client.netStatus {
		return true
	}
	req := &msg.LouMiaoLease{}
	req.Uid = int32(config.SERVER_NODE_UID)

	buff, _ := message.EncodeProBuff(0, "LouMiaoLease", req)
	Client.pInnerService.Send(buff)
	return true
}

//设置服务
//@prefix: key值
//@val: 设置的value
func PutService(prefix, val string) bool {
	if !Client.netStatus {
		llog.Errorf("EtcfClient.PutService: prefix = %s, server not connected", prefix)
		return false
	}
	req := &msg.LouMiaoPutValue{}
	req.Prefix = prefix
	req.Value = val
	req.Lease = LEASE_TIME

	buff, _ := message.EncodeProBuff(0, "LouMiaoPutValue", req)
	Client.pInnerService.Send(buff)

	return true
}

//设置值
//@prefix: key值
//@val: 设置的value
func SetValue(prefix, val string) bool {
	if !Client.netStatus {
		llog.Errorf("EtcfClient.SetValue: prefix = %s, server not connected", prefix)
		return false
	}
	req := &msg.LouMiaoPutValue{}
	req.Prefix = prefix
	req.Value = val
	req.Lease = 0

	buff, _ := message.EncodeProBuff(0, "LouMiaoPutValue", req)
	Client.pInnerService.Send(buff)

	return true
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

	buff, _ := message.EncodeProBuff(0, "LouMiaoAquireLock", req)
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
		llog.Warningf("EtcfClient.AquireLock timeout: %s",  key)
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

	buff, _ := message.EncodeProBuff(0, "LouMiaoReleaseLock", req)
	Client.pInnerService.Send(buff)

	llog.Infof("EtcfClient.UnLock: key = %s", key)
}

//选举leader，所有参与选举的人使用相同的value和prefix，leader负责设置value
//同redisdb的AquireLeader
//@prefix: 选举区分标识
//@value: 本次选举的值，每次发起选举，value应该和上次选举时的value不同
//@return: 是否是leader
func AquireLeader(key string, value string) (isleader bool) {
	if !Client.netStatus {
		llog.Errorf("EtcfClient.AquireLeader: key = %s, server not connected", key)
		return false
	}
	req := &msg.LouMiaoAquireLock{}
	req.Prefix = key
	req.TimeOut = 1000
	req.Value = value

	buff, _ := message.EncodeProBuff(0, "LouMiaoAquireLock", req)
	Client.pInnerService.Send(buff)

	llog.Infof("EtcfClient.AquireLeader: key = %s, value = %s", key, value)
	ch, ok := Client.mChanAquireLock.Load(key)
	if !ok {
		ch = make(chan int)
		Client.mChanAquireLock.Store(key, ch)
	}
	select {
	case left := <- ch.(chan int):
		return left > 0
	case <-time.After(time.Duration(1000) * time.Millisecond):
		llog.Warningf("EtcfClient.AquireLeader timeout: %s",  key)
		return false
	}
}

//获取值
//@prefix: key值
func GetOne(prefix string) string {
	if !Client.netStatus {
		llog.Errorf("EtcfClient.GetOne: prefix = %s, server not connected", prefix)
		return ""
	}

	req := &msg.LouMiaoGetValue{}
	req.Prefix = prefix
	buff, _ := message.EncodeProBuff(0, "LouMiaoGetValue", req)
	ch, ok := Client.mChanGetValue.Load(prefix)
	if !ok {
		ch = make(chan []ETKeyValue)
		Client.mChanGetValue.Store(prefix, ch)
	}

	Client.pInnerService.Send(buff)

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
//@prefix: key值
func GetAll(prefix string) []string {
	if !Client.netStatus {
		llog.Errorf("EtcfClient.GetAll: prefix = %s, server not connected", prefix)
		return nil
	}

	req := &msg.LouMiaoGetValue{}
	req.Prefix = prefix
	req.Prefixs = append(req.Prefixs, "m")

	buff, _ := message.EncodeProBuff(0, "LouMiaoGetValue", req)
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
func WatchKey(prefix string, hanlder hanlderFunc) bool {
	if !Client.netStatus {
		llog.Errorf("EtcfClient.WatchKey: prefix = %s, server not connected", prefix)
		return false
	}

	Client.otherFunc.Store(prefix, hanlder)

	req := &msg.LouMiaoWatchKey{}
	req.Prefix = prefix
	req.Opcode = msg.LouMiaoWatchKey_ADD

	buff, _ := message.EncodeProBuff(0, "LouMiaoWatchKey", req)
	Client.pInnerService.Send(buff)
	return true
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

	buff, _ := message.EncodeProBuff(0, "LouMiaoWatchKey", req)
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
	if message.UnPackProto(req, data) != nil {
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
		llog.Warningf("innerClientLouMiaoAquireLock timeout: %s",  req.Prefix)
	}
}

//value changed
func innerClientLouMiaoNoticeValue(socketId int, data []byte) {
	req := new(msg.LouMiaoNoticeValue)
	if message.UnPackProto(req, data) != nil {
		return
	}

	//llog.Debugf("innerClientLouMiaoNoticeValue: req = %v", req)
	Client.otherFunc.Range(func(key, value interface{}) bool {
		if strings.HasPrefix(req.Prefix, key.(string)) {
			cb := value.(hanlderFunc)
			if len(req.Value) > 0 {
				cb(req.Prefix, req.Value, true)
			} else {
				cb(req.Prefix, "", false)
			}
		}
		return true
	})
}

//get value
func innerClientLouMiaoGetValue(socketId int, data []byte) {
	req := new(msg.LouMiaoGetValue)
	if message.UnPackProto(req, data) != nil {
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
	//llog.Debugf("innerClientLouMiaoLease: socketId = %d",socketId)
	Client.leaseTime = 0
	Client.leaseCallBackTimes++
	if Client.leaseCallBackTimes >= PUT_STATUS_TIME {
		PutStatus()
		Client.leaseCallBackTimes = 0
	}
}


//goroutine unsafe
//net msg handler,this func belong to socket's goroutine
func packetClientFunc(socketid int, buff []byte, nlen int) bool {
	//llog.Debugf("packetFunc: socketid=%d, bufferlen=%d", socketid, nlen)
	_, name, buffbody, err := message.UnPackHead(buff, nlen)
	//llog.Debugf("packetFunc  %s", name)
	if nil != err {
		llog.Errorf("packetClientFunc Decode error: %s", err.Error())
		//This.closeClient(socketid)
	} else {
		handler, ok := handler_client_Map[name]
		if ok {
			go func() {
				defer func() {
					if r := recover(); r != nil {
						buf := make([]byte, 2048)
						l := runtime.Stack(buf, false)
						llog.Errorf("EtcfClient packetClientFunc:  %s", buf[:l])
					}
				}()
				handler(socketid, buffbody)
			}()
		} else {
			llog.Errorf("packetClientFunc handler is nil, drop it[%s]", name)
		}

	}
	return true
}
