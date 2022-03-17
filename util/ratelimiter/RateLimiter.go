package ratelimiter

import (
	"github.com/snowyyj001/loumiao/util"
	"sync"
	"time"
)

const(
	MaxPermits = 50000
)

/*
令牌桶算法，参考google RateLimiter
协程安全
如果纯粹web限流可以使用官方的LimitListener，更方便精确，LimitListener会在请求结束后返回计数
 */
type RateLimiter struct {
	permitsPerSecond int
	maxPermits int
	storedPermits int
	nextFreeTicketMicros int64
	stableIntervalMicros float64

	aquireChan 		chan struct{}
	permitLock      sync.RWMutex
}

//创建一个令牌桶
//@permitsPerSecond: QPS
//@maxPermits: 最大请求量
func Create(permitsPerSecond, maxPermits int) *RateLimiter {
	limiter := new(RateLimiter)
	limiter.aquireChan = make(chan struct{})
	limiter.permitsPerSecond = permitsPerSecond
	if maxPermits <= 0 {
		limiter.maxPermits = MaxPermits
	} else {
		limiter.maxPermits = maxPermits
	}
	limiter.storedPermits = limiter.maxPermits
	limiter.stableIntervalMicros = float64(time.Millisecond) / float64(permitsPerSecond)		//精确度到微妙
	limiter.nextFreeTicketMicros = time.Now().UnixMicro()
	return limiter
}

func (self *RateLimiter) coolDownIntervalMicros() float64 {
	return self.stableIntervalMicros
}

//计算获取令牌所需等待的时间
func (self *RateLimiter) reserve(permits int) int {
	defer self.permitLock.Unlock()
	self.permitLock.Lock()

	nowMicros := time.Now().UnixMicro()

	return int(self.reserveAndGetWaitLength(permits, nowMicros))
}

func (self *RateLimiter) reserveAndGetWaitLength(permits int, nowMicros int64) int64 {
	momentAvailable := self.reserveEarliestAvailable(permits, nowMicros)
	return util.Max64(momentAvailable - nowMicros, 0)
}

func (self *RateLimiter) reserveEarliestAvailable(requiredPermits int, nowMicros int64) int64 {
	//刷新令牌数
	self.resync(nowMicros)

	returnValue := self.nextFreeTicketMicros

	storedPermitsToSpend := util.Min(requiredPermits, self.storedPermits)
	freshPermits := requiredPermits - storedPermitsToSpend

	waitMicros := freshPermits * int(self.stableIntervalMicros)		//requiredPermits=1，不会出现溢出现象
	self.nextFreeTicketMicros  = self.nextFreeTicketMicros +  int64(waitMicros)		//因为预支取，这里加上等待时间

	self.storedPermits -= storedPermitsToSpend		//预支取

	return returnValue
}

//更新当前令牌数
func (self *RateLimiter) resync(nowMicros int64) {
	if nowMicros > self.nextFreeTicketMicros {
		newPermits := int(float64(nowMicros - self.nextFreeTicketMicros) / self.coolDownIntervalMicros())
		self.storedPermits = util.Min(self.maxPermits, self.storedPermits + newPermits)
		self.nextFreeTicketMicros = nowMicros
	}
}

//获取一个许可，该方法会被阻塞直到获取到请求
func (self *RateLimiter) Acquire() {
	microsToWait := self.reserve(1)
	if microsToWait > 0 {
		time.Sleep(time.Duration(microsToWait) * time.Microsecond)
	}
}

//从RateLimiter获取一个许可，该方法不会被阻塞
//@ret: 成功获取返回true
func (self *RateLimiter) TryAcquire() bool {
	defer self.permitLock.RUnlock()
	self.permitLock.RLock()
	if self.storedPermits <= 0 {
		return false
	}
	self.storedPermits--
	return true
}