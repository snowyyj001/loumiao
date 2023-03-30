package detour

import (
	"github.com/snowyyj001/loumiao/lbase/lmath"
	"github.com/snowyyj001/loumiao/lbase/vector"
	"testing"
)

//https://github.com/kbengine/unity3d_nav_critterai

func Test66(t *testing.T) {
	t.Log("测试巡径")
	dt := NewDetour(1000)
	dt.Load("CAIBakedNavmesh.navmesh")
	for j := 0; j < 10000; j++ {
		dt.FindPath(lmath.Point3F{-500, 0, 0}, lmath.Point3F{0, 0, 0}, vector.NewVector())
	}
}

func TestOne(t *testing.T) {
	dt := NewDetour(1000)
	errCode := dt.Load("CAIBakedNavmesh.nav")
	if errCode != 0 {
		t.Log("nav format error: ", errCode)
		return
	}
	vec := vector.NewVector()
	ok := dt.FindPath(lmath.Point3F{75, 0, 13}, lmath.Point3F{0, 0, 0}, vec)
	if ok {
		t.Log("测试ok", vec.Len(), vec.Size())
		iter := vec.Iterator()
		for iter.Next() {
			t.Log(iter.Value())
		}

	} else {
		t.Log("测试失败")
	}
}
