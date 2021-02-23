package config

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
)

const (
	ServerType_ETCD      = iota //0 etcd
	ServerType_Gate             //1 网关
	ServerType_Account          //2 账号
	ServerType_World            //3 世界
	ServerType_Zone             //4 地图
	ServerType_DB               //5 数据库
	ServerType_Log              //6 日志
	ServerType_IM               //7 聊天
	ServerType_WEB_GM           //8 web gm
	ServerType_WEB_LOGIN        //9 web login
)

var (
	NET_NODE_ID    = -1               //节点id(服标识)
	NET_NODE_TYPE  = -1               //节点类型ServerType_*
	NET_GATE_SADDR = "127.0.0.1:6789" //网关监听地址

	NET_PROTOCOL            = "PROTOBUF"  //消息协议格式："PROTOBUF" or "JSON"
	NET_WEBSOCKET           = false       //使用websocket or socket
	NET_MAX_CONNS           = 65535       //最大连接数
	NET_MAX_RPC_CONNS       = 1024        //rpc最大连接数
	NET_BUFFER_SIZE         = 1024 * 64   //最大消息包长度64k(对外)
	NET_CLUSTER_BUFFER_SIZE = 1024 * 1024 //最大消息包长度1M(对内)
	NET_MAX_NUMBER          = 30000       //pcu

	SERVER_GROUP     = "A"      //服务器分组
	SERVER_NAME      = "server" //服务器名字
	SERVER_NODE_UID  = 0        //服务器uid
	NET_LISTEN_SADDR = "0.0.0.0:6789"
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
	MaxNum    int    `json:"maxnum"`
	Group     string `json:"group"`
	LogFile   int    `json:"logfile"` //如果-1，代表输出到控制台
}

type ServerCfg struct {
	NetCfg   NetNode  `json:"net"`
	EtcdAddr []string `json:"etcd"`
	NatsAddr []string `json:"nats"`
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
	SERVER_NODE_UID = Cfg.NetCfg.Uid
	NET_NODE_TYPE = Cfg.NetCfg.Type
	NET_PROTOCOL = Cfg.NetCfg.Protocol
	NET_WEBSOCKET = Cfg.NetCfg.WebSocket == 1
	NET_MAX_NUMBER = Cfg.NetCfg.MaxNum
	SERVER_GROUP = Cfg.NetCfg.Group
	NET_GATE_SADDR = Cfg.NetCfg.SAddr
	NET_LISTEN_SADDR = NET_GATE_SADDR
	GAME_LOG_CONLOSE = Cfg.NetCfg.LogFile == -1
	if GAME_LOG_CONLOSE {
		GAME_LOG_LEVEL = 7
	} else {
		GAME_LOG_LEVEL = Cfg.NetCfg.LogFile
	}

	argv := len(os.Args)
	fmt.Println("启动参数个数argv: ", argv)
	fmt.Println("启动参数值argc：", os.Args)
	if argv > 6 {
		flag.StringVar(&SERVER_NAME, "n", "server", "server name")
		flag.StringVar(&NET_GATE_SADDR, "s", "127.0.0.1:6789", "server listen address") //	"127.0.0.1:6789"
		flag.IntVar(&Cfg.NetCfg.Uid, "u", 0, "server uid")

		flag.Parse() //parse之后参数才会被解析复制

		arrStr := strings.Split(NET_GATE_SADDR, ":")            //服发现使用正常的局域网ip
		NET_LISTEN_SADDR = fmt.Sprintf("0.0.0.0:%s", arrStr[1]) //socket监听,监听所有网卡绑定的ip，格式(0.0.0.0:port)(web监听格式也可以是(:port))
		Cfg.NetCfg.SAddr = NET_GATE_SADDR
		SERVER_NODE_UID = Cfg.NetCfg.Uid
	}
}
