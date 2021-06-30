package aoi

import (
	rand2 "math/rand"
	"sync"
)
//基于跳跃表的十字链表AOI
//位置移动，改变在链表中的位置，每层都要移动，如果聚集的玩家间相互移动频繁，效率会很低
//所以基于哨兵的十字链表应该比跳跃表效率高
//人为的给链表添加哨兵节点，哨兵节点固定，哨兵节点越多，查找效率越高
//移动时，只移动链表本身；查找时，先二分查找到哨兵节点，再去找该哨兵节点后的链表节点
const (
	ZSKIPLIST_MAXLEVEL = 32			//skiplist最高层
	ZSKIPLIST_P = 25			//晋升概率
)

var (
	cnodescache sync.Pool
)


type ICrossNode interface {
	SetNext(ICrossNode)
	GetNext() ICrossNode
	GetPre() ICrossNode
	SetPre(ICrossNode)
	GetUp() ICrossNode
	GetDown() ICrossNode
	GetVal() float32
}

type CrossLevelNode struct {		//非0层节点
	mNextNode ICrossNode
	mPreNode ICrossNode
	mUpNode ICrossNode
	mDownNode ICrossNode

	mVal float32
}

func (self *CrossLevelNode) GetVal() float32 {
	return self.mVal
}

func (self *CrossLevelNode) SetNext(node ICrossNode) {
	self.mNextNode = node
}

func (self *CrossLevelNode) GetNext() ICrossNode {
	return self.mNextNode
}

func (self *CrossLevelNode) GetPre() ICrossNode {
	return self.mPreNode
}

func (self *CrossLevelNode) SetPre(node ICrossNode) {
	self.mPreNode = node
}

func (self *CrossLevelNode) GetUp() ICrossNode {
	return self.mUpNode
}

func (self *CrossLevelNode) GetDown() ICrossNode {
	return self.mDownNode
}

type CrossNode struct {		//0层节点
	CrossLevelNode

	mUid int
}

type CrossEntity struct {		//0层节点
	mAxisx *CrossNode
	mAxisy *CrossNode
	mAxisz *CrossNode

	mViewSize float32		//视野
	mUid int
	mViwers []int
}

type SkipList struct {		//跳表
	mRoot [ZSKIPLIST_MAXLEVEL]ICrossNode
	mNowLevel int		//当前最大有效层数

}

func (self *SkipList) zslRandomLevel() int {
	var level int = 1
	for (rand2.Int31n(100) < ZSKIPLIST_P) {
		level += 1;
	}
	if level < ZSKIPLIST_MAXLEVEL {
		return level
	} else {
		return ZSKIPLIST_MAXLEVEL
	}
}

//搜索应该插入的节点位置
func (self *SkipList) SearchNode(val float32) ICrossNode {
	root := self.mRoot[self.mNowLevel]
	last := root
	for root != nil {
		last = root
		next := root.GetNext()
		if next == root {		//链表尾部，向下层搜索
			root = root.GetDown()
			last = root
		} else {
			for next != root {
				if val < next.GetVal() {	//要查找的值在下一层
					root = last.GetDown()
					break
				} else if val == next.GetVal() {		//精确找到该值，返回第0层该节点
					for next.GetDown() != nil {
						next = next.GetDown()
					}
					root = nil
					last = next
					break
				}
				last = next
				next = next.GetNext()
			}
		}
	}
	return last		//该节点的值<=val，在该节点位置之后插入
}

func (self *SkipList) AddNode(node *CrossNode) {
	root := self.SearchNode(node.mVal)		//寻找插入位置
	root.GetNext().SetPre(node)		//插入节点
	node.SetNext(root.GetNext())
	root.SetNext(node)
	node.SetPre(root)

	maxlevel := self.zslRandomLevel()		//插入层数
	if self.mNowLevel < maxlevel {
		self.mNowLevel = maxlevel
	}
	val := node.mVal
	for i:=1; i<maxlevel; i++ {		//逐层插入
		root = self.mRoot[i]
		last := root
		next := last.GetNext()

		lnode := cnodescache.Get().(*CrossLevelNode)
		lnode.mVal = val
		for next != root {
			if val < next.GetVal() {
				next.SetPre(lnode)
				break
			}
			last = next.(*CrossNode)
			next = next.GetNext()
		}
		lnode.SetNext(next)
		last.SetNext(lnode)
		lnode.SetPre(last)
	}
}

func (self *SkipList) DelNode(node *CrossNode) {
	node.mPreNode.SetNext(node.mNextNode)
	node.mNextNode.SetPre(node.mPreNode)

	upnode := node.mUpNode
	for upnode != nil {		//从下往上逐层删除
		upnode.GetPre().SetNext(upnode.GetNext())
		upnode.GetNext().SetPre(upnode.GetPre())
		cnodescache.Put(upnode)
		upnode = upnode.GetUp()
	}
}

func (self *SkipList) MoveNode(node *CrossNode) {
	val := node.GetVal()
	movenode := ICrossNode(node)
	var moded bool = false
	for i:=0; i<self.mNowLevel; i++ {		//向后移动
		next := movenode.GetNext()
		for val > next.GetVal() {
			next = next.GetNext()
		}
		changenode := next.GetPre()
		if changenode != movenode {
			tmp := movenode.GetPre()
			movenode.SetPre(changenode.GetPre())
			changenode.SetPre(tmp)
			tmp = movenode.GetNext()
			movenode.SetNext(changenode.GetNext())
			changenode.SetNext(tmp)

			moded = true
		} else {
			break
		}
		movenode = movenode.GetUp()
		if movenode == nil {
			break
		}
	}
	if moded {
		return
	}
	for i:=0; i<self.mNowLevel; i++ {		//向前移动
		next := movenode.GetPre()
		for val < next.GetVal() {
			next = next.GetPre()
		}
		changenode := next.GetNext()
		if changenode != movenode {
			tmp := movenode.GetPre()
			movenode.SetPre(changenode.GetPre())
			changenode.SetPre(tmp)
			tmp = movenode.GetNext()
			movenode.SetNext(changenode.GetNext())
			changenode.SetNext(tmp)
		} else {
			break
		}
		movenode = movenode.GetUp()
		if movenode == nil {
			return
		}
	}
}

type CrossList struct {		// xyz十字链
	mRootx *SkipList
	mRooty *SkipList
	mRootz *SkipList
}

func (self *CrossList) SearchViwers(mRoot *SkipList, node *CrossNode, view float32) []*CrossNode {
	next := node.GetNext()
	viewsz := node.GetVal() + view
	var tmps []*CrossNode
	for next != mRoot.mRoot[0] {		//向后搜索
		if next.GetVal() < viewsz{
			tmps = append(tmps, next.(*CrossNode))
			next = next.GetNext()
		} else {
			break
		}
	}
	viewsz = node.GetVal() - view
	pre := node.GetPre()
	for pre != mRoot.mRoot[0] {		//向前搜索
		if pre.GetVal() > viewsz {
			tmps = append(tmps, pre.(*CrossNode))
			pre = pre.GetPre()
		} else {
			break
		}
	}

	return tmps
}

func (self *CrossList) RefreshViwer(entity *CrossEntity) {
	xv := self.SearchViwers(self.mRootx, entity.mAxisx, entity.mViewSize)
	yv := self.SearchViwers(self.mRooty, entity.mAxisy, entity.mViewSize)
	zv := self.SearchViwers(self.mRootz, entity.mAxisz, entity.mViewSize)
	var viwers map[int]int = make(map[int]int)
	for _, p := range xv {
		viwers[p.mUid]++
	}

	for _, p := range yv {
		viwers[p.mUid]++
	}

	for _, p := range zv {
		viwers[p.mUid]++
	}

	entity.mViwers = []int{}
	for uid, n := range viwers {
		if n == 3 {			//xyz轴上都在视野范围内
			entity.mViwers = append(entity.mViwers, uid)
		}
	}
}

func inti()  {
	cnodescache.New = func() interface{} {
		return new(CrossLevelNode)
	}
}
