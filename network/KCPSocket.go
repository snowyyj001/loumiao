package network

import (
	"github.com/snowyyj001/loumiao/llog"
	"github.com/snowyyj001/loumiao/lutil"
	"github.com/snowyyj001/loumiao/nodemgr"
	"github.com/xtaci/kcp-go"
	"sync"
	"sync/atomic"
)

/*switch kcp-model {
case "normal":
	config.NoDelay, config.Interval, config.Resend, config.NoCongestion = 0, 40, 2, 1
case "fast":
	config.NoDelay, config.Interval, config.Resend, config.NoCongestion = 0, 30, 2, 1
case "fast2":
	config.NoDelay, config.Interval, config.Resend, config.NoCongestion = 1, 20, 2, 1
case "fast3":
	config.NoDelay, config.Interval, config.Resend, config.NoCongestion = 1, 10, 2, 1
}
*/

const (
	KCPReadBuffer   = 2 * 1024 * 1024
	KCPWriteBuffer  = 2 * 1024 * 1024
	KCPWinSedSize   = 256  //最大发送窗口，默认为32
	KCPWinRevSize   = 256  //最大接收窗口，默认为32
	KCPNoDelay      = 1    //kcp-model：是否启用 nNoDelayodelay模式，0不启用；1启用。
	KCPInterval     = 20   //kcp-model：协议内部工作的 interval，单位毫秒，比如 10ms或者 20ms
	KCPResend       = 2    //kcp-model：快速重传模式，默认0关闭，可以设置2（2次ACK跨越将会直接重传）
	KCPNoCongestion = 1    //kcp-model:是否关闭流控，默认是0代表不关闭，1代表关闭
	KCPStreamMode   = true //流模式，会组合数据包,需要处理粘包
	KCPMTU          = 1400 //纯算法协议并不负责探测 MTU，默认 mtu是1400字节
	KCPWriteDelay   = false
	KCPDSCP         = 0
	KCPAckNodelay   = true
	KCPTIMEOUT      = 6
	DATASHARD       = 10 //fec参数，禁用和PARITYSHARD同时设置为0
	PARITYSHARD     = 3  //fec参数
)

type KcpSocket struct {
	Socket
	m_nClientCount  int
	mMaxClients     int
	mMinClients     int
	m_nIdSeed       int32
	m_bShuttingDown bool
	m_ClientList    map[int]*KCPSocketClient
	m_ClientLocker  *sync.RWMutex
	m_Lock          sync.Mutex
	m_Listen        *kcp.Listener
}

func (self *KcpSocket) Init(saddr string) bool {
	self.Socket.Init(saddr)
	self.m_ClientList = make(map[int]*KCPSocketClient)
	self.m_ClientLocker = &sync.RWMutex{}
	self.m_bShuttingDown = true
	self.m_nState = SSF_INIT
	return true
}
func (self *KcpSocket) Start() bool {
	if self.m_nConnectType == 0 {
		llog.Error("KcpSocket.Start error : unknown socket type")
		return false
	}
	self.m_bShuttingDown = false

	if self.m_sAddr == "" {
		llog.Error("KcpSocket Start error, saddr is null")
		return false
	}

	ln, err := kcp.ListenWithOptions(self.m_sAddr, nil, DATASHARD, PARITYSHARD)
	if err != nil {
		llog.Errorf("%v", err)
		return false
	}
	ln.SetReadBuffer(KCPReadBuffer)
	ln.SetWriteBuffer(KCPWriteBuffer)
	ln.SetDSCP(KCPDSCP)
	llog.Infof("KcpSocket 启动监听，等待链接！%s", self.m_sAddr)
	self.m_Listen = ln
	self.m_nState = SSF_ACCEPT
	lutil.Go(func() {
		kcpRoutine(self)
	})
	return true
}

func (self *KcpSocket) AssignClientId() int {
	return int(atomic.AddInt32(&self.m_nIdSeed, 1))
}

func (self *KcpSocket) GetClientById(id int) *KCPSocketClient {
	self.m_ClientLocker.RLock()
	client, exist := self.m_ClientList[id]
	self.m_ClientLocker.RUnlock()
	if exist == true {
		return client
	}

	return nil
}

func (self *KcpSocket) ClientRemoteAddr(clientid int) string {
	pClinet := self.GetClientById(clientid)
	if pClinet != nil {
		return pClinet.m_KcpConn.RemoteAddr().String()
	}
	return ""
}

func (self *KcpSocket) DelClinet(pClient *KCPSocketClient) bool {
	self.m_ClientLocker.Lock()
	delete(self.m_ClientList, pClient.m_ClientId)
	llog.Debugf("KcpSocket 客户端：%s已断开连接[%d]", pClient.m_KcpConn.RemoteAddr().String(), pClient.m_ClientId)
	self.m_ClientLocker.Unlock()
	self.m_nClientCount--
	return true
}

func (self *KcpSocket) StopClient(id int) {
	pClinet := self.GetClientById(id)
	if pClinet != nil {
		pClinet.Close()
	}
}

func (self *KcpSocket) LoadClient() *KCPSocketClient {
	s := &KCPSocketClient{}
	s.m_MaxReceiveBufferSize = self.m_MaxReceiveBufferSize
	s.m_MaxSendBufferSize = self.m_MaxSendBufferSize
	return s
}

func (self *KcpSocket) BroadCast(buff []byte) {
	self.m_ClientLocker.RLock()
	for _, client := range self.m_ClientList {
		client.Send(buff)
	}
	self.m_ClientLocker.Unlock()
}

func (self *KcpSocket) SetMaxClients(maxnum int) {
	self.mMaxClients = maxnum
}

func (self *KcpSocket) SendById(id int, buff []byte) int {
	pClient := self.GetClientById(id)
	if pClient != nil {
		return pClient.Send(buff)
		//llog.Warningf("KcpSocket.SendById n = %d, buflen = %d", n, len(buff))
	} else {
		llog.Warningf("KcpSocket发送数据失败[%d]", id)
	}
	return 0
}

func (self *KcpSocket) Restart() bool {
	return true
}

func (self *KcpSocket) Connect() bool {
	return true
}

func (self *KcpSocket) Disconnect(bool) bool {
	return true
}

func (self *KcpSocket) Stop() bool {
	if self.m_bShuttingDown {
		return true
	}
	self.m_bShuttingDown = true
	self.Close()
	return true
}

func (self *KcpSocket) Close() {
	self.m_Listen.Close()
	self.Clear()
	nodemgr.ServerEnabled = false
}

func (self *KcpSocket) AddClinet(kcpConn *kcp.UDPSession, addr string, connectType int) *KCPSocketClient {
	pClient := self.LoadClient()
	if pClient != nil {
		pClient.Socket.Init(addr)
		pClient.m_pServer = self
		pClient.m_ClientId = self.AssignClientId()
		pClient.SetConnectType(connectType)
		pClient.SetKcpConn(kcpConn)
		pClient.BindPacketFunc(self.m_PacketFunc)
		self.m_ClientLocker.Lock()
		self.m_ClientList[pClient.m_ClientId] = pClient
		self.m_ClientLocker.Unlock()
		pClient.Start()
		self.m_nClientCount++
		llog.Debugf("KcpSocket 客户端：%s已连接[%d]", kcpConn.RemoteAddr().String(), pClient.m_ClientId)
		return pClient
	} else {
		kcpConn.Close()
		llog.Errorf("KcpSocket.AddClinet %s", "无法创建客户端连接对象")
	}
	return nil
}

func kcpRoutine(server *KcpSocket) {
	defer lutil.Recover()
	for {
		kcpConn, err := server.m_Listen.AcceptKCP()
		if err != nil {
			llog.Errorf("KcpSocket kcpRoutine listen err: %s", err.Error())
			continue
		}

		if server.m_nClientCount >= server.mMaxClients {
			kcpConn.Close()
			llog.Warning("kcpRoutine: too many conns")
			continue
		}

		// set kcp parameters
		kcpConn.SetWindowSize(KCPWinSedSize, KCPWinRevSize)
		kcpConn.SetNoDelay(KCPNoDelay, KCPInterval, KCPResend, KCPNoCongestion) //fast2
		kcpConn.SetStreamMode(KCPStreamMode)
		kcpConn.SetMtu(KCPMTU)
		kcpConn.SetWriteDelay(KCPWriteDelay)
		kcpConn.SetACKNoDelay(KCPAckNodelay)
		//kcpConn.SetReadDeadline(time.Now().Add(time.Second * KCPTIMEOUT))	//此函数不可靠，bug，不要使用
		handleKcpConn(server, kcpConn, kcpConn.RemoteAddr().String())
	}
	server.Stop()
}

func handleKcpConn(server *KcpSocket, kcpConn *kcp.UDPSession, addr string) bool {
	if kcpConn == nil {
		return false
	}

	pClient := server.AddClinet(kcpConn, addr, server.m_nConnectType)
	if pClient == nil {
		return false
	}

	return true
}
