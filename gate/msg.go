package gate

import (
	"github.com/snowyyj001/loumiao/message"
)

type LouMiaoHeartBeat struct {
	Uid   int
	Child int
}

type LouMiaoHandShake struct {
	Uid int
}

type LouMiaoRpcMsg struct {
	ClientId int
	Buffer   []byte
}

func init() {
	message.RegisterPacket(&LouMiaoHeartBeat{})
	message.RegisterPacket(&LouMiaoHandShake{})
	message.RegisterPacket(&LouMiaoRpcMsg{})
}
