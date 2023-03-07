package agent

import (
	"github.com/snowyyj001/loumiao/define"
)

const (
	CHAN_BUFFER_LEN = 1000 //channel缓冲数量
)

type LouMiaoAgent struct {
	UserId   int64
	GateUid  int
	LobbyUid int
	ZoneUid  int
	MatchUid int
	SocketId int
	NetState int //1:连接，2:断开

	WorldToken string
	ZoneToken  string
}

func (self *LouMiaoAgent) SetState(state int) {
	self.NetState = state
}

func (self *LouMiaoAgent) SetOffLine() {
	self.NetState = define.CLIENT_DISCONNECT
}

func (self *LouMiaoAgent) SetOnLine() {
	self.NetState = define.CLIENT_CONNECT
}

func (self *LouMiaoAgent) IsOnLine() bool {
	return self.NetState == define.CLIENT_CONNECT
}

func (self *LouMiaoAgent) OnDisConnect() {
	self.NetState = define.CLIENT_DISCONNECT
}
