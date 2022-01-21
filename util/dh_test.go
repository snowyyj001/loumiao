package util_test

import (
	"fmt"
	"testing"


	"github.com/snowyyj001/loumiao/util"
)

func TestDH(t *testing.T)  {
	util.Calc_DH_Yb()
	util.Calc_DH_A_k()

	util.Calc_DH_Ya()
	util.Calc_DH_B_k()

	fmt.Println("DH_YA", util.DH_YA)
	fmt.Println("DH_YB", util.DH_YB)
	fmt.Println("DH_KA", util.DH_KA)
	fmt.Println("DH_KB", util.DH_KB)

	if util.DH_KA != util.DH_KB {
		t.FailNow()
	}
}

