package detour

import (
	"github.com/snowyyj001/loumiao/base/lmath"
	"github.com/snowyyj001/loumiao/base/vector"
	"testing"
)

//https://github.com/kbengine/unity3d_nav_critterai

func Test66(t *testing.T) {
	t.Log("测试巡径")
	dt := NewDetour(1000)
	dt.Load("srv_CAIBakedNavmesh.navmesh")
	for j := 0; j < 10000; j++ {
		dt.FindPath(lmath.Point3F{-500, 0, 0}, lmath.Point3F{0, 0, 0}, vector.NewVector())
	}
}
