package message

import (
	"sync"

	"github.com/snowyyj001/loumiao/base/vector"
	"github.com/snowyyj001/loumiao/timer"
)

/*
sync.Map的性能高体现在读操作远多于写操作的时候。 极端情况下，只有读操作时，是普通map的性能的44.3倍。
反过来，如果是全写，没有读，那么sync.Map还不如加普通map+mutex锁呢。只有普通map性能的一半。
建议使用sync.Map时一定要考虑读写比例。当写操作只占总操作的<=1/10的时候，使用sync.Map性能会明显高很多
经测试，在几乎全是读的情况下，sync.Map的效率不如sync.RWMutex
*/
const (
	EXPIRE  = 60 * 1000 //每分钟，删除过多的缓存
	KEEPLEN = 640       //缓存超过KEEPLEN，开始清理
)

type BufferCache struct {
	vec   vector.Vector
	mutex sync.Mutex
}

var (
	buffers map[int]*BufferCache
	rdMutex sync.RWMutex
)

func init() {
	buffers = make(map[int]*BufferCache)
}

func delExpireCache(sz int) {
	timer.NewTimer(EXPIRE, func(dt int64) bool {
		rdMutex.RLock()
		cache, ok := buffers[sz]
		rdMutex.RUnlock()
		if ok {
			cache.mutex.Lock()
			if cache.vec.Len() > KEEPLEN {
				cache.vec.Release(KEEPLEN / 2) //每次清理KEEPLEN的一半
			}
			cache.mutex.Unlock()
			//fmt.Println("delExpireCache: ", sz, cache.vec.Len(), cache.vec.Size())
		}
		return true
	}, true)
}

//获取一个长度为sz的[]byte对象
func GetBuffer(sz int) []byte {
	rdMutex.RLock()
	cache, ok := buffers[sz]
	rdMutex.RUnlock()
	if ok {
		defer cache.mutex.Unlock()
		cache.mutex.Lock()
		if cache.vec.Empty() {
			buff := make([]byte, sz)
			//fmt.Println("GetBuffer0: ", sz, cache.vec.Len(), cache.vec.Size())
			return buff
		} else {
			buff := cache.vec.Back().([]byte)
			cache.vec.PopBack()
			//fmt.Println("GetBuffer1: ", sz, cache.vec.Len(), cache.vec.Size())
			return buff
		}
	} else {
		rdMutex.Lock()
		buffers[sz] = new(BufferCache)
		rdMutex.Unlock()
		delExpireCache(sz)
		buff := make([]byte, sz)
		return buff
	}
}

//缓存buff对象
func BackBuffer(buff []byte) {
	sz := len(buff)
	rdMutex.RLock()
	cache, ok := buffers[sz]
	rdMutex.RUnlock()
	if ok {
		cache.mutex.Lock()
		cache.vec.PushBack(buff)
		cache.mutex.Unlock()

		//fmt.Println("BackBuffer: ", sz, cache.vec.Len(), cache.vec.Size())
	}
}
