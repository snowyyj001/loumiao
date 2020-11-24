package base

import (
	"math"
	"unsafe"
)

const(
	size_int = int(unsafe.Sizeof(int(0))) * 8
)

type(
	BitMap struct {
		m_Bits []int
		m_Size int
	}

	IBitMap interface {
		Init(size int)
		Set(index int)//设置位
		Test(index int) bool//位是否被设置
		Clear(index int)//清楚位
		ClearAll()
	}
)

func (self *BitMap) Init(size int){
	self.m_Size = int(math.Ceil(float64(size) / float64(size_int) ))
	self.m_Bits = make([]int, self.m_Size)
}

func (self *BitMap) Set(index int){
	if index >= self.m_Size * size_int{
		return
	}

	self.m_Bits[index / size_int] |= 1 << uint(index % size_int)
}

func (self *BitMap) Test(index int) bool{
	if index >= self.m_Size * size_int{
		return false
	}

	return self.m_Bits[index / size_int] & (1 << uint(index % size_int)) != 0

}

func (self *BitMap) Clear(index int){
	if index >= self.m_Size * size_int{
		return
	}

	self.m_Bits[index / size_int] &= ^(1 << uint(index % size_int))
}

func (self *BitMap) ClearAll(){
	self.Init(self.m_Size * size_int)
}

func NewBitMap(size int) *BitMap{
	bitmap := &BitMap{}
	bitmap.Init(size)
	return bitmap
}