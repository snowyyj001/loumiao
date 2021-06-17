package define

//分布式锁-uid
const ETCD_LOCKUID string = "/lockuid/"

//服务器信息注册
const ETCD_NODEINFO string = "/nodeinfos/"

//node状态
const ETCD_NODESTATUS string = "/nodestatus/"

//分布式锁-world position
const KEY_LOCKWORLD string = "/lockworldpos/"

const (
	TIME_LOCK_WUID = 3000 //分布式锁超时时间-ETCD_LOCKWORLD
)

const (
	CLIENT_CONNECT    = 0 //客户端建立连接
	CLIENT_DISCONNECT = 1 //客户端断开连接
)
const ( //kafka消息topic
	TOPIC_SERVER_MAIL = "tp:servermail" //server关键信息，发送邮件
)

const ( //发送真实邮件类型
	MAIL_TYPE_ERR   = 0 //服务器发生error
	MAIL_TYPE_START = 1 //服务器启动
	MAIL_TYPE_STOP  = 2 //服务器关闭
	MAIL_SYS_WARN   = 3 //系统资源告警
)
