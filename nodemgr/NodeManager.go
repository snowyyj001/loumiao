package nodemgr

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"

	"github.com/snowyyj001/loumiao/config"
	"github.com/snowyyj001/loumiao/define"
	"github.com/snowyyj001/loumiao/etcd"
	"github.com/snowyyj001/loumiao/log"
	"github.com/snowyyj001/loumiao/util"

	"github.com/coreos/etcd/clientv3/concurrency"
)

type NodeInfo struct {
	config.NetNode      //所有的服务器列表，如果要删除，需要删除etcd里面的内容
	Number         int  //-1代表服务器未激活,服务器通过ETCD_NODESTATUS上报状态，就算激活，即number >= 0
	SocketActive   bool //该服务器是否处于激活状态
}

//负载均衡策略都是挑选number最小的
//gate和accout目前有监控服务器信息,account挑选gate和world给客户端使用
//gate挑选zone给客户端使用
var (
	node_Map map[string]*NodeInfo //服务器信息
	nodeLock sync.RWMutex
	NodeUid  int //本服务器节点的uid，就是gateserver通过GetServerUid获取的id
)

func init() {
	node_Map = make(map[string]*NodeInfo)

}

func CreateNode(saddr string, uid int, atype int, group string) *NodeInfo {
	node := &NodeInfo{Number: -1}
	node.Uid = uid
	node.Type = atype
	node.Group = group
	nodeLock.Lock()
	node_Map[saddr] = node
	nodeLock.Unlock()
	return node
}

func GetServerNode(saddr string) *NodeInfo {
	//log.Debugf("GetServerNode: %s", saddr)
	nodeLock.RLock()
	node, _ := node_Map[saddr]
	nodeLock.RUnlock()
	return node
}

//所有的服注册，都会通知到这里
func NewNodeDiscover(key string, val string, dis bool) {
	var saddr string
	arrStr := strings.Split(key, "/")
	saddr = arrStr[1]
	fmt.Sscanf(key, define.ETCD_NODEINFO+"%s", &saddr)
	log.Debugf("NewNodeDiscover: key=%s,val=%s,dis=%t,debug=%s", key, val, dis, saddr)

	if dis == true {
		nodeLock.Lock()
		node, _ := node_Map[saddr]
		if node == nil {
			node = &NodeInfo{Number: -1}
			node_Map[saddr] = node
		}
		json.Unmarshal([]byte(val), &node.NetNode)
		nodeLock.Unlock()
	} else {
		nodeLock.Lock()
		delete(node_Map, saddr)
		nodeLock.Unlock()
	}
}

//服务状态更新，间隔GAME_LEASE_TIME/3
func NodeStatusUpdate(key string, val string, dis bool) {
	var saddr string
	var group string
	arrStr := strings.Split(key, "/")
	saddr = arrStr[3]
	group = arrStr[2]

	//log.Debugf("NodeStatusUpdate: key=%s,val=%s,dis=%t", key, val, dis)
	if config.NET_NODE_TYPE == config.ServerType_Gate {
		if group != config.SERVER_GROUP {
			return
		}
	}

	nodeLock.RLock()
	node, _ := node_Map[saddr]
	nodeLock.RUnlock()
	if node == nil {
		nodeLock.Lock()
		if node == nil {
			node = &NodeInfo{Number: -1}
			node_Map[saddr] = node
		}
		nodeLock.Unlock()
	}

	if dis == true {
		node.Number = util.Atoi(val)

	} else {
		node.Number = -1
	}
}

//pick a gate and world for client
func GetBalanceServer(group string, onlyworld bool) (string, int) {
	//pick the gate and the world
	var saddr string
	var worlduid int

	var minNum int = 0x7fffffff
	var minNum_2 int = 0x7fffffff

	nodeLock.RLock()
	for val, node := range node_Map {
		if node.SocketActive && node.Group == group && node.Number != -1 && node.Number < minNum && node.Type == config.ServerType_Gate {
			saddr = val
			minNum = node.Number
		}
		if onlyworld == false {
			if node.SocketActive && node.Number != -1 && node.Number < minNum_2 && node.Type == config.ServerType_World {
				worlduid = node.Uid
				minNum_2 = node.Number
			}
		}
	}
	nodeLock.RUnlock()

	return saddr, worlduid
}

//pick a zone server
func GetBalanceZone(group string) int {
	var uid int

	var minNum int = 0x7fffffff

	nodeLock.RLock()
	for _, node := range node_Map {
		if node.SocketActive && node.Group == group && node.Number != -1 && node.Number < minNum && node.Type == config.ServerType_Zone {
			uid = node.Uid
			minNum = node.Number
		}
	}
	nodeLock.RUnlock()

	return uid
}

func DisableNode(uid int) {
	nodeLock.RLock()
	for _, node := range node_Map {
		if node.Uid == uid {
			node.Number = -1
			break
		}
	}
	nodeLock.RUnlock()
	return
}

func GetNode(uid int) (ret *NodeInfo) {
	nodeLock.RLock()
	for _, node := range node_Map {
		if node.Uid == uid {
			ret = node
			break
		}
	}
	nodeLock.RUnlock()
	return
}

func GetActiveNode(uid int) (ret *NodeInfo) {
	nodeLock.RLock()
	for _, node := range node_Map {
		if node.SocketActive && node.Number != -1 && node.Uid == uid {
			ret = node
			break
		}
	}
	nodeLock.RUnlock()
	return
}

//generate a server uid, ip+port <--> uid
func GetServerUid(cli etcd.IEtcdBase, key string) int {

	sk := fmt.Sprintf("%s%s", define.ETCD_NODEINFO, key)
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
		NodeUid = node.Uid
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
	NodeUid = topvalue
	return topvalue
}

func PackNodeInfos(stype int) []byte {
	st := struct {
		Nodes []*NodeInfo `json:"nodes"`
	}{}

	nodeLock.RLock()
	for _, node := range node_Map {
		if stype == 0 || node.Type == stype {
			st.Nodes = append(st.Nodes, node)
		}
	}
	nodeLock.RUnlock()

	buffer, err := json.Marshal(&st)
	util.CheckErr(err)
	return buffer
}
