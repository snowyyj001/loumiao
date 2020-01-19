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

func (this *WebSocketClient) Start() bool {
	if this.m_nState != SSF_SHUT_DOWN {
		return false
	}

	if this.m_pServer == nil {
		return false
	}

	this.m_nState = SSF_ACCEPT

	//this.OnNetConn()
	go wserverclientRoutine(this)
	//wserverclientRoutine(this)
	return true
}

func (this *WebSocketClient) Send(buff []byte) int {
	defer func() {
		if err := recover(); err != nil {
			log.Errorf("WebSocketClient Send", err)
		}
	}()

	if this.m_WsConn == nil {
		return 0
	}

	err := this.m_WsConn.WriteMessage(websocket.BinaryMessage, buff)
	handleError(err)
	return 0
}

func (this *WebSocketClient) OnNetConn() {
	buff, nLen := message.Encode("CONNECT", nil)
	//bufflittle := common.BigEngianToLittle(buff, nLen)
	this.HandlePacket(this.m_ClientId, buff, nLen)
}

func (this *WebSocketClient) OnNetFail(error int) {
	this.Stop()
	buff, nLen := message.Encode("DISCONNECT", nil)
	//bufflittle := common.BigEngianToLittle(buff, nLen)
	this.HandlePacket(this.m_ClientId, buff, nLen)
}

func (this *WebSocketClient) Close() {
	if this.m_WsConn != nil {
		this.m_WsConn.Close()
	}
	this.m_WsConn = nil
	this.Socket.Close()
	if this.m_pServer != nil {
		this.m_pServer.DelClinet(this)
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
	log.Infof("%s关闭连接", pClient.m_sIP)
	return true
}
