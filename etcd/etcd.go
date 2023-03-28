package etcd

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/snowyyj001/loumiao/lconfig"
	"github.com/snowyyj001/loumiao/ldefine"
	"github.com/snowyyj001/loumiao/lutil"
	"sync"
	"time"

	"github.com/snowyyj001/loumiao/etcf"
	"github.com/snowyyj001/loumiao/llog"
	"github.com/snowyyj001/loumiao/nodemgr"

	"github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/clientv3/concurrency"
	"github.com/coreos/etcd/mvcc/mvccpb"
)

const (
	useEtcf = true //使用自己内部的etcf替代方案，不支持集群
)

var (
	This   *clientv3.Client
	Client IClientDis
)

func NewClient() error {
	if useEtcf {
		cli, err := etcf.NewEtcf()
		if err == nil {
			Client = cli
		}
		return err
	} else {
		cli, err := NewEtcd()
		if err == nil {
			Client = cli
		}
		return err
	}
}

type IClientDis interface {
	PutStatus() error
	PutNode() error
	WatchCommon(prefix string, handler func(string, string, bool)) error
	RevokeLease() error
	SetValue(prefix, val string) error
	GetOne(key string) (string, error)
	GetAll(key string) ([]string, error)
}

type IEtcdBase interface {
	Put(key string, val string, withlease bool) error
	Delete(key string) error
	Get(key string) (*clientv3.GetResponse, error)
	GetClient() *clientv3.Client
	SetLeasefunc(call func(bool))
	SetLease(timeNum int64, keepalive bool) error
	RevokeLease() error
}

type EtcdBase struct {
	client        *clientv3.Client
	lease         clientv3.Lease
	leaseResp     *clientv3.LeaseGrantResponse
	canclefunc    func()
	leasefunc     func(bool)
	keepAliveChan <-chan *clientv3.LeaseKeepAliveResponse
}

func (self *EtcdBase) GetClient() *clientv3.Client {
	return self.client
}

// 设置租约回调
func (self *EtcdBase) SetLeasefunc(call func(bool)) {
	self.leasefunc = call
}

// 设置value
func (self *EtcdBase) Put(key string, val string, withlease bool) error {
	if withlease {
		_, err := self.client.Put(context.TODO(), key, val, clientv3.WithLease(self.leaseResp.ID))
		return err
	} else {
		_, err := self.client.Put(context.TODO(), key, val)
		return err
	}
}

// 删除value
func (self *EtcdBase) Delete(key string) error {
	llog.Debugf("etcd delete: %s", key)
	_, err := self.client.Delete(context.TODO(), key)
	return err
}

// 获得value
func (self *EtcdBase) Get(key string) (*clientv3.GetResponse, error) {
	gresp, err := self.client.Get(context.TODO(), key)
	return gresp, err
}

func (self *EtcdBase) GetOne(key string) (string, error) {
	resp, err := self.Get(key)
	if err != nil {
		return "", err
	}
	if resp == nil {
		return "", nil
	}
	for i := range resp.Kvs {
		if string(resp.Kvs[i].Key) == key {
			return string(resp.Kvs[i].Value), nil
		}
	}
	return "", nil
}

func (self *EtcdBase) GetAll(key string) ([]string, error) {
	resp, err := self.Get(key)
	if err != nil {
		return []string{}, err
	}
	if resp == nil {
		return []string{}, nil
	}
	rets := []string{}
	for i := range resp.Kvs {
		rets = append(rets, string(resp.Kvs[i].Value))
	}
	return rets, nil
}

// 设置租约
// @timeNum: 过期时间
// @keepalive: 是否自动续约
func (self *EtcdBase) SetLease(timeNum int64, keepalive bool) error {
	if self.lease != nil {
		return fmt.Errorf("lease error")
	}
	lease := clientv3.NewLease(self.client)

	//设置租约时间
	leaseResp, err := lease.Grant(context.TODO(), timeNum)
	if err != nil {
		llog.Warning("lease.Grant error")
		return err
	}

	//设置续租
	if keepalive {
		ctx, cancelFunc := context.WithCancel(context.TODO())
		leaseRespChan, err := lease.KeepAlive(ctx, leaseResp.ID)
		if err != nil {
			llog.Warning("lease.KeepAlive error")
			lease.Close()
			return err
		}
		self.canclefunc = cancelFunc
		self.keepAliveChan = leaseRespChan

		lutil.Go(func() {
			self.listenLeaseRespChan()
		})
	}

	self.lease = lease
	self.leaseResp = leaseResp

	return nil
}

// 监听 续租情况
func (self *EtcdBase) listenLeaseRespChan() {
	for {
		select {
		case leaseKeepResp := <-self.keepAliveChan:
			if leaseKeepResp == nil {
				llog.Warningf("EtcdBase.listenLeaseRespChan: 续租功能已经关闭")
				self.lease.Close()
				self.lease = nil
				if self.leasefunc != nil {
					self.leasefunc(false)
				}
				return
			} else {
				//fmt.Printf("续租成功\n")
				if self.leasefunc != nil {
					self.leasefunc(true)
				}
			}
		}
	}
}

// 撤销租约
func (self *EtcdBase) RevokeLease() error {
	if self.leaseResp == nil {
		return fmt.Errorf("RevokeLease: lease has already been ewvoked")
	}
	self.canclefunc()
	self.leasefunc = nil //主动撤销不再回调
	_, err := self.lease.Revoke(context.TODO(), self.leaseResp.ID)
	self.leaseResp = nil
	return err
}

// 获取一个分布式锁,expire毫秒后会超时返回nil
// @prefix: 锁key
// @expire: 超时时间,毫秒
func AquireLock(prefix string, expire int) *concurrency.Mutex {
	if This == nil {
		return nil
	}
	var session *concurrency.Session
	session, err := concurrency.NewSession(This, concurrency.WithTTL(10))
	if err != nil {
		llog.Error("EtcdBase Lock NewSession failed " + err.Error())
		return nil
	}
	m := concurrency.NewMutex(session, prefix)
	// 获取锁使用context.TODO()会一直获取锁直到获取成功
	// 如果这里使用context.WithTimeout(context.TODO(), expire*time.Second)
	// 表示获取锁expire秒如果没有获取成功则返回error
	//if err = m.Lock(context.TODO()); err != nil {
	ct, _ := context.WithTimeout(context.TODO(), time.Duration(expire)*time.Millisecond)
	if err = m.Lock(ct); err != nil {
		//llog.Error("EtcdBase Lock NewMutex Lock lock failed " + err.Error())
		session.Close()
		return nil
	}
	//注意在获取锁后要调用该函数在session的租约到期后才会释放锁
	//session.Orphan()
	return m
}

func UnLock(key string, lockval *concurrency.Mutex) {
	if lockval != nil {
		lockval.Unlock(context.TODO())
	}
}

// 选举leader，所有参与选举的人使用相同的value和prefix，leader负责设置value
// @prefix: 选举区分标识
// @value: 本次选举的值，每次发起选举，value应该和上次选举时的value不同
func AquireLeader(prefix string, value string) (isleader bool) {
	isleader = false
	mt := AquireLock(prefix, 200)
	gresp, err := This.Get(context.TODO(), prefix)
	if err != nil {
		return
	}
	if len(gresp.Kvs) == 0 { //没有值
		This.Put(context.TODO(), prefix, value) //set value，就是标记我是本次选举leader
		isleader = true
	} else {
		nowvalue := string(gresp.Kvs[0].Value)
		if value != nowvalue { //还未被设置该值
			This.Put(context.TODO(), prefix, value) //set value，就是标记我是本次选举leader
			isleader = true
		}
	}
	UnLock(prefix, mt)
	return
}

// /////////////////////////////////////////////////////////
// 服发现
type ClientDis struct {
	EtcdBase
	otherFunc sync.Map //[string]HandlerFunc

	etcdKey   string
	statusKey string
}

func (self *ClientDis) watchFuc(prefix, key, value string, put bool) {
	cb, ok := self.otherFunc.Load(prefix)
	if ok {
		cb.(func(string, string, bool))(key, value, put)
	} else {
		llog.Errorf("etcd WatchFuc prefix no handler: %s", prefix)
	}

}

func (self *ClientDis) watcher(prefix string) {
	rch := self.client.Watch(context.Background(), prefix, clientv3.WithPrefix())
	for wresp := range rch {
		for _, ev := range wresp.Events {
			switch ev.Type {
			case mvccpb.PUT:
				self.watchFuc(prefix, string(ev.Kv.Key), string(ev.Kv.Value), true)
			case mvccpb.DELETE:
				self.watchFuc(prefix, string(ev.Kv.Key), "", false)
			}
		}
	}
}

// 通用发现
// @prefix: 监听key值
// @handler: key值变化回调
func (self *ClientDis) WatchCommon(prefix string, handler func(string, string, bool)) error {
	resp, err := self.client.Get(context.Background(), prefix, clientv3.WithPrefix())
	if err != nil {
		return err
	}
	self.otherFunc.Store(prefix, handler)
	self.extractOthers(handler, resp)

	lutil.Go(func() {
		self.watcher(prefix)
	})
	return nil
}

func (self *ClientDis) extractOthers(handler func(string, string, bool), resp *clientv3.GetResponse) {
	if resp == nil || resp.Kvs == nil {
		return
	}
	for i := range resp.Kvs {
		if v := resp.Kvs[i].Value; v != nil {
			handler(string(resp.Kvs[i].Key), string(resp.Kvs[i].Value), true)
		}
	}
}

// 通过租约 注册服务
func (self *ClientDis) PutService(key, val string) error {
	kv := clientv3.NewKV(self.client)
	_, err := kv.Put(context.TODO(), key, val, clientv3.WithLease(self.leaseResp.ID))
	return err
}

// 删除服务，保留租约
func (self *ClientDis) DelService(key string) {
	self.Delete(key)
}

func (self *ClientDis) leaseCallBack(success bool) {
	if success { //成功续租
		//llog.Debugf("etcd lease 续租成功")
		if nodemgr.SelfNode != nil {
			self.PutStatus()
		}
	} else {
		llog.Errorf("etcd lease 续租失败")
	T:
		llog.Debug("尝试重新续租")
		err := self.SetLease(int64(lconfig.GAME_LEASE_TIME), true)
		if err != nil {
			llog.Debugf("尝试重新续租失败: err=%s", err.Error())
			time.Sleep(time.Second)
			goto T
		} else {
			llog.Debugf("尝试重新续租成功")
			err = self.PutNode()
			if err != nil {
				llog.Errorf("leaseCallBack PutService error: err=%v", err)
				self.RevokeLease()
			}
		}
	}
}

func (self *ClientDis) PutStatus() error {
	val := fmt.Sprintf("%d:%t", nodemgr.OnlineNum, nodemgr.ServerEnabled)
	return self.Put(self.statusKey, val, true)
}

func (self *ClientDis) PutNode() error {
	obj, _ := json.Marshal(&lconfig.Cfg.NetCfg)
	return self.PutService(self.etcdKey, string(obj))
}

func (self *ClientDis) SetValue(prefix, val string) error {
	return self.Put(prefix, val, false)
}

// 创建服务发现
func NewEtcd() (*ClientDis, error) {
	conf := clientv3.Config{
		Endpoints:   lconfig.Cfg.EtcdAddr,
		DialTimeout: 3 * time.Second,
	}
	if client, err := clientv3.New(conf); err == nil {
		This = client
		timeoutCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		_, err = client.Status(timeoutCtx, conf.Endpoints[0])
		if err != nil {
			This = nil
			return nil, err
		}
		llog.Infof("etcd connect success: %v", lconfig.Cfg.EtcdAddr)
		disEtcd := &ClientDis{
			EtcdBase: EtcdBase{client: client},
		}
		disEtcd.etcdKey = fmt.Sprintf("%s%d/%s", ldefine.ETCD_NODEINFO, lconfig.NET_NODE_ID, lconfig.NET_GATE_SADDR)
		disEtcd.statusKey = fmt.Sprintf("%s%d/%s", ldefine.ETCD_NODESTATUS, lconfig.NET_NODE_ID, lconfig.NET_GATE_SADDR)
		disEtcd.SetLeasefunc(disEtcd.leaseCallBack)
		disEtcd.SetLease(int64(lconfig.GAME_LEASE_TIME), true)

		return disEtcd, nil
	} else {
		return nil, err
	}
}
