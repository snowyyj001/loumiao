package network

import (
	"github.com/gorilla/websocket"
	"github.com/snowyyj001/loumiao/llog"
	"github.com/snowyyj001/loumiao/message"
	"runtime"
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
	self.m_bShuttingDown = false
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
	if self.pServer != nil {
		self.pServer.DelClinet(self)
		self.pServer = nil
	}
	self.Socket.Close()
}
func wsServerClientWriteRoutine(pClient *WebSocketClient) bool {
	defer func() {
		if r := recover(); r != nil {
			buf := make([]byte, 2048)
			l := runtime.Stack(buf, false)
			llog.Errorf("wsServerClientWriteRoutine %v: %s", r, buf[:l])
		}
	}()
	if pClient.m_Conn == nil {
		return false
	}

	bMsg := broadMsgArray[pClient.m_broadMsgId]
	for {
		if pClient.m_bShuttingDown {
			break
		}

		if pClient.m_nState == SSF_SHUT_DOWN {
			break
		}

		select {
		case m := <-pClient.m_WriteChan:
			pClient.sendClient(m)
		case <-bMsg.c:
			pClient.sendClient(bMsg.buffer)
			pClient.m_broadMsgId++
			bMsg = broadMsgArray[pClient.m_broadMsgId]
		}
	}

	return true
}

func wsServerClientReadRoutine(pClient *WebSocketClient) bool {
	defer func() {
		if r := recover(); r != nil {
			buf := make([]byte, 2048)
			l := runtime.Stack(buf, false)
			llog.Errorf("wsServerClientReadRoutine %v: %s", r, buf[:l])
		}
	}()
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

func wsServerClientRoutine(pClient *WebSocketClient) {
	go wsServerClientReadRoutine(pClient)
	go wsServerClientWriteRoutine(pClient)
}
