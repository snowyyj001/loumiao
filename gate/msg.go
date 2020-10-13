package gate

import (
	"github.com/snowyyj001/loumiao/message"
)

type LouMiaoLoginGate struct {
	TokenId int
	UserId  int
}

type LouMiaoKickOut struct {
}

type LouMiaoClientOffline struct {
	ClientId int
}

type LouMiaoRpcMsg struct {
	ClientId int
	FuncName string
	Buffer   []byte
}

type LouMiaoNetMsg struct {
	ClientId int
	Buffer   []byte
}

func init() {
	message.RegisterPacket(&LouMiaoLoginGate{})
	message.RegisterPacket(&LouMiaoKickOut{})
	message.RegisterPacket(&LouMiaoClientOffline{})
	message.RegisterPacket(&LouMiaoRpcMsg{})
	message.RegisterPacket(&LouMiaoNetMsg{})
}
