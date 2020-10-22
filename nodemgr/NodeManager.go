package nodemgr

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/snowyyj001/loumiao/etcd"

	"github.com/snowyyj001/loumiao/util"

	"github.com/snowyyj001/loumiao/config"
	"github.com/snowyyj001/loumiao/define"
	"github.com/snowyyj001/loumiao/log"

	"github.com/coreos/etcd/clientv3/concurrency"
)

var (
	node_Map   map[string]*config.NetNode //服务器节点
	status_Map map[string]int             //服务器状态

	lock sync.Mutex
)

func init() {
	node_Map = make(map[string]*config.NetNode)
	status_Map = make(map[string]int)
}

//所有的服注册，都会通知到这里
func NewNodeDiscover(key string, val string, dis bool) {
	lock.Lock()
	defer lock.Unlock()

	var saddr string
	fmt.Sscanf(key, define.ETCD_LOCKUID+"%s", &saddr)
	log.Debugf("NewNodeDiscover: key=%s,val=%s,dis=%t", key, val, dis)

	if dis == true {
		node := new(config.NetNode)
		json.Unmarshal([]byte(val), node)
		node_Map[saddr] = node
	} else {
		delete(node_Map, saddr)
	}
}

//服务状态更新，间隔GAME_LEASE_TIME/3
func NodeStatusUpdate(key string, val string, dis bool) {
	lock.Lock()
	defer lock.Unlock()

	var saddr string
	var number int
	fmt.Sscanf(key, define.ETCD_NODESTATUS+"%s/%d", &saddr, &number)
	//log.Debugf("NodeStatusUpdate: key=%s,val=%s,dis=%t", key, val, dis)

	if dis == true {
		status_Map[saddr] = number
	} else {
		delete(status_Map, saddr)
	}
}

//pick a gate for client
func GetBalanceGate() string {
	//pick the least gate
	var saddr string
	var minNum int = 0x7fffffff
	for val, number := range status_Map {
		if number < minNum && node_Map[val].Type == config.ServerType_Gate {
			saddr = val
		}
	}
	return saddr
}

//generate a server uid, ip+port <--> uid
func GetServerUid(cli etcd.IEtcdBase, key string) int {

	sk := fmt.Sprintf("%s%s", define.ETCD_LOCKUID, key)
	gresp, err := cli.Get(sk)
	if err != nil {
		log.Fatal("GetServerUid get failed " + err.Error())
	}
	if len(gresp.Kvs) > 0 { //server has been assigned value
		node := config.NetNode{}
		err = json.Unmarshal(gresp.Kvs[0].Value, &node)
		if err != nil {
			log.Fatal("GetServerUid etcd value bad " + string(gresp.Kvs[0].Value))
		}
		//log.Debugf("GetServerUid %s, %v", key, node)
		return node.Uid
	}

	var session *concurrency.Session
	session, err = concurrency.NewSession(cli.GetClient())
	if err != nil {
		log.Fatal("GetServerUid NewSession failed " + err.Error())
	}
	m := concurrency.NewMutex(session, define.ETCD_LOCKUID)
	if err = m.Lock(context.TODO()); err != nil {
		log.Fatal("GetServerUid NewMutex failed " + err.Error())
	}
	var topvalue int
	if config.Cfg.NetCfg.Uid != 0 { //手工分配了uid
		topvalue = config.Cfg.NetCfg.Uid
	} else {
		sk_reserve := fmt.Sprintf("%s%s", define.ETCD_LOCKUID, "0")
		gresp, err = cli.Get(sk_reserve)
		if len(gresp.Kvs) == 0 {
			topvalue = 1
			cli.Put(sk_reserve, util.Itoa(topvalue), false) //init value
		} else {
			topvalue = util.Atoi(string(gresp.Kvs[0].Value)) + 1 //inc value
			cli.Put(sk_reserve, util.Itoa(topvalue), false)      //inc value
		}
	}
	config.Cfg.NetCfg.Uid = topvalue
	obj, _ := json.Marshal(&config.Cfg.NetCfg)
	jsonstr := string(obj)
	cli.Put(sk, jsonstr, false) //set value
	m.Unlock(context.TODO())

	return topvalue
}
