package message

import (
	"encoding/binary"
	_ "fmt"
	"reflect"
	"sync"

	"github.com/snowyyj001/loumiao/config"
)

const (
	HEAD_SIZE    = 8  //包头大小
	MSGNAME_SIZE = 36 //消息名最大长度

	Flag_RPC int = 1 << 0
)

type MsgPool struct {
	name  string
	mtype reflect.Type
	cache sync.Pool
}

type ClassNewHandler func() interface{}

//消息结构创建不缓存了，经测试反射性能也还可以，见reflect_test.go
//开放了decode接口给逻辑层，已经不需要decode了，逻辑层根据消息直接创建对应的结构体
var (
	//Packet_CreateFactorStringMap map[string]*MsgPool
	Packet_CreateFactorStringMap map[string]ClassNewHandler
	filterWarning                map[string]bool
	MaxPacketSize                int //一个消息包的最大大小,如果一个消息超过该阀值，那么就需要分包
)

func init() {
	//Packet_CreateFactorStringMap = make(map[string]func() interface{})
	Packet_CreateFactorStringMap = make(map[string]ClassNewHandler)
	filterWarning = make(map[string]bool)
	filterWarning["CONNECT"] = true
	filterWarning["DISCONNECT"] = true
	filterWarning["C_CONNECT"] = true
	filterWarning["C_DISCONNECT"] = true
	MaxPacketSize = config.NET_BUFFER_SIZE
}

//注册消息
func RegisterPacket(packet interface{}) {
	packetName := GetMessageName(packet)
	//fmt.Println("RegisterPacket", packetName)
	pt := reflect.TypeOf(packet).Elem()
	packetFunc := func() interface{} {
		packet = reflect.New(pt).Interface()
		return packet
	}
	Packet_CreateFactorStringMap[packetName] = packetFunc
	//fmt.Println("RegisterPacket: " + packetName)
	/*mpool := &MsgPool{name: packetName, mtype: pt}
	mpool.cache.New = func() interface{} {
		return reflect.New(mpool.mtype).Interface()
	}
	Packet_CreateFactorStringMap[packetName] = mpool*/
}

func GetMessageName(packet interface{}) string {
	typeOfStruct := reflect.TypeOf(packet)
	elem := typeOfStruct.Elem()
	return elem.Name()
}

func GetPakcet(name string) interface{} {
	packetFunc, exist := Packet_CreateFactorStringMap[name]
	if exist {
		//return packetFunc.cache.Get()
		return packetFunc()
	}
	return nil
}

/*
func PutPakcet(name string, data interface{}) {
	packetFunc, exist := Packet_CreateFactorStringMap[name]
	if exist {
		packetFunc.cache.Put(data)
	}
}
*/

//替换消息包的target字段(5,6字节)
func ReplacePakcetTarget(target int32, buff []byte) {
	tmp := uint16(target)
	binary.BigEndian.PutUint16(buff[4:], tmp)
}
