package network

import (
	"io"
	"net"
	"runtime"

	"github.com/snowyyj001/loumiao/llog"
	"github.com/snowyyj001/loumiao/message"
)

type IServerSocketClient interface {
	ISocket
}

type ServerSocketClient struct {
	Socket
	m_pServer *ServerSocket
}

func (self *ServerSocketClient) Start() bool {
	if self.m_nConnectType == 0 {
		llog.Error("ServerSocketClient.Start error : unkonwen socket type")
		return false
	}
	if self.m_nState != SSF_SHUT_DOWN {
		return false
	}

	if self.m_pServer == nil {
		return false
	}
	self.m_bShuttingDown = false
	self.m_nState = SSF_CONNECT
	self.m_Conn.(*net.TCPConn).SetNoDelay(true)
	//self.m_Conn.SetKeepAlive(true)
	//self.m_Conn.SetKeepAlivePeriod(5*time.Second)
	self.OnNetConn()
	go serverclientRoutine(self)

	return true
}

func (self *ServerSocketClient) Send(buff []byte) int {
	n, err := self.m_Conn.Write(buff)
	handleError(err)
	if n > 0 {
		return n
	}
	return 0
}

func (self *ServerSocketClient) OnNetConn() {
	buff, nLen := message.Encode(0, "CONNECT", nil)
	self.HandlePacket(self.m_ClientId, buff, nLen)
}

func (self *ServerSocketClient) OnNetFail(errcode int) {
	buff, nLen := message.Encode(0, "DISCONNECT", nil)
	self.HandlePacket(self.m_ClientId, buff, nLen)
	self.Close()
}

func (self *ServerSocketClient) Close() {

	if self.m_pServer != nil {
		self.m_pServer.DelClinet(self)
		self.m_pServer = nil
	}
	self.Socket.Close()
}

func serverclientRoutine(pClient *ServerSocketClient) bool {
	defer func() {
		if r := recover(); r != nil {
			buf := make([]byte, 2048)
			l := runtime.Stack(buf, false)
			llog.Errorf("ServerSocketClient.serverclientRoutine %v: %s", r, buf[:l])
		}
	}()
	if pClient.m_Conn == nil {
		return false
	}
	var buff = make([]byte, pClient.m_MaxReceiveBufferSize)
	for {
		if pClient.m_bShuttingDown {
			llog.Infof("远程链接：%s已经被关闭！", pClient.GetSAddr())
			pClient.OnNetFail(0)
			break
		}

		n, err := pClient.m_Conn.Read(buff)
		if err == io.EOF {
			llog.Infof("远程m_Conn：%s已经关闭！", pClient.GetSAddr())
			pClient.OnNetFail(1)
			break
		}
		if err != nil {
			llog.Infof("远程read错误: %s！ %s", pClient.GetSAddr(), err.Error())
			pClient.OnNetFail(2)
			break
		}
		if n > 0 {
			ok := pClient.ReceivePacket(pClient.m_ClientId, buff[:n])
			if !ok {
				llog.Errorf("远程ReceivePacket错误: %s！", pClient.GetSAddr())
				pClient.OnNetFail(3)
				break
			}
		}
	}

	pClient.Close()
	return true
}
