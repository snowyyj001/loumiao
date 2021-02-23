package maps

import (
	"fmt"

	"github.com/snowyyj001/loumiao/base/containers"
)

// Map interface that all Maps implement
type IMap interface {
	containers.Container
	// Empty() bool
	// Size() int
	// Clear()
	// Values() []interface{}
}

func assertMapImplementation() {
	var _ IMap = (*Map)(nil)
}

type color bool

const (
	black, red color = true, false
)

// Map holds elements of the red-black tree
type Map struct {
	Root       *Node
	size       int
	Comparator containers.Comparator
}

// Node is a single element within the Map
type Node struct {
	Key    interface{}
	Value  interface{}
	color  color
	Left   *Node
	Right  *Node
	Parent *Node
}

// NewWith instantiates a red-black tree with the custom comparator.
func NewWith(comparator containers.Comparator) *Map {
	return &Map{Comparator: comparator}
}

// NewWithIntComparator instantiates a red-black tree with the IntComparator, i.e. keys are of type int.
func NewWithIntComparator() *Map {
	return &Map{Comparator: containers.IntComparator}
}

// NewWithIntComparator instantiates a red-black tree with the UInt32Comparator, i.e. keys are of type int.
func NewWithUInt32Comparator() *Map {
	return &Map{Comparator: containers.UInt32Comparator}
}

// NewWithStringComparator instantiates a red-black tree with the StringComparator, i.e. keys are of type string.
func NewWithStringComparator() *Map {
	return &Map{Comparator: containers.StringComparator}
}

// Put inserts node into the tree.
// Key should adhere to the comparator's type assertion, otherwise method panics.
func (self *Map) Put(key interface{}, value interface{}) {
	var insertedNode *Node
	if self.Root == nil {
		// Assert key is of comparator's type for initial tree
		self.Comparator(key, key)
		self.Root = &Node{Key: key, Value: value, color: red}
		insertedNode = self.Root
	} else {
		node := self.Root
		loop := true
		for loop {
			compare := self.Comparator(key, node.Key)
			switch {
			case compare == 0:
				node.Key = key
				node.Value = value
				return
			case compare < 0:
				if node.Left == nil {
					node.Left = &Node{Key: key, Value: value, color: red}
					insertedNode = node.Left
					loop = false
				} else {
					node = node.Left
				}
			case compare > 0:
				if node.Right == nil {
					node.Right = &Node{Key: key, Value: value, color: red}
					insertedNode = node.Right
					loop = false
				} else {
					node = node.Right
				}
			}
		}
		insertedNode.Parent = node
	}
	self.insertCase1(insertedNode)
	self.size++
}

// Get searches the node in the tree by key and returns its value or nil if key is not found in tree.
// Second return parameter is true if key was found, otherwise false.
// Key should adhere to the comparator's type assertion, otherwise method panics.
func (self *Map) Get(key interface{}) (value interface{}, found bool) {
	node := self.lookup(key)
	if node != nil {
		return node.Value, true
	}
	return nil, false
}

// Remove remove the node from the tree by key.
// Key should adhere to the comparator's type assertion, otherwise method panics.
func (self *Map) Remove(key interface{}) {
	var child *Node
	node := self.lookup(key)
	if node == nil {
		return
	}
	if node.Left != nil && node.Right != nil {
		pred := node.Left.maximumNode()
		node.Key = pred.Key
		node.Value = pred.Value
		node = pred
	}
	if node.Left == nil || node.Right == nil {
		if node.Right == nil {
			child = node.Left
		} else {
			child = node.Right
		}
		if node.color == black {
			node.color = nodeColor(child)
			self.deleteCase1(node)
		}
		self.replaceNode(node, child)
		if node.Parent == nil && child != nil {
			child.color = black
		}
	}
	self.size--
}

// Empty returns true if tree does not contain any nodes
func (self *Map) Empty() bool {
	return self.size == 0
}

// Size returns number of nodes in the tree.
func (self *Map) Size() int {
	return self.size
}

// Keys returns all keys in-order
func (self *Map) Keys() []interface{} {
	keys := make([]interface{}, self.size)
	it := self.Iterator()
	for i := 0; it.Next(); i++ {
		keys[i] = it.Key()
	}
	return keys
}

// Values returns all values in-order based on the key.
func (self *Map) Values() []interface{} {
	values := make([]interface{}, self.size)
	it := self.Iterator()
	for i := 0; it.Next(); i++ {
		values[i] = it.Value()
	}
	return values
}

// Left returns the left-most (min) node or nil if tree is empty.
func (self *Map) Left() *Node {
	var parent *Node
	current := self.Root
	for current != nil {
		parent = current
		current = current.Left
	}
	return parent
}

// Right returns the right-most (max) node or nil if tree is empty.
func (self *Map) Right() *Node {
	var parent *Node
	current := self.Root
	for current != nil {
		parent = current
		current = current.Right
	}
	return parent
}

// Floor Finds floor node of the input key, return the floor node or nil if no floor is found.
// Second return parameter is true if floor was found, otherwise false.
//
// Floor node is defined as the largest node that is smaller than or equal to the given node.
// A floor node may not be found, either because the tree is empty, or because
// all nodes in the tree are larger than the given node.
//
// Key should adhere to the comparator's type assertion, otherwise method panics.
func (self *Map) Floor(key interface{}) (floor *Node, found bool) {
	found = false
	node := self.Root
	for node != nil {
		compare := self.Comparator(key, node.Key)
		switch {
		case compare == 0:
			return node, true
		case compare < 0:
			node = node.Left
		case compare > 0:
			floor, found = node, true
			node = node.Right
		}
	}
	if found {
		return floor, true
	}
	return nil, false
}

// Ceiling finds ceiling node of the input key, return the ceiling node or nil if no ceiling is found.
// Second return parameter is true if ceiling was found, otherwise false.
//
// Ceiling node is defined as the smallest node that is larger than or equal to the given node.
// A ceiling node may not be found, either because the tree is empty, or because
// all nodes in the tree are smaller than the given node.
//
// Key should adhere to the comparator's type assertion, otherwise method panics.
func (self *Map) Ceiling(key interface{}) (ceiling *Node, found bool) {
	found = false
	node := self.Root
	for node != nil {
		compare := self.Comparator(key, node.Key)
		switch {
		case compare == 0:
			return node, true
		case compare < 0:
			ceiling, found = node, true
			node = node.Left
		case compare > 0:
			node = node.Right
		}
	}
	if found {
		return ceiling, true
	}
	return nil, false
}

// Clear removes all nodes from the tree.
func (self *Map) Clear() {
	self.Root = nil
	self.size = 0
}

// String returns a string representation of container
func (self *Map) String() string {
	str := "RedBlackTree\n"
	if !self.Empty() {
		output(self.Root, "", true, &str)
	}
	return str
}

func (node *Node) String() string {
	return fmt.Sprintf("%v", node.Key)
}

func output(node *Node, prefix string, isTail bool, str *string) {
	if node.Right != nil {
		newPrefix := prefix
		if isTail {
			newPrefix += "│   "
		} else {
			newPrefix += "    "
		}
		output(node.Right, newPrefix, false, str)
	}
	*str += prefix
	if isTail {
		*str += "└── "
	} else {
		*str += "┌── "
	}
	*str += node.String() + "\n"
	if node.Left != nil {
		newPrefix := prefix
		if isTail {
			newPrefix += "    "
		} else {
			newPrefix += "│   "
		}
		output(node.Left, newPrefix, true, str)
	}
}

func (self *Map) lookup(key interface{}) *Node {
	node := self.Root
	for node != nil {
		compare := self.Comparator(key, node.Key)
		switch {
		case compare == 0:
			return node
		case compare < 0:
			node = node.Left
		case compare > 0:
			node = node.Right
		}
	}
	return nil
}

func (node *Node) grandparent() *Node {
	if node != nil && node.Parent != nil {
		return node.Parent.Parent
	}
	return nil
}

func (node *Node) uncle() *Node {
	if node == nil || node.Parent == nil || node.Parent.Parent == nil {
		return nil
	}
	return node.Parent.sibling()
}

func (node *Node) sibling() *Node {
	if node == nil || node.Parent == nil {
		return nil
	}
	if node == node.Parent.Left {
		return node.Parent.Right
	}
	return node.Parent.Left
}

func (self *Map) rotateLeft(node *Node) {
	right := node.Right
	self.replaceNode(node, right)
	node.Right = right.Left
	if right.Left != nil {
		right.Left.Parent = node
	}
	right.Left = node
	node.Parent = right
}

func (self *Map) rotateRight(node *Node) {
	left := node.Left
	self.replaceNode(node, left)
	node.Left = left.Right
	if left.Right != nil {
		left.Right.Parent = node
	}
	left.Right = node
	node.Parent = left
}

func (self *Map) replaceNode(old *Node, new *Node) {
	if old.Parent == nil {
		self.Root = new
	} else {
		if old == old.Parent.Left {
			old.Parent.Left = new
		} else {
			old.Parent.Right = new
		}
	}
	if new != nil {
		new.Parent = old.Parent
	}
}

func (self *Map) insertCase1(node *Node) {
	if node.Parent == nil {
		node.color = black
	} else {
		self.insertCase2(node)
	}
}

func (self *Map) insertCase2(node *Node) {
	if nodeColor(node.Parent) == black {
		return
	}
	self.insertCase3(node)
}

func (self *Map) insertCase3(node *Node) {
	uncle := node.uncle()
	if nodeColor(uncle) == red {
		node.Parent.color = black
		uncle.color = black
		node.grandparent().color = red
		self.insertCase1(node.grandparent())
	} else {
		self.insertCase4(node)
	}
}

func (self *Map) insertCase4(node *Node) {
	grandparent := node.grandparent()
	if node == node.Parent.Right && node.Parent == grandparent.Left {
		self.rotateLeft(node.Parent)
		node = node.Left
	} else if node == node.Parent.Left && node.Parent == grandparent.Right {
		self.rotateRight(node.Parent)
		node = node.Right
	}
	self.insertCase5(node)
}

func (self *Map) insertCase5(node *Node) {
	node.Parent.color = black
	grandparent := node.grandparent()
	grandparent.color = red
	if node == node.Parent.Left && node.Parent == grandparent.Left {
		self.rotateRight(grandparent)
	} else if node == node.Parent.Right && node.Parent == grandparent.Right {
		self.rotateLeft(grandparent)
	}
}

func (node *Node) maximumNode() *Node {
	if node == nil {
		return nil
	}
	for node.Right != nil {
		node = node.Right
	}
	return node
}

func (self *Map) deleteCase1(node *Node) {
	if node.Parent == nil {
		return
	}
	self.deleteCase2(node)
}

func (self *Map) deleteCase2(node *Node) {
	sibling := node.sibling()
	if nodeColor(sibling) == red {
		node.Parent.color = red
		sibling.color = black
		if node == node.Parent.Left {
			self.rotateLeft(node.Parent)
		} else {
			self.rotateRight(node.Parent)
		}
	}
	self.deleteCase3(node)
}

func (self *Map) deleteCase3(node *Node) {
	sibling := node.sibling()
	if nodeColor(node.Parent) == black &&
		nodeColor(sibling) == black &&
		nodeColor(sibling.Left) == black &&
		nodeColor(sibling.Right) == black {
		sibling.color = red
		self.deleteCase1(node.Parent)
	} else {
		self.deleteCase4(node)
	}
}

func (self *Map) deleteCase4(node *Node) {
	sibling := node.sibling()
	if nodeColor(node.Parent) == red &&
		nodeColor(sibling) == black &&
		nodeColor(sibling.Left) == black &&
		nodeColor(sibling.Right) == black {
		sibling.color = red
		node.Parent.color = black
	} else {
		self.deleteCase5(node)
	}
}

func (self *Map) deleteCase5(node *Node) {
	sibling := node.sibling()
	if node == node.Parent.Left &&
		nodeColor(sibling) == black &&
		nodeColor(sibling.Left) == red &&
		nodeColor(sibling.Right) == black {
		sibling.color = red
		sibling.Left.color = black
		self.rotateRight(sibling)
	} else if node == node.Parent.Right &&
		nodeColor(sibling) == black &&
		nodeColor(sibling.Right) == red &&
		nodeColor(sibling.Left) == black {
		sibling.color = red
		sibling.Right.color = black
		self.rotateLeft(sibling)
	}
	self.deleteCase6(node)
}

func (self *Map) deleteCase6(node *Node) {
	sibling := node.sibling()
	sibling.color = nodeColor(node.Parent)
	node.Parent.color = black
	if node == node.Parent.Left && nodeColor(sibling.Right) == red {
		sibling.Right.color = black
		self.rotateLeft(node.Parent)
	} else if nodeColor(sibling.Left) == red {
		sibling.Left.color = black
		self.rotateRight(node.Parent)
	}
}

func nodeColor(node *Node) color {
	if node == nil {
		return black
	}
	return node.color
}
