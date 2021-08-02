package etcd

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/snowyyj001/loumiao/llog"
	"go.etcd.io/etcd/clientv3"
	"go.etcd.io/etcd/clientv3/concurrency"
	"go.etcd.io/etcd/mvcc/mvccpb"

	"github.com/snowyyj001/loumiao/define"
)

type HanlderFunc func(string, string, bool)

var (
	This *clientv3.Client
	EtcdClient *ClientDis
)

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

//
func (self *EtcdBase) GetClient() *clientv3.Client {
	return self.client
}

//设置租约回调
func (self *EtcdBase) SetLeasefunc(call func(bool)) {
	self.leasefunc = call
}

//设置value
func (self *EtcdBase) Put(key string, val string, withlease bool) error {
	if withlease {
		_, err := self.client.Put(context.TODO(), key, val, clientv3.WithLease(self.leaseResp.ID))
		return err
	} else {
		_, err := self.client.Put(context.TODO(), key, val)
		return err
	}
}

//删除value
func (self *EtcdBase) Delete(key string) error {
	llog.Debugf("etcd delete: %s", key)
	_, err := self.client.Delete(context.TODO(), key)
	return err
}

//获得value
func (self *EtcdBase) Get(key string) (*clientv3.GetResponse, error) {
	gresp, err := self.client.Get(context.TODO(), key)
	return gresp, err
}

//设置租约
//@timeNum: 过期时间
//@keepalive: 是否自动续约
func (self *EtcdBase) SetLease(timeNum int64, keepalive bool) error {
	if self.lease != nil {
		return fmt.Errorf("lease error")
	}
	lease := clientv3.NewLease(self.client)

	//设置租约时间
	leaseResp, err := lease.Grant(context.TODO(), timeNum)
	if err != nil {
		return err
	}

	//设置续租
	if keepalive {
		ctx, cancelFunc := context.WithCancel(context.TODO())
		leaseRespChan, err := lease.KeepAlive(ctx, leaseResp.ID)

		if err != nil {
			return err
		}
		self.canclefunc = cancelFunc
		self.keepAliveChan = leaseRespChan

		go self.listenLeaseRespChan()
	}

	self.lease = lease
	self.leaseResp = leaseResp

	return nil
}

//监听 续租情况
func (self *EtcdBase) listenLeaseRespChan() {
	for {
		select {
		case leaseKeepResp := <-self.keepAliveChan:
			if leaseKeepResp == nil {
				llog.Debugf("已经关闭续租功能")
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

//撤销租约
func (self *EtcdBase) RevokeLease() error {
	if self.leaseResp == nil {
		return fmt.Errorf("RevokeLease: lease has already been ewvoked")
	}
	self.canclefunc()
	_, err := self.lease.Revoke(context.TODO(), self.leaseResp.ID)
	self.leaseResp = nil
	return err
}

//获取一个分布式锁,expire毫秒后会超时返回nil
//@prefix: 锁key
//@expire: 超时时间,毫秒
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

//选举leader，所有参与选举的人使用相同的value和prefix，leader负责设置value
//@prefix: 选举区分标识
//@value: 本次选举的值，每次发起选举，value应该和上次选举时的value不同
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

//创建etcd服务
//@addr: etcd地址
//@timeNum: 连接超时时间
func NewEtcd(addr []string, timeNum int64) (*EtcdBase, error) {
	conf := clientv3.Config{
		Endpoints:   addr,
		DialTimeout: time.Duration(timeNum) * time.Second,
	}

	var (
		client *clientv3.Client
	)

	if clientTem, err := clientv3.New(conf); err == nil {
		client = clientTem
	} else {
		return nil, err
	}

	ser := &EtcdBase{
		client: client}

	This = client
	return ser, nil
}

///////////////////////////////////////////////////////////
//服发现
type ClientDis struct {
	EtcdBase
	otherFunc sync.Map //[string]HanlderFunc

	NodeListFuc   func(string, string, bool) //服发现
	StatusListFuc func(string, string, bool) //状态更新
}

func (self *ClientDis) watchFuc(prefix, key, value string, put bool) {
	switch prefix {
	case define.ETCD_NODEINFO:
		self.NodeListFuc(key, value, put)
	case define.ETCD_NODESTATUS:
		self.StatusListFuc(key, value, put)
	default:
		cb, ok := self.otherFunc.Load(prefix)
		if ok {
			cb.(HanlderFunc)(key, value, put)
		} else {
			llog.Errorf("etcd WatchFuc prefix no handler: %s", prefix)
		}
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

//通用发现
//@prefix: 监听key值
//@hanlder: key值变化回调
func (self *ClientDis) WatchCommon(prefix string, hanlder HanlderFunc) ([]string, error) {
	resp, err := self.client.Get(context.Background(), prefix, clientv3.WithPrefix())
	if err != nil {
		return nil, err
	}
	self.otherFunc.Store(prefix, hanlder)
	addrs := self.extractOthers(hanlder, resp)

	go self.watcher(prefix)
	return addrs, nil
}

func (self *ClientDis) extractOthers(hanlder HanlderFunc, resp *clientv3.GetResponse) []string {
	addrs := make([]string, 0)
	if resp == nil || resp.Kvs == nil {
		return addrs
	}
	for i := range resp.Kvs {
		if v := resp.Kvs[i].Value; v != nil {
			hanlder(string(resp.Kvs[i].Key), string(resp.Kvs[i].Value), true)
			addrs = append(addrs, string(v))
		}
	}
	return addrs
}

//服发现
//@prefix: 监听key值
//@hanlder: key值变化回调
func (self *ClientDis) WatchNodeList(prefix string, hanlder func(string, string, bool)) ([]string, error) {
	resp, err := self.client.Get(context.Background(), prefix, clientv3.WithPrefix())
	if err != nil {
		return nil, err
	}
	self.NodeListFuc = hanlder
	addrs := self.extractAddrs_All(resp)

	go self.watcher(prefix)
	return addrs, nil
}

func (self *ClientDis) extractAddrs_All(resp *clientv3.GetResponse) []string {
	addrs := make([]string, 0)
	if resp == nil || resp.Kvs == nil {
		return addrs
	}
	for i := range resp.Kvs {
		if v := resp.Kvs[i].Value; v != nil {
			self.NodeListFuc(string(resp.Kvs[i].Key), string(resp.Kvs[i].Value), true)
			addrs = append(addrs, string(v))
		}
	}
	return addrs
}

//通过租约 注册服务
func (self *ClientDis) PutService(key, val string) error {
	kv := clientv3.NewKV(self.client)
	_, err := kv.Put(context.TODO(), key, val, clientv3.WithLease(self.leaseResp.ID))
	return err
}

//删除服务，保留租约
func (self *ClientDis) DelService(key string) {
	self.Delete(key)
}

//状态更新
//@prefix: 监听key值
//@hanlder: key值变化回调
func (self *ClientDis) WatchStatusList(prefix string, hanlder func(string, string, bool)) error {
	self.StatusListFuc = hanlder

	go self.watcher(prefix)
	return nil
}

//创建服务发现
func NewClientDis(addr []string) (*ClientDis, error) {
	llog.Debugf("etcd connect: %v", addr)
	conf := clientv3.Config{
		Endpoints:   addr,
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
		llog.Infof("NewClientDis, etcd connect success: %v", addr)
		EtcdClient = &ClientDis{
			EtcdBase: EtcdBase{client: client},
		}
		return EtcdClient, nil
	} else {
		return nil, err
	}
}