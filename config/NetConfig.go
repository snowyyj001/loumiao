package config

import (
	"encoding/json"
	"io/ioutil"
)

var (
	NET_GATE_IP   = "127.0.0.1" //网关监听地址
	NET_GATE_PORT = 6789        //网关监听端口
	NET_BE_CHILD  = false       //作为分布式子网节点[false：单节点，true：分布式节点]
	NET_RPC_IP    = "127.0.0.1" //RPC监听地址
	NET_RPC_PORT  = 5678        //RPC监听端口

	NET_PROTOCOL    = "PROTOBUF" //消息协议格式："PROTOBUF" or "JSON"
	NET_WEBSOCKET   = false      //使用websocket or socket
	NET_MAX_CONNS   = 20000      //最大连接数
	NET_BUFFER_SIZE = 1024 * 64  //最大消息包长度64k
)

type NetNode struct {
	Id   int    `json:"id"`
	Ip   string `json:"ip"`
	Port int    `json:"port"`
	Name string `json:"string"`
}

type Server struct {
	ServerNodes []NetNode `json:"net"`
}

var ServerCfg Server

func init() {
	data, err := ioutil.ReadFile("config/cfg.json")
	if err != nil {
		return
	}
	err = json.Unmarshal(data, &ServerCfg)
	if err != nil {
		return
	}
}
