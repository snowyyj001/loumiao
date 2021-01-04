// mysql
package redisdb

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/gomodule/redigo/redis"

	"github.com/snowyyj001/loumiao/config"
	"github.com/snowyyj001/loumiao/log"
	"github.com/snowyyj001/loumiao/util"
)

const (
	POOL_SIZE = 10
)

var (
	pool *redis.Pool
)

//连接数据库
//url：数据库地址
///例如： Redis.Dial("127.0.0.1:6379")
func Dial(url string) error {
	log.Debugf("redis Dial: %s", config.DBCfg.RedisUri)
	pool = redis.NewPool(func() (redis.Conn, error) {
		c, err := redis.Dial("tcp", url)
		if err != nil {
			return nil, err
		}
		return c, nil
	}, POOL_SIZE)

	util.Assert(pool)
	c, err := pool.Dial()
	if util.CheckErr(err) {
		return fmt.Errorf("redis Dial error: %s", url)
	}
	c.Close()
	log.Infof("redis dail success: %s", url)

	return nil
}

//连接数据库,使用config-redis默认参数
func DialDefault() error {
	log.Debugf("redis DialDefault: %s", config.DBCfg.RedisUri)
	pool = redis.NewPool(func() (redis.Conn, error) {
		c, err := redis.Dial("tcp", config.DBCfg.RedisUri)
		if err != nil {
			return nil, err
		}
		return c, nil
	}, POOL_SIZE)
	util.Assert(pool)
	c, err := pool.Dial()
	if util.CheckErr(err) {
		return fmt.Errorf("redis DialDefault error: %s", config.DBCfg.RedisUri)
	}
	c.Close()
	log.Infof("redis DialDefault success: %s", config.DBCfg.RedisUri)

	return nil
}

//关闭连接
func Close() {
	pool.Close()
	pool = nil
}

///Do a func can do no defer close
func Do(database int, pFunc func(c redis.Conn) (reply interface{}, err error)) (reply interface{}, err error) {
	db := pool.Get()
	defer db.Close()

	return pFunc(db)
}

func Set(args ...interface{}) {
	db := pool.Get()
	defer db.Close()
	db.Do("SET", args...)
}

//redis分布式锁,尝试expiretime毫秒后拿不到锁就返回0,否则返回锁的随机值
//@key: 锁key
//@expiretime: 锁的过期时间,毫秒
func Lock(key string, expiretime int) int {
	db := pool.Get()
	defer db.Close()
	val := rand.Int()
	ret, err := db.Do("SET", key, val, "NX", "PX", expiretime)
	if ret == nil {
		nt := util.TimeStamp()
		for {
			time.Sleep(50 * time.Millisecond) //等待50ms再次尝试
			ret, err = db.Do("SET", key, val, "NX", "PX", expiretime)
			if ret != nil {
				break
			}
			if util.TimeStamp()-nt > int64(expiretime) {
				break
			}
		}
	}
	if err != nil {
		log.Errorf("redis Lock error: key=%d, error=%s", key, err.Error())
	}
	if ret != nil {
		return 0
	}
	return val
}

func UnLock(key string, val int) {
	db := pool.Get()
	defer db.Close()
	luascript := `
		if redis.call('get',KEYS[1]) == ARGV[1] then
			return redis.call('del',KEYS[1]) 
		else
	   		return 0
		end
	`
	db.Do("eval", luascript, 1, key, val)
}

func GetString(key string) (string, error) {
	db := pool.Get()
	defer db.Close()
	return redis.String(db.Do("GET", key))
}

func GetInt(key string) (int, error) {
	db := pool.Get()
	defer db.Close()
	return redis.Int(db.Do("GET", key))
}

func GetInt64(key string) (int64, error) {
	db := pool.Get()
	defer db.Close()
	return redis.Int64(db.Do("GET", key))
}

// 判断所在的 key 是否存在
func Exist(name string) (bool, error) {
	db := pool.Get()
	defer db.Close()
	v, err := redis.Bool(db.Do("EXISTS", name))
	if err != nil {
		log.Error(err.Error())
	}
	return v, err
}

// 自增
func StringIncr(name string) (int, error) {
	db := pool.Get()
	defer db.Close()
	v, err := redis.Int(db.Do("INCR", name))
	if err != nil {
		log.Error(err.Error())
	}
	return v, err
}

// 设置过期时间 （单位 秒）
func Expire(name string, newSecondsLifeTime int64) error {
	db := pool.Get()
	defer db.Close()
	// 设置key 的过期时间
	_, err := db.Do("EXPIRE", name, newSecondsLifeTime)
	if err != nil {
		log.Error(err.Error())
	}
	return err
}

// 删除指定的键
func Delete(keys ...interface{}) (bool, error) {
	db := pool.Get()
	defer db.Close()
	v, err := redis.Bool(db.Do("DEL", keys...))
	if err != nil {
		log.Error(err.Error())
	}
	return v, err
}

// 查看指定的长度
func StrLen(name string) (int, error) {
	db := pool.Get()
	defer db.Close()
	v, err := redis.Int(db.Do("STRLEN", name))
	if err != nil {
		log.Error(err.Error())
	}
	return v, err
}

// //  hash ///
// 删除指定的 hash 键
func HDel(name, key string) (bool, error) {
	db := pool.Get()
	defer db.Close()
	var err error
	v, err := redis.Bool(db.Do("HDEL", name, key))
	if err != nil {
		log.Error(err.Error())
	}
	return v, err
}

// //  hash ///
// 删除指定的 hash 键
func HMDel(name string, fields ...string) (bool, error) {
	db := pool.Get()
	defer db.Close()

	args := []interface{}{name}
	for _, field := range fields {
		args = append(args, field)
	}
	var err error
	v, err := redis.Bool(db.Do("HDEL", args...))
	if err != nil {
		log.Error(err.Error())
	}
	return v, err
}

// 查看hash 中指定是否存在
func HExists(name, field string) (bool, error) {
	db := pool.Get()
	defer db.Close()
	var err error
	v, err := redis.Bool(db.Do("HEXISTS", name, field))
	if err != nil {
		log.Error(err.Error())
	}
	return v, err
}

// 获取hash 的键的个数
func HLen(name string) (int, error) {
	db := pool.Get()
	defer db.Close()
	v, err := redis.Int(db.Do("HLEN", name))
	if err != nil {
		log.Error(err.Error())
	}
	return v, err
}

// 传入的 字段列表获得对应的值
func HMget(name string, fields ...string) ([]interface{}, error) {
	db := pool.Get()
	defer db.Close()
	args := []interface{}{name}
	for _, field := range fields {
		args = append(args, field)
	}
	value, err := redis.Values(db.Do("HMGET", args...))
	if err != nil {
		log.Error(err.Error())
	}
	return value, err
}

// 设置单个值, value 还可以是一个 map slice 等
func HSet(name string, key string, value interface{}) (err error) {
	db := pool.Get()
	defer db.Close()
	_, err = db.Do("HSET", name, key, value)
	if err != nil {
		log.Error(err.Error())
	}
	return
}

// 设置多个值 , obj 可以是指针 slice map struct
func HMSet(name string, obj interface{}) (err error) {
	db := pool.Get()
	defer db.Close()
	_, err = db.Do("HMSET", redis.Args{}.Add(name).AddFlat(obj)...)
	if err != nil {
		log.Error(err.Error())
	}
	return
}

// 获取单个hash 中的值
func HGetString(name, field string) (string, error) {
	db := pool.Get()
	defer db.Close()
	return redis.String(db.Do("HGET", name, field))
}

// 获取单个hash 中的值
func HGetInt(name, field string) (int, error) {
	db := pool.Get()
	defer db.Close()
	return redis.Int(db.Do("HGET", name, field))
}

// 获取单个hash 中的值
func HGetInt64(name, field string) (int64, error) {
	db := pool.Get()
	defer db.Close()
	return redis.Int64(db.Do("HGET", name, field))
}

func Int64(val interface{}) (int64, error) {
	return redis.Int64(val, nil)
}

func Int(val interface{}) (int, error) {
	return redis.Int(val, nil)
}

func String(val interface{}) (string, error) {
	return redis.String(val, nil)
}

// set 集合

// 获取 set 集合中所有的元素, 想要什么类型的自己指定
func Smembers(args ...interface{}) (interface{}, error) {
	db := pool.Get()
	defer db.Close()
	return db.Do("smembers", args)
}

// 获取集合中元素的个数
func ScardInt64s(name string) (int64, error) {
	db := pool.Get()
	defer db.Close()
	v, err := redis.Int64(db.Do("SCARD", name))
	if err != nil {
		log.Error(err.Error())
	}
	return v, err
}
