package define

//分布式锁-uid
const ETCD_LOCKUID string = "/lockuid/"

//服务器信息注册
const ETCD_NODEINFO string = "/nodeinfos/"

//node状态
const ETCD_NODESTATUS string = "/nodestatus/"

//分布式锁-world position
const ETCD_LOCKWORLD string = "/lockworldpos/"

const (
	TIME_LOCK_WUID = 3 //分布式锁超时时间-ETCD_LOCKWORLD
)

const (
	CLIENT_CONNECT    = 0 //客户端建立连接
	CLIENT_DISCONNECT = 1 //客户端断开连接
)
