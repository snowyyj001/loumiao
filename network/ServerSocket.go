package network

import (
	"net"
	"sync"
	"sync/atomic"

	"github.com/snowyyj001/loumiao/log"
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
	m_nMaxClients   int
	m_nMinClients   int
	m_nIdSeed       int32
	m_bShuttingDown bool
	m_bCanAccept    bool
	m_bNagle        bool
	m_ClientList    map[int]*ServerSocketClient
	m_ClientLocker  *sync.RWMutex
	m_Listen        *net.TCPListener
	m_Pool          sync.Pool
	m_Lock          sync.Mutex
}

func (self *ServerSocket) Init(saddr string) bool {
	self.Socket.Init(saddr)
	self.m_ClientList = make(map[int]*ServerSocketClient)
	self.m_ClientLocker = &sync.RWMutex{}
	self.m_Pool = sync.Pool{
		New: func() interface{} {
			var s = &ServerSocketClient{}
			return s
		},
	}
	return true
}
func (self *ServerSocket) Start() bool {
	self.m_bShuttingDown = false

	if self.m_sAddr == "" {
		log.Error("ServerSocket Start error, saddr is null")
		return false
	}

	tcpAddr, err := net.ResolveTCPAddr("tcp4", self.m_sAddr)
	if err != nil {
		log.Errorf("%v", err)
	}
	ln, err := net.ListenTCP("tcp4", tcpAddr)
	if err != nil {
		log.Errorf("%v", err)
		return false
	}

	log.Infof("ServerSocket 启动监听，等待链接！%s", self.m_sAddr)

	self.m_Listen = ln
	//延迟，监听关闭
	//defer ln.Close()
	self.m_nState = SSF_ACCEPT
	go serverRoutine(self)
	return true
}

func (self *ServerSocket) AssignClientId() int {
	return int(atomic.AddInt32(&self.m_nIdSeed, 1))
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
		pClient.m_pServer = self
		pClient.m_ClientId = self.AssignClientId()
		pClient.SetConnectType(connectType)
		pClient.SetTcpConn(tcpConn)
		pClient.BindPacketFunc(self.m_PacketFunc)
		self.m_ClientLocker.Lock()
		self.m_ClientList[pClient.m_ClientId] = pClient
		self.m_ClientLocker.Unlock()
		pClient.Start()
		self.m_nClientCount++
		log.Debugf("客户端：%s已连接[%d]", tcpConn.RemoteAddr().String(), pClient.m_ClientId)
		return pClient
	} else {
		log.Errorf("%s", "无法创建客户端连接对象")
	}
	return nil
}

func (self *ServerSocket) DelClinet(pClient *ServerSocketClient) bool {
	self.m_Pool.Put(pClient)
	self.m_ClientLocker.Lock()
	delete(self.m_ClientList, pClient.m_ClientId)
	log.Debugf("客户端：%s已断开连接[%d]", pClient.m_Conn.RemoteAddr().String(), pClient.m_ClientId)
	self.m_ClientLocker.Unlock()
	self.m_nClientCount--
	return true
}

func (self *ServerSocket) StopClient(id int) {
	log.Debugf("ServerSocket.StopClient: %d", id)
	pClinet := self.GetClientById(id)
	if pClinet != nil {
		pClinet.Stop()
	}
}

func (self *ServerSocket) LoadClient() *ServerSocketClient {
	s := self.m_Pool.Get().(*ServerSocketClient)
	s.m_MaxReceiveBufferSize = self.m_MaxReceiveBufferSize
	s.m_MaxSendBufferSize = self.m_MaxSendBufferSize
	return s
}

func (self *ServerSocket) Stop() bool {
	if self.m_bShuttingDown {
		return true
	}

	self.m_bShuttingDown = true
	self.m_nState = SSF_SHUT_DOWN
	return true
}

func (self *ServerSocket) SendById(id int, buff []byte) int {
	pClient := self.GetClientById(id)
	if pClient != nil {
		pClient.Send(buff)
	} else {
		log.Warningf("ServerSocket发送数据失败[%d]", id)
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
	defer self.m_Listen.Close()
	self.Clear()
	//self.m_Pool.Put(self)
}

func (self *ServerSocket) SetMaxClients(maxnum int) {
	self.m_nMaxClients = maxnum
}

func SendClient(pClient *ServerSocketClient, buff []byte) {
	defer func() {
		if err := recover(); err != nil {
			log.Errorf("SendRpc", err) // 这里的err其实就是panic传入的内容，55
		}
	}()

	if pClient != nil {
		pClient.Send(buff)
	}
}

func serverRoutine(server *ServerSocket) {
	for {
		tcpConn, err := server.m_Listen.AcceptTCP()
		handleError(err)
		if err != nil {
			return
		}

		if server.m_nClientCount >= server.m_nMaxClients {
			log.Warning("serverRoutine: too many conns")
			return
		}

		//延迟，关闭链接
		//defer tcpConn.Close()
		handleConn(server, tcpConn, tcpConn.RemoteAddr().String())
	}
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
