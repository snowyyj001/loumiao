package maps

import "loumiao/base/containers"

func assertIteratorImplementation() {
	var _ containers.ReverseIteratorWithKey = (*Iterator)(nil)
}

// Iterator holding the iterator's state
type Iterator struct {
	maps     *Map
	node     *Node
	position position
}

type position byte

const (
	begin, between, end position = 0, 1, 2
)

// Iterator returns a stateful iterator whose elements are key/value pairs.
func (self *Map) Iterator() Iterator {
	return Iterator{maps: self, node: nil, position: begin}
}

// Next moves the iterator to the next element and returns true if there was a next element in the container.
// If Next() returns true, then next element's key and value can be retrieved by Key() and Value().
// If Next() was called for the first time, then it will point the iterator to the first element if it exists.
// Modifies the state of the iterator.
func (self *Iterator) Next() bool {
	if self.position == end {
		goto end
	}
	if self.position == begin {
		left := self.maps.Left()
		if left == nil {
			goto end
		}
		self.node = left
		goto between
	}
	if self.node.Right != nil {
		self.node = self.node.Right
		for self.node.Left != nil {
			self.node = self.node.Left
		}
		goto between
	}
	if self.node.Parent != nil {
		node := self.node
		for self.node.Parent != nil {
			self.node = self.node.Parent
			if self.maps.Comparator(node.Key, self.node.Key) <= 0 {
				goto between
			}
		}
	}

end:
	self.node = nil
	self.position = end
	return false

between:
	self.position = between
	return true
}

// Prev moves the iterator to the previous element and returns true if there was a previous element in the container.
// If Prev() returns true, then previous element's key and value can be retrieved by Key() and Value().
// Modifies the state of the iterator.
func (self *Iterator) Prev() bool {
	if self.position == begin {
		goto begin
	}
	if self.position == end {
		right := self.maps.Right()
		if right == nil {
			goto begin
		}
		self.node = right
		goto between
	}
	if self.node.Left != nil {
		self.node = self.node.Left
		for self.node.Right != nil {
			self.node = self.node.Right
		}
		goto between
	}
	if self.node.Parent != nil {
		node := self.node
		for self.node.Parent != nil {
			self.node = self.node.Parent
			if self.maps.Comparator(node.Key, self.node.Key) >= 0 {
				goto between
			}
		}
	}

begin:
	self.node = nil
	self.position = begin
	return false

between:
	self.position = between
	return true
}

// Value returns the current element's value.
// Does not modify the state of the iterator.
func (self *Iterator) Value() interface{} {
	return self.node.Value
}

// Key returns the current element's key.
// Does not modify the state of the iterator.
func (self *Iterator) Key() interface{} {
	return self.node.Key
}

// Begin resets the iterator to its initial state (one-before-first)
// Call Next() to fetch the first element if any.
func (self *Iterator) Begin() {
	self.node = nil
	self.position = begin
}

// End moves the iterator past the last element (one-past-the-end).
// Call Prev() to fetch the last element if any.
func (self *Iterator) End() {
	self.node = nil
	self.position = end
}

// First moves the iterator to the first element and returns true if there was a first element in the container.
// If First() returns true, then first element's key and value can be retrieved by Key() and Value().
// Modifies the state of the iterator
func (self *Iterator) First() bool {
	self.Begin()
	return self.Next()
}

// Last moves the iterator to the last element and returns true if there was a last element in the container.
// If Last() returns true, then last element's key and value can be retrieved by Key() and Value().
// Modifies the state of the iterator.
func (self *Iterator) Last() bool {
	self.End()
	return self.Prev()
}
