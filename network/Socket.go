package network

import (
	"encoding/binary"
	"github.com/snowyyj001/loumiao/util"
	"net"
	"sync"

	"github.com/snowyyj001/loumiao/base"
	"github.com/xtaci/kcp-go"

	"github.com/snowyyj001/loumiao/config"
	"github.com/snowyyj001/loumiao/llog"

	"github.com/gorilla/websocket"
)

const (
	SSF_INIT = iota
	SSF_ACCEPT
	SSF_CONNECT
	SSF_SHUT_DOWN //已经关闭
)

const (
	CLIENT_CONNECT = iota + 1 //对外
	SERVER_CONNECT            //对内
	CHILD_CONNECT             //client
)

type BroadMsg struct {
	c      chan struct{}
	buffer []byte
}

var (
	broadMsgMapSync sync.Mutex
	broadMsgArray   [65536]*BroadMsg
	broadMsgId      uint16
)

func BoardSend(buffer []byte) {
	defer broadMsgMapSync.Unlock()
	broadMsgMapSync.Lock()

	c := make(chan struct{})
	bMsg := broadMsgArray[broadMsgId]
	broadMsgArray[broadMsgId+1] = &BroadMsg{c: c, buffer: buffer}
	broadMsgId++
	close(bMsg.c)
}

func handleError(err error) {
	if err == nil {
		return
	}
	llog.Warningf("handleError 错误：%s\n", err.Error())
}

type (
	HandleFunc func(int, []byte, int) error //回调函数
	Socket     struct {
		m_Conn                 net.Conn
		m_WsConn               *websocket.Conn
		m_KcpConn              *kcp.UDPSession
		m_sAddr                string
		m_nState               int
		m_nConnectType         int
		m_MaxReceiveBufferSize int
		m_MaxSendBufferSize    int

		m_ClientId int

		m_TotalNum     int
		m_AcceptedNum  int
		m_ConnectedNum int

		m_SendTimes     int
		m_ReceiveTimes  int
		m_bShuttingDown bool
		m_PacketFunc    HandleFunc

		m_pInBufferLen int
		m_pInBuffer    []byte
		m_WriteChan    chan []byte
		m_broadMsgId   uint16
	}

	ISocket interface {
		Init(string) bool
		Start() bool
		Restart() bool
		Connect() bool
		Disconnect(bool) bool
		OnNetConn(int)
		OnNetFail(int)
		Clear()
		Close()
		Send([]byte) int
		SendById(int, []byte) int
		BroadCast(buff []byte)
		GetSAddr() string

		GetState() int
		GetMaxSendBufferSize() int
		GetMaxReceiveBufferSize() int
		BindPacketFunc(HandleFunc)
		SetConnectType(int)
		SetTcpConn(net.Conn)
		ReceivePacket(int, []byte) bool
		HandlePacket(int, []byte, int) error
	}
)

// virtual
func (self *Socket) Init(saddr string) bool {
	self.m_sAddr = saddr
	self.Clear()
	return true
}

func (self *Socket) Start() bool {
	return true
}
func (self *Socket) Restart() bool {
	return true
}
func (self *Socket) Connect() bool {
	return true
}
func (self *Socket) Disconnect(bool) bool {
	return true
}
func (self *Socket) OnNetConn(int) {
}

func (self *Socket) OnNetFail(int) {
	self.Close()
}

func (self *Socket) GetState() int {
	return self.m_nState
}

func (self *Socket) SetState(state int) {
	self.m_nState = state
}

func (self *Socket) Send(buffer []byte) int {
	self.m_WriteChan <- buffer
	return 0
}

func (self *Socket) SendById(int, []byte) int {
	return 0
}

func (self *Socket) SetClientId(cid int) {
	self.m_ClientId = cid
}
func (self *Socket) GetClientId() int {
	return self.m_ClientId
}
func (self *Socket) GetSAddr() string {
	return self.m_sAddr
}

func (self *Socket) Clear() {
	self.m_nState = SSF_SHUT_DOWN
	self.m_Conn = nil
	self.m_WsConn = nil
	self.m_KcpConn = nil
	self.m_bShuttingDown = true
	self.m_nConnectType = -1
}

func (self *Socket) Close() {
	if self.m_nState == SSF_SHUT_DOWN {
		return
	}
	if self.m_Conn != nil {
		self.m_Conn.Close()
	}
	if self.m_WsConn != nil {
		self.m_WsConn.Close()
	}
	if self.m_KcpConn != nil {
		self.m_KcpConn.Close()
	}
	self.Clear()

}

func (self *Socket) GetMaxReceiveBufferSize() int {
	return self.m_MaxReceiveBufferSize
}

func (self *Socket) GetMaxSendBufferSize() int {
	return self.m_MaxSendBufferSize
}

func (self *Socket) SetConnectType(nType int) {
	self.m_nConnectType = nType
	if self.m_nConnectType == SERVER_CONNECT { //user for inner
		self.m_MaxSendBufferSize = config.NET_CLUSTER_BUFFER_SIZE
		self.m_MaxReceiveBufferSize = config.NET_CLUSTER_BUFFER_SIZE
	} else {
		self.m_MaxSendBufferSize = config.NET_BUFFER_SIZE
		self.m_MaxReceiveBufferSize = config.NET_BUFFER_SIZE
	}
	self.m_pInBuffer = make([]byte, self.m_MaxReceiveBufferSize) //预先申请一份内存来换取临时申请，减少gc但每个socket会申请2倍的m_MaxReceiveBufferSize内存大小
	self.m_WriteChan = make(chan []byte, config.NET_MAX_WRITE_CHANSIZE)
}

func (self *Socket) SetUdpConn(conn net.Conn) {
	self.m_Conn = conn
	//self.m_Reader = bufio.NewReader(conn)
	//self.m_Writer = bufio.NewWriter(conn)
}

func (self *Socket) SetTcpConn(conn net.Conn) {
	self.m_Conn = conn
	//self.m_Reader = bufio.NewReader(conn)
	//self.m_Writer = bufio.NewWriter(conn)
}

func (self *Socket) SetWsConn(conn *websocket.Conn) {
	self.m_WsConn = conn
	//self.m_Reader = bufio.NewReader(conn)
	//self.m_Writer = bufio.NewWriter(conn)
}

func (self *Socket) SetKcpConn(conn *kcp.UDPSession) {
	self.m_KcpConn = conn
}

func (self *Socket) BindPacketFunc(callfunc HandleFunc) {
	if callfunc == nil {
		llog.Error("BindPacketFunc: callfunc is nil") // 接受包错误
	}
	self.m_PacketFunc = callfunc
}

func (self *Socket) HandlePacket(Id int, buff []byte, nlen int) error {
	newbuff := make([]byte, nlen)
	copy(newbuff, buff[:nlen])
	return self.m_PacketFunc(Id, newbuff, nlen)
}

func (self *Socket) ReceivePacket(Id int, dat []byte) bool {
	defer util.Recover()
	//	llog.Debugf("收到消息包 %v %d", dat, len(dat))
	copy(self.m_pInBuffer[self.m_pInBufferLen:], dat)
	self.m_pInBufferLen += len(dat)
	for {
		if self.m_pInBufferLen < 8 {
			break
		}
		mbuff1 := self.m_pInBuffer[0:4]
		nLen := int(base.BytesToUInt32(mbuff1, binary.BigEndian)) //消息总长度
		//	llog.Debugf("当前消息包长度 %d", nLen, self.m_pInBufferLen)
		//	t, n, _, e := message.UnPackHead(self.m_pInBuffer, self.m_pInBufferLen)
		//	llog.Debugf("当前消息包名字 ", t, n, e)
		if nLen > self.m_pInBufferLen {
			break
		}
		if nLen > self.m_MaxReceiveBufferSize {
			llog.Errorf("ReceivePacket: 包长度越界[%d][%d]", nLen, self.m_MaxReceiveBufferSize) // 接受包错误
			self.Close()
			return false
		}

		err := self.HandlePacket(Id, self.m_pInBuffer, nLen)
		if err != nil {
			llog.Errorf("ReceivePacket HandlePacket error: %s", err.Error())
			return false
		}
		copy(self.m_pInBuffer, self.m_pInBuffer[nLen:self.m_pInBufferLen])
		self.m_pInBufferLen -= nLen
		//		llog.Debugf("剩余消息包长度 %d", self.m_pInBufferLen)
	}
	return true
}

func init() {
	broadMsgArray[broadMsgId] = &BroadMsg{c: make(chan struct{})}
}
