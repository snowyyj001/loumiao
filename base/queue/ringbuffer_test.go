package queue_test

import (
	"fmt"
	"github.com/snowyyj001/loumiao/base/queue"
	"sync"
	"testing"
	"time"
)

/*
go test -v ringbuffer_test.go
*/

func TestRingBuffer(t *testing.T)  {
	ring := queue.NewRing(1000)

	que := queue.New()

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		for i:=0; i<2000; i++ {
			ring.Push(i)
		}
		time.Sleep(time.Millisecond)
		wg.Done()
	}()

	wg.Add(1)
	go func() {
		lastnum := -1
		for i:=0; i<2000; i++ {
			elm, ok := ring.Pop()
			if ok {
				if elm.(int) != lastnum+1 {
					t.FailNow()
				} else {
					lastnum = elm.(int)
				}
				que.Add(elm)
			}
			time.Sleep(time.Millisecond)
		}
		wg.Done()
	}()

	wg.Wait()

	fmt.Println("total size = ", que.Length())
	for que.Length() > 0 {
		fmt.Print(" ", que.Peek())
		que.Remove()
	}

}


