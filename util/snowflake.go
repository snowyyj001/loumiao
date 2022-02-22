/*
* Snowflake
*
* 1                                         37                51              64
* +-----------------------------------------+-----------------+---------------+
* | timestamp(0.01s)                        |    workerid     |   sequence    |
* +-----------------------------------------+-----------------+---------------+
* | 0000000000 0000000000 0000000000 000000 | 0000000000 0000 | 0000000000 00 |
* +-----------------------------------------+-----------------+---------------+
* 0. 最高位1bit，其值始终是0
* 1. 36位时间截(0.01秒)，注意这是时间截的差值（当前时间截 - 开始时间截)。可以使用约21年: (1L << 36) / (100L * 60 * 60 * 24 * 365) = 21
* 2. 14位数据机器位，可以部署在16383(0x3FFF)个节点
* 3. 12位序列，0.01秒内的计数，同一机器，同一时间截并发4095个序号
 */
package util

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/snowyyj001/loumiao/config"
)

const (
	twepoch        = int64(16093440000000)             //开始时间截 (2020-12-31 00:00:00), 0.01秒
	workeridBits   = uint(14)                         //机器id所占的位数
	sequenceBits   = uint(12)                         //序列所占的位数
	workeridMax    = int64(-1 ^ (-1 << workeridBits)) //支持的最大机器id数量
	sequenceMask   = int64(-1 ^ (-1 << sequenceBits)) //
	workeridShift  = sequenceBits                     //机器id左移位数
	timestampShift = sequenceBits + workeridBits      //时间戳左移位数
	maxuid = 1000		//最大开服数
	timescale = 10000000		//UnixNano/timescale=10ms
)

var SnowFlakeInst *Snowflake

// A Snowflake struct holds the basic information needed for a snowflake generator worker
type Snowflake struct {
	sync.Mutex
	timestamp int64
	workerid  int64
	sequence  int64
}

// NewNode returns a new snowflake worker that can be used to generate snowflake IDs
func NewSnowflake(workerid int64) error {
	if SnowFlakeInst != nil {
		return errors.New("NewSnowflake has been created")
	}

	if workerid < 0 || workerid > workeridMax {
		return errors.New("workerid must be between 0 and 1023")
	}

	SnowFlakeInst = &Snowflake{
		timestamp: 0,
		workerid:  workerid,
		sequence:  0,
	}
	return nil
}

// Generate creates and returns a unique snowflake ID
func (s *Snowflake) Generate() int64 {

	s.Lock()

	now := time.Now().UnixNano() / timescale		//0.01秒

	if s.timestamp == now {
		s.sequence = (s.sequence + 1) & sequenceMask
		if s.sequence == 0 { //毫秒内序列溢出
			for now <= s.timestamp {
				now = time.Now().UnixNano() / timescale
			}
		}
	} else {
		s.sequence = 0
	}

	s.timestamp = now

	r := int64((now-twepoch)<<timestampShift | (s.workerid << workeridShift) | (s.sequence))

	s.Unlock()
	return r
}

func UUID() int64 { //该函数调用应该在config.SERVER_NODE_UID赋值之后
	if config.SERVER_NODE_UID <= 0 {
		fmt.Errorf("UUID: wrong server uid: %d", config.SERVER_NODE_UID)
		return 0
	}
	if config.SERVER_NODE_UID > maxuid {
		fmt.Errorf("UUID: too big(>%d) server uid: %d", maxuid, config.SERVER_NODE_UID)
		return 0
	}
	if SnowFlakeInst == nil {
		NewSnowflake(int64(config.SERVER_NODE_UID))
	}
	r := SnowFlakeInst.Generate()

	if r <= 0  {
		fmt.Errorf("UUID:uuid: %d", r)
		return 0
	}
	return r
}
