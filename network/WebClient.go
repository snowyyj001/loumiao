package network

import (
	"fmt"
	"github.com/snowyyj001/loumiao/message"
	"net/url"

	"github.com/gorilla/websocket"
)

type WebClient struct {
	Socket
	m_nMaxClients int
	m_nMinClients int
}

func (this *WebClient) Init(ip string, port int) bool {
	if this.m_nPort == port || this.m_sIP == ip {
		return false
	}

	this.Socket.Init(ip, port)
	fmt.Println("ClientSocket", ip, port)
	return true
}
func (this *WebClient) Start() bool {
	this.m_bShuttingDown = false

	if this.m_sIP == "" {
		this.m_sIP = "127.0.0.1"
	}

	if this.Connect() {
		go wsclientRoutine(this)
	}
	//延迟，监听关闭
	//defer ln.Close()
	return true
}

func (this *WebClient) Stop() bool {
	if this.m_bShuttingDown {
		return true
	}

	this.m_bShuttingDown = true
	return true
}

func (this *WebClient) SendMsg(name string, msg interface{}) {
	buff, _ := message.Encode(name, msg)
	this.Send(buff)
}

func (this *WebClient) Send(buff []byte) int {
	defer func() {
		if err := recover(); err != nil {
			fmt.Println("WebClient Send", err)
		}
	}()

	if this.m_WsConn == nil {
		return 0
	}
	err := this.m_WsConn.WriteMessage(websocket.BinaryMessage, buff)
	handleError(err)
	//this.m_Writer.Flush()
	return 0
}

func (this *WebClient) Restart() bool {
	return true
}

func (this *WebClient) Connect() bool {
	if this.m_nState == SSF_CONNECT {
		return false
	}

	var strRemote string = fmt.Sprintf("%s:%d", this.m_sIP, this.m_nPort)
	wsAddr := url.URL{Scheme: "ws", Host: strRemote, Path: "/ws"}
	conn, _, err := websocket.DefaultDialer.Dial(wsAddr.String(), nil)
	if err != nil {
		return false
	}
	this.m_nState = SSF_CONNECT
	this.SetWsConn(conn)
	this.OnNetConn()
	return true
}

func (this *WebClient) OnDisconnect() {
}

func (this *WebClient) OnNetConn() {
	buff, nLen := message.Encode("CONNECT", nil)
	this.HandlePacket(this.m_ClientId, buff, nLen)
}

func (this *WebClient) OnNetFail(int) {
	this.Stop()
	buff, nLen := message.Encode("DISCONNECT", nil)
	this.HandlePacket(this.m_ClientId, buff, nLen)
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
	fmt.Printf("%s关闭连接", pClient.m_sIP)
	return true
}
