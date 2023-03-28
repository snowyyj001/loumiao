package udpgate

import (
	"fmt"
	"github.com/snowyyj001/loumiao/gorpc"
	"github.com/snowyyj001/loumiao/lconfig"
	"github.com/snowyyj001/loumiao/llog"
	"github.com/snowyyj001/loumiao/lutil"
	"github.com/snowyyj001/loumiao/message"
	"github.com/snowyyj001/loumiao/network"
	"strings"
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
	Igo     gorpc.IGoRoutine
	Handler gorpc.HandlerNetFunc
}

type UdpSendSt struct {
	ClientId int
	Buffer   []byte
}

var (
	This         *UdpGateServer
	handlerMap   map[int64]*UdpHandlerIgo //userid -> HandlerNetFunc
	rwMutex      sync.RWMutex
	messagesChan chan *UdpSendSt
)

func (self *UdpGateServer) DoInit() bool {
	llog.Infof("%s DoInit", self.Name)
	This = self
	self.pService = new(network.UdpServerSocket)
	arr := strings.Split(self.ListenStr, ":")
	self.pService.Init(fmt.Sprintf("0.0.0.0:%s", arr[1]))
	self.pService.BindPacketFunc(packetFunc)
	self.pService.SetConnectType(network.CLIENT_CONNECT)
	self.pService.SetMaxClients(lconfig.NET_MAX_CONNS)

	messagesChan = make(chan *UdpSendSt, CHAN_SEND_LEN)

	if self.InitFunc != nil {
		self.InitFunc()
	}
	return true
}

func (self *UdpGateServer) DoRegister() {
	llog.Info("UdpGateServer DoRegister")
}

func (self *UdpGateServer) DoStart() {
	llog.Info("UdpGateServer DoStart")

	self.Id = lconfig.Cfg.NetCfg.Uid
	if self.pService.Start() == false {
		llog.Fatalf("UdpGateServer start error")
	}
	llog.Infof("UdpGateServer DoStart success: name=%s,saddr=%s,uid=%d", self.Name, lconfig.SERVER_PLATFORM, lconfig.SERVER_NODE_UID)

	lutil.Go(func() {
		BufferSend()
	})
}

func (self *UdpGateServer) DoDestroy() {
	llog.Info("UdpGateServer DoDestroy")
	self.pService.Close()
}

func (self *UdpGateServer) closeClient(clientid int) {
	self.pService.StopClient(clientid)
}

// RegisterHandler 注册udp消息处理actor
func RegisterHandler(userId int, igo gorpc.IGoRoutine, handler gorpc.HandlerNetFunc) {
	defer rwMutex.Unlock()
	rwMutex.Lock()
	handlerMap[int64(userId)] = &UdpHandlerIgo{Handler: handler, Igo: igo}
}

// UnRegisterHandler 取消注册udp消息处理actor
func UnRegisterHandler(userId int) {
	defer rwMutex.Unlock()
	rwMutex.Lock()
	delete(handlerMap, int64(userId))
}

// goroutine unsafe,此时已不涉及map的修改，直处理了，不用再去RecvPackMsg中处理
func packetFunc(socketid int, buff []byte, nlen int) error {
	defer lutil.Recover()
	llog.Debugf("udp packetFunc: socketid=%d", socketid)
	msgId, clientId, body := message.UnPackUdp(buff)
	rwMutex.RLock()
	actor, ok := handlerMap[clientId]
	rwMutex.RUnlock()
	if ok {
		actor.Handler(actor.Igo, int(clientId), body)
	} else {
		if !lconfig.SERVER_RELEASE {
			llog.Warningf("UdpGateServer packetFunc handler is nil, drop it[%d][%d]", msgId, socketid)
		}

	}
	return nil
}

// SendClient 给clientid发消息
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
