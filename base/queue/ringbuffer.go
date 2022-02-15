package queue

//无锁RingBuffer，单读单写协程安全
//IsFull和IsEmpty两个函数充分说明了无锁环形队列的核心思想

type RingBuffer struct {
	buf               []interface{}
	head, tail, count, size int
}

// New constructs and returns a new RingBuffer.
func NewRing(ringSize int) *RingBuffer {
	return &RingBuffer{
		size: ringSize,
		buf: make([]interface{}, ringSize),
		head: 0,
		tail: 1,
	}
}


// Length returns the number of elements currently stored in the queue.
func (q *RingBuffer) Length() int {
	return q.count
}


// whether the ring is full
func (q *RingBuffer) IsFull() bool {
	return q.tail + 1 - q.size == q.head
}

// whether the ring is empty
func (q *RingBuffer) IsEmpty() bool {
	return q.tail + 1 == q.head
}

// Add puts an element on the end of the queue.
func (q *RingBuffer) Push(elem interface{}) bool {
	if q.IsFull() {
		return false
	}

	q.buf[(q.tail+1)%q.size] = elem
	q.tail++
	q.count++

	return true
}

// Peek returns the element at the head of the queue. This call return nil
// if the queue is empty.
func (q *RingBuffer) Peek() interface{} {
	if q.count <= 0 {
		return nil
	}
	return q.buf[q.head]
}

// Pop removes and returns the element from the front of the queue. If the
// queue is empty, the call do nothing
func (q *RingBuffer) Pop() interface{} {
	if q.IsEmpty() {
		return nil
	}

	elm := q.buf[q.head%q.size]
	q.head++
	q.count--

	return elm
}
