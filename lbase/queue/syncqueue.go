package queue

import (
	"sync"
)

/*
Package queue provides a fast, ring-buffer queue based on the version suggested by Dariusz Górecki.
Using this instead of other, simpler, queue implementations (slice+append or linked list) provides
substantial memory and time benefits, and fewer GC pauses.
The queue implemented here is thread-safe
*/

// Queue represents a single instance of the queue data structure.
type SyncQueue struct {
	buf               []interface{}
	head, tail, count int
	locker            sync.Mutex
}

// New constructs and returns a new Queue.
func NewSync() *SyncQueue {
	return &SyncQueue{
		buf: make([]interface{}, minQueueLen),
	}
}

// Length returns the number of elements currently stored in the queue.
func (q *SyncQueue) Length() int {
	return q.count
}

// resizes the queue to fit exactly twice its current contents
// this can result in shrinking if the queue is less than half-full
func (q *SyncQueue) resize() {
	newBuf := make([]interface{}, q.count<<1)

	if q.tail > q.head {
		copy(newBuf, q.buf[q.head:q.tail])
	} else {
		n := copy(newBuf, q.buf[q.head:])
		copy(newBuf[n:], q.buf[:q.tail])
	}

	q.head = 0
	q.tail = q.count
	q.buf = newBuf
}

// Add puts an element on the end of the queue.
func (q *SyncQueue) Add(elem interface{}) int {
	defer q.locker.Unlock()
	q.locker.Lock()
	if q.count == len(q.buf) {
		q.resize()
	}

	q.buf[q.tail] = elem
	// bitwise modulus
	q.tail = (q.tail + 1) & (len(q.buf) - 1)
	q.count++
	return q.count
}

// Peek returns the element at the head of the queue. This call panics
// if the queue is empty.
func (q *SyncQueue) Peek() interface{} {
	defer q.locker.Unlock()
	q.locker.Lock()
	if q.count <= 0 {
		panic("queue: Peek() called on empty queue")
	}
	return q.buf[q.head]
}

// Get returns the element at index i in the queue. If the index is
// invalid, the call will panic. This method accepts both positive and
// negative index values. Index 0 refers to the first element, and
// index -1 refers to the last.
func (q *SyncQueue) Get(i int) interface{} {
	defer q.locker.Unlock()
	q.locker.Lock()
	// If indexing backwards, convert to positive index.
	if i < 0 {
		i += q.count
	}
	if i < 0 || i >= q.count {
		panic("queue: Get() called with index out of range")
	}
	// bitwise modulus
	return q.buf[(q.head+i)&(len(q.buf)-1)]
}

// Remove removes and returns the element from the front of the queue. If the
// queue is empty, the call will panic.
func (q *SyncQueue) Remove() (interface{}, int) {
	defer q.locker.Unlock()
	q.locker.Lock()
	if q.count <= 0 {
		panic("queue: Remove() called on empty queue")
	}
	ret := q.buf[q.head]
	q.buf[q.head] = nil
	// bitwise modulus
	q.head = (q.head + 1) & (len(q.buf) - 1)
	q.count--
	// Resize down if buffer 1/4 full.
	if len(q.buf) > minQueueLen && (q.count<<2) == len(q.buf) {
		q.resize()
	}
	return ret, q.count
}
