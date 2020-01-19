package network

import (
	"io"
	"github.com/snowyyj001/loumiao/log"
	"github.com/snowyyj001/loumiao/message"
	"net"
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

func (this *ServerSocketClient) Start() bool {
	if this.m_nState != SSF_SHUT_DOWN {
		return false
	}

	if this.m_pServer == nil {
		return false
	}

	this.m_nState = SSF_CONNECT
	this.m_Conn.(*net.TCPConn).SetNoDelay(true)
	//this.m_Conn.SetKeepAlive(true)
	//this.m_Conn.SetKeepAlivePeriod(5*time.Second)
	this.OnNetConn()
	go serverclientRoutine(this)

	return true
}

func (this *ServerSocketClient) Send(buff []byte) int {
	defer func() {
		if err := recover(); err != nil {
			log.Errorf("ServerSocketClient Send", err)
		}
	}()

	n, err := this.m_Conn.Write(buff)
	handleError(err)
	if n > 0 {
		return n
	}
	//this.m_Writer.Flush()
	return 0
}

func (this *ServerSocketClient) OnNetConn() {
	buff, nLen := message.Encode("CONNECT", nil)
	//bufflittle := common.BigEngianToLittle(buff, nLen)
	this.HandlePacket(this.m_ClientId, buff, nLen)
}

func (this *ServerSocketClient) OnNetFail(error int) {
	this.Stop()
	buff, nLen := message.Encode("DISCONNECT", nil)
	//bufflittle := common.BigEngianToLittle(buff, nLen)
	this.HandlePacket(this.m_ClientId, buff, nLen)
}

func (this *ServerSocketClient) Close() {
	this.Socket.Close()
	if this.m_pServer != nil {
		this.m_pServer.DelClinet(this)
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

	for {
		if pClient.m_bShuttingDown {
			break
		}

		var buff = make([]byte, pClient.m_MaxReceiveBufferSize)
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
	log.Debugf("%s关闭连接\n", pClient.m_sIP)
	return true
}
