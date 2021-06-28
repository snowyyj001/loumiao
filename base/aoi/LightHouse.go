package aoi

import "fmt"

//“九宫格法”——每个区域中记录的是处在本区域的实体
//“灯塔法”——每个区域中记录的是会观察到我的实体
//坐标原点在左下角
//灯塔范围半径>=玩家视野半径，这样每个玩家最多订阅四个灯塔，最少一个
//丫的，灯塔应该比九宫格效率低的，稍后测试下
type LightGrid struct {
	mEntitys map[int]IAOIEntity		//灯塔区域的玩家
}

func (self *LightGrid) init() {
	self.mEntitys = make(map[int]IAOIEntity)
}

func (self *LightGrid) removeEntity(entity IAOIEntity) {
	delete(self.mEntitys, entity.GetUid())
}

func (self *LightGrid) addEntity(entity IAOIEntity) {
	self.mEntitys[entity.GetUid()] = entity
}

type LightHouse struct {
	 mGrids [][]*LightGrid
	 mCellSize float32
	 mCellWidth int
	 mCellHeight int
}

//构建网格
//@cellsize 网格大小
//@width 地图宽
//@height 地图长
func (self *LightHouse) BuildMap(cellsize, width, height int) {
	nw := width / cellsize
	nh := height / cellsize
	self.mGrids = make([][]*LightGrid, nw)
	for i := 0; i < nw; i++ {
		self.mGrids[i] = make([]*LightGrid, nh)
	}
	for i := 0; i < nw; i++ {
		for j := 0; j < nh; j++ {
			grid := new(LightGrid)
			grid.init()
			self.mGrids[i][j] = grid
		}
	}


	self.mCellSize = float32(cellsize)
	self.mCellWidth = nw
	self.mCellHeight = nh
}

//获得x，y位置所在的灯塔
func (self *LightHouse) Pos2Grid(x, y float32) (int, int) {
	px := int(x / self.mCellSize)
	py := int(y / self.mCellSize)
	return px, py
}

//获得x，y位置所在的灯塔
func (self *LightHouse) GetGrid(x, y float32) *LightGrid {
	px := int(x / self.mCellSize)
	py := int(y / self.mCellSize)

	if px < 0 || px >= self.mCellWidth {
		return nil
	}
	if py < 0 || py >= self.mCellHeight {
		return nil
	}

	return nil
}

//刷新玩家可视灯塔
func (self *LightHouse) RefreshHouse(entity IAOIEntity) {
	px, py := entity.GetPos()
	vz := entity.GetViewSize()

	x, y := self.Pos2Grid(px, py)		//应该在哪个塔区域里
	bx, by := entity.GetGrid()		//当前所在塔区
	self.mGrids[bx][by].addEntity(entity)
	if bx != x || by != y {		//交换塔区
		self.mGrids[bx][by].removeEntity(entity)
		self.mGrids[x][y].addEntity(entity)
		entity.SetGrid(x, y)
	}

	//更新灯塔观察者
	entity.ResetWatch()

	entity.AddWatch(x, y)



	if (y - 1 >= 0) {	//down
		if (py - float32(y)*self.mCellSize < vz) {
			entity.AddWatch(x, y-1)
		}
	}

	if (y + 1 < self.mCellHeight) {	//up
		if (float32(y+1)*self.mCellSize - py < vz) {
			entity.AddWatch(x, y+1)
		}
	}

	if (x - 1 >= 0) {
		if (px - float32(x)*self.mCellSize < vz) {	//left
			entity.AddWatch(x-1, y)
			if (y - 1 >= 0) {	//left-down
				if (py - float32(y)*self.mCellSize < vz) {
					entity.AddWatch(x-1, y-1)
				}
			}
			if (y + 1 < self.mCellHeight) {	//left-up
				if (float32(y+1)*self.mCellSize - py < vz) {
					entity.AddWatch(x-1, y+1)
				}
			}
		}
	}

	if (x + 1 < self.mCellWidth) {
		if (float32(x+1)*self.mCellSize - px < vz) {	//right
			entity.AddWatch(x+1, y)
			if (y - 1 >= 0) {	//right-down
				if (py - float32(y)*self.mCellSize < vz) {
					entity.AddWatch(x+1, y-1)
				}
			}
			if (y + 1 < self.mCellHeight) {	//right-up
				if (float32(y+1)*self.mCellSize - py < vz) {
					entity.AddWatch(x+1, y+1)
				}
			}
		}
	}
}

//检查距离视野
func (self *LightHouse) CheckDistance(watcher, bewatcher IAOIEntity) bool {
	x1, y1 := watcher.GetPos()
	x2, y2 := bewatcher.GetPos()
	vz := watcher.GetViewSize()
	vz *= vz
	l := (x1 - x2) * (x1 - x2) + (y1 - y2) * (y1 - y2)
	return vz >= l
}

//刷新玩家视野
func (self *LightHouse) RefreshViwer(entity IAOIEntity) {
	entity.ResetViwers()
	watcher := entity.GetWatch()
	for _, towerpox := range watcher {
		tower := self.mGrids[towerpox.X][towerpox.Y]
		for _, be := range tower.mEntitys {
			if be.GetUid() != entity.GetUid() && self.CheckDistance(entity, be) {
				px, py := be.GetPos()
				fmt.Println(be.GetUid(), px, py)
				entity.AddViwers(be)
			}
		}
	}
}
