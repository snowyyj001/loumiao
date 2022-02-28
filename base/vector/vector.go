package vector

import (
	"log"

	"github.com/snowyyj001/loumiao/base/containers"
)

type PVector *Vector

func assert(x bool, y string) {
	if bool(x) == false {
		log.Fatalf("Assert: %s", y)
	}
}

const (
	VectorBlockSize = 16
)

type (
	Vector struct {
		mElementCount int
		mArraySize    int
		mArray        []interface{}
	}

	IVector interface {
		containers.Container
		insert(int)
		increment()
		decrement()

		Erase(int)
		PushFront(interface{})
		PushBack(interface{})
		PopFront()
		PopBack()
		Front() interface{}
		Back() interface{}
		Len() int
		Get(int) interface{}
		Swap(i, j int)
		Less(i, j int) bool
	}
)

func (self *Vector) insert(index int) {
	//assert(index <= self.mElementCount, "Vector<T>::insert - out of bounds index.")

	if self.mElementCount == self.mArraySize {
		self.resize(self.mElementCount + 1)
	} else {
		self.mElementCount++
	}

	for i := self.mElementCount - 1; i > index; i-- {
		self.mArray[i] = self.mArray[i-1]
	}
}

func (self *Vector) increment() {
	if self.mElementCount == self.mArraySize {
		self.resize(self.mElementCount + 1)
	} else {
		self.mElementCount++
	}
}

func (self *Vector) decrement() {
	//assert(self.mElementCount != 0, "Vector<T>::decrement - cannot decrement zero-length vector.")
	self.mElementCount--
}

func (self *Vector) resize(newCount int) bool {
	if newCount > 0 {
		blocks := newCount / VectorBlockSize
		if newCount%VectorBlockSize != 0 {
			blocks++
		}

		self.mElementCount = newCount
		self.mArraySize = blocks * VectorBlockSize
		self.mArray = append(self.mArray, make([]interface{}, VectorBlockSize)...)
	}
	return true
}

func (self *Vector) Erase(index int) {
	//assert(index < self.mElementCount, "Vector<T>::erase - out of bounds index.")
	if index < self.mElementCount-1 {
		for i := index; i < self.mElementCount-1; i++ {
			self.mArray[i] = self.mArray[i+1]
		}
	}

	self.mElementCount--
}

func (self *Vector) PushFront(value interface{}) {
	self.insert(0)
	self.mArray[0] = value
}

func (self *Vector) PushBack(value interface{}) {
	self.increment()
	self.mArray[self.mElementCount-1] = value
}

func (self *Vector) PopFront() {
	//assert(self.mElementCount != 0, "Vector<T>::pop_front - cannot pop the front of a zero-length vector.")
	self.Erase(0)
}

func (self *Vector) PopBack() {
	//assert(self.mElementCount != 0, "Vector<T>::pop_back - cannot pop the back of a zero-length vector.")
	self.decrement()
}

// Check that the index is within bounds of the list
func (self *Vector) withinRange(index int) bool {
	return index >= 0 && index < self.mElementCount
}

func (self *Vector) Front() interface{} {
	//assert(self.mElementCount != 0, "Vector<T>::first - Error, no first element of a zero sized array! (const)")
	return self.mArray[0]
}

func (self *Vector) Back() interface{} {
	//assert(self.mElementCount != 0, "Vector<T>::last - Error, no last element of a zero sized array! (const)")
	return self.mArray[self.mElementCount-1]
}

func (self *Vector) Empty() bool {
	return self.mElementCount == 0
}

func (self *Vector) Size() int {
	return self.mArraySize
}

func (self *Vector) Clear() {
	self.mElementCount = 0
}

func (self *Vector) Release(rsz int) {
	if self.mElementCount < rsz {
		return
	}
	self.mElementCount = self.mElementCount - rsz
	blocks := self.mElementCount / VectorBlockSize
	if self.mElementCount%VectorBlockSize != 0 {
		blocks++
	}
	sz := (blocks + 1) * VectorBlockSize
	if self.mArraySize <= sz {
		return
	}
	for i := self.mArraySize - 1; i >= sz; i-- {
		self.mArray[i] = nil
	}
	self.mArraySize = sz
}

func (self *Vector) Len() int {
	return self.mElementCount
}

func (self *Vector) Get(index int) interface{} {
	//assert(index < self.mElementCount, "Vector<T>::operator[] - out of bounds array access!")
	return self.mArray[index]
}

func (self *Vector) Values() []interface{} {
	return self.mArray[0:self.mElementCount]
}

func (self *Vector) Swap(i, j int) {
	self.mArray[i], self.mArray[j] = self.mArray[j], self.mArray[i]
}

func (self *Vector) Less(i, j int) bool {
	return true
}

func NewVector() *Vector {
	return &Vector{}
}
