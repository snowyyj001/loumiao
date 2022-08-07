package util

import (
	"fmt"
	"github.com/snowyyj001/loumiao/config"
	"testing"
)

func Test_Uuid(t *testing.T) {
	config.SERVER_NODE_UID = 1
	nt1 := TimeStamp()
	for i := 0; i < 10000; i++ {
		UUID()
	}
	nt2 := TimeStamp()
	fmt.Println("Test_Uuid", nt2-nt1)
}

func Benchmark_Uuid(b *testing.B) {
	for i := 0; i < b.N; i++ {
		UUID()
	}
}
