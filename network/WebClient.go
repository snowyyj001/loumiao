package network

import (
	"net/url"
	"runtime"

	"github.com/gorilla/websocket"
	"github.com/snowyyj001/loumiao/llog"
	"github.com/snowyyj001/loumiao/message"
)

type WebClient struct {
	Socket
	mMaxClients int
	mMinClients int
}

func (self *WebClient) Init(saddr string) bool {
	if self.m_sAddr == saddr {
		return false
	}

	self.Socket.Init(saddr)
	return true
}
func (self *WebClient) Start() bool {
	if self.m_nConnectType == 0 {
		llog.Error("WebClient.Start error : unkonwen socket type")
		return false
	}
	self.m_bShuttingDown = false

	if self.m_sAddr == "" {
		return false
	}

	if self.Connect() {
		go wsclientRoutine(self)
	} else {
		llog.Errorf("WebClient.Start error : can not connect %s", self.m_sAddr)
	}
	//延迟，监听关闭
	//defer ln.Close()
	return true
}

func (self *WebClient) Stop() bool {
	if self.m_bShuttingDown {
		return true
	}
	self.m_bShuttingDown = true
	self.Close()
	return true
}

func (self *WebClient) Send(buff []byte) int {
	if self.m_WsConn == nil {
		return 0
	}
	err := self.m_WsConn.WriteMessage(websocket.BinaryMessage, buff)
	handleError(err)
	//self.m_Writer.Flush()
	return 0
}

func (self *WebClient) Restart() bool {
	return true
}

func (self *WebClient) Connect() bool {
	if self.m_nState == SSF_CONNECT {
		return false
	}
	//fmt.Println("11111111111111")
	wsAddr := url.URL{Scheme: "ws", Host: self.m_sAddr, Path: "/ws"}
	conn, _, err := websocket.DefaultDialer.Dial(wsAddr.String(), nil)
	//fmt.Println("2222222222")
	if err != nil {
		return false
	}
	self.m_nState = SSF_CONNECT
	self.SetWsConn(conn)
	self.OnNetConn()
	return true
}

func (self *WebClient) OnDisconnect() {
}

func (self *WebClient) OnNetConn() {
	buff, nLen := message.Encode(0, "C_CONNECT", nil)
	self.HandlePacket(self.m_ClientId, buff, nLen)
}

func (self *WebClient) OnNetFail(int) {
	buff, nLen := message.Encode(0, "C_DISCONNECT", nil)
	self.HandlePacket(self.m_ClientId, buff, nLen)
	self.Close()
}

func wsclientRoutine(pClient *WebClient) bool {
	defer func() {
		if r := recover(); r != nil {
			buf := make([]byte, 2048)
			l := runtime.Stack(buf, false)
			llog.Errorf("WebClient.wsclientRoutine %v: %s", r, buf[:l])
		}
	}()
	if pClient.m_WsConn == nil {
		return false
	}

	for {
		if pClient.m_bShuttingDown {
			break
		}

		mt, message, err := pClient.m_WsConn.ReadMessage()

		if err != nil {
			handleError(err)
			pClient.OnNetFail(0)
			break
		}
		if mt != websocket.BinaryMessage {
			pClient.OnNetFail(1)
			break
		}

		ok := pClient.ReceivePacket(pClient.m_ClientId, message)
		if !ok {
			pClient.OnNetFail(2)
			break
		}

	}

	pClient.Close()
	return true
}
