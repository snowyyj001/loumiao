package udpgate

import (
	"github.com/snowyyj001/loumiao/config"
	"github.com/snowyyj001/loumiao/gorpc"
	"github.com/snowyyj001/loumiao/llog"
	"github.com/snowyyj001/loumiao/message"
	"github.com/snowyyj001/loumiao/network"
	"github.com/snowyyj001/loumiao/util"
	"sync"
)

const (
	CHAN_SEND_LEN = 20000
)

type UdpGateServer struct {
	gorpc.GoRoutineLogic

	Id        int
	pService  *network.UdpServerSocket
	ListenStr string

	InitFunc func() //需要额外处理的函数回调
}

type UdpHandlerIgo struct {
	Igo      gorpc.IGoRoutine
	FuncNmae string
}

type UdpSendSt struct {
	ClientId int
	Buffer   []byte
}

var (
	This         *UdpGateServer
	handlerMap   map[int64]*UdpHandlerIgo //userid ->igo
	rwMutex      sync.RWMutex
	messagesChan chan *UdpSendSt
)

func (self *UdpGateServer) DoInit() bool {
	llog.Infof("%s DoInit", self.Name)
	This = self
	self.pService = new(network.UdpServerSocket)
	self.pService.Init(self.ListenStr)
	self.pService.BindPacketFunc(packetFunc)
	self.pService.SetConnectType(network.CLIENT_CONNECT)
	self.pService.SetMaxClients(config.NET_MAX_CONNS)

	messagesChan = make(chan *UdpSendSt, CHAN_SEND_LEN)

	if self.InitFunc != nil {
		self.InitFunc()
	}
	return true
}

func (self *UdpGateServer) DoRegsiter() {
	llog.Info("UdpGateServer DoRegsiter")
}

func (self *UdpGateServer) DoStart() {
	llog.Info("UdpGateServer DoStart")

	self.Id = config.Cfg.NetCfg.Uid
	if self.pService.Start() == false {
		llog.Fatalf("UdpGateServer start error")
	}
	llog.Infof("UdpGateServer DoStart success: name=%s,saddr=%s,uid=%d", self.Name, config.SERVER_PLATFORM, config.SERVER_NODE_UID)

	go BufferSend()
}

func (self *UdpGateServer) DoDestory() {
	llog.Info("UdpGateServer DoDestory")
	self.pService.Close()
}

func (self *UdpGateServer) closeClient(clientid int) {
	self.pService.StopClient(clientid)
}

// RegisterHandler 注册udp消息处理actor
func RegisterHandler(userId int, funcName string, igo gorpc.IGoRoutine) {
	defer rwMutex.Unlock()
	rwMutex.Lock()
	handlerMap[int64(userId)] = &UdpHandlerIgo{FuncNmae: funcName, Igo: igo}
}

// UnRegisterHandler 取消注册udp消息处理actor
func UnRegisterHandler(userId int) {
	defer rwMutex.Unlock()
	rwMutex.Lock()
	delete(handlerMap, int64(userId))
}

//goroutine unsafe,此时已不涉及map的修改，直处理了，不用再去RecvPackMsg中处理
func packetFunc(socketid int, buff []byte, nlen int) error {
	defer util.Recover()
	llog.Debugf("udp packetFunc: socketid=%d, bufferlen=%d", socketid, nlen)
	msgId, clientId, body := message.UpPackUdp(buff)
	rwMutex.RLock()
	handler, ok := handlerMap[clientId]
	rwMutex.RUnlock()

	if ok {
		nm := &gorpc.M{Id: int(clientId), Name: handler.FuncNmae, Data: body}
		handler.Igo.Send("ServiceHandler", nm)
	} else {
		llog.Warningf("UdpGateServer packetFunc handler is nil, drop it[%d]", msgId)
	}
	return nil
}

//SendClient 给clientid发消息
func SendClient(clientid int, buff []byte) {
	st := &UdpSendSt{ClientId: clientid, Buffer: buff}
	messagesChan <- st
}

func BufferSend() {
	for {
		msg := <-messagesChan
		This.pService.SendById(msg.ClientId, msg.Buffer)
	}
}

func init() {
	handlerMap = make(map[int64]*UdpHandlerIgo)
}
