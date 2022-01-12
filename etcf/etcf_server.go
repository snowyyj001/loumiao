// 配置中心服务
package etcf

import (
	"github.com/snowyyj001/loumiao/message"
	"github.com/snowyyj001/loumiao/msg"
	"strings"
	"fmt"

	"github.com/snowyyj001/loumiao/config"
	"github.com/snowyyj001/loumiao/gorpc"
	"github.com/snowyyj001/loumiao/llog"
	"github.com/snowyyj001/loumiao/network"
	"github.com/snowyyj001/loumiao/util"
)

var (
	This        *EtcfServer
	handler_Map map[string]string
)

type EtcfServer struct {
	gorpc.GoRoutineLogic

	pInnerService network.ISocket

	mWatchKeys   map[string]map[int]bool //prefix -> [socket id][true]
	mStoreValues map[string]string       //prefix -> value
	mStoreLocks map[string]int		//prefix -> time
	mStoreLockWaiters map[string][]int	//prefix -> uids
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

	self.mWatchKeys = make(map[string]map[int]bool)
	self.mStoreValues = make(map[string]string)
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

}

func (self *EtcfServer) DoStart() {
	llog.Info("EtcfServer DoStart")

	util.Assert(self.pInnerService.Start(), fmt.Sprintf("EtcfServer listen failed: saddr=%s", self.pInnerService.GetSAddr()))
}

func (self *EtcfServer) DoDestory() {
	llog.Info("EtcfServer DoDestory")

}

func (self *EtcfServer) addWatch(prefix string, sid int) {
	llog.Infof("EtcfServer addWatch: prefix = %s, sid = %d", prefix, sid)
	vals, ok := self.mWatchKeys[prefix]
	if !ok {
		vals = make(map[int]bool)
		self.mWatchKeys[prefix] = vals
	}
	vals[sid] = true
	
	if value, ok := self.mStoreValues[prefix]; ok {
		self.broadCastValue(prefix, value)
	}
}

func (self *EtcfServer) removeWatch(prefix string, sid int) {
	llog.Infof("EtcfServer removeWatch: prefix = %s, sid = %d", prefix, sid)
	vals, ok := self.mWatchKeys[prefix]
	if !ok {
		return
	}
	delete(vals, sid)
}

func (self *EtcfServer) putValue(prefix string, value string) {
	llog.Infof("EtcfServer putValue: prefix = %s, value = %s", prefix, value)
	if _, ok := self.mStoreValues[prefix]; ok {
		return
	}
	self.mStoreValues[prefix] = value
	self.broadCastValue(prefix, value)
}

func (self *EtcfServer) removeValue(prefix string) {
	llog.Infof("EtcfServer removeValue: prefix = %s", prefix)
	if _, ok := self.mStoreValues[prefix]; ok {
		delete(self.mStoreValues, prefix)
		self.broadCastValue(prefix, "")
	}
}


func (self *EtcfServer) broadCastValue(prefix string, value string) {
	req := new(msg.LouMiaoNoticeValue)
	req.Prefix = prefix
	req.Value = value
	buff, _ := message.Encode(0, "LouMiaoNoticeValue", req)

	for key, vals := range self.mWatchKeys {
		if strings.HasPrefix(key, prefix) {
			for sid, _ := range vals {
				This.pInnerService.SendById(sid, buff)
			}
		}
	}
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
		buff, _ := message.Encode(0, "LouMiaoAquireLock", req)
		for i:=0; i<len(arr); i++ {
			socketId := arr[i]
			This.pInnerService.SendById(socketId, buff)
		}
		delete(self.mStoreLockWaiters, prefix)
	}
}


//goroutine unsafe
//net msg handler,this func belong to socket's goroutine
func packetFunc(socketid int, buff []byte, nlen int) bool {
	//llog.Debugf("packetFunc: socketid=%d, bufferlen=%d", socketid, nlen)
	_, name, buffbody, err := message.UnPackHead(buff, nlen)
	//llog.Debugf("packetFunc  %s %v", name, pm)
	if nil != err {
		llog.Errorf("packetFunc Decode error: %s", err.Error())
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
	return true
}
