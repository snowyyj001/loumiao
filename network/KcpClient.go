package network

import (
	"github.com/snowyyj001/loumiao/llog"
	"github.com/snowyyj001/loumiao/message"
	"github.com/xtaci/kcp-go"
	"io"
	"runtime"
)

type KcpClient struct {
	Socket
}

func (self *KcpClient) Init(saddr string) bool {
	if self.m_sAddr == saddr {
		return false
	}
	self.Socket.Init(saddr)
	return true
}
func (self *KcpClient) Start() bool {
	if self.m_nConnectType == 0 {
		llog.Error("KcpClient.Start error : unkonwen socket type")
		return false
	}
	self.m_bShuttingDown = false
	if self.m_sAddr == "" {
		return false
	}

	if self.Connect() {
		go clientKcpRoutine(self)
		return true
	}
	return false
}

func (self *KcpClient) Send(buff []byte) int {
	if self.m_KcpConn == nil {
		return 0
	}
	//llog.Debugf("发送消息 %v", buff)
	n, err := self.m_KcpConn.Write(buff)
	if err != nil {
		llog.Errorf("KcpClient.Send error : %s", err.Error())
		return 0
	}
	//self.m_Writer.Flush()
	return n
}

func (self *KcpClient) Restart() bool {
	return true
}

func (self *KcpClient) Connect() bool {
	if self.m_nState == SSF_CONNECT {
		return false
	}
	kcpConn, err := kcp.DialWithOptions(self.m_sAddr, nil, 0, 0)
	if err != nil {
		llog.Errorf("KcpClient DialWithOptions[%s] error: %s", self.m_sAddr, err.Error())
		return false
	}
	// solve dead link problem:
	// physical disconnection without any communcation between client and server
	// will cause the read to block FOREVER, so a timeout is a rescue.
	//kcpConn.SetReadDeadline(time.Now().Add(time.Second * KCPTIMEOUT))
	// set kcp parameters
	kcpConn.SetWindowSize(KCPWinSedSize, KCPWinRevSize)
	kcpConn.SetNoDelay(KCPNoDelay, KCPInterval, KCPResend, KCPNoCongestion) //fast2
	kcpConn.SetStreamMode(KCPStreamMode)
	kcpConn.SetMtu(KCPMTU)
	kcpConn.SetWriteDelay(KCPWriteDelay)
	kcpConn.SetACKNoDelay(KCPAckNodelay)
	if err := kcpConn.SetReadBuffer(KCPReadBuffer); err != nil {
		llog.Errorf("KcpClient.SetReadBuffer: error %s", err.Error())
		return false
	}
	if err := kcpConn.SetWriteBuffer(KCPWriteBuffer); err != nil {
		llog.Errorf("KcpClient.Connect:  error %s", err.Error())
		return false
	}
	if err := kcpConn.SetDSCP(KCPDSCP); err != nil {
		kcpConn.Close()
		llog.Errorf("KcpClient.SetDSCP: error %s", err.Error())
		return false
	}
	self.m_nState = SSF_CONNECT
	self.SetKcpConn(kcpConn)
	self.OnNetConn()

	return true
}

func (self *KcpClient) OnDisconnect() {
}

func (self *KcpClient) OnNetConn() {
	buff, nLen := message.Encode(0, "C_CONNECT", nil)
	self.HandlePacket(self.m_ClientId, buff, nLen)
}

func (self *KcpClient) OnNetFail(int) {
	buff, nLen := message.Encode(0, "C_DISCONNECT", nil)
	self.HandlePacket(self.m_ClientId, buff, nLen)
	self.Close()
}

func clientKcpRoutine(pClient *KcpClient) bool {
	defer func() {
		if r := recover(); r != nil {
			buf := make([]byte, 2048)
			l := runtime.Stack(buf, false)
			llog.Errorf("KcpClient.clientRoutine %v: %s", r, buf[:l])
		}
	}()
	if pClient.m_KcpConn == nil {
		return false
	}
	var buff = make([]byte, pClient.m_MaxReceiveBufferSize)
	for {
		if pClient.m_bShuttingDown {
			break
		}
		n, err := pClient.m_KcpConn.Read(buff)
		if err == io.EOF {
			llog.Debugf("0.KcpClient远程链接：%s已经关闭: %s", pClient.m_KcpConn.RemoteAddr().String(), err.Error())
			pClient.OnNetFail(0)
			break
		}
		if err != nil {
			llog.Debugf("1.KcpClient远程链接：%s已经关闭: %s", pClient.m_KcpConn.RemoteAddr().String(), err.Error())
			pClient.OnNetFail(1)
			break
		}
		if n > 0 {
			ok := pClient.ReceivePacket(pClient.m_ClientId, buff[:n])
			if !ok {
				llog.Debugf("2.KcpClient远程链接：%s已经关闭: %d", pClient.m_KcpConn.RemoteAddr().String(), n)
				pClient.OnNetFail(2)
				break
			}
		}
	}

	pClient.Close()
	return true
}
