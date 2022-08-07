package define

//分布式锁-uid
const ETCD_LOCKUID string = "/lockuid/"

//服务器信息注册
const ETCD_NODEINFO string = "/nodeinfos/" // /nodeinfos/areaid/host/

//node状态
const ETCD_NODESTATUS string = "/nodestatus/" // /nodestatus/areaid/host/

//分布式锁-world position
const KEY_LOCKWORLD string = "/lockworldpos/"

const (
	TIME_LOCK_WUID = 3000 //分布式锁超时时间-ETCD_LOCKWORLD
)

const (
	CLIENT_NONE       = iota
	CLIENT_CONNECT    //客户端建立连接
	CLIENT_DISCONNECT //客户端断开连接
)
const ( //kafka消息topic
	TOPIC_SERVER_MAIL = "tp:servermail" //server关键信息，发送邮件
	TOPIC_SERVER_LOG  = "tp:loglevel"   //日志级别调整
)

const ( //发送真实邮件类型
	MAIL_TYPE_ERR    = 0 //服务器发生error
	MAIL_TYPE_WARING = 1 //服务器发生需要注意的警告
	MAIL_TYPE_START  = 5 //服务器启动
	MAIL_TYPE_STOP   = 6 //服务器关闭
	MAIL_SYS_WARN    = 7 //系统资源告警
)

const (
	RPCMSG_FLAG_CALL  = 1 << 0 //rpc调用是call方式,否则就是异步send
	RPCMSG_FLAG_RESP  = 1 << 1 //rpc调用是远端call返回
	RPCMSG_FLAG_BROAD = 1 << 2 //rpc广播
)

var (
	RESERVED_PORT = map[int]bool{ //保留的端口，自己服务不使用这些
		3306:  true, //mysql
		6379:  true, //redis
		27017: true, //mongodb
		2379:  true, //etcd
		2380:  true, //etcd
		22379: true, //etcd
		22380: true, //etcd
		32379: true, //etcd
		32380: true, //etcd
		4222:  true, //nats
		6222:  true, //nats
		8222:  true, //nats
		5044:  true, //logstash
		8000:  true, //web
		80:    true, //web
		8080:  true, //web
		9000:  true, //docker Portainer
		9001:  true, //docker Portainer
	}
)
