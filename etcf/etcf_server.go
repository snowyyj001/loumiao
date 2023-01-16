// 配置中心服务
package etcf

import (
	"fmt"
	"github.com/snowyyj001/loumiao/base"
	"strings"
	"sync"

	"github.com/snowyyj001/loumiao/config"
	"github.com/snowyyj001/loumiao/gorpc"
	"github.com/snowyyj001/loumiao/llog"
	"github.com/snowyyj001/loumiao/message"
	"github.com/snowyyj001/loumiao/msg"
	"github.com/snowyyj001/loumiao/network"
	"github.com/snowyyj001/loumiao/util"
)

/*
* 纯内存的etcd
 */

var (
	This        *EtcfServer
	handler_Map map[string]string
	siduid_Map  map[int]int //socket id -> uid，目前只是为了调试信息查看
)

type ETKeyLease struct {
	Key      string
	SocketId int
}

type EtcfServer struct {
	gorpc.GoRoutineLogic

	pInnerService network.ISocket

	mWatchKeys            map[string]map[int]bool //prefix -> [socket id][true]
	mStoreValues          map[string]string       //prefix -> value
	mStoreValuesLease     map[string]ETKeyLease   //prefix -> ETKeyLease
	mStoreValuesLeaseTime map[int]int64           //socketid -> stmp
	mStoreLeaderValues    sync.Map                //prefix -> bool
	mStoreLocks           map[string]int          //prefix -> time
	mStoreLockWaiters     map[string][]int        //prefix -> uids
}

func (self *EtcfServer) DoInit() bool {
	llog.Info("EtcfServer DoInit")
	This = self

	self.pInnerService = new(network.ServerSocket)
	self.pInnerService.(*network.ServerSocket).SetMaxClients(config.NET_MAX_CONNS)
	self.pInnerService.Init(config.NET_LISTEN_SADDR)
	self.pInnerService.BindPacketFunc(packetFunc)
	self.pInnerService.SetConnectType(network.SERVER_CONNECT)

	handler_Map = make(map[string]string)
	siduid_Map = make(map[int]int)

	self.mWatchKeys = make(map[string]map[int]bool)
	self.mStoreValues = make(map[string]string)
	self.mStoreValuesLease = make(map[string]ETKeyLease)
	self.mStoreValuesLeaseTime = make(map[int]int64)
	self.mStoreLocks = make(map[string]int)
	self.mStoreLockWaiters = make(map[string][]int)

	return true
}

func (self *EtcfServer) DoRegsiter() {
	llog.Info("EtcfServer DoRegsiter")

	handler_Map["CONNECT"] = "EtcfServer" //server connect with rpc server
	self.RegisterGate("CONNECT", innerConnect)

	handler_Map["DISCONNECT"] = "EtcfServer"
	self.RegisterGate("DISCONNECT", innerDisConnect)

	handler_Map["LouMiaoWatchKey"] = "EtcfServer"
	self.RegisterGate("LouMiaoWatchKey", innerLouMiaoWatchKey)

	handler_Map["LouMiaoPutValue"] = "EtcfServer"
	self.RegisterGate("LouMiaoPutValue", innerLouMiaoPutValue)

	handler_Map["LouMiaoGetValue"] = "EtcfServer"
	self.RegisterGate("LouMiaoGetValue", innerLouMiaoGetValue)

	handler_Map["LouMiaoAquireLock"] = "EtcfServer"
	self.RegisterGate("LouMiaoAquireLock", innerLouMiaoAquireLock)

	handler_Map["LouMiaoReleaseLock"] = "EtcfServer"
	self.RegisterGate("LouMiaoReleaseLock", innerLouMiaoReleaseLock)

	handler_Map["LouMiaoLease"] = "EtcfServer"
	self.RegisterGate("LouMiaoLease", innerLouMiaoLease)

}

func (self *EtcfServer) DoStart() {
	llog.Info("EtcfServer DoStart")

	util.Assert(self.pInnerService.Start(), fmt.Sprintf("EtcfServer listen failed: saddr=%s", self.pInnerService.GetSAddr()))

	//self.RunTimer(1000, self.update_1000)
}

func (self *EtcfServer) DoDestory() {
	llog.Info("EtcfServer DoDestory")

}

// //////////////////////////////////////////////////////////////////////////////
// 每秒钟更新一次
func (self *EtcfServer) update_1000(dt int64) {
	//llog.Debugf("%s update_1000: %d", self.Name, dt)
	for sid, stmp := range self.mStoreValuesLeaseTime {
		if base.TimeStampSec()-stmp >= LEASE_SERVER_TIMEOUT { //超时
			uid, _ := siduid_Map[sid]
			llog.Warningf("EtcfServer.update_1000: lease timeout, sid = %d, uid = %d, timeout = %d", sid, uid, base.TimeStampSec()-stmp)
			//self.removeAllLeaseById(sid)
			//delete(self.mStoreValuesLeaseTime, sid)
		}
	}
}

func (self *EtcfServer) addWatch(prefix string, sid int) {
	llog.Debugf("EtcfServer addWatch: prefix = %s, sid = %d", prefix, sid)
	vals, ok := self.mWatchKeys[prefix]
	if !ok {
		vals = make(map[int]bool)
		self.mWatchKeys[prefix] = vals
	}
	vals[sid] = true

	for key, value := range This.mStoreValues {
		if strings.HasPrefix(key, prefix) {
			self.noticeValue(sid, key, value) //将已经存在的值发送给目标server
		}
	}
}

func (self *EtcfServer) removeWatch(prefix string, sid int) {
	llog.Debugf("EtcfServer removeWatch: prefix = %s, sid = %d", prefix, sid)
	vals, ok := self.mWatchKeys[prefix]
	if !ok {
		return
	}
	delete(vals, sid)
}

func (self *EtcfServer) removeAllLeaseById(sid int) {
	llog.Debugf("EtcfServer removeAllLeaseById: socketid = %d", sid)
	for prefix, v := range self.mStoreValuesLease {
		if v.SocketId == sid {
			delete(self.mStoreValuesLease, prefix)
			delete(self.mStoreValuesLeaseTime, sid)
			self.removeValue(prefix)
		}
	}
	//llog.Debugf("removeAllLeaseById: %v", self.mWatchKeys)
}

func (self *EtcfServer) removeAllWatchById(sid int) {
	llog.Debugf("EtcfServer removeAllWatchById: socketid = %d", sid)
	for _, vals := range self.mWatchKeys {
		for k, _ := range vals {
			if k == sid {
				delete(vals, k)
				break
			}
		}
	}
	//llog.Debugf("removeAllWatchById: %v", self.mWatchKeys)
}

func (self *EtcfServer) putValue(prefix string, value string) {
	//llog.Debugf("EtcfServer putValue: prefix = %s, value = %s", prefix, value)
	self.mStoreValues[prefix] = value
	self.broadCastValue(prefix, value)
}

func (self *EtcfServer) removeValue(prefix string) {
	//llog.Debugf("EtcfServer removeValue: prefix = %s", prefix)
	if _, ok := self.mStoreValues[prefix]; ok {
		delete(self.mStoreValues, prefix)
		self.broadCastValue(prefix, "")
	}
}

func (self *EtcfServer) broadCastValue(prefix string, value string) {
	req := new(msg.LouMiaoNoticeValue)
	req.Prefix = prefix
	req.Value = value
	buff, _ := message.EncodeProBuff(0, "LouMiaoNoticeValue", req)

	for key, vals := range self.mWatchKeys {
		if strings.HasPrefix(prefix, key) {
			for sid, _ := range vals {
				This.pInnerService.SendById(sid, buff)
			}
		}
	}
}

func (self *EtcfServer) noticeValue(sid int, prefix string, value string) {
	req := new(msg.LouMiaoNoticeValue)
	req.Prefix = prefix
	req.Value = value
	buff, _ := message.EncodeProBuff(0, "LouMiaoNoticeValue", req)
	n := This.pInnerService.SendById(sid, buff)
	llog.Debugf("EtcfServer.noticeValue: %d %d ", n, len(buff))
}

func (self *EtcfServer) lockTimeout(param interface{}) {
	prefix := param.(string)
	if _, ok := self.mStoreLocks[prefix]; !ok {
		return
	}
	delete(self.mStoreLocks, prefix)
	arr, ok := self.mStoreLockWaiters[prefix]
	if ok {
		//通知还在等待的锁超时
		req := &msg.LouMiaoAquireLock{Prefix: prefix, TimeOut: 0}
		buff, _ := message.EncodeProBuff(0, "LouMiaoAquireLock", req)
		for i := 0; i < len(arr); i++ {
			socketId := arr[i]
			This.pInnerService.SendById(socketId, buff)
		}
		delete(self.mStoreLockWaiters, prefix)
	}
}

// goroutine unsafe
// net msg handler,this func belong to socket's goroutine
func packetFunc(socketid int, buff []byte, nlen int) error {
	//llog.Debugf("packetFunc: socketid=%d, bufferlen=%d", socketid, nlen)
	_, name, buffbody, err := message.UnPackHead(buff, nlen)
	//llog.Debugf("packetFunc  %s", name)
	if nil != err {
		return fmt.Errorf("packetFunc Decode error: %s", err.Error())
		//This.closeClient(socketid)
	} else {
		handler, ok := handler_Map[name]
		if ok {
			nm := &gorpc.M{Id: socketid, Name: name, Data: buffbody}
			gorpc.MGR.Send(handler, "ServiceHandler", nm)
		} else {
			llog.Errorf("packetFunc handler is nil, drop it[%s]", name)
		}

	}
	return nil
}
