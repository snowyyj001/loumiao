package config

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
)

const (
	ServerType_ETCD    = iota //0 etcd
	ServerType_Gate           //1 网关
	ServerType_Account        //2 账号
	ServerType_World          //3 世界
	ServerType_Zone           //4 地图
	ServerType_DB             //5 数据库
	ServerType_Log            //6 日志
	ServerType_IM             //7 聊天
)

var (
	NET_NODE_ID    = -1               //节点id(机器标识,1~1024)
	NET_NODE_TYPE  = -1               //节点类型ServerType_*
	NET_GATE_SADDR = "127.0.0.1:6789" //网关监听地址

	NET_PROTOCOL      = "PROTOBUF" //消息协议格式："PROTOBUF" or "JSON"
	NET_WEBSOCKET     = false      //使用websocket or socket
	NET_MAX_CONNS     = 65535      //最大连接数
	NET_MAX_RPC_CONNS = 1024       //rpc最大连接数
	NET_BUFFER_SIZE   = 1024 * 64  //最大消息包长度64k
)

//uid通过etcd自动分配，一般不要手动分配uid，除非清楚知道自己在做什么,参考GetServerUid
//uid和SAddr是一一对应的,可以通过删除ETCD_LOCKUID来重置uid的分配
type NetNode struct {
	Id        int    `json:"id"`
	Type      int    `json:"type"`
	SAddr     string `json:"saddr"`
	Protocol  string `json:"protocol"`
	WebSocket int    `json:"websocket"`
	Uid       int    `json:"uid"`
}

type ServerCfg struct {
	NetCfg   NetNode  `json:"net"`
	EtcdAddr []string `json:"etcd"`
	NsqAddr  []string `json:"nsqd"`
}

var Cfg ServerCfg

func init() {
	data, err := ioutil.ReadFile("config/cfg.json")
	if err != nil {
		fmt.Println(err)
		return
	}

	err = json.Unmarshal(data, &Cfg)
	if err != nil {
		fmt.Println(err)
		return
	}

	NET_NODE_ID = Cfg.NetCfg.Id
	NET_NODE_TYPE = Cfg.NetCfg.Type
	NET_PROTOCOL = Cfg.NetCfg.Protocol
	NET_WEBSOCKET = Cfg.NetCfg.WebSocket == 1
	NET_GATE_SADDR = Cfg.NetCfg.SAddr

}
