package etcf

import (
	"github.com/snowyyj001/loumiao/gorpc"
	"github.com/snowyyj001/loumiao/llog"
	"github.com/snowyyj001/loumiao/message"
	"github.com/snowyyj001/loumiao/msg"
	"github.com/snowyyj001/loumiao/timer"
	"github.com/snowyyj001/loumiao/util"
	"strings"
)

//client connect
func innerConnect(igo gorpc.IGoRoutine, socketId int, data []byte) {
	llog.Debugf("etcf server innerConnect: %d", socketId)

}

//client disconnect
func innerDisConnect(igo gorpc.IGoRoutine, socketId int, data []byte) {
	llog.Debugf("etcf server innerDisConnect: %d", socketId)
	This.removeAllWatchById(socketId)
	This.removeAllLeaseById(socketId)
	delete(This.mStoreValuesLeaseTime, socketId)
}

//watch/remove key
func innerLouMiaoWatchKey(igo gorpc.IGoRoutine, socketId int, data []byte) {
	req := &msg.LouMiaoWatchKey{}
	if message.UnPackProto(req, data) != nil {
		return
	}
	llog.Debugf("innerLouMiaoWatchKey: %v", req)
	siduid_Map[socketId] = int(req.Uid)
	if req.Opcode == msg.LouMiaoWatchKey_ADD {
		This.addWatch(req.Prefix, socketId)
	} else {
		This.removeWatch(req.Prefix, socketId)
	}
}

//put/remove value
func innerLouMiaoPutValue(igo gorpc.IGoRoutine, socketId int, data []byte) {
	req := &msg.LouMiaoPutValue{}
	if message.UnPackProto(req, data) != nil {
		return
	}
	//llog.Debugf("innerLouMiaoPutValue: %v", req)
	if len(req.Value) > 0 {
		if req.Lease > 0 {
			if _, ok := This.mStoreValuesLease[req.Prefix]; ok {
				llog.Errorf("innerLouMiaoPutValue: prefix has already been set, %s", req.Prefix)
			} else {
				This.putValue(req.Prefix, req.Value)
				This.mStoreValuesLease[req.Prefix] = ETKeyLease{req.Prefix, socketId}
				This.mStoreValuesLeaseTime[socketId] = util.TimeStampSec()
			}
		} else {
			This.putValue(req.Prefix, req.Value)
		}
	} else {
		This.removeValue(req.Prefix)
	}
}

//get value
func innerLouMiaoGetValue(igo gorpc.IGoRoutine, socketId int, data []byte) {
	req := &msg.LouMiaoGetValue{}
	if message.UnPackProto(req, data) != nil {
		return
	}
	llog.Debugf("innerLouMiaoGetValue: %v", req)
	isMul := len(req.Prefixs)
	req.Prefixs = make([]string, 0)
	for key, value := range This.mStoreValues {
		if strings.HasPrefix(key, req.Prefix) {
			req.Prefixs = append(req.Prefixs, key)
			req.Values = append(req.Values, value)
			if isMul == 0 {
				break
			}
		}
	}

	buff, _ := message.EncodeProBuff(0, "LouMiaoGetValue", req)
	This.pInnerService.SendById(socketId, buff)
}

//aquire lock
func innerLouMiaoAquireLock(igo gorpc.IGoRoutine, socketId int, data []byte) {
	req := &msg.LouMiaoAquireLock{}
	if message.UnPackProto(req, data) != nil {
		return
	}
	if req.TimeOut < 1000 { //加个时间精度控制
		req.TimeOut = 1000
	}
	llog.Debugf("innerLouMiaoAquireLock: %v", req)
	if len(req.Value) > 0 { //选举leader
		key := "leader:" + req.Prefix + ":" + req.Value
		_, ok := This.mStoreLeaderValues.LoadOrStore(key, true)
		if ok {
			req.TimeOut = 0 //slave
			buff, _ := message.EncodeProBuff(0, "LouMiaoAquireLock", req)
			This.pInnerService.SendById(socketId, buff)
		} else { //master
			buff, _ := message.EncodeProBuff(0, "LouMiaoAquireLock", req)
			This.pInnerService.SendById(socketId, buff)
			timer.DelayJob(60*60*1000, func() { //一小时后删除这个leader key，主要是为了清理内存
				llog.Warningf("innerLouMiaoAquireLock: %s", key)
				This.mStoreLeaderValues.Delete(key)
			}, false)
		}
	} else {
		nt := util.TimeStamp()
		stam, ok := This.mStoreLocks[req.Prefix]
		if !ok {
			stam = int(nt) + int(req.TimeOut)
			This.mStoreLocks[req.Prefix] = stam //直接拿到锁，先到先得

			This.RunDelayJob(int(req.TimeOut), This.lockTimeout, req.Prefix)

			buff, _ := message.EncodeProBuff(0, "LouMiaoAquireLock", req)
			This.pInnerService.SendById(socketId, buff)
		} else { //等待锁的释放
			This.mStoreLockWaiters[req.Prefix] = append(This.mStoreLockWaiters[req.Prefix], socketId)
		}
	}
}

//release lock
func innerLouMiaoReleaseLock(igo gorpc.IGoRoutine, socketId int, data []byte) {
	req := &msg.LouMiaoReleaseLock{}
	if message.UnPackProto(req, data) != nil {
		return
	}
	llog.Debugf("LouMiaoReleaseLock: %v", req)

	stam, ok := This.mStoreLocks[req.Prefix]
	if !ok {
		return
	}

	nt := int(util.TimeStamp())

	arr, ok := This.mStoreLockWaiters[req.Prefix]
	if ok {
		//通知还在等待的最后一个对象锁可用(这样数组移除最后一个元素就行了)
		req := &msg.LouMiaoAquireLock{Prefix: req.Prefix, TimeOut: int32(stam - nt)}
		buff, _ := message.EncodeProBuff(0, "LouMiaoAquireLock", req)
		i := len(arr) - 1
		socketId := arr[i]
		This.pInnerService.SendById(socketId, buff)
		if i == 0 {
			delete(This.mStoreLockWaiters, req.Prefix)
			delete(This.mStoreLocks, req.Prefix)
		} else {
			arr = append(arr[:i], arr[i+1:]...)
			This.mStoreLockWaiters[req.Prefix] = arr
		}
	}
}

//lease
func innerLouMiaoLease(igo gorpc.IGoRoutine, socketId int, data []byte) {
	req := &msg.LouMiaoLease{}
	if message.UnPackProto(req, data) != nil {
		return
	}
	//llog.Debugf("innerLouMiaoLease: %v", req)

	This.mStoreValuesLeaseTime[socketId] = util.TimeStampSec()

	resp := &msg.LouMiaoLease{}
	resp.Uid = req.Uid
	buff, _ := message.EncodeProBuff(0, "LouMiaoLease", resp)
	This.pInnerService.SendById(socketId, buff)
}
