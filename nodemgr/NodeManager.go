package nodemgr

import (
	"encoding/json"
	"fmt"
	"strings"
	"sync"

	"github.com/snowyyj001/loumiao/llog"

	"github.com/snowyyj001/loumiao/config"
	"github.com/snowyyj001/loumiao/util"
)

//负载均衡策略都是挑选number最小的
//gate和accout目前有监控服务器信息,account挑选gate和world给客户端使用
//gate挑选zone给客户端使用
var (
	node_Map      map[string]*NodeInfo //服务器信息
	saddr_uid_Map map[int]string       //saddr -> uid
	nodeLock      sync.RWMutex
	ServerEnabled bool
	OnlineNum     int
)

func init() {
	node_Map = make(map[string]*NodeInfo)
	saddr_uid_Map = make(map[int]string)
	ServerEnabled = false
}

func AddNode(node *NodeInfo) {
	defer nodeLock.Unlock()
	nodeLock.Lock()
	node_Map[node.SAddr] = node
	saddr_uid_Map[node.Uid] = node.SAddr
}

func RemoveNodeById(uid int) {
	defer nodeLock.Unlock()
	nodeLock.Lock()
	saddr, ok := saddr_uid_Map[uid]
	if ok {
		delete(node_Map, saddr)
		delete(saddr_uid_Map, uid)
	}
}

func RemoveNode(saddr string) *NodeInfo {
	defer nodeLock.Unlock()
	nodeLock.Lock()
	node, _ := node_Map[saddr]
	if node != nil {
		delete(node_Map, saddr)
		delete(saddr_uid_Map, node.Uid)
	}
	return node
}

func GetNodeByAddr(saddr string) *NodeInfo {
	//llog.Debugf("GetNodeByAddr: %s", saddr)
	defer nodeLock.RUnlock()
	nodeLock.RLock()
	node, _ := node_Map[saddr]
	return node
}

//服务状态更新，间隔GAME_LEASE_TIME/3
func NodeStatusUpdate(key string, val string, dis bool) *NodeInfo {
	if !dis {
		return nil
	}
	//var group string
	arrStr := strings.Split(key, "/")
	/*areaid := util.Atoi(arrStr[2])
	if areaid != config.NET_NODE_ID { //不关心其他服的情况
		return
	}*/
	if len(arrStr) < 3 {
		llog.Errorf("NodeStatusUpdate illegal status prefix: %s", key)
		return nil
	}

	saddr := arrStr[3]
	//llog.Debugf("NodeStatusUpdate: key=%s,val=%s,dis=%t,saddr=%s", key, val, dis, saddr)
	var number int
	var socketActive bool
	fmt.Sscanf(val, "%d:%t", &number, &socketActive)

	defer nodeLock.Unlock()
	nodeLock.Lock()
	node, _ := node_Map[saddr]
	if node == nil {
		if socketActive {
			node = new(NodeInfo)
			node.SAddr = saddr
			node_Map[node.SAddr] = node
		} else {
			return nil
		}
	}
	node.Number = number
	node.SocketActive = socketActive

	//llog.Debugf("NodeStatusUpdate: saddr=%s,active=%t", node.SAddr, node.SocketActive)
	return node
}

//服发现
func NodeDiscover(key string, val string, dis bool) *NodeInfo {
	arrStr := strings.Split(key, "/")
	if len(arrStr) < 3 {
		llog.Errorf("NodeDiscover error fromat key : %s", key)
		return nil
	}
	/*areaid := util.Atoi(arrStr[2])
	if areaid != config.NET_NODE_ID { //不关心其他服的情况
		return
	}*/

	saddr := arrStr[3]
	llog.Debugf("NodeDiscover: key=%s,val=%s,dis=%t,saddr=%s", key, val, dis, saddr)
	if dis == true {
		node := GetNodeByAddr(saddr)
		if node == nil { //maybe, etcf still have older data, etcf has a huge delay
			node = new(NodeInfo)
			node.SocketActive = true
		}
		json.Unmarshal([]byte(val), node)
		AddNode(node)
		return node

	} else {
		return RemoveNode(saddr)

	}
}

//pick a gate and world for client
func GetBalanceServer(widthgate bool, widthworld bool) (string, int) {
	//pick the gate and the world
	var saddr []string
	var worlduid []int

	var minNum int = 0x7fffffff
	var minNum_2 int = 0x7fffffff
	defer nodeLock.RUnlock()
	nodeLock.RLock()
	for val, node := range node_Map {
		//llog.Debugf("GetBalanceServer %t, %d, %d", node.SocketActive , node.Number, node.Type)
		if widthgate {
			if node.SocketActive && node.Number <= minNum && node.Type == config.ServerType_Gate {
				if node.Number == minNum {
					saddr = append(saddr, val)
				} else {
					saddr = []string{val}
					minNum = node.Number
				}
			}
		}
		if widthworld == true {
			if node.SocketActive && node.Number <= minNum_2 && node.Type == config.ServerType_World {
				if node.Number == minNum_2 {
					worlduid = append(worlduid, node.Uid)
				} else {
					worlduid = []int{node.Uid}
					minNum_2 = node.Number
				}
			}
		}
	}

	sz := len(saddr)
	var retsaddr string
	if sz > 0 {
		retsaddr = saddr[util.Random(sz)]
	}

	var retuid int
	sz = len(worlduid)
	if sz > 0 {
		retuid = worlduid[util.Random(sz)]
	}

	return retsaddr, retuid
}

//pick a zone server by random
func GetBalanceZone() *NodeInfo  {
	defer nodeLock.RUnlock()
	nodeLock.RLock()

	nodes := make([]*NodeInfo, 0)		//这里不要利用map访问的随机性，因为map每次的访问虽然是随机的但并不均匀
	for _, node := range node_Map {
		if node.SocketActive && node.Type == config.ServerType_Zone {
			nodes = append(nodes, node)
		}
	}
	sz := len(nodes)
	if sz > 0 {
		node := nodes[util.Random(1000)%sz]
		return node

	} else {
		return nil
	}
}

func GetNode(uid int) (ret *NodeInfo) {
	defer nodeLock.RUnlock()
	nodeLock.RLock()
	saddr, ok := saddr_uid_Map[uid]
	if ok {
		ret, _ = node_Map[saddr]
	}
	return
}

func GetActiveNode(uid int) (ret *NodeInfo) {
	node := GetNode(uid)
	if node != nil && node.SocketActive {
		ret = node
	}
	return
}

func DisableNode(uid int) {
	node := GetNode(uid)
	if node != nil {
		node.SocketActive = false
	}
	if uid == config.SERVER_NODE_UID {
		ServerEnabled = false
	}
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
	defer nodeLock.RUnlock()
	nodeLock.RLock()
	for _, node := range node_Map {
		if (node.Group == group || group == "") && (stype == 0 || node.Type == stype) {
			st.Nodes = append(st.Nodes, node)
		}
	}

	buffer, err := json.Marshal(&st)
	util.CheckErr(err)
	return buffer
}
