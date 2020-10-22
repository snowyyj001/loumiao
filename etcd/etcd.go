package etcd

import (
	"context"
	"time"

	"github.com/snowyyj001/loumiao/log"

	"github.com/snowyyj001/loumiao/define"

	"github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/mvcc/mvccpb"
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
				//fmt.Printf("已经关闭续租功能\n")
				if self.leasefunc != nil {
					self.leasefunc(false)
				}
				return
			} else {
				//	fmt.Printf("续租成功\n")
				if self.leasefunc != nil {
					self.leasefunc(true)
				}
			}
		}
	}
}

//撤销租约
func (self *EtcdBase) RevokeLease() error {
	self.canclefunc()
	_, err := self.lease.Revoke(context.TODO(), self.leaseResp.ID)
	return err
}

//创建服务注册
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

	return ser, nil
}

///////////////////////////////////////////////////////////
//服务注册
type ServiceReg struct {
	EtcdBase

	TimeOut int64 //续租时间
}

//通过租约 注册服务
func (self *ServiceReg) PutService(key, val string) error {
	kv := clientv3.NewKV(self.client)
	_, err := kv.Put(context.TODO(), key, val, clientv3.WithLease(self.leaseResp.ID))
	return err
}

//创建服务注册
func NewServiceReg(addr []string, timeNum int64) (*ServiceReg, error) {
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

	ser := &ServiceReg{
		EtcdBase: EtcdBase{client: client},
		TimeOut:  timeNum}

	if err := ser.SetLease(timeNum, true); err != nil {
		return nil, err
	}
	return ser, nil
}

///////////////////////////////////////////////////////////
//服务监听
type ClientDis struct {
	EtcdBase

	ServerListFuc func(string, string, bool) //服发现
	RpcListFuc    func(string, string, bool) //rpc发现
	NodeListFuc   func(string, string, bool) //所有服发现
	StatusListFuc func(string, string, bool) //状态更新
}

func (self *ClientDis) watchFuc(prefix, key, value string, put bool) {
	switch prefix {
	case define.ETCD_SADDR:
		self.ServerListFuc(key, value, put)
	case define.ETCD_RPCADDR:
		self.RpcListFuc(key, value, put)
	case define.ETCD_LOCKUID:
		self.NodeListFuc(key, value, put)
	case define.ETCD_NODESTATUS:
		self.StatusListFuc(key, value, put)
	default:
		log.Warningf("etcd WatchFuc prefix no handler: %s", prefix)
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

//服发现
//@prefix: 监听key值
//@hanlder: key值变化回调
func (self *ClientDis) WatchServerList(prefix string, hanlder func(string, string, bool)) ([]string, error) {
	resp, err := self.client.Get(context.Background(), prefix, clientv3.WithPrefix())
	if err != nil {
		return nil, err
	}
	self.ServerListFuc = hanlder
	addrs := self.extractAddrs(resp)

	go self.watcher(prefix)
	return addrs, nil
}

func (self *ClientDis) extractAddrs(resp *clientv3.GetResponse) []string {
	addrs := make([]string, 0)
	if resp == nil || resp.Kvs == nil {
		return addrs
	}
	for i := range resp.Kvs {
		if v := resp.Kvs[i].Value; v != nil {
			self.ServerListFuc(string(resp.Kvs[i].Key), string(resp.Kvs[i].Value), true)
			addrs = append(addrs, string(v))
		}
	}
	return addrs
}

//rpc发现
//@prefix: 监听key值
//@hanlder: key值变化回调
func (self *ClientDis) WatchRpcList(prefix string, hanlder func(string, string, bool)) ([]string, error) {
	resp, err := self.client.Get(context.Background(), prefix, clientv3.WithPrefix())
	if err != nil {
		return nil, err
	}
	self.RpcListFuc = hanlder
	addrs := self.extractRpcs(resp)

	go self.watcher(prefix)
	return addrs, nil
}

func (self *ClientDis) extractRpcs(resp *clientv3.GetResponse) []string {
	addrs := make([]string, 0)
	if resp == nil || resp.Kvs == nil {
		return addrs
	}
	for i := range resp.Kvs {
		if v := resp.Kvs[i].Value; v != nil {
			self.RpcListFuc(string(resp.Kvs[i].Key), string(resp.Kvs[i].Value), true)
			addrs = append(addrs, string(v))
		}
	}
	return addrs
}

//所有服发现
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
	conf := clientv3.Config{
		Endpoints:   addr,
		DialTimeout: 5 * time.Second,
	}
	if client, err := clientv3.New(conf); err == nil {
		return &ClientDis{
			EtcdBase: EtcdBase{client: client},
		}, nil
	} else {
		return nil, err
	}
}
