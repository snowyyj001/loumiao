package config

var (
	NET_GATE_IP     = "127.0.0.1" //网关监听地址
	NET_GATE_PORT   = 6789        //网关监听端口
	NET_PROTOCOL    = "PROTOBUF"  //消息协议格式："PROTOBUF" or "JSON"
	NET_WEBSOCKET   = false       //使用websocket or socket
	NET_MAX_CONNS   = 20000       //最大连接数
	NET_BUFFER_SIZE = 1024 * 64   //最大消息包长度64k
)
