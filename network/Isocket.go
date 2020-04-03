package network

import (
	"encoding/binary"
	"net"

	"github.com/snowyyj001/loumiao/config"
	"github.com/snowyyj001/loumiao/log"
	"github.com/snowyyj001/loumiao/util"

	"github.com/gorilla/websocket"
)

const (
	SSF_ACCEPT = iota
	SSF_CONNECT
	SSF_SHUT_DOWN //已经关闭
)

const (
	CLIENT_CONNECT = iota //对外
	SERVER_CONNECT        //对内
	CHILD_CONNECT         //client
)

const (
	MAX_WRITE_CHAN = 32
)

type (
	HandleFunc func(int, []byte, int) bool //回调函数
	Socket     struct {
		m_Conn                 net.Conn
		m_WsConn               *websocket.Conn
		m_nPort                int
		m_sIP                  string
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
	}

	ISocket interface {
		Init(string, int) bool
		Start() bool
		Stop() bool
		Restart() bool
		Connect() bool
		Disconnect(bool) bool
		OnNetConn(int)
		OnNetFail(int)
		Clear()
		Close()
		Send([]byte) int
		SendById(int, []byte) int
		SendMsg(string, ...interface{})
		SendMsgById(int, string, ...interface{})

		GetState() int
		SetMaxSendBufferSize(int)
		GetMaxSendBufferSize() int
		SetMaxReceiveBufferSize(int)
		GetMaxReceiveBufferSize() int
		BindPacketFunc(HandleFunc)
		SetConnectType(int)
		SetTcpConn(net.Conn)
		ReceivePacket(int, []byte) bool
		HandlePacket(int, []byte, int) bool
	}
)

// virtual
func (self *Socket) Init(ip string, port int) bool {
	self.m_nPort = port
	self.m_sIP = ip
	self.m_nState = SSF_SHUT_DOWN
	self.m_MaxSendBufferSize = config.NET_BUFFER_SIZE
	self.m_MaxReceiveBufferSize = config.NET_BUFFER_SIZE
	self.m_nConnectType = SERVER_CONNECT
	return true
}

func (self *Socket) Start() bool {
	return true
}
func (self *Socket) Stop() bool {
	self.m_bShuttingDown = true
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
	self.Stop()
}

func (self *Socket) GetState() int {
	return self.m_nState
}

func (self *Socket) Send([]byte) int {
	return 0
}

func (self *Socket) SendById(int, []byte) int {
	return 0
}

func (self *Socket) SendMsg(funcName string, params ...interface{}) {
}

func (self *Socket) SendMsgById(int, string, ...interface{}) {
}

func (self *Socket) SetClientId(cid int) {
	self.m_ClientId = cid
}
func (self *Socket) GetClientId() int {
	return self.m_ClientId
}
func (self *Socket) GetIP() string {
	return self.m_sIP
}
func (self *Socket) GetPort() int {
	return self.m_nPort
}

func (self *Socket) Clear() {
	self.m_nState = SSF_SHUT_DOWN
	self.m_Conn = nil
	self.m_WsConn = nil
	self.m_bShuttingDown = false
}

func (self *Socket) Close() {
	if self.m_Conn != nil {
		self.m_Conn.Close()
	}
	if self.m_WsConn != nil {
		self.m_WsConn.Close()
	}
	self.Clear()
}

func (self *Socket) GetMaxReceiveBufferSize() int {
	return self.m_MaxReceiveBufferSize
}

func (self *Socket) SetMaxReceiveBufferSize(maxReceiveSize int) {
	self.m_MaxReceiveBufferSize = maxReceiveSize
}

func (self *Socket) GetMaxSendBufferSize() int {
	return self.m_MaxSendBufferSize
}

func (self *Socket) SetMaxSendBufferSize(maxSendSize int) {
	self.m_MaxSendBufferSize = maxSendSize
}

func (self *Socket) SetConnectType(nType int) {
	self.m_nConnectType = nType
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

func (self *Socket) BindPacketFunc(callfunc HandleFunc) {
	if callfunc == nil {
		log.Error("BindPacketFunc: callfunc is nil") // 接受包错误
	}
	self.m_PacketFunc = callfunc
}

func (self *Socket) HandlePacket(Id int, buff []byte, nlen int) bool {
	return self.m_PacketFunc(Id, buff, nlen)
}

func (self *Socket) ReceivePacket(Id int, dat []byte) bool {
	defer func() {
		if err := recover(); err != nil {
			log.Errorf("ReceivePacket", err) // 接受包错误
		}
	}()
	//log.Debugf("收到消息包 %v", dat)
	self.m_pInBufferLen += len(dat)
	self.m_pInBuffer = append(self.m_pInBuffer, dat...)
	for {
		if self.m_pInBufferLen < 4 {
			break
		}
		mbuff1 := self.m_pInBuffer[0:2]
		nLen1 := int(util.BytesToUInt16(mbuff1, binary.BigEndian)) //消息总长度
		if nLen1 > self.m_pInBufferLen {
			break
		}
		if nLen1 > self.m_MaxReceiveBufferSize {
			log.Errorf("ReceivePacket: 包长度越界[%d][%d]", nLen1, self.m_MaxReceiveBufferSize) // 接受包错误
			return false
		}

		ok := self.HandlePacket(Id, self.m_pInBuffer, int(nLen1))
		if ok == false {
			return false
		}
		self.m_pInBufferLen -= int(nLen1)
		if self.m_pInBufferLen > 0 {
			self.m_pInBuffer = self.m_pInBuffer[int(nLen1):]
		} else {
			self.m_pInBuffer = []byte{}
		}
	}
	return true
}
