package message

import (
	"reflect"
)

var (
	Packet_CreateFactorStringMap map[string]func() interface{}
)

func init() {
	Packet_CreateFactorStringMap = make(map[string]func() interface{})
}

//注册网络消息
func RegisterPacket(packet interface{}) {
	packetName := GetMessageName(packet)
	packetFunc := func() interface{} {
		packet := reflect.New(reflect.ValueOf(packet).Elem().Type()).Interface()
		return packet
	}

	Packet_CreateFactorStringMap[packetName] = packetFunc
}

func GetMessageName(packet interface{}) string {
	typeOfStruct := reflect.TypeOf(packet)
	typeOfStruct = typeOfStruct.Elem()
	return typeOfStruct.Name()
}

func GetPakcet(name string) interface{} {
	packetFunc, exist := Packet_CreateFactorStringMap[name]
	if exist {
		return packetFunc()
	}

	return nil
}
