package etcf

import (
	"encoding/json"
	"fmt"
	"github.com/snowyyj001/loumiao/lconfig"
	"github.com/snowyyj001/loumiao/ldefine"
	"github.com/snowyyj001/loumiao/llog"
	"github.com/snowyyj001/loumiao/lutil"
	"github.com/snowyyj001/loumiao/message"
	"github.com/snowyyj001/loumiao/network"
	"github.com/snowyyj001/loumiao/nodemgr"
	"github.com/snowyyj001/loumiao/pbmsg"
	"github.com/snowyyj001/loumiao/timer"
	"strings"
	"sync"
	"time"
)

type HandlerNetFunc func(clientid int, buffer []byte)

var (
	Client             *EtcfClient
	handler_client_Map map[string]HandlerNetFunc
)

type ETKeyValue struct {
	Key   string
	Value string
}

const (
	GET_VALUE_TIMEOUT   = 3    //GetValue超时时间,秒
	WRITE_VALUE_TIMEOUT = 30   //写操作超时时间,毫秒
	LEASE_TIME          = 1000 //租约心跳时间,毫秒
	PUT_STATUS_TIME     = 2    //刷新服务状态时间

	LEASE_SERVER_TIMEOUT = 5 //租约超时时间，秒
)

func newEtcfClient() (*EtcfClient, error) {
	if Client == nil {
		Client = &EtcfClient{}
	}
	Client.Init()
	err := Client.Start(false)

	Client.etcdKey = fmt.Sprintf("%s%d/%s", ldefine.ETCD_NODEINFO, lconfig.NET_NODE_ID, lconfig.NET_GATE_SADDR)
	Client.statusKey = fmt.Sprintf("%s%d/%s", ldefine.ETCD_NODESTATUS, lconfig.NET_NODE_ID, lconfig.NET_GATE_SADDR)

	return Client, err
}

type EtcfClient struct {
	pInnerService *network.ClientSocket

	leaseTime          int64
	netStatus          bool
	leaseCallBackTimes int

	etcdKey   string
	statusKey string

	otherFunc       sync.Map //[string]HandlerFunc
	mChanGetValue   sync.Map //map[string]chan []KeyValue
	mChanAquireLock sync.Map //map[string]chan bool
}

// 初始化
func (self *EtcfClient) Init() {
	llog.Infof("EtcfClient Init")

	handler_client_Map = make(map[string]HandlerNetFunc)

	handler_client_Map["C_CONNECT"] = innerClientConnect
	handler_client_Map["C_DISCONNECT"] = innerClientDisConnect
	handler_client_Map["LouMiaoAquireLock"] = innerClientLouMiaoAquireLock
	handler_client_Map["LouMiaoNoticeValue"] = innerClientLouMiaoNoticeValue
	handler_client_Map["LouMiaoGetValue"] = innerClientLouMiaoGetValue
	handler_client_Map["LouMiaoLease"] = innerClientLouMiaoLease
}

// 连接etcf server
func (self *EtcfClient) Start(reconnect bool) error {
	llog.Infof("EtcfClient Start: reconnect = %t", reconnect)
	if reconnect {
		self.pInnerService.Stop()
	}
	client := new(network.ClientSocket)
	client.SetClientId(lconfig.SERVER_NODE_UID)
	client.Init(lconfig.Cfg.EtcdAddr[0]) //不支持集群，想要集群使用etcd
	client.SetConnectType(network.CHILD_CONNECT)
	client.BindPacketFunc(packetClientFunc)
	client.Uid = lconfig.SERVER_NODE_UID

	self.pInnerService = client

	if client.Start() {
		llog.Infof("EtcfClient connected successed: %s ", client.GetSAddr())
	} else {
		return fmt.Errorf("EtcfClient connect failed: %s", client.GetSAddr())
	}
	if !reconnect {
		timer.NewTimer(LEASE_TIME, leaseCallBack, true)
	}
	Client.netStatus = client.GetState() == network.SSF_CONNECT
	return nil
}

// 重新监听key
func ReWatchKey() {
	Client.otherFunc.Range(func(key, value interface{}) bool {
		req := &pbmsg.LouMiaoWatchKey{}
		req.Prefix = key.(string)
		req.Opcode = pbmsg.LouMiaoWatchKey_ADD

		buff, _ := message.EncodeProBuff(0, "LouMiaoWatchKey", req)
		Client.pInnerService.Send(buff)
		return true
	})
}

// 写服务状态
func PutStatus() error {
	val := fmt.Sprintf("%d:%t", nodemgr.OnlineNum, nodemgr.ServerEnabled)
	return SetValue(Client.statusKey, val)
}

// 写服务信息
func PutNode() error {
	obj, _ := json.Marshal(&lconfig.Cfg.NetCfg)
	return PutService(Client.etcdKey, string(obj))
}

func leaseCallBack(dt int64) bool {
	Client.leaseTime++
	//llog.Debugf("leaseCallBack: %d", Client.leaseTime)
	if Client.leaseTime > LEASE_SERVER_TIMEOUT {
		llog.Errorf("etcf lease 续租失败: leaseTime = %d", Client.leaseTime)
		llog.Debug("尝试重新续租")
		Client.Start(true)
		if Client.netStatus {
			Client.leaseTime = 0
			llog.Debugf("etcf lease 尝试重新续租成功")
			PutNode()    //再次写入本服务信息
			ReWatchKey() //重新注册要监听的key
		}
	}
	if !Client.netStatus {
		return true
	}
	req := &pbmsg.LouMiaoLease{}
	req.Uid = int32(lconfig.SERVER_NODE_UID)

	buff, _ := message.EncodeProBuff(0, "LouMiaoLease", req)
	Client.pInnerService.Send(buff)
	return true
}

// 设置服务
// @prefix: key值
// @val: 设置的value
func PutService(prefix, val string) error {
	if !Client.netStatus {
		return fmt.Errorf("EtcfClient.PutService: prefix = %s, server not connected", prefix)
	}
	req := &pbmsg.LouMiaoPutValue{}
	req.Prefix = prefix
	req.Value = val
	req.Lease = LEASE_TIME

	buff, _ := message.EncodeProBuff(0, "LouMiaoPutValue", req)
	Client.pInnerService.Send(buff)

	return nil
}

// 设置值
// @prefix: key值
// @val: 设置的value
func SetValue(prefix, val string) error {
	if !Client.netStatus {
		return fmt.Errorf("EtcfClient.SetValue: prefix = %s, server not connected", prefix)
	}
	req := &pbmsg.LouMiaoPutValue{}
	req.Prefix = prefix
	req.Value = val
	req.Lease = 0

	buff, _ := message.EncodeProBuff(0, "LouMiaoPutValue", req)
	Client.pInnerService.Send(buff)

	return nil
}

// AquireLock 获取一个分布式锁,expire毫秒后会超时返回0
// @prefix: 锁key
// @expire: 超时时间,毫秒
// @retutn: 锁剩余时间
func AquireLock(key string, expire int) int {
	if !Client.netStatus {
		llog.Errorf("EtcfClient.AquireLock: key = %s, server not connected", key)
		return 0
	}
	req := &pbmsg.LouMiaoAquireLock{}
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
	case left := <-ch.(chan int):
		return left
	case <-time.After(time.Duration(expire) * time.Millisecond):
		llog.Warningf("EtcfClient.AquireLock timeout: %s", key)
		return 0
	}
}

// 释放锁
func UnLock(key string) {
	Client.mChanAquireLock.Delete(key)

	if !Client.netStatus {
		llog.Errorf("EtcfClient.UnLock: key = %s, server not connected", key)
		return
	}
	req := &pbmsg.LouMiaoReleaseLock{}
	req.Prefix = key

	buff, _ := message.EncodeProBuff(0, "LouMiaoReleaseLock", req)
	Client.pInnerService.Send(buff)

	llog.Infof("EtcfClient.UnLock: key = %s", key)
}

// AquireLeader 选举leader，所有参与选举的人使用相同的value和prefix，leader负责设置value
// 同redisdb的AquireLeader
// @prefix: 选举区分标识
// @value: 本次选举的值，每次发起选举，value应该和上次选举时的value不同
// @return: 是否是leader
func AquireLeader(key string, value string) (isleader bool) {
	if !Client.netStatus {
		llog.Errorf("EtcfClient.AquireLeader: key = %s, server not connected", key)
		return false
	}
	req := &pbmsg.LouMiaoAquireLock{}
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
	case left := <-ch.(chan int):
		return left > 0
	case <-time.After(time.Duration(1000) * time.Millisecond):
		llog.Warningf("EtcfClient.AquireLeader timeout: %s", key)
		return false
	}
}

// GetOne 获取值
// @prefix: key值
func GetOne(prefix string) (string, error) {
	if !Client.netStatus {
		return "", fmt.Errorf("EtcfClient.GetOne: prefix = %s, server not connected", prefix)
	}

	req := &pbmsg.LouMiaoGetValue{}
	req.Prefix = prefix
	buff, _ := message.EncodeProBuff(0, "LouMiaoGetValue", req)
	ch, ok := Client.mChanGetValue.Load(prefix)
	if !ok {
		ch = make(chan []ETKeyValue)
		Client.mChanGetValue.Store(prefix, ch)
	}

	Client.pInnerService.Send(buff)

	select {
	case kvs := <-ch.(chan []ETKeyValue):
		if len(kvs) > 0 {
			for _, v := range kvs {
				if v.Key == prefix {
					return v.Value, nil
				}
			}
		}
		return "", nil
	case <-time.After(GET_VALUE_TIMEOUT * time.Second):
		return "", fmt.Errorf("EtcfClient.GetOne timeout: %s", prefix)
	}
}

// 获取值
// @prefix: key值
func GetAll(prefix string) ([]string, error) {
	if !Client.netStatus {
		return nil, fmt.Errorf("EtcfClient.GetAll: prefix = %s, server not connected", prefix)
	}

	req := &pbmsg.LouMiaoGetValue{}
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
	case kvs := <-ch.(chan []ETKeyValue):
		if len(kvs) > 0 {
			arr := make([]string, 0)
			for _, v := range kvs {
				arr = append(arr, v.Value)
			}
			return arr, nil
		}
		return nil, nil
	case <-time.After(GET_VALUE_TIMEOUT * time.Second):
		return nil, fmt.Errorf("EtcfClient.GetAll timeout: %s", prefix)
	}

}

// 监听key
// @prefix: 监听key值
// @handler: key值变化回调
func WatchKey(prefix string, handler func(string, string, bool)) bool {
	if !Client.netStatus {
		llog.Errorf("EtcfClient.WatchKey: prefix = %s, server not connected", prefix)
		return false
	}

	Client.otherFunc.Store(prefix, handler)

	req := &pbmsg.LouMiaoWatchKey{}
	req.Prefix = prefix
	req.Opcode = pbmsg.LouMiaoWatchKey_ADD
	req.Uid = int32(lconfig.SERVER_NODE_UID)

	buff, _ := message.EncodeProBuff(0, "LouMiaoWatchKey", req)
	Client.pInnerService.Send(buff)
	return true
}

// 删除key值
// @prefix: 监听key值
func RemoveKey(prefix string) {
	Client.otherFunc.Delete(prefix)

	if !Client.netStatus {
		llog.Errorf("EtcfClient.WatchKey: prefix = %s, server not connected", prefix)
		return
	}
	req := &pbmsg.LouMiaoWatchKey{}
	req.Prefix = prefix
	req.Opcode = pbmsg.LouMiaoWatchKey_DEL
	req.Uid = int32(lconfig.SERVER_NODE_UID)

	buff, _ := message.EncodeProBuff(0, "LouMiaoWatchKey", req)
	Client.pInnerService.Send(buff)
}

// client connect
func innerClientConnect(socketId int, data []byte) {
	llog.Debugf("etcf client innerConnect: %d", socketId)
}

// client disconnect
func innerClientDisConnect(socketId int, data []byte) {
	llog.Debugf("etcf client innerDisConnect: %d", socketId)
	Client.netStatus = false
}

// aquire lock
func innerClientLouMiaoAquireLock(socketId int, data []byte) {
	req := &pbmsg.LouMiaoAquireLock{}
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
		llog.Warningf("innerClientLouMiaoAquireLock timeout: %s", req.Prefix)
	}
}

// value changed
func innerClientLouMiaoNoticeValue(socketId int, data []byte) {
	req := new(pbmsg.LouMiaoNoticeValue)
	if message.UnPackProto(req, data) != nil {
		return
	}

	//llog.Debugf("innerClientLouMiaoNoticeValue: req = %v", req)
	Client.otherFunc.Range(func(key, value interface{}) bool {
		if strings.HasPrefix(req.Prefix, key.(string)) {
			cb := value.(func(string, string, bool))
			if len(req.Value) > 0 {
				cb(req.Prefix, req.Value, true)
			} else {
				cb(req.Prefix, "", false)
			}
		}
		return true
	})
}

// get value
func innerClientLouMiaoGetValue(socketId int, data []byte) {
	req := new(pbmsg.LouMiaoGetValue)
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
	for i := 0; i < len(req.Prefixs); i++ {
		kv := ETKeyValue{Key: req.Prefixs[i], Value: req.Values[i]}
		arr = append(arr, kv)
	}

	select {
	case chval <- arr:
	case <-time.After(WRITE_VALUE_TIMEOUT * time.Millisecond):
		llog.Errorf("innerClientLouMiaoGetValue timeout: %s", req.Prefix)
	}
}

// lease
func innerClientLouMiaoLease(socketId int, data []byte) {
	//llog.Debugf("innerClientLouMiaoLease: socketId = %d",socketId)
	Client.leaseTime = 0
	Client.leaseCallBackTimes++
	if Client.leaseCallBackTimes >= PUT_STATUS_TIME {
		PutStatus()
		Client.leaseCallBackTimes = 0
	}
}

// goroutine unsafe
// net msg handler,this func belong to socket's goroutine
func packetClientFunc(socketid int, buff []byte, nlen int) error {
	//llog.Debugf("packetFunc: socketid=%d, bufferlen=%d", socketid, nlen)
	_, name, buffbody, err := message.UnPackHead(buff, nlen)
	//llog.Debugf("packetFunc  %s", name)
	if nil != err {
		return fmt.Errorf("packetClientFunc Decode error: %s", err.Error())
		//This.closeClient(socketid)
	} else {
		handler, ok := handler_client_Map[name]
		if ok {
			lutil.Go(func() {
				handler(socketid, buffbody)
			})
		} else {
			llog.Errorf("packetClientFunc handler is nil, drop it[%s]", name)
		}

	}
	return nil
}
