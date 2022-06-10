package config

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"
)

const (
	ServerType_None      = iota //0 物理机控制节点
	ServerType_Gate             //1 网关
	ServerType_Account          //2 账号
	ServerType_World            //3 世界
	ServerType_Zone             //4 地图
	ServerType_DB               //5 数据库
	ServerType_Log              //6 日志
	ServerType_Public           //7 唯一公共服
	ServerType_WEB_GM           //8 web gm
	ServerType_WEB_LOGIN        //9 web login
	ServerType_RPCGate          //10 rpc gate
	ServerType_ETCF             //11 配置中心
	ServerType_LOGINQUEUE		//12 排队
	ServerType_Robot			//13 机器人
)

var (
	ServerNames = map[int]string{
		ServerType_None:      "machine",
		ServerType_Gate:      "gate",
		ServerType_Account:   "login",
		ServerType_World:     "lobby",
		ServerType_Zone:      "zone",
		ServerType_DB:        "db",
		ServerType_Log:       "logserver",
		ServerType_Public:    "publicserver",
		ServerType_WEB_GM:    "webserver",
		ServerType_WEB_LOGIN: "weblogin",
		ServerType_RPCGate:   "rpcserver",
		ServerType_ETCF:      "etcfserver",
		ServerType_LOGINQUEUE: "queueserver",
		ServerType_Robot: "robot",
	}
)

var (
	NET_NODE_AREAID = "-1"
	NET_NODE_ID     = -1               //节点id(服标识)
	NET_NODE_TYPE   = -1               //节点类型ServerType_*
	NET_GATE_SADDR  = "127.0.0.1:6789" //网关监听地址

	NET_PROTOCOL            = "PROTOBUF"      //消息协议格式："PROTOBUF" or "JSON"
	NET_WEBSOCKET           = false           //使用websocket or socket
	NET_MAX_CONNS           = 65535           //最大连接数
	NET_MAX_RPC_CONNS       = 1024            //rpc最大连接数
	NET_BUFFER_SIZE         = 1024 * 32       //最大消息包长度32k(对外)
	NET_CLUSTER_BUFFER_SIZE = 5 * 1024 * 1024 //最大消息包长度5M(对内)
	NET_MAX_NUMBER          = 10000           //pcu

	SERVER_GROUP     = "A"            //服务器分组
	SERVER_NAME      = "server"       //服务器名字
	SERVER_TYPE_NAME = "server"       //服务器类型名字
	SERVER_NODE_UID  = 0              //服务器uid
	NET_LISTEN_SADDR = "0.0.0.0:6789" //内网tcp监听地址
	SERVER_PARAM     = ""             //启动参数
	SERVER_RELEASE	 = false		  //配置上区分一下release和debug，方便开发期间的一些coding
	SERVER_DEBUGPORT	 = 0			  //pprof的监听端口,0不监听

)

//uid通过etcd自动分配，一般不要手动分配uid，除非清楚知道自己在做什么,参考GetServerUid
//uid和SAddr是一一对应的,可以通过删除ETCD_LOCKUID来重置uid的分配
type NetNode struct {
	Id        int    `json:"id"`
	Type      int    `json:"type"`
	SAddr     string `json:"saddr"`
	Param     string `json:"param"` //可选的启动参数，server根据自己的特殊需求配置具体内容
	Protocol  string `json:"protocol"`
	WebSocket int    `json:"websocket"`
	Uid       int    `json:"uid"`
	MaxNum    int    `json:"maxnum"`
	Group     string `json:"group"`
	LogFile   int    `json:"logfile"` //如果-1，代表输出到控制台
	Release   bool   `json:"release"`
	DebugPort int 	 `json:"debugport"`

}

type DBNode struct {
	SqlUri string `json:"sqluri"`
	Master int    `json:"master"`
}

type ServerCfg struct {
	NetCfg   NetNode  `json:"net"`
	EtcdAddr []string `json:"etcd"`
	NatsAddr []string `json:"nats"`
	RedisUri string   `json:"redisuri"`
	SqlCfg   []DBNode `json:"db"`
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

	argv := len(os.Args)
	fmt.Println("启动参数个数argv: ", argv)
	fmt.Println("启动参数值argc：", os.Args)
	if argv > 2 {
		var param string
		flag.StringVar(&param, "v", "{}", "json param cfg.json") //
		flag.Parse()                                             //parse之后参数才会被解析复制

		var scfg ServerCfg
		err = json.Unmarshal([]byte(param), &scfg)
		if err != nil {
			log.Fatalf("cfg fromat error: %s", err.Error())
		}
		Cfg.NetCfg.Id = scfg.NetCfg.Id
		Cfg.NetCfg.Uid = scfg.NetCfg.Uid
		Cfg.NetCfg.SAddr = scfg.NetCfg.SAddr
		Cfg.NetCfg.LogFile = scfg.NetCfg.LogFile
		Cfg.NetCfg.Param = scfg.NetCfg.Param
		Cfg.NetCfg.Release = scfg.NetCfg.Release
		Cfg.RedisUri = scfg.RedisUri
		Cfg.SqlCfg = scfg.SqlCfg
	}
	if Cfg.NetCfg.Uid == 0 {
		log.Fatalf("cfg content uid error: %d", Cfg.NetCfg.Uid)
	}

	NET_NODE_ID = Cfg.NetCfg.Id      //区服id
	SERVER_NODE_UID = Cfg.NetCfg.Uid //server uid
	NET_NODE_TYPE = Cfg.NetCfg.Type
	NET_PROTOCOL = Cfg.NetCfg.Protocol
	NET_WEBSOCKET = Cfg.NetCfg.WebSocket == 1
	NET_MAX_NUMBER = Cfg.NetCfg.MaxNum
	SERVER_GROUP = Cfg.NetCfg.Group
	NET_GATE_SADDR = Cfg.NetCfg.SAddr
	NET_LISTEN_SADDR = NET_GATE_SADDR
	SERVER_PARAM = Cfg.NetCfg.Param
	SERVER_RELEASE = Cfg.NetCfg.Release
	SERVER_DEBUGPORT = Cfg.NetCfg.DebugPort

	GAME_LOG_CONLOSE = Cfg.NetCfg.LogFile == -1 //-1log也输出到控制台，外网不需要输出到控制台
	if GAME_LOG_CONLOSE {
		GAME_LOG_LEVEL = 0
	} else {
		GAME_LOG_LEVEL = Cfg.NetCfg.LogFile
	}

	if typeName, ok := ServerNames[NET_NODE_TYPE]; ok {
		SERVER_TYPE_NAME = typeName
	}
	SERVER_NAME = fmt.Sprintf("%s-%d-%d", SERVER_TYPE_NAME, NET_NODE_TYPE, SERVER_NODE_UID)


	arrStr := strings.Split(NET_GATE_SADDR, ":")            //服发现使用正常的ip,例如 192.168.32.15:6789 127.0.0.1:6789
	NET_LISTEN_SADDR = fmt.Sprintf("0.0.0.0:%s", arrStr[1]) //socket监听,监听所有网卡绑定的ip，格式(0.0.0.0:port)(web监听格式也可以是(:port))
	Cfg.NetCfg.SAddr = NET_GATE_SADDR
	NET_NODE_AREAID = fmt.Sprintf("%d", NET_NODE_ID) //just for simple when need string type
}
