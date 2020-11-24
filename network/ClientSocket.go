package network

import (
	"fmt"
	"io"
	"net"
	"runtime"

	"github.com/snowyyj001/loumiao/log"

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
	defer func() {
		if err := recover(); err != nil {

			fmt.Println("ClientSocket Send", err)
		}
	}()

	if self.m_Conn == nil {
		return 0
	}
	//log.Debugf("发送消息 %v", buff)
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
		log.Warningf("ClientSocket address error", self.m_sAddr)
		return false
	}
	//ln, err1 := net.DialTimeout("tcp4", self.m_sAddr, 5*time.Second)
	ln, err1 := net.DialTCP("tcp4", nil, tcpAddr)
	if err1 != nil {
		log.Errorf("ClientSocket DialTCP  %v", err1)
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
	buff, nLen := message.Encode(0, 0, "C_CONNECT", nil)
	self.HandlePacket(self.m_ClientId, buff, nLen)
}

func (self *ClientSocket) OnNetFail(int) {
	self.Stop()
	buff, nLen := message.Encode(0, 0, "C_DISCONNECT", nil)
	self.HandlePacket(self.m_ClientId, buff, nLen)
}

func clientRoutine(pClient *ClientSocket) bool {
	if pClient.m_Conn == nil {
		return false
	}

	defer func() {
		if err := recover(); err != nil {
			var buf [4096]byte
			n := runtime.Stack(buf[:], false)
			data := string(buf[:n])
			log.Error(data)
			log.Errorf("clientRoutine error: %v", err)
		}
	}()
	var buff = make([]byte, pClient.m_MaxReceiveBufferSize)
	for {
		if pClient.m_bShuttingDown {
			break
		}
		n, err := pClient.m_Conn.Read(buff)
		if err == io.EOF {
			log.Debugf("0.远程链接：%s已经关闭: %s", pClient.m_Conn.RemoteAddr().String(), err.Error())
			pClient.OnNetFail(0)
			break
		}
		if err != nil {
			log.Debugf("1.远程链接：%s已经关闭: %s", pClient.m_Conn.RemoteAddr().String(), err.Error())
			pClient.OnNetFail(1)
			break
		}
		if n > 0 {
			ok := pClient.ReceivePacket(pClient.m_ClientId, buff[:n])
			if !ok {
				log.Debugf("2.远程链接：%s已经关闭: %s", pClient.m_Conn.RemoteAddr().String(), err.Error())
				pClient.OnNetFail(2)
				break
			}
		}
	}

	pClient.Clear()
	return true
}
