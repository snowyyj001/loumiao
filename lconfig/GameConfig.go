package lconfig

var (
	// send update to all channels
	//	nextKeepAlive := time.Now().Add((time.Duration(karesp.TTL) * time.Second) / 3.0)
	//	ka.deadline = time.Now().Add(time.Duration(karesp.TTL) * time.Second)
	GAME_LEASE_TIME  = 9    //etcd租约过期时间，续租时间是GAME_LEASE_TIME/3=3秒, karesp.TTL == GAME_LEASE_TIME
	GAME_LOG_CONLOSE = true //log是否输出到控制台
	GAME_LOG_LEVEL   = 0    //log输出级别

	NET_MAX_CONNS           = 50000            //最大连接数
	NET_MAX_RPC_CONNS       = 1024             //rpc最大连接数
	NET_BUFFER_SIZE         = 1024 * 32        //最大消息包长度32k(对外)
	NET_CLUSTER_BUFFER_SIZE = 16 * 1024 * 1024 //最大消息包长度16M(对内)
	NET_MAX_NUMBER          = 10000            //pcu
	NET_MAX_WRITE_CHANSIZE  = 20000            //socket写缓冲
)
