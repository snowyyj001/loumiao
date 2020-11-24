package define

//服务器地址
const ETCD_SADDR string = "/serveraddr/"

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
