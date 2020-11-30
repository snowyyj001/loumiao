package msg

import (
	"github.com/snowyyj001/loumiao/message"
)

func init() {
	message.RegisterPacket(&LouMiaoLoginGate{})
	message.RegisterPacket(&LouMiaoKickOut{})
	message.RegisterPacket(&LouMiaoClientConnect{})
	message.RegisterPacket(&LouMiaoRpcRegister{})
	message.RegisterPacket(&LouMiaoRpcMsg{})
	message.RegisterPacket(&LouMiaoNetMsg{})
	message.RegisterPacket(&LouMiaoBindGate{})
}
