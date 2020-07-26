package network

import (
	"io"
	"net"

	"github.com/snowyyj001/loumiao/log"
	"github.com/snowyyj001/loumiao/message"
)

type IServerSocketClient interface {
	ISocket
}

type ServerSocketClient struct {
	Socket
	m_pServer *ServerSocket
}

func handleError(err error) {
	if err == nil {
		return
	}
	log.Errorf("错误：%s\n", err.Error())
}

func (self *ServerSocketClient) Start() bool {
	if self.m_nState != SSF_SHUT_DOWN {
		return false
	}

	if self.m_pServer == nil {
		return false
	}

	self.m_nState = SSF_CONNECT
	self.m_Conn.(*net.TCPConn).SetNoDelay(true)
	//self.m_Conn.SetKeepAlive(true)
	//self.m_Conn.SetKeepAlivePeriod(5*time.Second)
	self.OnNetConn()
	go serverclientRoutine(self)

	return true
}

func (self *ServerSocketClient) Send(buff []byte) int {
	defer func() {
		if err := recover(); err != nil {
			log.Errorf("ServerSocketClient Send", err)
		}
	}()

	n, err := self.m_Conn.Write(buff)
	handleError(err)
	if n > 0 {
		return n
	}
	return 0
}

func (self *ServerSocketClient) OnNetConn() {
	buff, nLen := message.Encode(-1, 0, "CONNECT", nil)
	//bufflittle := common.BigEngianToLittle(buff, nLen)
	self.HandlePacket(self.m_ClientId, buff, nLen)
}

func (self *ServerSocketClient) OnNetFail(error int) {
	self.Stop()
	buff, nLen := message.Encode(-1, 0, "DISCONNECT", nil)
	//bufflittle := common.BigEngianToLittle(buff, nLen)
	self.HandlePacket(self.m_ClientId, buff, nLen)
}

func (self *ServerSocketClient) Close() {
	self.Socket.Close()
	if self.m_pServer != nil {
		self.m_pServer.DelClinet(self)
	}
}

func serverclientRoutine(pClient *ServerSocketClient) bool {
	if pClient.m_Conn == nil {
		return false
	}

	defer func() {
		if err := recover(); err != nil {
			log.Errorf("serverclientRoutine", err)
		}
	}()
	var buff = make([]byte, pClient.m_MaxReceiveBufferSize)
	for {
		if pClient.m_bShuttingDown {
			break
		}

		n, err := pClient.m_Conn.Read(buff)
		if err == io.EOF {
			log.Debugf("远程链接：%s已经关闭！\n", pClient.m_Conn.RemoteAddr().String())
			pClient.OnNetFail(0)
			break
		}
		if err != nil {
			handleError(err)
			pClient.OnNetFail(1)
			break
		}
		if n > 0 {
			ok := pClient.ReceivePacket(pClient.m_ClientId, buff[:n])
			if !ok {
				pClient.OnNetFail(2)
				break
			}
		}
	}

	pClient.Close()
	return true
}
