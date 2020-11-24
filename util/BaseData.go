package util

type (
	BaseDataRes struct {
		DataMap map[interface{}]interface{}
	}

	IBaseDataRes interface {
		Close()
		Clear()
		Init()
		AddData(int, interface{})
		GetBaseData(int) interface{}
		Read() bool
	}
)

func (self *BaseDataRes) Close() {
	self.Clear()
}

func (self *BaseDataRes) Clear() {
	for i, _ := range self.DataMap {
		delete(self.DataMap, i)
	}
}

func (self *BaseDataRes) AddData(id int, pData interface{}) {
	self.DataMap[id] = pData
}

func (self *BaseDataRes) GetBaseData(id int) interface{} {
	pData, exist := self.DataMap[id]
	if exist {
		return pData
	}
	return nil
}

func (self *BaseDataRes) Init() {
	self.DataMap = make(map[interface{}]interface{})
}

func (self *BaseDataRes) Read() bool {
	return true
}
