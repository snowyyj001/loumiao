package gate

import (
	"github.com/snowyyj001/loumiao/message"
)

type LouMiaoRegisterRpc struct {
	Uid  int
	Name string
}

type LouMiaoHandShake struct {
	Uid int
}

type LouMiaoLoginGate struct {
	TokenId int
	UserId  int
}

type LouMiaoRpcMsg struct {
	ClientId int
	Buffer   []byte
}

type LouMiaoKickOut struct {
}

type LouMiaoClientOffline struct {
	ClientId int
}

func init() {
	message.RegisterPacket(&LouMiaoHandShake{})
	message.RegisterPacket(&LouMiaoRpcMsg{})
	message.RegisterPacket(&LouMiaoRegisterRpc{})
	message.RegisterPacket(&LouMiaoLoginGate{})
	message.RegisterPacket(&LouMiaoClientOffline{})
}
