package nodemgr

import (
	"encoding/json"
	"strings"
	"sync"

	"github.com/snowyyj001/loumiao/config"
	"github.com/snowyyj001/loumiao/util"
)

type NodeInfo struct {
	config.NetNode      //所有的服务器列表，如果要删除，需要删除etcd里面的内容
	Number         int  //-1代表服务器未激活,服务器通过ETCD_NODESTATUS上报状态，就算激活，即number >= 0，socket断开或etcd断开都意味着节点不可用，已经关闭
	SocketActive   bool //服务被主动关闭，节点还可用（因为节点是有状态的，为了不丢失数据），但节点不再被集群主动发现使用，
}

//负载均衡策略都是挑选number最小的
//gate和accout目前有监控服务器信息,account挑选gate和world给客户端使用
//gate挑选zone给客户端使用
var (
	node_Map      map[string]*NodeInfo //服务器信息
	saddr_uid_Map map[int]string       //saddr -> uid
	nodeLock      sync.RWMutex
)

func init() {
	node_Map = make(map[string]*NodeInfo)
	saddr_uid_Map = make(map[int]string)
}

func AddNode(node *NodeInfo) {
	nodeLock.Lock()
	node_Map[node.SAddr] = node
	saddr_uid_Map[node.Uid] = node.SAddr
	nodeLock.Unlock()
}

func RemoveNode(saddr string) {
	nodeLock.Lock()
	node, _ := node_Map[saddr]
	if node != nil {
		delete(node_Map, saddr)
		delete(saddr_uid_Map, node.Uid)
	}
	nodeLock.Unlock()
}

func GetNodeByAddr(saddr string) *NodeInfo {
	//llog.Debugf("GetNodeByAddr: %s", saddr)
	nodeLock.RLock()
	node, _ := node_Map[saddr]
	nodeLock.RUnlock()
	return node
}

//服务状态更新，间隔GAME_LEASE_TIME/3
func NodeStatusUpdate(key string, val string, dis bool) {
	var saddr string
	//var group string
	arrStr := strings.Split(key, "/")
	saddr = arrStr[2]
	//group = arrStr[2]

	//llog.Debugf("NodeStatusUpdate: key=%s,val=%s,dis=%t", key, val, dis)
	/*	if config.NET_NODE_TYPE == config.ServerType_Gate {
		if group != config.SERVER_GROUP {
			return
		}
	}*/

	nodeLock.RLock()
	node, _ := node_Map[saddr]
	nodeLock.RUnlock()
	if node == nil {
		return
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
		if node.SocketActive && node.Number != -1 && node.Number < minNum && node.Type == config.ServerType_Gate {
			saddr = val
			minNum = node.Number
		}
		if onlyworld == false {
			if node.SocketActive && /* && node.Group == group*/ node.Number != -1 && node.Number < minNum_2 && node.Type == config.ServerType_World {
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
		if node.SocketActive /* && node.Group == group */ && node.Number != -1 && node.Number < minNum && node.Type == config.ServerType_Zone {
			uid = node.Uid
			minNum = node.Number
		}
	}
	nodeLock.RUnlock()

	return uid
}

func DisableNode(uid int) {
	node := GetNode(uid)
	if node != nil {
		node.Number = -1
	}
}

func GetNode(uid int) (ret *NodeInfo) {
	nodeLock.RLock()
	saddr, ok := saddr_uid_Map[uid]
	if ok {
		ret, _ = node_Map[saddr]
	}
	nodeLock.RUnlock()
	return
}

func GetActiveNode(uid int) (ret *NodeInfo) {
	node := GetNode(uid)
	if node != nil && node.SocketActive && node.Number != -1 {
		ret = node
	}
	return
}

/*
//generate a server uid, ip+port <--> uid
func GetServerUid(cli etcd.IEtcdBase, key string) int {

	sk := fmt.Sprintf("%s%s", define.ETCD_NODEINFO, key)
	gresp, err := cli.Get(sk)
	if err != nil {
		llog.Fatal("GetServerUid get failed " + err.Error())
	}
	if len(gresp.Kvs) > 0 { //server has been assigned value
		node := config.NetNode{}
		err = json.Unmarshal(gresp.Kvs[0].Value, &node)
		if err != nil {
			llog.Fatal("GetServerUid etcd value bad " + string(gresp.Kvs[0].Value))
		}
		NodeUid = node.Uid
		//llog.Debugf("GetServerUid %s, %v", key, node)
		return node.Uid
	}

	var session *concurrency.Session
	session, err = concurrency.NewSession(cli.GetClient())
	if err != nil {
		llog.Fatal("GetServerUid NewSession failed " + err.Error())
	}
	m := concurrency.NewMutex(session, define.ETCD_LOCKUID)
	if err = m.Lock(context.TODO()); err != nil {
		llog.Fatal("GetServerUid NewMutex failed " + err.Error())
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
*/
func PackNodeInfos(group string, stype int) []byte {
	st := struct {
		Nodes []*NodeInfo `json:"nodes"`
	}{}

	nodeLock.RLock()
	for _, node := range node_Map {
		if (node.Group == group || group == "") && (stype == 0 || node.Type == stype) {
			st.Nodes = append(st.Nodes, node)
		}
	}
	nodeLock.RUnlock()

	buffer, err := json.Marshal(&st)
	util.CheckErr(err)
	return buffer
}
