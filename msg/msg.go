package msg

import (
	"github.com/snowyyj001/loumiao/message"
)

func init() {
	message.RegisterPacket(&LouMiaoLoginGate{})
	message.RegisterPacket(&LouMiaoHeartBeat{})
	message.RegisterPacket(&LouMiaoKickOut{})
	message.RegisterPacket(&LouMiaoClientConnect{})
	message.RegisterPacket(&LouMiaoClientDisConnect{})
	message.RegisterPacket(&LouMiaoRpcRegister{})
	message.RegisterPacket(&LouMiaoRpcMsg{})
	message.RegisterPacket(&LouMiaoNetMsg{})
	message.RegisterPacket(&LouMiaoBindGate{})
}
