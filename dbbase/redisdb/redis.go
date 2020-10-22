// mysql
package redisdb

import (
	"github.com/snowyyj001/loumiao/config"
	"github.com/snowyyj001/loumiao/log"
	"github.com/snowyyj001/loumiao/util"

	"github.com/gomodule/redigo/redis"
)

var (
	DB redis.Conn
)

//连接数据库
//url：数据库地址
///例如： Redis.Dial("127.0.0.1:6379")
func Dial(url string) error {
	conn, err := redis.Dial("tcp", url)
	if util.CheckErr(err) == false {
		DB = conn
	}
	log.Infof("redis dail success: %s", url)
	return err
}

//连接数据库,使用config-redis默认参数
func DialDefault() error {
	conn, err := redis.Dial("tcp", config.REDIS_URI)
	if util.CheckErr(err) {
		panic(err)
	}
	DB = conn

	log.Infof("redis DialDefault success: %s", config.REDIS_URI)

	return err
}
