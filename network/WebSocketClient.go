package network

import (
	"github.com/gorilla/websocket"
	"github.com/snowyyj001/loumiao/lconfig"
	"github.com/snowyyj001/loumiao/llog"
	"github.com/snowyyj001/loumiao/lutil"
	"github.com/snowyyj001/loumiao/message"
)

type IWebSocketClient interface {
	ISocket
}

type WebSocketClient struct {
	Socket
	pServer      *WebSocket
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

	if self.pServer == nil {
		return false
	}
	self.m_nState = SSF_ACCEPT

	//self.m_WsConn.SetReadDeadline(time.Now().Add(10 * time.Second))

	self.OnNetConn()
	wsServerClientRoutine(self)
	return true
}

func (self *WebSocketClient) sendClient(buff []byte) int {
	if self.m_WsConn == nil {
		return 0
	}

	err := self.m_WsConn.WriteMessage(websocket.BinaryMessage, buff)
	handleError(err)
	return 0
}

func (self *WebSocketClient) OnNetConn() {
	buff, nLen := message.Encode(0, "CONNECT", nil)
	self.HandlePacket(self.m_ClientId, buff, nLen)
}

func (self *WebSocketClient) OnNetFail(error int) {
	buff, nLen := message.Encode(0, "DISCONNECT", nil)
	self.HandlePacket(self.m_ClientId, buff, nLen)
}

func (self *WebSocketClient) Stop() {
	if self.pServer != nil {
		self.pServer.DelClient(self)
		self.pServer = nil
	}
	self.m_WriteChan <- []byte{}
}

func wsServerClientWriteRoutine(pClient *WebSocketClient) bool {
	defer lutil.Recover()
	if pClient.m_Conn == nil {
		return false
	}

	bMsg := broadMsgArray[pClient.m_broadMsgId]
	for {
		select {
		case m := <-pClient.m_WriteChan:
			if m == nil || len(m) == 0 {
				goto EndLabel
			}
			if lconfig.NET_NODE_TYPE == lconfig.ServerType_Gate {
				message.EncryptBuffer(m, len(m))
			}
			pClient.sendClient(m)
		case <-bMsg.c:
			if lconfig.NET_NODE_TYPE == lconfig.ServerType_Gate {
				message.EncryptBuffer(bMsg.buffer, len(bMsg.buffer))
			}
			pClient.sendClient(bMsg.buffer)
			pClient.m_broadMsgId++
			bMsg = broadMsgArray[pClient.m_broadMsgId]
		}
	}

EndLabel:
	pClient.Close()
	return true
}

func wsServerClientReadRoutine(pClient *WebSocketClient) bool {
	defer lutil.Recover()
	if pClient.m_WsConn == nil {
		return false
	}

	for {
		mt, buff, err := pClient.m_WsConn.ReadMessage()
		if err != nil {
			llog.Infof("远程链接：%s已经关闭！%v\n", pClient.GetSAddr(), err)
			pClient.OnNetFail(1)
			break
		}
		if mt != websocket.BinaryMessage {
			llog.Infof("远程read内容格式错误: %s！ %s", pClient.GetSAddr(), string(buff))
			pClient.OnNetFail(2)
			break
		}

		if lconfig.NET_NODE_TYPE == lconfig.ServerType_Gate {
			message.DecryptBuffer(buff, len(buff))
		}
		ok := pClient.ReceivePacket(pClient.m_ClientId, buff)
		if !ok {
			llog.Errorf("远程ReceivePacket错误: %s！", pClient.GetSAddr())
			pClient.OnNetFail(3)
			break
		}
	}

	pClient.Close()
	return true
}

func wsServerClientRoutine(pClient *WebSocketClient) {
	lutil.Go(func() {
		wsServerClientReadRoutine(pClient)
	})
	lutil.Go(func() {
		wsServerClientWriteRoutine(pClient)
	})
}
