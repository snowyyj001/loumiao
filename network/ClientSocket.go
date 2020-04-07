package network

import (
	"fmt"
	"io"
	"net"

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
	Uuid          string
	Uid           int
	Type          int
	SendTimes     int
}

func (self *ClientSocket) Init(ip string, port int) bool {
	if self.m_nPort == port || self.m_sIP == ip {
		return false
	}

	self.Socket.Init(ip, port)
	return true
}
func (self *ClientSocket) Start() bool {
	self.m_bShuttingDown = false

	if self.m_sIP == "" {
		self.m_sIP = "127.0.0.1"
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
	return true
}

func (self *ClientSocket) SendMsg(name string, msg interface{}) {
	buff, _ := message.Encode(name, msg)
	self.Send(buff)
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

	var strRemote = fmt.Sprintf("%s:%d", self.m_sIP, self.m_nPort)
	tcpAddr, err := net.ResolveTCPAddr("tcp4", strRemote)
	if err != nil {
		log.Warningf("ClientSocket address error", strRemote)
		return false
	}

Label:
	ln, err1 := net.DialTCP("tcp4", nil, tcpAddr)
	if err1 != nil {
		goto Label
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
	buff, nLen := message.Encode("CONNECT", nil)
	self.HandlePacket(self.m_ClientId, buff, nLen)
}

func (self *ClientSocket) OnNetFail(int) {
	self.Stop()
	buff, nLen := message.Encode("DISCONNECT", nil)
	self.HandlePacket(self.m_ClientId, buff, nLen)
}

func clientRoutine(pClient *ClientSocket) bool {
	if pClient.m_Conn == nil {
		return false
	}

	defer func() {
		if err := recover(); err != nil {
			fmt.Println("clientRoutine", err)
		}
	}()

	for {
		if pClient.m_bShuttingDown {
			break
		}

		var buff = make([]byte, pClient.m_MaxReceiveBufferSize)
		n, err := pClient.m_Conn.Read(buff)
		if err == io.EOF {
			fmt.Printf("0.远程链接：%s已经关闭！\n", pClient.m_Conn.RemoteAddr().String())
			pClient.OnNetFail(0)
			break
		}
		if err != nil {
			handleError(err)
			fmt.Printf("1.远程链接：%s已经关闭！\n", pClient.m_Conn.RemoteAddr().String())
			pClient.OnNetFail(1)
			break
		}
		if n > 0 {
			ok := pClient.ReceivePacket(pClient.m_ClientId, buff[:n])
			if !ok {
				fmt.Printf("2.远程链接：%s已经关闭！\n", pClient.m_Conn.RemoteAddr().String())
				pClient.OnNetFail(2)
				break
			}
		}
	}

	pClient.Close()
	return true
}
