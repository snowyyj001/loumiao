package vector

import "github.com/snowyyj001/loumiao/base/containers"

func assertIteratorImplementation() {
	var _ containers.ReverseIteratorWithIndex = (*Iterator)(nil)
}

// Iterator holding the iterator's state
type Iterator struct {
	vec   *Vector
	index int
}

// Iterator returns a stateful iterator whose values can be fetched by an index.
func (self *Vector) Iterator() Iterator {
	return Iterator{vec: self, index: -1}
}

// Next moves the iterator to the next element and returns true if there was a next element in the container.
// If Next() returns true, then next element's index and value can be retrieved by Index() and Value().
// If Next() was called for the first time, then it will point the iterator to the first element if it exists.
// Modifies the state of the iterator.
func (self *Iterator) Next() bool {
	if self.index < self.vec.mElementCount {
		self.index++
	}
	return self.vec.withinRange(self.index)
}

// Prev moves the iterator to the previous element and returns true if there was a previous element in the container.
// If Prev() returns true, then previous element's index and value can be retrieved by Index() and Value().
// Modifies the state of the iterator.
func (self *Iterator) Prev() bool {
	if self.index >= 0 {
		self.index--
	}
	return self.vec.withinRange(self.index)
}

// Value returns the current element's value.
// Does not modify the state of the iterator.
func (self *Iterator) Value() interface{} {
	return self.vec.Get(self.index)
}

// Index returns the current element's index.
// Does not modify the state of the iterator.
func (self *Iterator) Index() int {
	return self.index
}

// Begin resets the iterator to its initial state (one-before-first)
// Call Next() to fetch the first element if any.
func (self *Iterator) Begin() {
	self.index = -1
}

// End moves the iterator past the last element (one-past-the-end).
// Call Prev() to fetch the last element if any.
func (self *Iterator) End() {
	self.index = self.vec.mElementCount
}

// First moves the iterator to the first element and returns true if there was a first element in the container.
// If First() returns true, then first element's index and value can be retrieved by Index() and Value().
// Modifies the state of the iterator.
func (self *Iterator) First() bool {
	self.Begin()
	return self.Next()
}

// Last moves the iterator to the last element and returns true if there was a last element in the container.
// If Last() returns true, then last element's index and value can be retrieved by Index() and Value().
// Modifies the state of the iterator.
func (self *Iterator) Last() bool {
	self.End()
	return self.Prev()
}
