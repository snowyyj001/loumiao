package base

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/snowyyj001/loumiao/base/vector"
)

//datatype
const (
	DType_none        = iota
	DType_String      = iota
	DType_Enum        = iota
	DType_S8          = iota
	DType_S16         = iota
	DType_S32         = iota
	DType_F32         = iota
	DType_F64         = iota
	DType_S64         = iota
	DType_StringArray = iota
	DType_S8Array     = iota
	DType_S16Array    = iota
	DType_S32Array    = iota
	DType_F32Array    = iota
	DType_F64Array    = iota
	DType_S64Array    = iota
)

type (
	RData struct {
		m_Type int

		m_String      string
		m_Enum        int
		m_S8          int8
		m_S16         int16
		m_S32         int
		m_F32         float32
		m_F64         float64
		m_S64         int64
		m_StringArray []string
		m_S8Array     []int8
		m_S16Array    []int16
		m_S32Array    []int
		m_F32Array    []float32
		m_F64Array    []float64
		m_S64Array    []int64
	}

	CDataFile struct {
		RecordNum int //记录数量
		ColumNum  int //列数量

		fstream            *BitStream
		readstep           int //控制读的总数量
		dataTypes          vector.Vector
		currentColumnIndex int
	}

	IDateFile interface {
		ReadDataFile(string) bool
		GetData(*RData) bool
		ReadDataInit()
	}
)

func (self *CDataFile) ReadDataInit() {
	self.ColumNum = 0
	self.RecordNum = 0
	self.readstep = 0
	self.fstream = nil
}

func (self *CDataFile) ReadDataFile(fileName string) bool {
	self.dataTypes.Clear()
	self.currentColumnIndex = 0

	file, err := os.Open(fileName)
	if err != nil {
		fmt.Printf("[%s] open failed", fileName)
		return false
	}
	defer file.Close()
	buf, err := ioutil.ReadAll(file)
	if err != nil {
		return false
	}
	self.fstream = NewBitStream(buf, len(buf))

	for {
		tchr := self.fstream.ReadInt(8)
		if tchr == '@' { //找到数据文件的开头
			tchr = self.fstream.ReadInt(8) //这个是换行字符
			//fmt.Println(tchr)
			break
		}
	}
	//得到记录总数
	self.RecordNum = self.fstream.ReadInt(32)
	//得到列的总数
	self.ColumNum = self.fstream.ReadInt(32)
	//sheet name
	self.fstream.ReadString()

	self.readstep = self.RecordNum * self.ColumNum
	for nColumnIndex := 0; nColumnIndex < self.ColumNum; nColumnIndex++ {
		//col name
		//self.fstream.ReadString()
		//nDataType := self.fstream.ReadInt(8)
		nDataType := 1
		self.dataTypes.PushBack(int(nDataType))
	}
	return true
}

/****************************
	格式:
	头文件:
	1、总记录数(int)
	2、总字段数(int)
	字段格式:
	1、字段长度(int)
	2、字读数据类型(int->2,string->1,enum->3,float->4)
	3、字段内容(int,string)
*************************/
func (self *CDataFile) GetData(pData *RData) bool {
	if self.readstep == 0 || self.fstream == nil {
		return false
	}

	switch self.dataTypes.Get(self.currentColumnIndex).(int) {
	case DType_String:
		pData.m_String = self.fstream.ReadString()
	case DType_S8:
		pData.m_S8 = int8(self.fstream.ReadInt(8))
	case DType_S16:
		pData.m_S16 = int16(self.fstream.ReadInt(16))
	case DType_S32:
		pData.m_S32 = self.fstream.ReadInt(32)
	case DType_Enum:
		pData.m_Enum = self.fstream.ReadInt(16)
	case DType_F32:
		pData.m_F32 = self.fstream.ReadFloat()
	case DType_F64:
		pData.m_F64 = self.fstream.ReadFloat64()
	case DType_S64:
		pData.m_S64 = self.fstream.ReadInt64()

	case DType_StringArray:
		nLen := self.fstream.ReadInt(8)
		pData.m_StringArray = make([]string, nLen)
		for i := 0; i < nLen; i++ {
			pData.m_StringArray[i] = self.fstream.ReadString()
		}
	case DType_S8Array:
		nLen := self.fstream.ReadInt(8)
		pData.m_S8Array = make([]int8, nLen)
		for i := 0; i < nLen; i++ {
			pData.m_S8Array[i] = int8(self.fstream.ReadInt(8))
		}
	case DType_S16Array:
		nLen := self.fstream.ReadInt(8)
		pData.m_S16Array = make([]int16, nLen)
		for i := 0; i < nLen; i++ {
			pData.m_S16Array[i] = int16(self.fstream.ReadInt(16))
		}
	case DType_S32Array:
		nLen := self.fstream.ReadInt(8)
		pData.m_S32Array = make([]int, nLen)
		for i := 0; i < nLen; i++ {
			pData.m_S32Array[i] = self.fstream.ReadInt(32)
		}
	case DType_F32Array:
		nLen := self.fstream.ReadInt(8)
		pData.m_F32Array = make([]float32, nLen)
		for i := 0; i < nLen; i++ {
			pData.m_F32Array[i] = self.fstream.ReadFloat()
		}
	case DType_F64Array:
		nLen := self.fstream.ReadInt(8)
		pData.m_F64Array = make([]float64, nLen)
		for i := 0; i < nLen; i++ {
			pData.m_F64Array[i] = self.fstream.ReadFloat64()
		}
	case DType_S64Array:
		nLen := self.fstream.ReadInt(8)
		pData.m_S64Array = make([]int64, nLen)
		for i := 0; i < nLen; i++ {
			pData.m_S64Array[i] = self.fstream.ReadInt64()
		}
	}

	pData.m_Type = self.dataTypes.Get(self.currentColumnIndex).(int)
	self.currentColumnIndex = (self.currentColumnIndex + 1) % self.ColumNum
	self.readstep--

	return true
}

/****************************
	RData funciton
****************************/
func (self *RData) String(dataname, datacol string) string {
	Assert(self.m_Type == DType_String, fmt.Sprintf("read [%s] col[%s] type[%d]error", dataname, datacol, self.m_Type))
	return self.m_String
}

func (self *RData) Enum(dataname, datacol string) int {
	Assert(self.m_Type == DType_Enum, fmt.Sprintf("read [%s] col[%s] error", dataname, datacol))
	return self.m_Enum
}

func (self *RData) Int8(dataname, datacol string) int8 {
	Assert(self.m_Type == DType_S8, fmt.Sprintf("read [%s] col[%s] error", dataname, datacol))
	return self.m_S8
}

func (self *RData) Int16(dataname, datacol string) int16 {
	Assert(self.m_Type == DType_S16, fmt.Sprintf("read [%s] col[%s] error", dataname, datacol))
	return self.m_S16
}

func (self *RData) Int(dataname, datacol string) int {
	Assert(self.m_Type == DType_S32, fmt.Sprintf("read [%s] col[%s] error", dataname, datacol))
	return self.m_S32
}

func (self *RData) Float32(dataname, datacol string) float32 {
	Assert(self.m_Type == DType_F32, fmt.Sprintf("read [%s] col[%s] error", dataname, datacol))
	return self.m_F32
}

func (self *RData) Float64(dataname, datacol string) float64 {
	Assert(self.m_Type == DType_F64, fmt.Sprintf("read [%s] col[%s] error", dataname, datacol))
	return self.m_F64
}

func (self *RData) Int64(dataname, datacol string) int64 {
	Assert(self.m_Type == DType_S64, fmt.Sprintf("read [%s] col[%s] error", dataname, datacol))
	return self.m_S64
}

func (self *RData) StringArray(dataname, datacol string) []string {
	Assert(self.m_Type == DType_StringArray, fmt.Sprintf("read [%s] col[%s] error", dataname, datacol))
	return self.m_StringArray
}

func (self *RData) Int8Array(dataname, datacol string) []int8 {
	Assert(self.m_Type == DType_S8Array, fmt.Sprintf("read [%s] col[%s] error", dataname, datacol))
	return self.m_S8Array
}

func (self *RData) Int16Array(dataname, datacol string) []int16 {
	Assert(self.m_Type == DType_S16Array, fmt.Sprintf("read [%s] col[%s] error", dataname, datacol))
	return self.m_S16Array
}

func (self *RData) IntArray(dataname, datacol string) []int {
	Assert(self.m_Type == DType_S32Array, fmt.Sprintf("read [%s] col[%s] error", dataname, datacol))
	return self.m_S32Array
}

func (self *RData) Float32Array(dataname, datacol string) []float32 {
	Assert(self.m_Type == DType_F32Array, fmt.Sprintf("read [%s] col[%s] error", dataname, datacol))
	return self.m_F32Array
}

func (self *RData) Float64Array(dataname, datacol string) []float64 {
	Assert(self.m_Type == DType_F64Array, fmt.Sprintf("read [%s] col[%s] error", dataname, datacol))
	return self.m_F64Array
}

func (self *RData) Int64Array(dataname, datacol string) []int64 {
	Assert(self.m_Type == DType_S64Array, fmt.Sprintf("read [%s] col[%s] error", dataname, datacol))
	return self.m_S64Array
}
