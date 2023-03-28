package lutil_test

import (
	"fmt"
	"github.com/snowyyj001/loumiao/lutil"
	"testing"
)

func TestDH(t *testing.T) {
	lutil.Calc_DH_Yb()
	lutil.Calc_DH_A_k()

	lutil.Calc_DH_Ya()
	lutil.Calc_DH_B_k()

	fmt.Println("DH_YA", lutil.DH_YA)
	fmt.Println("DH_YB", lutil.DH_YB)
	fmt.Println("DH_KA", lutil.DH_KA)
	fmt.Println("DH_KB", lutil.DH_KB)

	if lutil.DH_KA != lutil.DH_KB {
		t.FailNow()
	}
}
