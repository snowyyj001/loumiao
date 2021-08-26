// mysql
package redisdb

import (
	"fmt"
	"math/rand"
	"runtime"
	"time"

	"github.com/gomodule/redigo/redis"

	"github.com/snowyyj001/loumiao/config"
	"github.com/snowyyj001/loumiao/llog"
	"github.com/snowyyj001/loumiao/util"
)

var (
	pool *redis.Pool
)

//连接数据库
//url：数据库地址
///例如： Redis.Dial("127.0.0.1:6379")
func Dial(url string) error {
	llog.Debugf("redis Dial: %s", config.DBCfg.RedisUri)
	cpuNum := runtime.NumCPU()
	pool =& redis.Pool{
		MaxIdle:     cpuNum,
		Dial: func () (redis.Conn, error) {
			return redis.Dial("tcp", url)
		},
		TestOnBorrow: func(c redis.Conn, t time.Time) error {
			_, err := c.Do("PING")
			return err
		},
	}

	c, err := pool.Dial()
	if util.CheckErr(err) {
		return fmt.Errorf("redis Dial error: %s", url)
	}
	c.Close()
	llog.Infof("redis dail success: %s", url)

	return nil
}

//连接数据库,使用config-redis默认参数
func DialDefault() error {
	llog.Debugf("redis DialDefault: %s", config.DBCfg.RedisUri)
	return Dial(config.DBCfg.RedisUri)
}

//关闭连接
func Close() {
	pool.Close()
	pool = nil
}

///Do a command
func Do(command string ,args ...interface{}) (interface{},  error) {
	db := pool.Get()
	defer db.Close()
	return db.Do(command, args...)
}

func Set(args ...interface{}) {
	db := pool.Get()
	defer db.Close()
	db.Do("SET", args...)
}

//redis分布式锁,尝试expiretime毫秒后拿不到锁就返回0,否则返回锁的随机值
//@key: 锁key
//@expiretime: 锁的过期时间,毫秒,0代表立即返回锁结果
func AquireLock(key string, expiretime int) int {
	db := pool.Get()
	defer db.Close()
	val := rand.Intn(20000000) + 1
	et := expiretime
	if et <= 0 { //默认任何锁都是2s的过期时间
		et = 2000
	}
	ret, err := db.Do("SET", key, val, "NX", "PX", et)
	if ret == nil && expiretime > 0 {
		nt := util.TimeStamp()
		for {
			time.Sleep(50 * time.Millisecond) //等待50ms再次尝试
			ret, err = db.Do("SET", key, val, "NX", "PX", et)
			if ret != nil {
				break
			}
			if util.TimeStamp()-nt > int64(expiretime) {
				break
			}
		}
	}
	if err != nil {
		llog.Errorf("redis Lock error: key=%d, error=%s", key, err.Error())
	}
	if ret != nil {
		return val
	}
	return 0 			//没有拿到锁
}

func UnLock(key string, val int) {
	if val <= 0 {
		return
	}
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
		llog.Error(err.Error())
	}
	return v, err
}

// 自增
func StringIncr(name string) (int, error) {
	db := pool.Get()
	defer db.Close()
	v, err := redis.Int(db.Do("INCR", name))
	if err != nil {
		llog.Error(err.Error())
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
		llog.Error(err.Error())
	}
	return err
}

// 删除指定的键
func Delete(keys ...interface{}) (bool, error) {
	db := pool.Get()
	defer db.Close()
	v, err := redis.Bool(db.Do("DEL", keys...))
	if err != nil {
		llog.Error(err.Error())
	}
	return v, err
}

// 查看指定的长度
func StrLen(name string) (int, error) {
	db := pool.Get()
	defer db.Close()
	v, err := redis.Int(db.Do("STRLEN", name))
	if err != nil {
		llog.Error(err.Error())
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
		llog.Error(err.Error())
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
		llog.Error(err.Error())
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
		llog.Error(err.Error())
	}
	return v, err
}

// 获取hash 的键的个数
func HLen(name string) (int, error) {
	db := pool.Get()
	defer db.Close()
	v, err := redis.Int(db.Do("HLEN", name))
	if err != nil {
		llog.Error(err.Error())
	}
	return v, err
}

// 传入的 字段列表获得对应的值
func HMget(name string, fields ...string) ([]string, error) {
	db := pool.Get()
	defer db.Close()
	args := []interface{}{name}
	for _, field := range fields {
		args = append(args, field)
	}
	value, err := redis.Values(db.Do("HMGET", args...))
	if err != nil {
		llog.Error(err.Error())
	}
	return redis.Strings(value, err)
}

// 设置单个值, value 还可以是一个 map slice 等
func HSet(name string, key string, value interface{}) (err error) {
	db := pool.Get()
	defer db.Close()
	_, err = db.Do("HSET", name, key, value)
	if err != nil {
		llog.Error(err.Error())
	}
	return
}

// 设置多个值
func HMSet(name string, fields ...interface{}) (err error) {
	db := pool.Get()
	defer db.Close()
	//fmt.Println("HMSet", redis.Args{}.Add(name).Add(fields...))
	_, err = db.Do("HMSET", redis.Args{}.Add(name).Add(fields...)...)
	if err != nil {
		llog.Error(err.Error())
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
	return db.Do("SMEMBES", args)
}

// 获取集合中元素的个数
func ScardInt64s(name string) (int64, error) {
	db := pool.Get()
	defer db.Close()
	v, err := redis.Int64(db.Do("SCARD", name))
	if err != nil {
		llog.Error(err.Error())
	}
	return v, err
}

// sort set 有序集合
// 获取 sort set 集合中指定范围元素, 想要什么类型的自己指定, 递减排列
func ZRevrangeInt64(key string, start, stop int) ([]int64, error) {
	db := pool.Get()
	defer db.Close()
	return redis.Int64s(db.Do("ZREVRANGE", key, start, stop, "WITHSCORES"))
}

// 向 sort set 集合中添加一个或多个成员，或者更新已存在成员的分数
func ZAdd(key string, args ...interface{}) (int, error) {
	db := pool.Get()
	defer db.Close()
	return redis.Int(db.Do("ZADD", redis.Args{}.Add(key).AddFlat(args)...))
}

//选举leader，所有参与选举的人使用相同的value和prefix，leader负责设置value
//@prefix: 选举区分标识
//@value: 本次选举的值，每次发起选举，value应该和上次选举时的value不同
func AquireLeader(prefix string, value int) (isleader bool) {
	isleader = false
	val := AquireLock(prefix, 200)
	if val > 0 { 			//拿到锁了
		key := prefix + "leader"
		val, _ = GetInt(key)
		if val != value {		//还未被设置
			Set(key, value)
			isleader = true
		}
		UnLock(prefix, val)
	}
	return
}
