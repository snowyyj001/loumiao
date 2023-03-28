package skiplist

import (
	"fmt"
	"math"
	rand2 "math/rand"
	"testing"
	"time"
)

/*
go test -v skipList_test.go skipList.go
go test -v -bench=. skipList_test.go skipList.go
*/

type Item struct {
	Score int
}

func (self *Item) Less(v interface{}) bool {
	vv := v.(*Item)
	return self.Score < vv.Score
}

func (self *Item) Equal(v interface{}) bool {
	vv := v.(*Item)
	return self.Score == vv.Score
}

func (self *Item) Greater(v interface{}) bool {
	vv := v.(*Item)
	return self.Score > vv.Score
}

func TestSkipList(t *testing.T) {
	rand2.Seed(time.Now().UnixNano())
	skiplist := NewSkipList(&Item{Score: math.MinInt}, &Item{Score: math.MaxInt}, true)
	var finditem *Item
	for i := 0; i < 10; i++ {
		item := new(Item)
		item.Score = int(rand2.Int31n(int32(100)))
		skiplist.AddNode(item)
		fmt.Println(item.Score)
		finditem = item
	}
	for i := 0; i < skiplist.mNowLevel+1; i++ {
		node := skiplist.mRoot[i].GetNext()
		for node.GetNext() != nil {
			fmt.Print(" ", node.GetVal().(*Item).Score)
			node = node.GetNext()
		}
		fmt.Println("  ")
	}
	fmt.Println("find item", finditem.Score)
	si := skiplist.Find(finditem)
	fmt.Println("find result ", si.GetVal())

	fmt.Println("del item", finditem.Score)
	skiplist.DelNode(si)
	fmt.Println("del result ")
	for i := 0; i < skiplist.mNowLevel+1; i++ {
		node := skiplist.mRoot[i].GetNext()
		for node.GetNext() != nil {
			fmt.Print(" ", node.GetVal().(*Item).Score)
			node = node.GetNext()
		}
		fmt.Println("  ")
	}
}

func Benchmark_SkipList(b *testing.B) {
	skiplist := NewSkipList(&Item{Score: math.MinInt}, &Item{Score: math.MaxInt}, true)
	for i := 0; i < 1000; i++ {
		item := new(Item)
		skiplist.AddNode(item)
	}
}

func Benchmark_Find(b *testing.B) {
	skiplist := NewSkipList(&Item{Score: math.MinInt}, &Item{Score: math.MaxInt}, true)
	for i := 0; i < 1000; i++ {
		item := new(Item)
		item.Score = rand2.Int() % 100000
		skiplist.AddNode(item)
	}
	for i := 0; i < b.N; i++ {
		skiplist.Find(&Item{Score: rand2.Int() % 100000})
	}
}
