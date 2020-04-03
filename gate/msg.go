package gate

import (
	"github.com/snowyyj001/loumiao/message"
)

type LouMiaoRegisterRpc struct {
	Uid  int
	Name string
}

type LouMiaoHeartBeat struct {
	Uid   int
	Child int
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

func init() {
	message.RegisterPacket(&LouMiaoHeartBeat{})
	message.RegisterPacket(&LouMiaoHandShake{})
	message.RegisterPacket(&LouMiaoRpcMsg{})
	message.RegisterPacket(&LouMiaoRegisterRpc{})
	message.RegisterPacket(&LouMiaoLoginGate{})

}
