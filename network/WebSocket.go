package network

import (
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	"github.com/snowyyj001/loumiao/llog"

	"github.com/gorilla/websocket"
)

type IWebSocket interface {
	ISocket

	AssignClientId() int
	GetClientById(int) *WebSocketClient
	LoadClient() *WebSocketClient
	AddClinet(*websocket.Conn, string, int) *WebSocketClient
	DelClinet(*WebSocketClient) bool
	StopClient(int)
	ClientRemoteAddr(clientid int) string
}

type WebSocket struct {
	Socket
	m_nClientCount  int
	m_nMaxClients   int
	m_nMinClients   int
	m_nIdSeed       int32
	m_bShuttingDown bool
	m_bCanAccept    bool
	m_bNagle        bool
	m_ClientList    map[int]*WebSocketClient
	m_ClientLocker  *sync.RWMutex
	m_httpServer    *http.Server
	m_Lock          sync.Mutex
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
} // use default options
var This *WebSocket

func (self *WebSocket) Init(saddr string) bool {
	This = self
	self.Socket.Init(saddr)
	self.m_ClientList = make(map[int]*WebSocketClient)
	self.m_ClientLocker = &sync.RWMutex{}
	self.m_bShuttingDown = true
	self.m_nState = SSF_INIT
	return true
}
func (self *WebSocket) Start() bool {
	llog.Debug("WebSocket.Start")
	if self.m_nConnectType == 0 {
		llog.Error("WebSocket.Start error : unkonwen socket type")
		return false
	}
	self.m_bShuttingDown = false

	if self.m_sAddr == "" {
		return false
	}

	http.HandleFunc("/", serveHome)
	http.HandleFunc("/ws", serveWs)
	self.m_httpServer = &http.Server{Addr: self.m_sAddr}
	succ := true
	go func() {
		err := self.m_httpServer.ListenAndServe()
		if err != nil {
			succ = false
			llog.Errorf("WebSocket ListenAndServe: %v", err)
			return
		}
	}()
	time.Sleep(time.Second)
	if succ == false {
		return false
	}

	llog.Infof("websocket 启动监听，等待链接！%s", self.m_sAddr)

	//延迟，监听关闭
	//defer ln.Close()
	self.m_nState = SSF_ACCEPT
	return true
}

func (self *WebSocket) AssignClientId() int {
	return int(atomic.AddInt32(&self.m_nIdSeed, 1))
}

func (self *WebSocket) GetClientById(id int) *WebSocketClient {
	self.m_ClientLocker.RLock()
	client, exist := self.m_ClientList[id]
	self.m_ClientLocker.RUnlock()
	if exist == true {
		return client
	}

	return nil
}

func (self *WebSocket) ClientRemoteAddr(clientid int) string {
	pClinet := self.GetClientById(clientid)
	if pClinet != nil {
		return pClinet.m_Conn.RemoteAddr().String()
	}
	return ""
}

func (self *WebSocket) AddClinet(wConn *websocket.Conn, addr string, connectType int) *WebSocketClient {
	pClient := self.LoadClient()
	if pClient != nil {
		pClient.Socket.Init(addr)
		pClient.m_pServer = self
		pClient.m_ClientId = self.AssignClientId()
		pClient.SetConnectType(connectType)
		pClient.SetWsConn(wConn)
		pClient.BindPacketFunc(self.m_PacketFunc)
		self.m_ClientLocker.Lock()
		self.m_ClientList[pClient.m_ClientId] = pClient
		self.m_ClientLocker.Unlock()
		pClient.Start()
		self.m_nClientCount++
		llog.Debugf("客户端：%s已连接[%d]！", wConn.RemoteAddr().String(), pClient.m_ClientId)
		return pClient
	} else {
		llog.Errorf("WebSocket.AddClinet %s", "无法创建客户端连接对象")
	}
	return nil
}

func (self *WebSocket) DelClinet(pClient *WebSocketClient) bool {
	self.m_ClientLocker.Lock()
	delete(self.m_ClientList, pClient.m_ClientId)
	llog.Debugf("客户端：%s已断开连接[%d]！", pClient.m_WsConn.RemoteAddr().String(), pClient.m_ClientId)
	self.m_ClientLocker.Unlock()
	self.m_nClientCount--
	return true
}

func (self *WebSocket) StopClient(id int) {
	pClinet := self.GetClientById(id)
	if pClinet != nil {
		pClinet.Close()
	}
}

func (self *WebSocket) LoadClient() *WebSocketClient {
	s := &WebSocketClient{}
	s.m_MaxReceiveBufferSize = self.m_MaxReceiveBufferSize
	s.m_MaxSendBufferSize = self.m_MaxSendBufferSize
	return s
}

func (self *WebSocket) SendById(id int, buff []byte) int {
	pClient := self.GetClientById(id)
	if pClient != nil {
		pClient.Send(buff)
	} else {
		llog.Warningf("WebSocket发送数据失败[%d]", id)
	}
	return 0
}

func (self *WebSocket) BroadCast(buff []byte) {
	self.m_ClientLocker.RLock()
	for _, client := range self.m_ClientList {
		client.Send(buff)
	}
	self.m_ClientLocker.Unlock()
}

func (self *WebSocket) Restart() bool {
	return true
}

func (self *WebSocket) Connect() bool {
	return true
}

func (self *WebSocket) Disconnect(bool) bool {
	return true
}

func (self *WebSocket) OnNetFail(int) {
}

func (self *WebSocket) Close() {
	self.m_httpServer.Close()
	self.Clear()
}

func (self *WebSocket) SetMaxClients(maxnum int) {
	self.m_nMaxClients = maxnum
}

func serveWs(w http.ResponseWriter, r *http.Request) {
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		llog.Errorf("serveWs upgrade: %s", err.Error())
		return
	}
	pClient := This.AddClinet(c, r.RemoteAddr, This.m_nConnectType)
	pClient.Start()
}

func serveHome(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "Not alowed", http.StatusNotFound)
}
