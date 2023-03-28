package lutil

import (
	"fmt"
	"sync"
	"testing"
)

/*
go test -v -bench="." -benchtime=3s chantrans_test.go
goos: windows
goarch: amd64
cpu: Intel(R) Core(TM) i7-4790 CPU @ 3.60GHz
Benchmark_ST
0
99
9999
999999
31708561
Benchmark_ST-8          31708562               111.7 ns/op
Benchmark_PTR
0
99
9999
999999
22675421
Benchmark_PTR-8         22675422               156.2 ns/op
PASS
ok      command-line-arguments  7.437s

经以上测试可知，chan传递结构体效率更高
再使用pprof查看，chan传递结构体使用的内存峰值和heap大小也更低，gc更少，因为没有发生内存逃逸
*/

type Stu0 struct {
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

func Benchmark_ST(b *testing.B) {
	var wg sync.WaitGroup
	var ch = make(chan Stu, 100)
	wg.Add(1)
	go func() {
		for i := 0; i < b.N; i++ {
			st := Stu{}
			st.Id = i
			ch <- st
		}
		wg.Done()
	}()
	wg.Add(1)
	go func() {
		var cnt int
		for i := 0; i < b.N; i++ {
			st := <-ch
			cnt = st.Id
		}
		wg.Done()
		fmt.Println(cnt)
	}()
	wg.Wait()
}

func Benchmark_PTR(b *testing.B) {
	var wg sync.WaitGroup
	var ch = make(chan *Stu, 100)
	wg.Add(1)
	go func() {
		for i := 0; i < b.N; i++ {
			st := &Stu{}
			st.Id = i
			ch <- st
		}
		wg.Done()
	}()
	wg.Add(1)
	go func() {
		var cnt int
		for i := 0; i < b.N; i++ {
			st := <-ch
			cnt = st.Id
		}
		wg.Done()
		fmt.Println(cnt)
	}()
	wg.Wait()
}
