package etcf

import (
	"github.com/snowyyj001/loumiao/util"
	"github.com/snowyyj001/loumiao/gorpc"
	"github.com/snowyyj001/loumiao/llog"
	"github.com/snowyyj001/loumiao/message"
	"github.com/snowyyj001/loumiao/msg"
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
}

//watch/remove key
func innerLouMiaoWatchKey(igo gorpc.IGoRoutine, socketId int, data []byte) {
	req := &msg.LouMiaoWatchKey{}
	if message.UnPack(req, data) != nil {
		return
	}
	llog.Debugf("innerLouMiaoWatchKey: %v", req)
	if req.Opcode == msg.LouMiaoWatchKey_ADD {
		This.addWatch(req.Prefix, socketId)
	} else {
		This.removeWatch(req.Prefix, socketId)
	}
}

//put/remove value
func innerLouMiaoPutValue(igo gorpc.IGoRoutine, socketId int, data []byte) {
	req := &msg.LouMiaoPutValue{}
	if message.UnPack(req, data) != nil {
		return
	}
	//llog.Debugf("innerLouMiaoPutValue: %v", req)
	if len(req.Value) > 0 {
		This.putValue(req.Prefix, req.Value)
	} else {
		This.removeValue(req.Prefix)
	}
}

//get value
func innerLouMiaoGetValue(igo gorpc.IGoRoutine, socketId int, data []byte) {
	req := &msg.LouMiaoGetValue{}
	if message.UnPack(req, data) != nil {
		return
	}
	llog.Debugf("innerLouMiaoGetValue: %v", req)
	
	for key, value := range This.mStoreValues {
		if strings.HasPrefix(key, req.Prefix) {
			req.Prefixs = append(req.Prefixs, key)
			req.Values = append(req.Values, value)
		}
	}

	buff, _ := message.Encode(0, "LouMiaoGetValue", req)
	This.pInnerService.SendById(socketId, buff)
}

//aquire lock
func innerLouMiaoAquireLock(igo gorpc.IGoRoutine, socketId int, data []byte) {
	req := &msg.LouMiaoAquireLock{}
	if message.UnPack(req, data) != nil {
		return
	}
	if req.TimeOut < 1000 {		//加个时间精度控制
		req.TimeOut = 1000
	}
	llog.Debugf("innerLouMiaoAquireLock: %v", req)
	nt := util.TimeStamp()
	stam, ok := This.mStoreLocks[req.Prefix]
	if !ok {
		stam = int(nt) + int(req.TimeOut)
		This.mStoreLocks[req.Prefix] = stam		//直接拿到锁，先到先得

		This.RunTicker(int(req.TimeOut), This.lockTimeout, req.Prefix)

		buff, _ := message.Encode(0, "LouMiaoAquireLock", req)
		This.pInnerService.SendById(socketId, buff)
	} else {		//等待锁的释放
		This.mStoreLockWaiters[req.Prefix] = append(This.mStoreLockWaiters[req.Prefix], socketId)
	}
}

//release lock
func innerLouMiaoReleaseLock(igo gorpc.IGoRoutine, socketId int, data []byte) {
	req := &msg.LouMiaoReleaseLock{}
	if message.UnPack(req, data) != nil {
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
		req := &msg.LouMiaoAquireLock{Prefix: req.Prefix, TimeOut: int32(stam-nt)}
		buff, _ := message.Encode(0, "LouMiaoAquireLock", req)
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
	if message.UnPack(req, data) != nil {
		return
	}
	//llog.Debugf("innerLouMiaoLease: %v", req)

	resp := &msg.LouMiaoLease{}
	resp.Uid = req.Uid
	buff, _ := message.Encode(0, "LouMiaoLease", resp)
	This.pInnerService.SendById(socketId, buff)
}
