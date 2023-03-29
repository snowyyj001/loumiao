package message

import (
	_ "fmt"
	"github.com/snowyyj001/loumiao/lconfig"
	"github.com/snowyyj001/loumiao/lutil"
	"reflect"
	"sync"
)

const (
	HEAD_SIZE    = 8  //包头大小
	MSGNAME_SIZE = 36 //消息名最大长度

	Flag_RPC int = 1 << 0
)

var (
	ENCRYPTION_KEY uint32 = 123 //消息报加密key
)

type MsgPool struct {
	name  string
	mtype reflect.Type
	cache sync.Pool
}

type ClassNewHandler func() interface{}

// 消息结构创建不缓存了，经测试反射性能也还可以，见reflect_test.go
// 开放了decode接口给逻辑层，已经不需要decode了，逻辑层根据消息直接创建对应的结构体
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
	MaxPacketSize = lconfig.NET_BUFFER_SIZE
}

// 注册消息
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
func PutPacket(name string, data interface{}) {
	packetFunc, exist := Packet_CreateFactorStringMap[name]
	if exist {
		packetFunc.cache.Put(data)
	}
}
*/

func EnableEncryptDecrypt() {
	ENCRYPTION_KEY = uint32(lutil.Random(10000))
}

func EncryptBuffer(buff []byte, sz int) {
	if ENCRYPTION_KEY == 0 {
		return
	}
	buff[5] = 1
	for i := 6; i < sz; i++ {
		buff[i] = byte(uint32(buff[i]) ^ ENCRYPTION_KEY)
	}
}

func DecryptBuffer(buff []byte, sz int) {
	if ENCRYPTION_KEY == 0 {
		return
	}
	if buff[5] == 0 {
		return
	}
	for i := 6; i < sz; i++ {
		buff[i] = byte(uint32(buff[i]) ^ ENCRYPTION_KEY)
	}
}
