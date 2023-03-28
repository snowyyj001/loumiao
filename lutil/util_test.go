package lutil

import (
	"github.com/snowyyj001/loumiao/lconfig"
	"testing"
)

func Test_Uuid(t *testing.T) {
	lconfig.SERVER_NODE_UID = 1
	for i := 0; i < 10000; i++ {
		UUID()
	}
}

func Benchmark_Uuid(b *testing.B) {
	for i := 0; i < b.N; i++ {
		UUID()
	}
}
