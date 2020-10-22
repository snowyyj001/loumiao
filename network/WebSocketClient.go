package network

import (
	"github.com/snowyyj001/loumiao/log"
	"github.com/snowyyj001/loumiao/message"

	"github.com/gorilla/websocket"
)

type IWebSocketClient interface {
	ISocket
}

type WebSocketClient struct {
	Socket
	m_pServer *WebSocket
}

func (self *WebSocketClient) Start() bool {
	if self.m_nState != SSF_SHUT_DOWN {
		return false
	}

	if self.m_pServer == nil {
		return false
	}

	self.m_nState = SSF_ACCEPT

	//self.OnNetConn()
	go wserverclientRoutine(self)
	return true
}

func (self *WebSocketClient) Send(buff []byte) int {
	defer func() {
		if err := recover(); err != nil {
			log.Errorf("WebSocketClient Send", err)
		}
	}()

	if self.m_WsConn == nil {
		return 0
	}

	err := self.m_WsConn.WriteMessage(websocket.BinaryMessage, buff)
	handleError(err)
	return 0
}

func (self *WebSocketClient) OnNetConn() {
	buff, nLen := message.Encode(0, 0, "CONNECT", nil)
	//bufflittle := common.BigEngianToLittle(buff, nLen)
	self.HandlePacket(self.m_ClientId, buff, nLen)
}

func (self *WebSocketClient) OnNetFail(error int) {
	self.Stop()
	buff, nLen := message.Encode(0, 0, "DISCONNECT", nil)
	//bufflittle := common.BigEngianToLittle(buff, nLen)
	self.HandlePacket(self.m_ClientId, buff, nLen)
}

func (self *WebSocketClient) Close() {
	if self.m_WsConn != nil {
		self.m_WsConn.Close()
	}
	self.m_WsConn = nil
	self.Socket.Close()
	if self.m_pServer != nil {
		self.m_pServer.DelClinet(self)
	}
}

func wserverclientRoutine(pClient *WebSocketClient) bool {
	if pClient.m_WsConn == nil {
		return false
	}

	for {
		if pClient.m_bShuttingDown {
			break
		}

		mt, message, err := pClient.m_WsConn.ReadMessage()
		if err != nil {
			log.Debugf("远程链接：%s已经关闭！%v\n", pClient.m_WsConn.RemoteAddr().String(), err)
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
