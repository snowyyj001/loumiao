package config

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
)

const (
	ServerType_ETCD  = 0 //etcd
	ServerType_World = 1 //世界服-->集群
	ServerType_Gate  = 2 //网关服-->集群
	ServerType_Login = 3 //登录服-->集群
	ServerType_DB    = 4 //数据库服-->集群
	ServerType_Zoon  = 5 //地图服-->集群
	ServerType_IM    = 6 //聊天服-->集群
)

var (
	NET_NODE_ID    = -1               //节点id(机器标识,1~1024)
	NET_NODE_TYPE  = -1               //节点类型(World, Gate, Login, Zoon, DB)
	NET_GATE_SADDR = "127.0.0.1:6789" //网关监听地址
	NET_RPC_ADDR   = "127.0.0.1:5678" //RPC监听地址

	NET_PROTOCOL      = "PROTOBUF" //消息协议格式："PROTOBUF" or "JSON"
	NET_WEBSOCKET     = false      //使用websocket or socket
	NET_MAX_CONNS     = 65535      //最大连接数
	NET_MAX_RPC_CONNS = 1024       //rpc最大连接数
	NET_BUFFER_SIZE   = 1024 * 64  //最大消息包长度64k
)

type NetNode struct {
	Id        int    `json:"id"`
	Type      int    `json:"type"`
	SAddr     string `json:"saddr"`
	Protocol  string `json:"protocol"`
	RpcAddr   string `json:"rpcaddr"`
	WebSocket int    `json:"websocket"`
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
	NET_RPC_ADDR = Cfg.NetCfg.RpcAddr
}
