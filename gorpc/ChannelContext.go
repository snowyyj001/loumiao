// ChannelContext
package gorpc

//"fmt"

// 声明一个函数类型
type HanlderFunc func(igo IGoRoutine, data interface{}) interface{}
type HanlderNetFunc func(igo IGoRoutine, clientid int, data interface{})

// 声明一个数据类型
type M struct {
	Id   int
	Name string
	Data interface{}
}

// 声明一个数据类型
type MS struct {
	Ids  []int
	Name string
	Data interface{}
}

type ChannelContext struct {
	Handler  string              //处理函数名字
	Data     interface{}         //传送携带数据
	ReadChan chan ChannelContext //读取chan
	Cb       HanlderFunc         //回调
}
