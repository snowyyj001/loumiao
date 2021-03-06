package config

var (
	GAME_LEASE_TIME  = 3     //etcd租约过期时间，续租时间是GAME_LEASE_TIME/3=1秒
	GAME_RPC_LENGTH  = 20000 //rpc之间的chan缓冲大小
	GAME_LOG_CONLOSE = true  //log是否输出到控制台
	GAME_LOG_JSON    = false //log格式是否使用json
	GAME_LOG_LEVEL   = 7     //log输出级别
	/**llog level
	LOGGER_LEVEL_EMERGENCY = iota
	LOGGER_LEVEL_CRITICAL
	LOGGER_LEVEL_ERROR
	LOGGER_LEVEL_WARNING
	LOGGER_LEVEL_NOTICE
	LOGGER_LEVEL_INFO
	LOGGER_LEVEL_DEBUG
	*/
)
