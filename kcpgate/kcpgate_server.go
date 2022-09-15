package kcpgate

import (
	"fmt"
	"github.com/snowyyj001/loumiao/config"
	"github.com/snowyyj001/loumiao/gorpc"
	"github.com/snowyyj001/loumiao/llog"
	"github.com/snowyyj001/loumiao/message"
	"github.com/snowyyj001/loumiao/network"
)

type KcpGateServer struct {
	gorpc.GoRoutineLogic

	Id        int
	pService  *network.KcpSocket
	ListenStr string

	InitFunc func() //需要额外处理的函数回调
}

var (
	This        *KcpGateServer
	handler_Map map[string]string
)

func (self *KcpGateServer) DoInit() bool {
	llog.Infof("%s DoInit", self.Name)
	This = self
	self.pService = new(network.KcpSocket)
	self.pService.Init(self.ListenStr)
	self.pService.BindPacketFunc(packetFunc)
	self.pService.SetConnectType(network.CLIENT_CONNECT)
	self.pService.SetMaxClients(config.NET_MAX_CONNS)

	if self.InitFunc != nil {
		self.InitFunc()
	}

	handler_Map = make(map[string]string)

	return true
}

func (self *KcpGateServer) DoRegsiter() {
	llog.Info("KcpGateServer DoRegsiter")

	//self.Register("RegisterNet", RegisterNet)

	self.RegisterSelfNet("CONNECT", innerConnect)
	self.RegisterSelfNet("DISCONNECT", innerDisConnect)

}

func (self *KcpGateServer) DoStart() {
	llog.Info("KcpGateServer DoStart")

	self.Id = config.Cfg.NetCfg.Uid
	if self.pService.Start() == false {
		llog.Fatalf("KcpGateServer start error")
	}
}

func (self *KcpGateServer) DoDestory() {
	llog.Info("KcpGateServer DoDestory")
}

func (self *KcpGateServer) closeClient(clientid int) {
	self.pService.StopClient(clientid)
}

// simple register self net hanlder, this func can only be called before igo started
func (self *KcpGateServer) RegisterSelfNet(hanlderName string, hanlderFunc gorpc.HanlderNetFunc) {
	handler_Map[hanlderName] = "KcpGateServer"
	self.RegisterGate(hanlderName, hanlderFunc)
}

// goroutine unsafe,此时已不涉及map的修改，直处理了，不用再去RecvPackMsg中处理
func packetFunc(socketid int, buff []byte, nlen int) error {
	//llog.Debugf("packetFunc: socketid=%d, bufferlen=%d", socketid, nlen)
	target, name, buffbody, err := message.UnPackHead(buff, nlen)
	if err != nil {
		return fmt.Errorf("KcpGateServer packetFunc Decode error: %s", err.Error())
		//This.closeClient(socketid)
	}
	if target != config.NET_NODE_TYPE && target > 0 {
		return fmt.Errorf("KcpGateServer packetFunc target error: target = %d, my = %d, name = %s", target, config.NET_NODE_TYPE, name)
		//This.closeClient(socketid)
	}
	handler, ok := handler_Map[name]
	if ok {
		if handler == This.Name {
			cb, ok := This.NetHandler[name]
			if ok {
				cb(This, socketid, buffbody)
			} else {
				llog.Errorf("KcpGateServer packetFunc[%s] handler is nil: %s", name, This.Name)
			}
		} else {
			nm := &gorpc.M{Id: socketid, Name: name, Data: buffbody}
			gorpc.MGR.Send(handler, "ServiceHandler", nm)
		}
	} else {
		llog.Errorf("KcpGateServer recvPackMsg self handler is nil, drop it[%s]", name)
	}
	return nil
}

// 这两个函数加了锁，可以直接调用，就不需要像gate那样通过actor调用了
func (self *KcpGateServer) SendClient(clientid int, buff []byte) {
	self.pService.SendById(clientid, buff)
}

// 给clientids广播消息
func (self *KcpGateServer) SendMulClient(clientids []int, buff []byte) {
	for _, socketId := range clientids {
		self.pService.SendById(socketId, buff)
	}
}
