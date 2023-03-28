package network

import (
	"github.com/snowyyj001/loumiao/llog"
	"github.com/snowyyj001/loumiao/lutil"
	"github.com/snowyyj001/loumiao/message"
	"io"
)

type IServerSocketClient interface {
	ISocket
}

type ServerSocketClient struct {
	Socket
	pServer *ServerSocket
}

func (self *ServerSocketClient) Start() bool {
	if self.m_nConnectType == 0 {
		llog.Error("ServerSocketClient.Start error : unkonwen socket type")
		return false
	}
	if self.m_nState != SSF_SHUT_DOWN {
		return false
	}

	if self.pServer == nil {
		return false
	}
	self.m_bShuttingDown = false
	self.m_nState = SSF_CONNECT
	//self.m_Conn.(*net.TCPConn).SetNoDelay(true)		//default is true，禁用Nagle算法
	//self.m_Conn.(*net.TCPConn).SetKeepAlive(true)		//default is enable，链接检测
	//self.m_Conn.(*net.TCPConn).SetKeepAlivePeriod(5*time.Second)		//default is 15 second (defaultTCPKeepAlive)，linux默认是7200秒
	//self.m_Conn.(*net.TCPConn).SetLinger(-1);		//default is < 0, 内核缺省close操作是立即返回，如果有数据残留在socket缓冲区中则系统将试着将这些数据发 送给对方。
	// cat /proc/sys/net/ipv4/tcp_max_syn_backlog		//半连接队列大小，在Linux内核2.2之后，分离为两个backlog来分别限制半连接（SYN_RCVD状态）队列大小和全连接（ESTABLISHED状态）队列大小
	// cat /proc/sys/net/core/somaxconn	//全连接队列大小,int listen(int sockfd, int backlog),值取somaxconn和backlog的最小的
	// cat /proc/sys/net/core/rmem_max		//可设置的最大读缓冲区大小
	// cat /proc/sys/net/core/wmem_max		//可设置的最大写缓冲区大小
	// cat /proc/sys/net/core/rmem_default		//默认的读缓冲区大小
	// cat /proc/sys/net/core/wmem_default		//默认的写缓冲区大小
	//SetReadBuffer()
	//SetWriteBuffer()
	self.OnNetConn()
	serverClientRoutine(self)

	return true
}

func (self *ServerSocketClient) sendClient(buff []byte) int {
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

	if self.pServer != nil {
		self.pServer.DelClinet(self)
		self.pServer = nil
	}
	self.Socket.Close()
}

// write msg
func serverWriteRoutine(pClient *ServerSocketClient) bool {
	defer lutil.Recover()
	if pClient.m_Conn == nil {
		return false
	}

	bMsg := broadMsgArray[pClient.m_broadMsgId]
	for {
		if pClient.m_bShuttingDown {
			break
		}

		if pClient.m_nState == SSF_SHUT_DOWN {
			break
		}

		select {
		case m := <-pClient.m_WriteChan:
			pClient.sendClient(m)
		case <-bMsg.c:
			pClient.sendClient(bMsg.buffer)
			pClient.m_broadMsgId++
			bMsg = broadMsgArray[pClient.m_broadMsgId]
		}
	}

	return true
}

// read msg
func serverReadRoutine(pClient *ServerSocketClient) bool {
	defer lutil.Recover()
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

func serverClientRoutine(pClient *ServerSocketClient) {
	lutil.Go(func() {
		serverReadRoutine(pClient)
	})
	lutil.Go(func() {
		serverWriteRoutine(pClient)
	})
}
