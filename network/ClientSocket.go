package network

import (
	"io"
	"net"
	"runtime"

	"github.com/snowyyj001/loumiao/llog"

	"github.com/snowyyj001/loumiao/message"
)

type IClientSocket interface {
	ISocket
}

type ClientSocket struct {
	Socket
	m_nMaxClients int
	m_nMinClients int
	Uid           int
	SendTimes     int
}

func (self *ClientSocket) Init(saddr string) bool {
	if self.m_sAddr == saddr {
		return false
	}
	self.Socket.Init(saddr)
	return true
}
func (self *ClientSocket) Start() bool {
	if self.m_nConnectType == 0 {
		llog.Error("ClientSocket.Start error : unkonwen socket type")
		return false
	}
	self.m_bShuttingDown = false
	if self.m_sAddr == "" {
		return false
	}

	if self.Connect() {
		self.m_Conn.(*net.TCPConn).SetNoDelay(true)
		go clientRoutine(self)
		return true
	}
	return false
}

func (self *ClientSocket) Stop() bool {
	if self.m_bShuttingDown {
		return true
	}
	self.m_bShuttingDown = true
	self.Close()
	return true
}

func (self *ClientSocket) Send(buff []byte) int {
	if self.m_Conn == nil {
		return 0
	}
	//llog.Debugf("发送消息 %v", buff)
	n, err := self.m_Conn.Write(buff)
	handleError(err)
	if n > 0 {
		return n
	}
	//self.m_Writer.Flush()
	return 0
}

func (self *ClientSocket) Restart() bool {
	return true
}

func (self *ClientSocket) Connect() bool {
	if self.m_nState == SSF_CONNECT {
		return false
	}

	tcpAddr, err := net.ResolveTCPAddr("tcp4", self.m_sAddr)
	if err != nil {
		llog.Errorf("ClientSocket address error: %s", self.m_sAddr)
		return false
	}
	//ln, err1 := net.DialTimeout("tcp4", self.m_sAddr, 5*time.Second)
	ln, err1 := net.DialTCP("tcp4", nil, tcpAddr)
	if err1 != nil {
		llog.Errorf("ClientSocket DialTCP  %v", err1)
		return false
	}

	self.m_nState = SSF_CONNECT
	self.SetTcpConn(ln)
	self.OnNetConn()

	return true
}

func (self *ClientSocket) OnDisconnect() {
}

func (self *ClientSocket) OnNetConn() {
	buff, nLen := message.Encode(0, "C_CONNECT", nil)
	self.HandlePacket(self.m_ClientId, buff, nLen)
}

func (self *ClientSocket) OnNetFail(int) {
	self.Stop()
	buff, nLen := message.Encode(0, "C_DISCONNECT", nil)
	self.HandlePacket(self.m_ClientId, buff, nLen)
}

func clientRoutine(pClient *ClientSocket) bool {
	defer func() {
		if r := recover(); r != nil {
			buf := make([]byte, 2048)
			l := runtime.Stack(buf, false)
			llog.Errorf("ClientSocket.clientRoutine %v: %s", r, buf[:l])
		}
	}()
	if pClient.m_Conn == nil {
		return false
	}
	var buff = make([]byte, pClient.m_MaxReceiveBufferSize)
	for {
		if pClient.m_bShuttingDown {
			break
		}
		n, err := pClient.m_Conn.Read(buff)
		if err == io.EOF {
			llog.Debugf("0.远程链接：%s已经关闭: %s", pClient.m_Conn.RemoteAddr().String(), err.Error())
			pClient.OnNetFail(0)
			break
		}
		if err != nil {
			llog.Debugf("1.远程链接：%s已经关闭: %s", pClient.m_Conn.RemoteAddr().String(), err.Error())
			pClient.OnNetFail(1)
			break
		}
		if n > 0 {
			ok := pClient.ReceivePacket(pClient.m_ClientId, buff[:n])
			if !ok {
				llog.Debugf("2.远程链接：%s已经关闭: %d", pClient.m_Conn.RemoteAddr().String(), n)
				pClient.OnNetFail(2)
				break
			}
		}
	}

	pClient.Stop()
	return true
}
