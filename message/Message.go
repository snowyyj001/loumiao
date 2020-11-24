package message

import (
	_ "fmt"
	"reflect"
	"sync"
)

const (
	Flag_RPC int = 1 << 0
)

type MsgPool struct {
	name  string
	mtype reflect.Type
	cache sync.Pool
}

var (
	Packet_CreateFactorStringMap map[string]*MsgPool
)

func init() {
	//Packet_CreateFactorStringMap = make(map[string]func() interface{})
	Packet_CreateFactorStringMap = make(map[string]*MsgPool)
}

//注册网络消息
func RegisterPacket(packet interface{}) {
	packetName := GetMessageName(packet)
	pt := reflect.TypeOf(packet).Elem()
	/*	packetFunc := func() interface{} {
		packet := reflect.New(pt).Interface()
		return packet
	}*/
	//fmt.Println("RegisterPacket: " + packetName)
	mpool := &MsgPool{name: packetName, mtype: pt}
	mpool.cache.New = func() interface{} {
		return reflect.New(mpool.mtype).Interface()
	}
	Packet_CreateFactorStringMap[packetName] = mpool
}

func GetMessageName(packet interface{}) string {
	typeOfStruct := reflect.TypeOf(packet)
	elem := typeOfStruct.Elem()
	return elem.Name()
}

/*
func GetPakcet(name string) interface{} {
	packetFunc, exist := Packet_CreateFactorStringMap[name]
	if exist {
		return packetFunc()
	}
	return nil
}
*/

func GetPakcet(name string) interface{} {
	packetFunc, exist := Packet_CreateFactorStringMap[name]
	if exist {
		return packetFunc.cache.Get()
	}
	return nil
}
