package network

import (
	"encoding/binary"
	"fmt"
	"github.com/snowyyj001/loumiao/base"
	"github.com/snowyyj001/loumiao/llog"
	"github.com/snowyyj001/loumiao/util"
	"net"
	"strings"
	"sync"
)

var (
	ThisUdpServerSocket *UdpServerSocket
)

type IUdpServerSocket interface {
	ISocket

	AssignClientId() int
	GetClientById(int) *UdpSocketClient
	LoadClient() *UdpSocketClient
	AddClinet(*net.TCPConn, string, int) *UdpSocketClient
	DelClinet(*UdpSocketClient) bool
	StopClient(int)
	ClientRemoteAddr(clientid int) string
}

type UdpServerSocket struct {
	Socket
	m_nClientCount  int
	m_nMaxClients   int
	m_nMinClients   int
	m_nIdSeed       int64
	m_bShuttingDown bool
	mClientList     map[int]*UdpSocketClient //clientid -> UdpSocketClient
	m_ClientLocker  *sync.RWMutex
	mUpdConn        *net.UDPConn
	m_Lock          sync.Mutex
	mBuffChan       chan UdpBufferTransport
}

func (self *UdpServerSocket) Init(saddr string) bool {
	self.Socket.Init(saddr)
	self.mClientList = make(map[int]*UdpSocketClient)
	self.m_ClientLocker = &sync.RWMutex{}
	self.m_bShuttingDown = true
	self.m_nState = SSF_INIT
	self.mBuffChan = make(chan UdpBufferTransport, self.m_nMaxClients)

	ThisUdpServerSocket = self
	return true
}
func (self *UdpServerSocket) Start() bool {
	llog.Debug("UdpServerSocket.Start")
	if self.m_nConnectType == 0 {
		llog.Error("UdpServerSocket.Start error : unkonwen socket type")
		return false
	}
	self.m_bShuttingDown = false

	if self.m_sAddr == "" {
		llog.Error("UdpServerSocket Start error, saddr is null")
		return false
	}
	arr := strings.Split(self.m_sAddr, ":")
	// udp server
	listenUdp, err := net.ListenUDP("udp", &net.UDPAddr{
		IP:   net.ParseIP(arr[0]),
		Port: util.Atoi(arr[1]),
	})
	if err != nil {
		llog.Fatalf("ListenUDP %v", err)
		return false
	}

	llog.Infof("UdpServerSocket 启动监听，等待链接！%s", self.m_sAddr)
	self.mUpdConn = listenUdp
	//延迟，监听关闭
	//defer ln.Close()
	self.m_nState = SSF_ACCEPT
	go udpserverRoutine()
	go handlerUdpMsg()
	return true
}

func (self *UdpServerSocket) AssignClientId() int {
	self.m_nIdSeed++
	return int(self.m_nIdSeed)
}

func (self *UdpServerSocket) GetClientById(id int) *UdpSocketClient {
	defer self.m_ClientLocker.RUnlock()
	self.m_ClientLocker.RLock()
	client, _ := self.mClientList[id]
	return client
}

func (self *UdpServerSocket) ClientRemoteAddr(clientid int) string {
	return ""
}

func (self *UdpServerSocket) AddClient(clientId int, addr *net.UDPAddr) *UdpSocketClient {
	pClient := self.LoadClient()
	if pClient != nil {
		pClient.ClientId = clientId
		pClient.RemoteAddr = addr
		self.m_ClientLocker.Lock()
		self.mClientList[clientId] = pClient
		self.m_ClientLocker.Unlock()
		self.m_nClientCount++
		llog.Debugf("udp 客户端：%s已连接[%d]", addr.String(), clientId)
		return pClient
	}
	return nil
}

func (self *UdpServerSocket) DelClinet(clientid int) bool {
	defer self.m_ClientLocker.Unlock()
	self.m_ClientLocker.Lock()
	delete(self.mClientList, clientid)
	llog.Debugf("udp客户端：%s已断开连接[%d]", clientid)
	self.m_nClientCount--
	return true
}

func (self *UdpServerSocket) StopClient(id int) {

}

func (self *UdpServerSocket) LoadClient() *UdpSocketClient {
	s := &UdpSocketClient{}
	return s
}

func (self *UdpServerSocket) SendById(id int, buff []byte) int {
	pClient := self.GetClientById(id)
	if pClient != nil {
		ThisUdpServerSocket.mUpdConn.WriteTo(buff, pClient.RemoteAddr)
	} else {
		llog.Warningf("UdpServerSocket SendById: no client [%d]", id)
	}
	return 0
}

func (self *UdpServerSocket) BroadCast(buff []byte) {
	self.m_ClientLocker.RLock()
	for _, client := range self.mClientList {
		self.mUpdConn.WriteTo(buff, client.RemoteAddr)
	}
	self.m_ClientLocker.Unlock()
}

func (self *UdpServerSocket) Restart() bool {
	return true
}

func (self *UdpServerSocket) Connect() bool {
	return true
}

func (self *UdpServerSocket) Disconnect(bool) bool {
	return true
}

func (self *UdpServerSocket) OnNetFail(int) {
}

func (self *UdpServerSocket) Close() {
	self.mUpdConn.Close()
	self.Clear()

}

func (self *UdpServerSocket) SetMaxClients(maxnum int) {
	self.m_nMaxClients = maxnum
}

func handlerUdpMsg() {
	for {
		st := <-ThisUdpServerSocket.mBuffChan
		ThisUdpServerSocket.m_PacketFunc(st.ClientId, st.Buff, len(st.Buff))
	}
}

func udpserverRoutine() {
	//defer util.Recover()
	var buff = make([]byte, ThisUdpServerSocket.m_MaxReceiveBufferSize)
	for {
		n, udpAddr, err := ThisUdpServerSocket.mUpdConn.ReadFromUDP(buff)
		fmt.Println("udpserverRoutine: ", n, udpAddr.String())
		if err != nil || n <= 10 {
			llog.Errorf("udpserverRoutine ReadFromUDP error: %s, read = %d", err.Error(), n)
			continue
		}
		clientid := int(base.BytesToInt64(buff, binary.BigEndian))
		fmt.Println("udpserverRoutine clientid ", clientid)
		client, ok := ThisUdpServerSocket.mClientList[clientid]
		if !ok {
			ThisUdpServerSocket.AddClient(clientid, udpAddr)
		} else {
			client.RemoteAddr = udpAddr
		}

		st := UdpBufferTransport{ClientId: clientid}
		st.Buff = make([]byte, n, n)
		copy(st.Buff, buff[:n])
		ThisUdpServerSocket.mBuffChan <- st
	}
}
