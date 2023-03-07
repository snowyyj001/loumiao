package base

import (
	"fmt"
	"os"
	"strings"
)

const (
	DATASHEET_BEGIN_ROW = 4
)

// datatype
const (
	DType_none          = ""
	DType_Enum          = "enum"
	DType_String        = "string"
	DType_S32           = "int"
	DType_S64           = "long"
	DType_KVString      = "kvs"
	DType_KVNumber      = "kvn"
	DType_StringArray   = "[string]"
	DType_S32Array      = "[int]"
	DType_S64Array      = "[long]"
	DType_KVStringArray = "[kvs]"
	DType_KVNumberArray = "[kvn]"
)

type (
	KeyValueS struct {
		Key   string
		Value string
	}

	KeyValueN struct {
		Key   int
		Value int
	}

	CDataFile struct {
		RecordNum int //记录数量
		ColumNum  int //列数量
		SheetName string

		fstream            *BitStream
		readstep           int //控制读的总数量
		DataNames          []string
		DataTypes          []string
		currentColumnIndex int
		currentRowIndex    int
	}

	IDateFile interface {
		ReadDataFile(string) bool
		ReadDataInit()
	}
)

type BufferLoader func(file string) ([]byte, error)

func DataBufferLoader(file string) []byte {
	fmt.Println("DataBufferLoader: ", file)
	if bytes, err := os.ReadFile(file); err != nil {
		return nil
	} else {
		return bytes
	}
}
func (self *CDataFile) ReadDataInit() {
	self.ColumNum = 0
	self.RecordNum = 0
	self.readstep = 0
	self.fstream = nil
}

func (self *CDataFile) SetWriterInit(colNum, rowNum int, stream *BitStream) {
	self.ColumNum = colNum
	self.RecordNum = rowNum
	self.readstep = 0
	self.fstream = stream
	self.readstep = self.RecordNum * self.ColumNum
}

/*
excel 格式：
row 0 : 注释
row 1 : 变量名
row 2 : 变量类型
row 3 : 导出标志
*/
func (self *CDataFile) WriteDataFile(sheetName string, dataTypes, dataNames, comments, exports []string) error {
	self.DataNames = dataNames
	self.DataTypes = dataTypes
	self.fstream.WriteInt32(self.RecordNum)
	self.fstream.WriteInt32(self.ColumNum)
	self.fstream.WriteString(sheetName)

	for i := 0; i < self.ColumNum; i++ {
		self.fstream.WriteString(comments[i])
	}
	for i := 0; i < self.ColumNum; i++ {
		self.fstream.WriteString(dataNames[i])
	}
	for i := 0; i < self.ColumNum; i++ {
		self.fstream.WriteString(dataTypes[i])
	}
	for i := 0; i < self.ColumNum; i++ {
		self.fstream.WriteString(exports[i])
	}
	return nil
}

func (self *CDataFile) WriteData(val string) error {
	dataType := self.DataTypes[self.currentColumnIndex]
	switch dataType {
	case DType_String:
		self.fstream.WriteString(val)
	case DType_S32:
		v, err := Int(val)
		if err != nil {
			return err
		}
		self.fstream.WriteInt32(v)
	case DType_Enum:
		v, err := Int(val)
		if err != nil {
			return err
		}
		self.fstream.WriteInt32(v)
	case DType_S64:
		v, err := Int(val)
		if err != nil {
			return err
		}
		self.fstream.WriteInt64(int64(v))
	case DType_KVString:
		arr := strings.Split(val, ":")
		if len(arr) != 2 {
			return fmt.Errorf("WriteData: col = %d, row = %d, %s is not a format (*:*)", self.currentColumnIndex, self.currentRowIndex, val)
		}
		self.fstream.WriteString(val)
	case DType_KVNumber:
		arr := strings.Split(val, ":")
		if len(arr) != 2 {
			return fmt.Errorf("WriteData: col = %d, row = %d, %s is not a format (*:*)", self.currentColumnIndex, self.currentRowIndex, val)
		}
		_, err := Int(arr[0])
		if err != nil {
			return fmt.Errorf("WriteData: col = %d, row = %d, %s", self.currentColumnIndex, self.currentRowIndex, err.Error())
		}
		_, err = Int(arr[1])
		if err != nil {
			return fmt.Errorf("WriteData: col = %d, row = %d, %s", self.currentColumnIndex, self.currentRowIndex, err.Error())
		}
		self.fstream.WriteString(val)
	case DType_StringArray:
		self.fstream.WriteString(val)
	case DType_S32Array:
		arr := strings.Split(val, ",")
		sz := len(arr)
		for i := 0; i < sz; i++ {
			_, err := Int(arr[i])
			if err != nil {
				return fmt.Errorf("WriteData: col = %d, row = %d, %s", self.currentColumnIndex, self.currentRowIndex, err.Error())
			}
		}
		self.fstream.WriteString(val)
	case DType_S64Array:
		arr := strings.Split(val, ",")
		sz := len(arr)
		for i := 0; i < sz; i++ {
			_, err := Int(arr[i])
			if err != nil {
				return fmt.Errorf("WriteData: col = %d, row = %d, %s", self.currentColumnIndex, self.currentRowIndex, err.Error())
			}
		}
		self.fstream.WriteString(val)
	case DType_KVStringArray:
		arr := strings.Split(val, ",")
		sz := len(arr)
		for i := 0; i < sz; i++ {
			narr := strings.Split(arr[i], ":")
			if len(narr) != 2 {
				return fmt.Errorf("WriteData: col = %d, row = %d, err = %s is not a format (*;*), %s", self.currentColumnIndex, self.currentRowIndex, val, arr[i])
			}
		}
		self.fstream.WriteString(val)
	case DType_KVNumberArray:
		arr := strings.Split(val, ",")
		sz := len(arr)

		for i := 0; i < sz; i++ {
			narr := strings.Split(arr[i], ":")
			if len(narr) != 2 {
				return fmt.Errorf("WriteData: col = %d, row = %d, %s is not a format (*:*), %s", self.currentColumnIndex, self.currentRowIndex, val, arr[i])
			}
			_, err := Int(narr[0])
			if err != nil {
				return fmt.Errorf("WriteData: col = %d, row = %d, %s", self.currentColumnIndex, self.currentRowIndex, err.Error())
			}
			_, err = Int(narr[1])
			if err != nil {
				return fmt.Errorf("WriteData: col = %d, row = %d, %s", self.currentColumnIndex, self.currentRowIndex, err.Error())
			}
		}
		self.fstream.WriteString(val)
	default:
		return fmt.Errorf("WriteData: col = %d, row = %d, data type[%d]", self.currentColumnIndex, self.currentRowIndex, dataType)
	}
	self.currentColumnIndex++
	if self.currentColumnIndex == self.ColumNum {
		self.currentRowIndex++
		self.currentColumnIndex = 0
	}
	self.readstep--
	return nil
}

func (self *CDataFile) GetData() (interface{}, error) {
	dataType := self.DataTypes[self.currentColumnIndex]

	self.currentColumnIndex++
	if self.currentColumnIndex == self.ColumNum {
		self.currentRowIndex++
		self.currentColumnIndex = 0
	}
	self.readstep--

	switch dataType {
	case DType_String:
		return self.fstream.ReadString(), nil
	case DType_S32:
		return self.fstream.ReadInt32(), nil
	case DType_Enum:
		return self.fstream.ReadInt32(), nil
	case DType_S64:
		return self.fstream.ReadInt64(), nil
	case DType_KVString:
		str := self.fstream.ReadString()
		arr := strings.Split(str, ":")
		if len(arr) != 2 {
			return nil, fmt.Errorf("GetData: col = %d, row = %d, err = %s is not a format (*:*)", self.currentColumnIndex, self.currentRowIndex, str)
		}
		kv := KeyValueS{arr[0], arr[1]}
		return kv, nil
	case DType_KVNumber:
		str := self.fstream.ReadString()
		arr := strings.Split(str, ":")
		if len(arr) != 2 {
			return nil, fmt.Errorf("GetData: col = %d, row = %d, err = %s is not a format (*:*)", self.currentColumnIndex, self.currentRowIndex, str)
		}
		v1, err := Int(arr[0])
		if err != nil {
			return nil, fmt.Errorf("GetData: col = %d, row = %d, err = %s", self.currentColumnIndex, self.currentRowIndex, err.Error())
		}
		v2, err := Int(arr[1])
		if err != nil {
			return nil, fmt.Errorf("GetData: col = %d, row = %d, err = %s", self.currentColumnIndex, self.currentRowIndex, err.Error())
		}
		kv := KeyValueN{v1, v2}
		return kv, nil
	case DType_StringArray:
		str := self.fstream.ReadString()
		arr := strings.Split(str, ",")
		return arr, nil
	case DType_S32Array:
		str := self.fstream.ReadString()
		arr := strings.Split(str, ",")
		sz := len(arr)
		vals := make([]int, sz, sz)
		for i := 0; i < sz; i++ {
			v, err := Int(arr[i])
			if err != nil {
				return nil, fmt.Errorf("GetData: col = %d, row = %d, err = %s", self.currentColumnIndex, self.currentRowIndex, err.Error())
			}
			vals[i] = v
		}
		return vals, nil
	case DType_S64Array:
		str := self.fstream.ReadString()
		arr := strings.Split(str, ",")
		sz := len(arr)
		vals := make([]int64, sz, sz)
		for i := 0; i < sz; i++ {
			v, err := Int(arr[i])
			if err != nil {
				return nil, fmt.Errorf("GetData: col = %d, row = %d, err = %s", self.currentColumnIndex, self.currentRowIndex, err.Error())
			}
			vals[i] = int64(v)
		}
		return vals, nil
	case DType_KVStringArray:
		str := self.fstream.ReadString()
		arr := strings.Split(str, ",")
		sz := len(arr)
		var r []*KeyValueS
		for i := 0; i < sz; i++ {
			narr := strings.Split(arr[i], ":")
			if len(narr) != 2 {
				return nil, fmt.Errorf("GetData: col = %d, row = %d, err = %s is not a format (*;*), %s", self.currentColumnIndex, self.currentRowIndex, str, arr[i])
			}
			r = append(r, &KeyValueS{narr[0], narr[1]})
		}
		return r, nil
	case DType_KVNumberArray:
		str := self.fstream.ReadString()
		arr := strings.Split(str, ",")
		sz := len(arr)
		var r []*KeyValueN
		for i := 0; i < sz; i++ {
			narr := strings.Split(arr[i], ":")
			if len(narr) != 2 {
				return nil, fmt.Errorf("GetData: col = %d, row = %d, err = %s is not a format (*;*), %s", self.currentColumnIndex, self.currentRowIndex, str, arr[i])
			}
			v1, err := Int(narr[0])
			if err != nil {
				return nil, fmt.Errorf("GetData: col = %d, row = %d, err = %s", self.currentColumnIndex, self.currentRowIndex, err.Error())
			}
			v2, err := Int(narr[1])
			if err != nil {
				return nil, fmt.Errorf("GetData: col = %d, row = %d, err = %s", self.currentColumnIndex, self.currentRowIndex, err.Error())
			}
			r = append(r, &KeyValueN{v1, v2})
		}
		return r, nil
	}
	return nil, fmt.Errorf("GetData: col = %d, row = %d, err = data type[%d]", self.currentColumnIndex, self.currentRowIndex, dataType)
}

func (self *CDataFile) ReadDataHead(buff []byte) {
	self.fstream = NewBitStreamR(buff)
	self.RecordNum = self.fstream.ReadInt32()
	self.ColumNum = self.fstream.ReadInt32()
	self.SheetName = self.fstream.ReadString()

	for i := 0; i < self.ColumNum; i++ {
		self.fstream.ReadString()
	}
	for i := 0; i < self.ColumNum; i++ {
		self.DataNames = append(self.DataNames, self.fstream.ReadString())
	}
	for i := 0; i < self.ColumNum; i++ {
		self.DataTypes = append(self.DataTypes, self.fstream.ReadString())
	}
	for i := 0; i < self.ColumNum; i++ {
		self.fstream.ReadString()
	}
}

func (self *CDataFile) ReadDataHeadCol(buff []byte) {
	self.fstream = NewBitStreamR(buff)
	self.RecordNum = self.fstream.ReadInt32()
	self.ColumNum = self.fstream.ReadInt32()
	self.SheetName = self.fstream.ReadString()

	for i := 0; i < self.ColumNum; i++ {
		self.fstream.ReadString()
	}
	for i := 0; i < self.ColumNum; i++ {
		self.DataNames = append(self.DataNames, self.fstream.ReadString())
	}
	for i := 0; i < self.ColumNum; i++ {
		self.DataTypes = append(self.DataTypes, self.fstream.ReadString())
	}
	for i := 0; i < self.ColumNum; i++ {
		self.fstream.ReadString()
	}
}
