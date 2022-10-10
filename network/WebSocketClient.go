package network

import (
	"github.com/snowyyj001/loumiao/llog"
	"github.com/snowyyj001/loumiao/message"

	"github.com/gorilla/websocket"
)

type IWebSocketClient interface {
	ISocket
}

type WebSocketClient struct {
	Socket
	m_pServer    *WebSocket
	m_ClientAddr string
}

func (self *WebSocketClient) Start() bool {
	if self.m_nConnectType == 0 {
		llog.Error("WebSocketClient.Start error : unkonwen socket type")
		return false
	}
	if self.m_nState != SSF_SHUT_DOWN {
		return false
	}

	if self.m_pServer == nil {
		return false
	}
	self.m_bShuttingDown = false
	self.m_nState = SSF_ACCEPT

	self.OnNetConn()
	go wserverclientRoutine(self)
	return true
}

func (self *WebSocketClient) Send(buff []byte) int {
	if self.m_WsConn == nil {
		return 0
	}

	err := self.m_WsConn.WriteMessage(websocket.BinaryMessage, buff)
	handleError(err)
	return 0
}

func (self *WebSocketClient) OnNetConn() {
	buff, nLen := message.Encode(0, "CONNECT", nil)
	//bufflittle := common.BigEngianToLittle(buff, nLen)
	self.HandlePacket(self.m_ClientId, buff, nLen)
}

func (self *WebSocketClient) OnNetFail(error int) {
	buff, nLen := message.Encode(0, "DISCONNECT", nil)
	//bufflittle := common.BigEngianToLittle(buff, nLen)
	self.HandlePacket(self.m_ClientId, buff, nLen)
	self.Close()
}

func (self *WebSocketClient) Close() {
	if self.m_pServer != nil {
		self.m_pServer.DelClinet(self)
		self.m_pServer = nil
	}
	self.Socket.Close()
}

func wserverclientRoutine(pClient *WebSocketClient) bool {
	if pClient.m_WsConn == nil {
		return false
	}

	for {
		if pClient.m_bShuttingDown {
			llog.Infof("远程链接：%s已经被关闭！", pClient.GetSAddr())
			pClient.OnNetFail(0)
			break
		}

		mt, message, err := pClient.m_WsConn.ReadMessage()
		if err != nil {
			llog.Infof("远程链接：%s已经关闭！%v\n", pClient.GetSAddr(), err)
			pClient.OnNetFail(1)
			break
		}
		if mt != websocket.BinaryMessage {
			llog.Infof("远程read内容格式错误: %s！ %s", pClient.GetSAddr(), string(message))
			pClient.OnNetFail(2)
			break
		}

		ok := pClient.ReceivePacket(pClient.m_ClientId, message)
		if !ok {
			llog.Errorf("远程ReceivePacket错误: %s！", pClient.GetSAddr())
			pClient.OnNetFail(3)
			break
		}
	}

	pClient.Close()
	return true
}
