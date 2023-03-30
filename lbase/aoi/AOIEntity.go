package aoi

import (
	"github.com/snowyyj001/loumiao/lbase/lmath"
	"sync"
)

const (
	WATCH_MAX = 4
)

var (
	towerscache sync.Pool
)

type IAOIEntity interface {
	GetUid() int                //entity uid
	GetPos() (float32, float32) //位置
	GetGrid() (int, int)        //所属网格
	SetGrid(int, int)
	SetViewSize(sz float32) //视野大小
	GetViewSize() float32
	ResetViwers() //可视玩家
	AddViwers(IAOIEntity)
	GetViwers() []IAOIEntity
	AddWatch(x, y int)
	ResetWatch()
	GetWatch() []*lmath.Vector2
}

type AOIEntity struct {
	mGridx       int //
	mGridy       int //
	mViewSize    float32
	mViwers      []IAOIEntity     //在视野内的entity
	mWatchTowers []*lmath.Vector2 //订阅塔坐标

	//方便测试，实际应该由子类entity提供这些值
	Posx float32
	Posy float32
	Uid  int
}

func (self *AOIEntity) GetUid() int {
	return self.Uid
}
func (self *AOIEntity) GetPos() (float32, float32) {
	return self.Posx, self.Posy
}
func (self *AOIEntity) SetViewSize(sz float32) {
	self.mViewSize = sz
}
func (self *AOIEntity) GetViewSize() float32 {
	return self.mViewSize
}
func (self *AOIEntity) GetGrid() (int, int) {
	return self.mGridx, self.mGridy
}
func (self *AOIEntity) SetGrid(x, y int) {
	self.mGridx = x
	self.mGridy = y
}

func (self *AOIEntity) GetViwers() []IAOIEntity {
	return self.mViwers
}

func (self *AOIEntity) ResetViwers() {
	self.mViwers = []IAOIEntity{}
}
func (self *AOIEntity) AddViwers(v IAOIEntity) {
	self.mViwers = append(self.mViwers, v)
}
func (self *AOIEntity) ResetWatch() {
	for _, v := range self.mWatchTowers {
		towerscache.Put(v)
	}
	self.mWatchTowers = []*lmath.Vector2{}
}
func (self *AOIEntity) AddWatch(x, y int) {
	v := towerscache.Get().(*lmath.Vector2)
	v.X = x
	v.Y = y
	self.mWatchTowers = append(self.mWatchTowers, v)
}

func (self *AOIEntity) GetWatch() []*lmath.Vector2 {
	return self.mWatchTowers
}
func init() {
	towerscache.New = func() interface{} {
		return new(lmath.Vector2)
	}
}
