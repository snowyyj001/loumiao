package nodemgr

import "github.com/snowyyj001/loumiao/config"

type NodeInfo struct {
	config.NetNode      //所有的服务器列表，如果要删除，需要删除etcd里面的内容
	Number         int  	`json:"number"` //服务器通过ETCD_NODESTATUS上报状态，即number >= 0
	SocketActive   bool 	`json:"socketactive"` //服务被主动关闭，节点还可用（因为节点是有状态的，为了不丢失数据），但节点不再被集群主动发现使用，
}