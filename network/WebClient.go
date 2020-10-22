package network

import (
	"fmt"
	"net/url"

	"github.com/snowyyj001/loumiao/message"

	"github.com/gorilla/websocket"
)

type WebClient struct {
	Socket
	m_nMaxClients int
	m_nMinClients int
}

func (self *WebClient) Init(saddr string) bool {
	if self.m_sAddr == saddr {
		return false
	}

	self.Socket.Init(saddr)
	fmt.Println("ClientSocket", saddr)
	return true
}
func (self *WebClient) Start() bool {
	self.m_bShuttingDown = false

	if self.m_sAddr == "" {
		return false
	}

	if self.Connect() {
		go wsclientRoutine(self)
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
	return true
}

func (self *WebClient) Send(buff []byte) int {
	defer func() {
		if err := recover(); err != nil {
			fmt.Println("WebClient Send", err)
		}
	}()

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

	wsAddr := url.URL{Scheme: "ws", Host: self.m_sAddr, Path: "/ws"}
	conn, _, err := websocket.DefaultDialer.Dial(wsAddr.String(), nil)
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
	buff, nLen := message.Encode(0, 0, "CONNECT", nil)
	self.HandlePacket(self.m_ClientId, buff, nLen)
}

func (self *WebClient) OnNetFail(int) {
	self.Stop()
	buff, nLen := message.Encode(0, 0, "DISCONNECT", nil)
	self.HandlePacket(self.m_ClientId, buff, nLen)
}

func wsclientRoutine(pClient *WebClient) bool {
	if pClient.m_WsConn == nil {
		return false
	}

	defer func() {
		if err := recover(); err != nil {
			fmt.Println("wsclientRoutine", err)
		}
	}()

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
