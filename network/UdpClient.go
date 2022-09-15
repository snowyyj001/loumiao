package network

import (
	"github.com/snowyyj001/loumiao/llog"
	"github.com/snowyyj001/loumiao/message"
	"github.com/snowyyj001/loumiao/util"
	"net"
	"strings"
)

type UdpClient struct {
	Socket
	mMaxClients int
	mMinClients int
	Uid         int
	SendTimes   int
}

func (self *UdpClient) Init(saddr string) bool {
	if self.m_sAddr == saddr {
		return false
	}
	self.Socket.Init(saddr)
	return true
}
func (self *UdpClient) Start() bool {
	if self.m_nConnectType == 0 {
		llog.Error("UdpClient.Start error : unkonwen socket type")
		return false
	}
	self.m_bShuttingDown = false
	if self.m_sAddr == "" {
		return false
	}

	if self.Connect() {
		go clientUdpRoutine(self)
		return true
	}
	return false
}

func (self *UdpClient) Stop() bool {
	if self.m_bShuttingDown {
		return true
	}
	self.m_bShuttingDown = true
	self.Close()
	return true
}

func (self *UdpClient) Send(buff []byte) int {
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
	return n
}

func (self *UdpClient) Restart() bool {
	return true
}

func (self *UdpClient) Connect() bool {
	if self.m_nState == SSF_CONNECT {
		return false
	}
	arr := strings.Split(self.m_sAddr, ":")
	udpConn, err := net.DialUDP("udp", nil, &net.UDPAddr{
		IP:   net.ParseIP(arr[0]),
		Port: util.Atoi(arr[1]),
	})
	if err != nil {
		llog.Errorf("UdpClient address error: %s, %s", self.m_sAddr, err.Error())
		return false
	}
	self.m_nState = SSF_CONNECT
	self.SetTcpConn(udpConn)
	self.OnNetConn()

	return true
}

func (self *UdpClient) OnDisconnect() {
}

func (self *UdpClient) OnNetConn() {

}

func (self *UdpClient) OnNetFail(int) {

}

func clientUdpRoutine(pClient *UdpClient) {
	defer util.Recover()
	var buff = make([]byte, pClient.m_MaxReceiveBufferSize)
	for {
		n, err := pClient.m_Conn.Read(buff)
		if err != nil {
			llog.Debugf("udp read error: %s", pClient.m_Conn.RemoteAddr().String(), err.Error())
			pClient.OnNetFail(0)
			break
		}
		if n < 10 {
			llog.Debugf("udp read msg too little: %d", n)
			continue
		}
		msgId, _, body := message.UnPackUdp(buff)
		err = pClient.m_PacketFunc(int(msgId), body, n-10)
		if err != nil {
			llog.Debugf("udp 处理消息错误: %s", err.Error())
			pClient.OnNetFail(1)
			break
		}
	}
}
