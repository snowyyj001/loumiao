package network

import (
	"github.com/snowyyj001/loumiao/util"
	"time"

	"github.com/snowyyj001/loumiao/llog"

	"github.com/snowyyj001/loumiao/message"
)

type KCPSocketClient struct {
	Socket
	m_pServer   *KcpSocket
	mHeartTimer *time.Timer
	mMsgRecved  bool
	mHeartDone  chan bool
}

func (self *KCPSocketClient) Start() bool {
	if self.m_nConnectType == 0 {
		llog.Error("KCPSocketClient.Start error : unkonwen socket type")
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

	self.OnNetConn()
	util.Go(func() {
		kcpclientRoutine(self)
	})
	return true
}

func (self *KCPSocketClient) Send(buff []byte) int {
	n, err := self.m_KcpConn.Write(buff)
	if err != nil {
		llog.Errorf("KCPSocketClient.Send error: %s", err.Error())
		return 0
	}
	if n > 0 {
		return n
	}
	return 0
}

func (self *KCPSocketClient) OnNetConn() {
	buff, nLen := message.Encode(0, "CONNECT", nil)
	self.HandlePacket(self.m_ClientId, buff, nLen)

	delat := time.Duration(KCPTIMEOUT) * time.Second
	self.mHeartTimer = time.NewTimer(delat)
	self.mHeartDone = make(chan bool)

	util.Go(func() {
		for {
			select {
			case <-self.mHeartTimer.C:
				if self.mMsgRecved {
					self.mHeartTimer.Reset(delat)
					self.mMsgRecved = false
				} else {
					if self.m_KcpConn != nil {
						self.m_KcpConn.Close()
					}
					return
				}
			case <-self.mHeartDone:
				return
			}
		}
	})
}

func (self *KCPSocketClient) OnNetFail(errcode int) {
	buff, nLen := message.Encode(0, "DISCONNECT", nil)
	self.HandlePacket(self.m_ClientId, buff, nLen)
	self.Close()
}

func (self *KCPSocketClient) Close() {
	if self.m_pServer != nil {
		self.m_pServer.DelClinet(self)
		self.m_pServer = nil
	}
	self.Socket.Close()
	self.mHeartDone <- true
}

func kcpclientRoutine(pClient *KCPSocketClient) bool {
	defer util.Recover()
	if pClient.m_KcpConn == nil {
		return false
	}
	var buff = make([]byte, pClient.m_MaxReceiveBufferSize)
	for {
		if pClient.m_bShuttingDown {
			llog.Infof("KCPSocketClient远程链接：%s已经被关闭！", pClient.GetSAddr())
			pClient.OnNetFail(0)
			break
		}

		n, err := pClient.m_KcpConn.Read(buff)
		if err != nil {
			llog.Infof("KCPSocketClient远程read错误: %s！ %s", pClient.GetSAddr(), err.Error())
			pClient.OnNetFail(2)
			break
		}
		//fmt.Println("kcpclientRoutine ", n)
		if n > 0 {
			ok := pClient.ReceivePacket(pClient.m_ClientId, buff[:n])
			if !ok {
				llog.Errorf("KCPSocketClient远程ReceivePacket错误: %s！", pClient.GetSAddr())
				pClient.OnNetFail(3)
				break
			}
		}
		pClient.mMsgRecved = true
	}

	pClient.Close()
	return true
}
