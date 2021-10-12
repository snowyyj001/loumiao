package config

var (
	// send update to all channels
	//	nextKeepAlive := time.Now().Add((time.Duration(karesp.TTL) * time.Second) / 3.0)
	//	ka.deadline = time.Now().Add(time.Duration(karesp.TTL) * time.Second)
	GAME_LEASE_TIME  = 9     //etcd租约过期时间，续租时间是GAME_LEASE_TIME/3=3秒, karesp.TTL == GAME_LEASE_TIME
	GAME_RPC_LENGTH  = 20000 //rpc之间的chan缓冲大小
	GAME_LOG_CONLOSE = true  //log是否输出到控制台
	GAME_LOG_LEVEL   = 0     //log输出级别
)
