// ChannelContext
package gorpc

//"fmt"

// 声明一个函数类型
type HanlderFunc func(igo IGoRoutine, data interface{}) interface{}
type HanlderNetFunc func(igo IGoRoutine, clientid int, data interface{}) interface{}

// 声明一个数据类型
type M map[string]interface{}

type ChannelContext struct {
	Handler  string               //处理函数名字
	Data     interface{}          //传送携带数据
	ReadChan chan *ChannelContext //读取chan
	Cb       HanlderFunc          //回调
}

func SimpleM(b bool) M {
	return M{"ret": b}
}

func SimpleO(data interface{}) M {
	return M{"data": data}
}

func SimpleGO(ct M) interface{} {
	data, ok := ct["data"]
	if ok {
		return data
	} else {
		return nil
	}
}

func SimpleD(name string, data interface{}) M {
	return M{"name": name, "data": data}
}

func SimpleGD(ct M) (string, interface{}) {
	name, _ := ct["name"]
	data, _ := ct["data"]
	return name.(string), data
}

func SimpleGK(ct M, key string) interface{} {
	data, ok := ct[key]
	if ok {
		return data
	} else {
		return nil
	}
}

func SimpleNet(clientid int, name string, data interface{}) M {
	return M{"cid": clientid, "pname": name, "data": data}
}

func SimpleGNet(m M) (int, string, interface{}) {
	name, _ := m["pname"].(string)
	data, _ := m["data"]
	cid, _ := m["cid"].(int)
	return cid, name, data
}
