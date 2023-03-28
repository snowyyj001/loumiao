package lutil

// 数据AB切换可能隐藏的问题：
// 在一个逻辑里，多次调用GetBaseData没有可能拿到不同的数据
// 所以，应该在一个逻辑里使用一次GetBaseData，然后使用返回的值
type (
	BaseDataRes struct {
		flag      bool //A:true, B:false
		dataMap_A map[interface{}]interface{}
		dataMap_B map[interface{}]interface{}
	}

	IBaseDataRes interface {
		Close()
		Clear()
		Init()
		AddData(int, interface{})
		GetBaseData(int) interface{}
		Len() int
		Read() error
		ReadDone()
		Name() string
		Reset()
	}
)

func (self *BaseDataRes) Close() {
	self.Clear()
}

func (self *BaseDataRes) Clear() {
	self.dataMap_A = nil
	self.dataMap_B = nil
}

func (self *BaseDataRes) Reset() {
	if self.flag { //a is using, reset b
		self.dataMap_B = make(map[interface{}]interface{})
	} else {
		self.dataMap_A = make(map[interface{}]interface{})
	}
}

func (self *BaseDataRes) AddData(id int, pData interface{}) {
	if self.flag { //a is using, read to b
		self.dataMap_B[id] = pData
	} else {
		self.dataMap_A[id] = pData
	}
}

func (self *BaseDataRes) GetBaseData(id int) interface{} {
	if self.flag { //a is using, return a
		pData, exist := self.dataMap_A[id]
		if exist {
			return pData
		}
		return nil
	} else {
		pData, exist := self.dataMap_B[id]
		if exist {
			return pData
		}
		return nil
	}
}

func (self *BaseDataRes) Len() int {
	if self.flag { //a is using
		return len(self.dataMap_A)
	} else {
		return len(self.dataMap_B)
	}
}

func (self *BaseDataRes) GetAllData() map[interface{}]interface{} {
	if self.flag { //a is using, return a
		return self.dataMap_A
	} else {
		return self.dataMap_B
	}
}

func (self *BaseDataRes) Init() {
	if self.dataMap_A != nil {
		return
	}
	self.dataMap_A = make(map[interface{}]interface{})
	self.dataMap_B = make(map[interface{}]interface{})
	self.flag = false
}

func (self *BaseDataRes) Read() error {
	return nil
}

func (self *BaseDataRes) ReadDone() {
	self.flag = !self.flag
}
