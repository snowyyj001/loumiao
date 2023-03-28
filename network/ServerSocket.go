package network

import (
	"github.com/snowyyj001/loumiao/lutil"
	"github.com/snowyyj001/loumiao/nodemgr"
	"net"
	"sync"
	"sync/atomic"
	"time"

	"github.com/snowyyj001/loumiao/llog"
)

type IServerSocket interface {
	ISocket

	AssignClientId() int
	GetClientById(int) *ServerSocketClient
	LoadClient() *ServerSocketClient
	AddClinet(*net.TCPConn, string, int) *ServerSocketClient
	DelClinet(*ServerSocketClient) bool
	StopClient(int)
	ClientRemoteAddr(clientid int) string
}

type ServerSocket struct {
	Socket
	m_nClientCount  int
	mMaxClients     int
	mMinClients     int
	m_nIdSeed       int64
	m_bShuttingDown bool
	m_ClientList    map[int]*ServerSocketClient
	m_ClientLocker  *sync.RWMutex
	m_Listen        *net.TCPListener
	m_Lock          sync.Mutex
}

func (self *ServerSocket) Init(saddr string) bool {
	self.Socket.Init(saddr)
	self.m_ClientList = make(map[int]*ServerSocketClient)
	self.m_ClientLocker = &sync.RWMutex{}
	self.m_bShuttingDown = true
	self.m_nState = SSF_INIT
	return true
}
func (self *ServerSocket) Start() bool {
	llog.Debug("ServerSocket.Start")
	if self.m_nConnectType == 0 {
		llog.Error("ServerSocket.Start error : unkonwen socket type")
		return false
	}
	self.m_bShuttingDown = false

	if self.m_sAddr == "" {
		llog.Error("ServerSocket Start error, saddr is null")
		return false
	}

	tcpAddr, err := net.ResolveTCPAddr("tcp4", self.m_sAddr)
	if err != nil {
		llog.Errorf("%v", err)
	}
	//setDefaultListenerSockopts SO_REUSEADDR 默认启用
	ln, err := net.ListenTCP("tcp4", tcpAddr)
	if err != nil {
		llog.Errorf("%v", err)
		return false
	}

	llog.Infof("ServerSocket 启动监听，等待链接！%s", self.m_sAddr)
	self.m_Listen = ln
	//延迟，监听关闭
	//defer ln.Close()
	self.m_nState = SSF_ACCEPT
	lutil.Go(func() {
		serverRoutine(self)
	})
	return true
}

func (self *ServerSocket) AssignClientId() int {
	return int(atomic.AddInt64(&self.m_nIdSeed, 1))
}

func (self *ServerSocket) GetClientById(id int) *ServerSocketClient {
	self.m_ClientLocker.RLock()
	client, exist := self.m_ClientList[id]
	self.m_ClientLocker.RUnlock()
	if exist == true {
		return client
	}

	return nil
}

func (self *ServerSocket) GetClientNumber() int {
	return self.m_nClientCount
}

func (self *ServerSocket) ClientRemoteAddr(clientid int) string {
	pClinet := self.GetClientById(clientid)
	if pClinet != nil {
		return pClinet.m_Conn.RemoteAddr().String()
	}
	return ""
}

func (self *ServerSocket) AddClinet(tcpConn *net.TCPConn, addr string, connectType int) *ServerSocketClient {
	pClient := self.LoadClient()
	if pClient != nil {
		pClient.Socket.Init(addr)
		pClient.pServer = self
		pClient.m_ClientId = self.AssignClientId()
		pClient.SetConnectType(connectType)
		pClient.SetTcpConn(tcpConn)
		pClient.BindPacketFunc(self.m_PacketFunc)
		self.m_ClientLocker.Lock()
		self.m_ClientList[pClient.m_ClientId] = pClient
		self.m_ClientLocker.Unlock()
		pClient.Start()
		self.m_nClientCount++
		llog.Debugf("客户端：%s已连接[%d]", tcpConn.RemoteAddr().String(), pClient.m_ClientId)
		return pClient
	} else {
		tcpConn.Close()
		llog.Errorf("ServerSocket.AddClinet %s", "无法创建客户端连接对象")
	}
	return nil
}

func (self *ServerSocket) DelClinet(pClient *ServerSocketClient) bool {
	self.m_ClientLocker.Lock()
	delete(self.m_ClientList, pClient.m_ClientId)
	llog.Debugf("客户端：已断开连接[%d]", pClient.m_ClientId)
	self.m_ClientLocker.Unlock()
	self.m_nClientCount--
	return true
}

func (self *ServerSocket) StopClient(id int) {
	pClinet := self.GetClientById(id)
	if pClinet != nil {
		pClinet.Close()
	}
}

func (self *ServerSocket) LoadClient() *ServerSocketClient {
	s := &ServerSocketClient{}
	s.m_MaxReceiveBufferSize = self.m_MaxReceiveBufferSize
	s.m_MaxSendBufferSize = self.m_MaxSendBufferSize
	return s
}

func (self *ServerSocket) Send(buffer []byte) int {
	llog.Error("ServerSocket should not call this func")
	return 0
}

func (self *ServerSocket) SendById(id int, buff []byte) int {
	pClient := self.GetClientById(id)
	if pClient != nil {
		return pClient.Send(buff)
	} else {
		llog.Warningf("ServerSocket发送数据失败[%d]", id)
	}
	return 0
}

func (self *ServerSocket) BroadCast(buff []byte) {
	self.m_ClientLocker.RLock()
	for _, client := range self.m_ClientList {
		client.Send(buff)
	}
	self.m_ClientLocker.Unlock()
}

func (self *ServerSocket) Restart() bool {
	return true
}

func (self *ServerSocket) Connect() bool {
	return true
}

func (self *ServerSocket) Disconnect(bool) bool {
	return true
}

func (self *ServerSocket) OnNetFail(int) {
}

func (self *ServerSocket) Close() {
	self.m_Listen.Close()
	self.Clear()
	nodemgr.ServerEnabled = false
}

func (self *ServerSocket) SetMaxClients(maxnum int) {
	self.mMaxClients = maxnum
}

func serverRoutine(server *ServerSocket) {
	defer lutil.Recover()
	var tempDelay time.Duration
	for {
		tcpConn, err := server.m_Listen.AcceptTCP()
		if err != nil {
			if ne, ok := err.(net.Error); ok && ne.Timeout() {
				if tempDelay == 0 {
					tempDelay = 1 * time.Millisecond
				} else {
					tempDelay++
				}
				if max := 3 * time.Second; tempDelay > max {
					break
				}
				llog.Errorf("serverRoutine accept error: %v", err)
				time.Sleep(tempDelay)
				continue
			}
		}
		tempDelay = 0
		if server.m_nClientCount >= server.mMaxClients {
			tcpConn.Close()
			llog.Warning("serverRoutine: too many conns")
			continue
		}
		handleConn(server, tcpConn, tcpConn.RemoteAddr().String())
	}
	server.Close()
}

func handleConn(server *ServerSocket, tcpConn *net.TCPConn, addr string) bool {
	if tcpConn == nil {
		return false
	}

	pClient := server.AddClinet(tcpConn, addr, server.m_nConnectType)
	if pClient == nil {
		return false
	}

	return true
}
