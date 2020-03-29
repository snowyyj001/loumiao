package config

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
)

var (
	NET_GATE_IP   = "127.0.0.1" //网关监听地址
	NET_GATE_PORT = 6789        //网关监听端口
	NET_BE_CHILD  = 0           //作为分布式子网节点[0：单节点，1：分布式节点client，2：分布式节点server]
	NET_RPC_IP    = "127.0.0.1" //RPC监听地址
	NET_RPC_PORT  = 5678        //RPC监听端口

	NET_PROTOCOL      = "PROTOBUF" //消息协议格式："PROTOBUF" or "JSON"
	NET_WEBSOCKET     = false      //使用websocket or socket
	NET_MAX_CONNS     = 65535      //最大连接数
	NET_MAX_RPC_CONNS = 1024       //rpc最大连接数
	NET_BUFFER_SIZE   = 1024 * 64  //最大消息包长度64k
)

type SelfNetCfg struct {
	Ip        string `json:"ip"`
	Port      int    `json:"port"`
	BeChild   int    `json:"bechild"` //0:single,1:client,2:server
	Protocol  string `json:"protocol"`
	WebSocket int    `json:"websocket"`
	RpcIp     string `json:"rpcip"`
	RpcPort   int    `json:"rpcport"`
}

type NetNode struct {
	Id   int    `json:"id"`
	Ip   string `json:"ip"`
	Port int    `json:"port"`
	Name string `json:"string"`
}

type Server struct {
	ServerNodes []NetNode  `json:"net"`
	NetCfg      SelfNetCfg `json:"selfnet"`
}

var ServerCfg Server

func init() {
	data, err := ioutil.ReadFile("config/cfg.json")
	if err != nil {
		fmt.Println(err)
		return
	}
	err = json.Unmarshal(data, &ServerCfg)
	if err != nil {
		fmt.Println(err)
		return
	}

	NET_PROTOCOL = ServerCfg.NetCfg.Protocol
	NET_WEBSOCKET = ServerCfg.NetCfg.WebSocket == 1
	NET_GATE_PORT = ServerCfg.NetCfg.Port
	NET_GATE_IP = ServerCfg.NetCfg.Ip
	NET_BE_CHILD = ServerCfg.NetCfg.BeChild
	NET_RPC_PORT = ServerCfg.NetCfg.RpcPort
	NET_RPC_IP = ServerCfg.NetCfg.RpcIp
}
