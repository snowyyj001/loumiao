package lagent

import (
	"github.com/snowyyj001/loumiao/ldefine"
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
	self.NetState = ldefine.CLIENT_DISCONNECT
}

func (self *LouMiaoAgent) SetOnLine() {
	self.NetState = ldefine.CLIENT_CONNECT
}

func (self *LouMiaoAgent) IsOnLine() bool {
	return self.NetState == ldefine.CLIENT_CONNECT
}

func (self *LouMiaoAgent) OnDisConnect() {
	self.NetState = ldefine.CLIENT_DISCONNECT
}
