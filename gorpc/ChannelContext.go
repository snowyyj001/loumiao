// ChannelContext
package gorpc

//"fmt"

// 声明一个rpc函数类型，
type HandlerRpcFunc func([]byte) []byte

// 声明一个函数类型，
type HandlerFunc func(igo IGoRoutine, data interface{}) interface{}

// data是一个nil或指针类型，需要
type HandlerNetFunc func(igo IGoRoutine, clientId int, buffer []byte)

// 声明一个数据类型
type M struct {
	Id    int
	Name  string
	Param int
	Data  interface{}
	Flag  bool //true, 标记M类型的有效数据是Data，否则是自己本身
}

// 声明一个数据类型
type MA struct {
	Id    int
	Param int
	Data  interface{}
}

// 声明一个数据类型
type MS struct {
	Ids  []int
	Data interface{}
}

// 声明一个数据类型
type MI struct {
	Id   int
	Data interface{}
}

// 声明一个数据类型
type MM struct {
	Id   string
	Data interface{}
}

type ChannelContext struct {
	Handler  string              //处理函数名字
	Data     M                   //传送携带数据(如果使用Data interface{}，Data会escapes to heap)
	ReadChan chan ChannelContext //读取chan
	Cb       HandlerFunc         //回调
}
