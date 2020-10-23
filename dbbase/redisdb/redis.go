// mysql
package redisdb

import (
	"github.com/snowyyj001/loumiao/config"
	"github.com/snowyyj001/loumiao/log"
	"github.com/snowyyj001/loumiao/util"

	"github.com/gomodule/redigo/redis"
)

var (
	db redis.Conn
)

//连接数据库
//url：数据库地址
///例如： Redis.Dial("127.0.0.1:6379")
func Dial(url string) error {
	conn, err := redis.Dial("tcp", url)
	if util.CheckErr(err) == false {
		db = conn
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
	db = conn

	log.Infof("redis DialDefault success: %s", config.REDIS_URI)

	return err
}

//关闭连接
func Close() {
	if db != nil {
		db.Flush()
		db.Close()
		db = nil
	}
}

func Set(args ...interface{}) {
	db.Do("SET", args...)
}

func GetString(key string) (string, error) {
	return redis.String(db.Do("GET", key))
}

func GetInt(key string) (int, error) {
	return redis.Int(db.Do("GET", key))
}

func GetInt64(key string) (int64, error) {
	return redis.Int64(db.Do("GET", key))
}

// 判断所在的 key 是否存在
func Exist(name string) (bool, error) {
	v, err := redis.Bool(db.Do("EXISTS", name))
	return v, err
}

// 自增
func StringIncr(name string) (int, error) {
	v, err := redis.Int(db.Do("INCR", name))
	return v, err
}

// 设置过期时间 （单位 秒）
func Expire(name string, newSecondsLifeTime int64) error {
	// 设置key 的过期时间
	_, err := db.Do("EXPIRE", name, newSecondsLifeTime)
	return err
}

// 删除指定的键
func Delete(keys ...interface{}) (bool, error) {
	v, err := redis.Bool(db.Do("DEL", keys...))
	return v, err
}

// 查看指定的长度
func StrLen(name string) (int, error) {
	v, err := redis.Int(db.Do("STRLEN", name))
	return v, err
}

// //  hash ///
// 删除指定的 hash 键
func HDel(name, key string) (bool, error) {
	var err error
	v, err := redis.Bool(db.Do("HDEL", name, key))
	return v, err
}

// 查看hash 中指定是否存在
func HExists(name, field string) (bool, error) {
	var err error
	v, err := redis.Bool(db.Do("HEXISTS", name, field))
	return v, err
}

// 获取hash 的键的个数
func HLen(name string) (int, error) {
	v, err := redis.Int(db.Do("HLEN", name))
	return v, err
}

// 传入的 字段列表获得对应的值
func HMget(name string, fields ...string) ([]interface{}, error) {
	args := []interface{}{name}
	for _, field := range fields {
		args = append(args, field)
	}
	value, err := redis.Values(db.Do("HMGET", args...))

	return value, err
}

// 设置单个值, value 还可以是一个 map slice 等
func HSet(name string, key string, value interface{}) (err error) {
	_, err = db.Do("HSET", name, key, value)
	return
}

// 设置多个值 , obj 可以是指针 slice map struct
func HMSet(name string, obj interface{}) (err error) {
	_, err = db.Do("HSET", redis.Args{}.Add(name).AddFlat(&obj)...)
	return
}

// 获取单个hash 中的值
func HGetString(name, field string) (string, error) {
	return redis.String(db.Do("HGET", name, field))
}

// 获取单个hash 中的值
func HGetInt(name, field string) (int, error) {
	return redis.Int(db.Do("HGET", name, field))
}

// 获取单个hash 中的值
func HGetInt64(name, field string) (int64, error) {
	return redis.Int64(db.Do("HGET", name, field))
}

// set 集合

// 获取 set 集合中所有的元素, 想要什么类型的自己指定
func Smembers(args ...interface{}) (interface{}, error) {
	return db.Do("smembers", args)
}

// 获取集合中元素的个数
func ScardInt64s(name string) (int64, error) {
	v, err := redis.Int64(db.Do("SCARD", name))
	return v, err
}
