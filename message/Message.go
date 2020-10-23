package message

import (
	_ "fmt"
	"reflect"
)

const (
	Flag_RPC int = 1 << 0
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
	pt := reflect.TypeOf(packet).Elem()
	packetFunc := func() interface{} {
		packet := reflect.New(pt).Interface()
		return packet
	}
	//	fmt.Println("RegisterPacket: " + packetName)
	Packet_CreateFactorStringMap[packetName] = packetFunc
}

func GetMessageName(packet interface{}) string {
	typeOfStruct := reflect.TypeOf(packet)
	elem := typeOfStruct.Elem()
	return elem.Name()
}

func GetPakcet(name string) interface{} {
	packetFunc, exist := Packet_CreateFactorStringMap[name]
	if exist {
		return packetFunc()
	}
	return nil
}
