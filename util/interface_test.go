package util

import (
	"testing"
)

/*
go test -v -bench="." interface_test.go

goos: windows
goarch: amd64
cpu: Intel(R) Core(TM) i7-4790 CPU @ 3.60GHz
Benchmark_if
Benchmark_if-8           2621606               453.7 ns/op
Benchmark_ifi
Benchmark_ifi-8          2798382               431.1 ns/op
PASS
ok      command-line-arguments  3.442s

经以上测试可知，map记录接口IStu比空的interface{}效率要高，但非常接近，可以忽略不计
*/

type IStu1 interface {
	Reset()
	Get() int
}

type Stu1 struct {
	Id   int
	Buff []byte
	Name string
	Id2  int
	Id3  int
	Id4  int
	Id5  int
	Id6  int
	Id7  int
	Id8  int
}

func (self *Stu1) Reset() {
	self.Id = -1
}

func (self *Stu1) Get() int {
	return self.Id
}

func Benchmark_if(b *testing.B) {
	mm := make(map[int]interface{})
	for i := 0; i < b.N; i++ {
		st := &Stu1{}
		st.Id = i
		mm[i] = st
	}
	for i := 0; i < b.N; i++ {
		st, ok := mm[i]
		if ok {
			st.(*Stu1).Get()
		}
	}
}

func Benchmark_ifi(b *testing.B) {
	mm := make(map[int]IStu1)
	for i := 0; i < b.N; i++ {
		st := &Stu1{}
		st.Id = i
		mm[i] = st
	}
	for i := 0; i < b.N; i++ {
		st, ok := mm[i]
		if ok {
			st.Reset()
		}
	}
}
