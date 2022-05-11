package skiplist

import (
	"math"
	rand2 "math/rand"
	"sync"
)

const (
	ZSKIPLIST_MAXLEVEL = 32			//skiplist最高层
	ZSKIPLIST_P = 25			//晋升概率
)

type ISkipVal interface {
	Less(v interface{}) bool
	Equal(v interface{}) bool
	Greater(v interface{}) bool
}

type ISkipNode interface {
	SetNext(ISkipNode)
	GetNext() ISkipNode
	GetPre() ISkipNode
	SetPre(ISkipNode)
	SetUp(ISkipNode)
	GetUp() ISkipNode
	SetDown(ISkipNode)
	GetDown() ISkipNode
	Clear()
	GetVal() ISkipVal

}

type SkipLevelNode struct {		//非0层节点
	mNextNode ISkipNode
	mPreNode  ISkipNode
	mUpNode   ISkipNode
	mDownNode ISkipNode

	mVal ISkipVal
}

func (self *SkipLevelNode) SetVal(v ISkipVal)  {
	self.mVal = v
}

func (self *SkipLevelNode) GetVal() ISkipVal {
	return self.mVal
}

func (self *SkipLevelNode) SetNext(node ISkipNode) {
	self.mNextNode = node
}

func (self *SkipLevelNode) GetNext() ISkipNode {
	return self.mNextNode
}

func (self *SkipLevelNode) GetPre() ISkipNode {
	return self.mPreNode
}

func (self *SkipLevelNode) SetPre(node ISkipNode) {
	self.mPreNode = node
}

func (self *SkipLevelNode) SetUp(node ISkipNode)  {
	self.mUpNode = node
}

func (self *SkipLevelNode) GetUp() ISkipNode {
	return self.mUpNode
}

func (self *SkipLevelNode) SetDown(node ISkipNode) {
	self.mDownNode = node
}

func (self *SkipLevelNode) GetDown() ISkipNode {
	return self.mDownNode
}

func (self *SkipLevelNode) Clear() {
	self.mNextNode = nil
	self.mPreNode = nil
	self.mUpNode = nil
	self.mDownNode = nil
	self.mVal = nil
}

type SkipList struct {		//跳表
	mRoot [ZSKIPLIST_MAXLEVEL]ISkipNode
	mNowLevel int		//当前最大有效层数
	mCnodescache sync.Pool
	mEnableCache bool
}

//构造一个跳跃表
//@cache: 是否缓存节点
func NewSkipList(minItem ISkipVal, maxItem ISkipVal, cache bool) *SkipList {
	list := new(SkipList)
	for i:=0; i<ZSKIPLIST_MAXLEVEL; i++ {
		rt1 := new(SkipLevelNode)
		rt1.SetVal(minItem)
		list.mRoot[i] = rt1

		rt2 := new(SkipLevelNode)
		rt2.SetVal(maxItem)
		rt2.SetPre(rt1)
		rt1.SetNext(rt2)
	}
	for i:=ZSKIPLIST_MAXLEVEL-1; i>0; i-- {
		list.mRoot[i].SetDown(list.mRoot[i-1])
		list.mRoot[i].GetNext().SetDown(list.mRoot[i-1].GetNext())

		list.mRoot[i-1].SetUp(list.mRoot[i])
		list.mRoot[i-1].GetNext().SetUp(list.mRoot[i].GetNext())
	}
	list.mNowLevel = 0
	list.mEnableCache = cache
	if cache {
		list.mCnodescache.New = func() interface{} {
			return new(SkipLevelNode)
		}
	}
	return list
}


func (self *SkipList) GetAllValues(n int) []ISkipVal {
	cnt := 0
	rets := make([]ISkipVal, 0)
	last := self.mRoot[0].GetNext()
	if n <= 0 {
		n = math.MaxInt
	}
	for last.GetNext() != nil {
		rets = append(rets, last.GetVal())
		last = last.GetNext()
		cnt++
		if cnt >= n {
			return rets
		}
	}
	return rets
}

//redis的高度晋升函数
func (self *SkipList) zslRandomLevel() int {
	var level = 1
	for rand2.Int31n(100) < ZSKIPLIST_P {
		level += 1
	}
	if level < ZSKIPLIST_MAXLEVEL {
		return level
	} else {
		return ZSKIPLIST_MAXLEVEL
	}
}

//搜索应该插入的节点位置
func (self *SkipList) searchNode(val ISkipVal) ISkipNode {
	root := self.mRoot[self.mNowLevel]
	last := root
	for root != nil {
		last = root
		next := root.GetNext()
		for next != nil {
			if val.Less(next.GetVal()) {	//要查找的值在下一层
				root = last.GetDown()
				break
			} else if val.Equal(next.GetVal()) {		//精确找到该值，返回第0层该节点
				root = nil
				last = next
				break
			} else {		//向后查找
				last = next
				next = next.GetNext()
			}
		}
	}
	for last.GetDown() != nil {
		last = last.GetDown()
	}
	return last		//该节点的值<=val，在该节点位置之后插入
}

func (self *SkipList) AddNode(valnode ISkipVal) {

	node := new(SkipLevelNode)
	node.mVal = valnode


	root := self.searchNode(node.GetVal())		//寻找插入位置

	root.GetNext().SetPre(node)		//插入节点
	node.SetNext(root.GetNext())
	root.SetNext(node)
	node.SetPre(root)

	maxlevel := self.zslRandomLevel()		//插入层数

	if self.mNowLevel < maxlevel {
		self.mNowLevel = maxlevel - 1
	}
	var downNode = node
	for i:=1; i<maxlevel; i++ {		//逐层插入
		root = self.mRoot[i]
		last := root
		next := last.GetNext()
		var lnode *SkipLevelNode
		if self.mEnableCache {
			lnode = self.mCnodescache.Get().(*SkipLevelNode)
		} else {
			lnode = new(SkipLevelNode)
		}
		lnode.mVal = valnode

		for next != nil {
			if valnode.Less(next.GetVal()) {
				next.SetPre(lnode)
				break
			}
			last = next.(*SkipLevelNode)
			next = next.GetNext()
		}

		downNode.SetUp(lnode)
		lnode.SetDown(downNode)
		lnode.SetNext(next)
		last.SetNext(lnode)
		lnode.SetPre(last)
		downNode = lnode
	}
}

func (self *SkipList) DelNode(node ISkipNode) {
	node.GetPre().SetNext(node.GetNext())
	node.GetNext().SetPre(node.GetPre())

	upnode := node.GetUp()
	for upnode != nil {		//从下往上逐层删除
		upnode.GetPre().SetNext(upnode.GetNext())
		upnode.GetNext().SetPre(upnode.GetPre())
		if self.mEnableCache {
			upnode.Clear()
			self.mCnodescache.Put(upnode)
		}
		upnode = upnode.GetUp()
	}
}

func (self *SkipList) Find(val ISkipVal) ISkipNode {
	root := self.mRoot[self.mNowLevel]
	last := root
	found := false
	for root != nil {
		last = root
		next := root.GetNext()
		for next != nil {
			if val.Less(next.GetVal()) {	//要查找的值在下一层
				root = last.GetDown()
				break
			} else if val.Equal(next.GetVal()) {		//精确找到该值，返回第0层该节点
				root = nil
				last = next
				found = true
				break
			} else {		//向后查找
				last = next
				next = next.GetNext()
			}
		}
	}
	if !found {
		return nil
	}
	for last.GetDown() != nil {		//返回的节点必须是0层节点
		last = last.GetDown()
	}
	return last
}