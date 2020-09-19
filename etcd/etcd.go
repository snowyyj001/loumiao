package etcd

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/snowyyj001/loumiao/log"

	"github.com/snowyyj001/loumiao/define"
	"github.com/snowyyj001/loumiao/util"

	"github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/clientv3/concurrency"
	"github.com/coreos/etcd/mvcc/mvccpb"
)

//创建租约注册服务
type ServiceReg struct {
	client        *clientv3.Client
	lease         clientv3.Lease
	leaseResp     *clientv3.LeaseGrantResponse
	canclefunc    func()
	keepAliveChan <-chan *clientv3.LeaseKeepAliveResponse
	key           string
	CallBack      func(string, string, bool)
}

//
func (self *ServiceReg) GetClient() *clientv3.Client {
	return self.client
}

//timeNum 租约心跳间隔
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
		client: client,
	}

	if err := ser.setLease(timeNum); err != nil {
		return nil, err
	}
	go ser.ListenLeaseRespChan()
	return ser, nil
}

//设置租约
func (self *ServiceReg) setLease(timeNum int64) error {
	lease := clientv3.NewLease(self.client)

	//设置租约时间
	leaseResp, err := lease.Grant(context.TODO(), timeNum)
	if err != nil {
		return err
	}

	//设置续租
	ctx, cancelFunc := context.WithCancel(context.TODO())
	leaseRespChan, err := lease.KeepAlive(ctx, leaseResp.ID)

	if err != nil {
		return err
	}

	self.lease = lease
	self.leaseResp = leaseResp
	self.canclefunc = cancelFunc
	self.keepAliveChan = leaseRespChan
	return nil
}

//监听 续租情况
func (self *ServiceReg) ListenLeaseRespChan() {
	for {
		select {
		case leaseKeepResp := <-self.keepAliveChan:
			if leaseKeepResp == nil {
				fmt.Printf("已经关闭续租功能\n")
				return
			} else {
				fmt.Printf("续租成功\n")
			}
		}
	}
}

//通过租约 注册服务
func (self *ServiceReg) PutService(key, val string) error {
	kv := clientv3.NewKV(self.client)
	_, err := kv.Put(context.TODO(), key, val, clientv3.WithLease(self.leaseResp.ID))
	return err
}

//撤销租约
func (self *ServiceReg) RevokeLease() error {
	self.canclefunc()
	_, err := self.lease.Revoke(context.TODO(), self.leaseResp.ID)
	return err
}

func (self *ServiceReg) Watch(prefix string, hanlder func(string, string, bool)) {
	resp, err := self.client.Get(context.Background(), prefix, clientv3.WithPrefix())
	if err != nil {
		log.Errorf("ServiceReg Watch %s error: %s", prefix, err.Error())
		return
	}
	self.CallBack = hanlder
	self.extractValues(resp)

	go self.watcher(prefix)
}

func (self *ServiceReg) extractValues(resp *clientv3.GetResponse) {
	if resp == nil || resp.Kvs == nil {
		return
	}
	for i := range resp.Kvs {
		if v := resp.Kvs[i].Value; v != nil {
			if self.CallBack != nil {
				self.CallBack(string(resp.Kvs[i].Key), string(resp.Kvs[i].Value), true)
			}
		}
	}
}

func (self *ServiceReg) watcher(prefix string) {
	rch := self.client.Watch(context.Background(), prefix, clientv3.WithPrefix())
	for wresp := range rch {
		for _, ev := range wresp.Events {
			switch ev.Type {
			case mvccpb.PUT:
				if self.CallBack != nil {
					self.CallBack(string(ev.Kv.Key), string(ev.Kv.Value), true)
				}
			case mvccpb.DELETE:
				if self.CallBack != nil {
					self.CallBack(string(ev.Kv.Key), "", false)
				}
			}
		}
	}
}

///////////////////////////////////////////////////////////
//client
type ClientDis struct {
	client     *clientv3.Client
	serverList map[string]string
	CallBack   func(string, string, bool)
	lock       sync.Mutex
}

//
func (self *ClientDis) GetClient() *clientv3.Client {
	return self.client
}

func NewClientDis(addr []string) (*ClientDis, error) {
	conf := clientv3.Config{
		Endpoints:   addr,
		DialTimeout: 5 * time.Second,
	}
	if client, err := clientv3.New(conf); err == nil {
		return &ClientDis{
			client:     client,
			serverList: make(map[string]string),
		}, nil
	} else {
		return nil, err
	}
}

func (self *ClientDis) Watch(prefix string, hanlder func(string, string, bool)) ([]string, error) {
	resp, err := self.client.Get(context.Background(), prefix, clientv3.WithPrefix())
	if err != nil {
		return nil, err
	}
	addrs := self.extractAddrs(resp)
	self.CallBack = hanlder

	go self.watcher(prefix)
	return addrs, nil
}

func (self *ClientDis) watcher(prefix string) {
	rch := self.client.Watch(context.Background(), prefix, clientv3.WithPrefix())
	for wresp := range rch {
		for _, ev := range wresp.Events {
			switch ev.Type {
			case mvccpb.PUT:
				self.SetServiceList(string(ev.Kv.Key), string(ev.Kv.Value))
			case mvccpb.DELETE:
				self.DelServiceList(string(ev.Kv.Key))
			}
		}
	}
}

func (self *ClientDis) extractAddrs(resp *clientv3.GetResponse) []string {
	addrs := make([]string, 0)
	if resp == nil || resp.Kvs == nil {
		return addrs
	}
	for i := range resp.Kvs {
		if v := resp.Kvs[i].Value; v != nil {
			self.SetServiceList(string(resp.Kvs[i].Key), string(resp.Kvs[i].Value))
			addrs = append(addrs, string(v))
		}
	}
	return addrs
}

func (self *ClientDis) SetServiceList(key, val string) {
	self.lock.Lock()
	defer self.lock.Unlock()
	self.serverList[key] = string(val)
	if self.CallBack != nil {
		self.CallBack(key, val, true)
	}
}

func (self *ClientDis) DelServiceList(key string) {
	self.lock.Lock()
	defer self.lock.Unlock()
	delete(self.serverList, key)
	if self.CallBack != nil {
		self.CallBack(key, "", true)
	}
}

func (self *ClientDis) SerList2Array() []string {
	self.lock.Lock()
	defer self.lock.Unlock()
	addrs := make([]string, 0)

	for _, v := range self.serverList {
		addrs = append(addrs, v)
	}
	return addrs
}

////////////////////////////////////////////////////////////
//

//设置value
func Put(cli *clientv3.Client, key string, val string) error {
	_, err := cli.Put(context.TODO(), key, val)
	return err
}

//删除value
func Delete(cli *clientv3.Client, key string) error {
	_, err := cli.Delete(context.TODO(), key)
	return err
}

//获得value
func Get(cli *clientv3.Client, key string) (*clientv3.GetResponse, error) {
	gresp, err := cli.Get(context.TODO(), key)
	return gresp, err
}

//获取server uid
func GetServerUid(cli *clientv3.Client, key string) int {
	sk := fmt.Sprintf("%s%s", define.ETCD_LOCKUID, key)
	gresp, err := Get(cli, sk)
	if err != nil {
		log.Fatal("GetServerUid get failed " + err.Error())
	}
	if len(gresp.Kvs) > 0 { //server has been assigned value
		return util.Atoi(string(gresp.Kvs[0].Value))
	}

	var session *concurrency.Session
	session, err = concurrency.NewSession(cli)
	if err != nil {
		log.Fatal("GetServerUid NewSession failed " + err.Error())
	}
	m := concurrency.NewMutex(session, define.ETCD_LOCKUID)
	if err = m.Lock(context.TODO()); err != nil {
		log.Fatal("GetServerUid NewMutex failed " + err.Error())
	}
	var topvalue int
	sk_reserve := fmt.Sprintf("%s%s", define.ETCD_LOCKUID, "0")
	gresp, err = Get(cli, sk_reserve)
	if len(gresp.Kvs) == 0 {
		topvalue = 1
		Put(cli, sk_reserve, util.Itoa(topvalue)) //init value
	} else {
		topvalue = util.Atoi(string(gresp.Kvs[0].Value)) + 1 //inc value
		Put(cli, sk_reserve, util.Itoa(topvalue))            //inc value
	}
	Put(cli, sk, util.Itoa(topvalue)) //set value
	m.Unlock(context.TODO())

	return topvalue
}
