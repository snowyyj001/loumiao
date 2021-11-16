package util

import (
	"reflect"
	"testing"
)

/*
go test -v -bench="." -benchtime=1s reflect_test.go
输出如下：可见创建一个Stu需要200ns
goos: windows
goarch: amd64
cpu: Intel(R) Core(TM) i7-4790 CPU @ 3.60GHz
Benchmark_NewReflect
Benchmark_NewReflect-8           8699092               209.9 ns/op
Benchmark_New
Benchmark_New-8                 32040946                37.24 ns/op
PASS
ok      command-line-arguments  4.294s
*/

type Stu struct {
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

var baseSt Stu
var baseStPtr *Stu

func InitStu(st *Stu) {
	st.Id = 12345
	st.Buff = make([]byte, 10)
	st.Name = "abcd"
}

func Benchmark_NewReflect(b *testing.B) {
	baseStPtr = &baseSt
	for i := 0; i < b.N; i++ {
		pt := reflect.TypeOf(baseStPtr).Elem()
		packet := reflect.New(pt).Interface()
		InitStu(packet.(*Stu))
	}
}

func Benchmark_New(b *testing.B) {
	for i := 0; i < b.N; i++ {
		st := baseSt
		InitStu(&st)
	}
}
