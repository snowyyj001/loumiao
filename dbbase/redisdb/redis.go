// mysql
package redisdb

import (
	"fmt"
	"math/rand"
	"runtime"
	"strconv"
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
	llog.Debugf("redis Dial: %s", config.Cfg.RedisUri)
	cpuNum := runtime.NumCPU()
	pool = &redis.Pool{
		MaxIdle: cpuNum,
		Dial: func() (redis.Conn, error) {
			return redis.Dial("tcp", url)
		},
		TestOnBorrow: func(c redis.Conn, t time.Time) error {
			if time.Since(t) < time.Second {
				return nil
			}
			_, err := c.Do("PING")
			if err != nil {
				llog.Fatalf("redis[%s] ping error: %s", url, err.Error())
			}
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
	llog.Debugf("redis DialDefault: %s", config.Cfg.RedisUri)
	return Dial(config.Cfg.RedisUri)
}

//关闭连接
func Close() {
	pool.Close()
	pool = nil
}

///Do a command
func Do(command string, args ...interface{}) (interface{}, error) {
	db := pool.Get()
	defer db.Close()
	v, err := db.Do(command, args...)
	if err != nil {
		llog.Errorf("Do: command = %s, args = %v, err = %s", command, args, err.Error())
	}
	return v, err
}

func Set(args ...interface{}) (interface{}, error) {
	db := pool.Get()
	defer db.Close()
	v, err := db.Do("SET", args...)
	if err != nil {
		llog.Errorf("Set: args = %v, err = %s", args, err.Error())
	}
	return v, err
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
		llog.Errorf("redis Lock error: key=%s, error=%s", key, err.Error())
	}
	if ret != nil {
		return val
	}
	return 0 //没有拿到锁
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

// String is a helper that converts a command reply to a string. If err is not
// equal to nil, then String returns "", err. Otherwise String converts the
// reply to a string as follows:
//
//  Reply type      Result
//  bulk string     string(reply), nil
//  simple string   reply, nil
//  nil             "",  nil
//  other           "",  error
func String(reply interface{}, err error) (string, error) {
	if err != nil {
		return "", err
	}
	switch reply := reply.(type) {
	case []byte:
		return string(reply), nil
	case string:
		return reply, nil
	case nil:
		return "", nil		//不使用redis.String，这里行为不一样
	case redis.Error:
		return "", reply
	}
	return "", fmt.Errorf("redisdb: unexpected type for String, got type %T", reply)
}

func GetString(key string) (string, error) {
	db := pool.Get()
	defer db.Close()
	reply, err := db.Do("GET", key)
	v, err := String(reply, err)
	if err != nil {
		llog.Errorf("GetString: key = %s, err = %s", key, err.Error())
	}
	return v, err
}

// Int is a helper that converts a command reply to an integer. If err is not
// equal to nil, then Int returns 0, err. Otherwise, Int converts the
// reply to an int as follows:
//
//  Reply type    Result
//  integer       int(reply), nil
//  bulk string   parsed reply, nil
//  nil           0, nil
//  other         0, error
func Int(reply interface{}, err error) (int, error) {
	if err != nil {
		return 0, err
	}
	switch reply := reply.(type) {
	case int64:
		x := int(reply)
		if int64(x) != reply {
			return 0, strconv.ErrRange
		}
		return x, nil
	case []byte:
		n, err := strconv.ParseInt(string(reply), 10, 0)
		return int(n), err
	case nil:
		return 0, nil		//不使用redis.Int，这里行为不一样
	case redis.Error:
		return 0, reply
	}
	return 0, fmt.Errorf("redigo: unexpected type for Int, got type %T", reply)
}

func GetInt(key string) (int, error) {
	db := pool.Get()
	defer db.Close()
	v, err := Int(db.Do("GET", key))
	if err != nil {
		llog.Errorf("GetInt: key = %s, err = %s", key, err.Error())
	}
	return v, err
}

// Int64 is a helper that converts a command reply to 64 bit integer. If err is
// not equal to nil, then Int64 returns 0, err. Otherwise, Int64 converts the
// reply to an int64 as follows:
//
//  Reply type    Result
//  integer       reply, nil
//  bulk string   parsed reply, nil
//  nil           0, nil
//  other         0, error
func Int64(reply interface{}, err error) (int64, error) {
	if err != nil {
		return 0, err
	}
	switch reply := reply.(type) {
	case int64:
		return reply, nil
	case []byte:
		n, err := strconv.ParseInt(string(reply), 10, 64)
		return n, err
	case nil:
		return 0, nil
	case redis.Error:
		return 0, reply
	}
	return 0, fmt.Errorf("redigo: unexpected type for Int64, got type %T", reply)
}

func GetInt64(key string) (int64, error) {
	db := pool.Get()
	defer db.Close()
	v, err := Int64(db.Do("GET", key))
	if err != nil {
		llog.Errorf("GetInt64: key = %s, err = %s", key, err.Error())
	}
	return v, err
}

// 判断所在的 key 是否存在
func Exist(key string) (bool, error) {
	db := pool.Get()
	defer db.Close()
	v, err := redis.Bool(db.Do("EXISTS", key))
	if err != nil {
		llog.Errorf("Exist: name = %s, err = %s", key, err.Error())
	}
	return v, err
}

// 自增
func StringIncr(key string) (int, error) {
	db := pool.Get()
	defer db.Close()
	v, err := Int(db.Do("INCR", key))
	if err != nil {
		llog.Errorf("StringIncr: name = %s, err = %s", key, err.Error())
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
		llog.Errorf("Delete: name = %s, newSecondsLifeTime = %d, err = %s", name, newSecondsLifeTime, err.Error())
	}
	return err
}

// Bool is a helper that converts a command reply to a boolean. If err is not
// equal to nil, then Bool returns false, err. Otherwise Bool converts the
// reply to boolean as follows:
//
//  Reply type      Result
//  integer         value != 0, nil
//  bulk string     strconv.ParseBool(reply)
//  nil             false, ErrNil
//  other           false, error
func Bool(reply interface{}, err error) (bool, error) {
	if err != nil {
		return false, err
	}
	switch reply := reply.(type) {
	case int64:
		return reply != 0, nil
	case []byte:
		return strconv.ParseBool(string(reply))
	case nil:
		return false, nil
	case redis.Error:
		return false, reply
	}
	return false, fmt.Errorf("redigo: unexpected type for Bool, got type %T", reply)
}

// 删除指定的键
func Delete(keys ...interface{}) (bool, error) {
	db := pool.Get()
	defer db.Close()
	v, err := Bool(db.Do("DEL", keys...))
	if err != nil {
		llog.Errorf("Delete: keys = %v, err = %s", keys, err.Error())
	}
	return v, err
}

// 查看指定的长度
func StrLen(name string) (int, error) {
	db := pool.Get()
	defer db.Close()
	v, err := Int(db.Do("STRLEN", name))
	if err != nil {
		llog.Errorf("StrLen: name =%s, err = %s", name, err.Error())
	}
	return v, err
}

// //  hash ///
// 删除指定的 hash 键
func HDel(name, key string) (bool, error) {
	db := pool.Get()
	defer db.Close()
	var err error
	v, err := Bool(db.Do("HDEL", name, key))
	if err != nil {
		llog.Errorf("HMDel: name =%s, key = %s, err = %s", name, key, err.Error())
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
	v, err := Bool(db.Do("HDEL", args...))
	if err != nil {
		llog.Errorf("HMDel: name =%s, fields = %v, err = %s", name, fields, err.Error())
	}
	return v, err
}

// 查看hash 中指定是否存在
func HExists(name, field string) (bool, error) {
	db := pool.Get()
	defer db.Close()
	var err error
	v, err := Bool(db.Do("HEXISTS", name, field))
	if err != nil {
		llog.Errorf("HExists: name =%s, field = %s, err = %s", name, field, err.Error())
	}
	return v, err
}

// 获取hash 的键的个数
func HLen(name string) (int, error) {
	db := pool.Get()
	defer db.Close()
	v, err := Int(db.Do("HLEN", name))
	if err != nil {
		llog.Errorf("HMget: name =%s, err = %s", name, err.Error())
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
		llog.Errorf("HMget: name =%s, fields = %v, err = %s", name, fields, err.Error())
	}
	return redis.Strings(value, err)
}

// 设置单个值, value 还可以是一个 map slice 等
func HSet(name string, key string, value interface{}) (err error) {
	db := pool.Get()
	defer db.Close()
	_, err = db.Do("HSET", name, key, value)
	if err != nil {
		llog.Errorf("HSet: name =%s, key = %s, value = %v, err = %s", name, key, value, err.Error())
	}
	return
}

// 设置多个值
func HMSet(key string, fields ...interface{}) (err error) {
	db := pool.Get()
	defer db.Close()
	//fmt.Println("HMSet", redis.Args{}.Add(name).Add(fields...))
	_, err = db.Do("HMSET", redis.Args{}.Add(key).Add(fields...)...)
	if err != nil {
		llog.Errorf("HMSet: key = %s, fields = %v, err = %s", key, fields, err.Error())
	}
	return
}

// 获取单个hash 中的值
func HGetString(key, field string) (string, error) {
	db := pool.Get()
	defer db.Close()
	v, err := String(db.Do("HGET", key, field))
	if err != nil {
		llog.Errorf("HGET: key = %s, field = %s, err = %s", key, field, err.Error())
	}
	return v, err
}

// 获取单个hash 中的值
func HGetInt(key, field string) (int, error) {
	db := pool.Get()
	defer db.Close()
	v, err := Int(db.Do("HGET", key, field))
	if err != nil {
		llog.Errorf("HGetInt: key = %s, field = %s, err = %s", key, field, err.Error())
	}
	return v, err
}

// 获取单个hash 中的值
func HGetInt64(key, field string) (int64, error) {
	db := pool.Get()
	defer db.Close()
	v, err := Int64(db.Do("HGET", key, field))
	if err != nil {
		llog.Errorf("HGetInt64: key = %s, field = %s, err = %s", key, field, err.Error())
	}
	return v, err
}

// set 集合
// 获取 set 集合中所有的元素, 想要什么类型的自己指定
func Smembers(args ...interface{}) (interface{}, error) {
	db := pool.Get()
	defer db.Close()
	v, err := db.Do("SMEMBES", args)
	if err != nil {
		llog.Errorf("Smembers:  args = %v, err = %s", args, err.Error())
	}
	return v, err
}

// 向 set 集合中添加一个或多个成员
func SAdd(key string, args ...interface{}) (int, error) {
	db := pool.Get()
	defer db.Close()
	v, err := Int(db.Do("SADD", redis.Args{}.Add(key).AddFlat(args)...))
	if err != nil {
		llog.Errorf("SAdd: key = %s, args = %v, err = %s", key, args, err.Error())
	}
	return v, err
}

//删除 set 集合中一个或多个成员
func SRem(key string, args ...interface{}) (int, error) {
	db := pool.Get()
	defer db.Close()
	v, err := Int(db.Do("SREM ", redis.Args{}.Add(key).AddFlat(args)...))
	if err != nil {
		llog.Errorf("SRem: key = %s, args = %v, err = %s", key, args, err.Error())
	}
	return v, err
}

// 获取集合中元素的个数
func ScardInt(key string) (int, error) {
	db := pool.Get()
	defer db.Close()
	v, err := Int(db.Do("SCARD", key))
	if err != nil {
		llog.Errorf("ScardInt64s: key = %s, err = %s", key, err.Error())
	}
	return v, err
}

//判断 member 元素是否是集合 key 的成员
func SisMember(key string, member interface{}) (int, error) {
	db := pool.Get()
	defer db.Close()
	v, err := Int(db.Do("SISMEMBER ", key, member))
	if err != nil {
		llog.Errorf("SisMember: key = %s, member = %v, err = %s", key, member, err.Error())
	}
	return v, err
}

// 迭代集合中键的元素
func Sscan(key string, cursor int) ([]string, int) {
	db := pool.Get()
	defer db.Close()
	results := make([]string, 0)
	v, err := redis.Values(db.Do("SSCAN ", key, cursor, "COUNT", 100))
	if err != nil {
		llog.Errorf("Sscan: key = %s, SSCAN %s", key, err.Error())
		return results, 0
	}
	v1, _ := redis.Values(v[1], nil)
	if err = redis.ScanSlice(v1, &results); err != nil {
		llog.Errorf("Sscan:  key = %s, ScanSlice %s", key, err.Error())
		return results, 0
	}
	if n, err := Int(v[0], nil); err != nil {
		llog.Errorf("Sscan: key = %s, Int %s", key, err.Error())
		return results, 0
	} else {
		return results, n
	}
}

// 迭代集合中的元素
func SscanSimple(key string, call func(members string) ) {
	var c int
	for {
		res, n := Sscan(key, c)
		for _, st := range res {
			call(st)
		}
		c = n
		if c == 0 {
			break
		}
	}
}

// sort set 有序集合
// 获取 sort set 集合中指定范围元素, 想要什么类型的自己指定, 递减排列
func ZRevrangeInt64(key string, start, stop int) ([]int64, error) {
	db := pool.Get()
	defer db.Close()
	v, err := redis.Int64s(db.Do("ZREVRANGE", key, start, stop, "WITHSCORES"))
	if err != nil && err != redis.ErrNil {
		llog.Errorf("ZRevrangeInt64: key = %s, start = %d, stop = %d, err = %s", key, start, stop, err.Error())
	}
	return v, err
}

// 向 sort set 集合中添加一个或多个成员，或者更新已存在成员的分数
func ZAdd(key string, args ...interface{}) (int, error) {
	db := pool.Get()
	defer db.Close()
	v, err := Int(db.Do("ZADD", redis.Args{}.Add(key).AddFlat(args)...))
	if err != nil {
		llog.Errorf("ZAdd: key = %s, args = %v, err = %s", key, args, err.Error())
	}
	return v, err
}


// 移除有序集中的一个或多个成员，不存在的成员将被忽略。
func Zrem (key string, args ...interface{}) (int, error) {
	db := pool.Get()
	defer db.Close()
	v, err := Int(db.Do("ZREM ", redis.Args{}.Add(key).AddFlat(args)...))
	if err != nil {
		llog.Errorf("Zrem: key = %s, args = %v, err = %s", key, args, err.Error())
	}
	return v, err
}

// 返回有序集中，成员的分数值
func Zscore(key string, member interface{}) (int, error) {
	db := pool.Get()
	defer db.Close()
	v, err := Int(db.Do("ZSCORE", key, member))
	if err != nil {
		llog.Errorf("Zscore: key = %s, member = %v, err = %s", key, member, err.Error())
	}
	return v, err
}

type ZSetUnit struct {
	Key string
	Score int
}

// 迭代有序集合中的元素（包括元素成员和元素分值）
func Zscan (key string, cursor int) ([]ZSetUnit, int) {
	db := pool.Get()
	defer db.Close()
	results := make([]ZSetUnit, 0)
	v, err := redis.Values(db.Do("ZSCAN", key, cursor, "COUNT", 100))
	if err != nil {
		llog.Errorf("Zscan: ZSCAN key = %s, err = %s", key, err.Error())
		return results, 0
	}
	v1, _ := redis.Values(v[1], nil)
	if err = redis.ScanSlice(v1, &results); err != nil {
		llog.Errorf("Zscan: ScanSlice key = %s, err = %s", key, err.Error())
		return results, 0
	}
	if n, err := Int(v[0], nil); err != nil {
		llog.Errorf("Zscan: Int key = %s, err = %s", key, err.Error())
		return results, 0
	} else {
		return results, n
	}
}

// 迭代有序集合中的元素（包括元素成员和元素分值）
func ZscanSimple(key string, call func(key string, score int) )  {
	var c int
	for {
		res, n := Zscan(key, c)
		for _, st := range res {
			call(st.Key, st.Score)
		}
		c = n
		if c == 0 {
			break
		}
	}
}


//选举leader，所有参与选举的人使用相同的value和prefix，leader负责设置value
//@prefix: 选举区分标识
//@value: 本次选举的值，每次发起选举，value应该和上次选举时的value不同
func AquireLeader(prefix string, value int) (isleader bool) {
	isleader = false
	val := AquireLock(prefix, 200)
	if val > 0 { //拿到锁了
		key := prefix + "leader"
		val, _ = GetInt(key)
		if val != value { //还未被设置
			Set(key, value)
			isleader = true
		}
		UnLock(prefix, val)
	}
	return
}
