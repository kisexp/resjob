package queue

/**
包队列基于Dariusz Górecki建议的版本提供了一个快速的环形缓冲队列。
使用这个代替其他更简单的队列实现(切片+追加或链表)提供了
显著的内存和时间优势，以及更少的垃圾收集暂停。
这里实现的队列之所以如此之快，还有一个原因:它不是线程安全的。
*/

// minQueueLen是队列可能具有的最小容量。
//按位模数必须是2的幂:x % n == x & (n - 1)。
const minQueueLen = 16

//队列表示队列数据结构的单个实例。
type Queue struct {
	buf []interface{}
	// 头，尾，个数
	head, tail, count int
}

//新构造并返回一个新队列。
func New() *Queue {
	return &Queue{
		buf: make([]interface{}, minQueueLen),
	}
}

// Length返回当前存储在队列中的元素数。
func (this *Queue) Length() int {
	return this.count
}

//调整队列大小，使其正好是当前内容的两倍
//如果队列不到半满，这可能会导致收缩
func (this *Queue) resize() {
	newBuf := make([]interface{}, this.count<<1)

	if this.tail > this.head {
		copy(newBuf, this.buf[this.head:this.tail])
	} else {
		n := copy(newBuf, this.buf[this.head:])
		copy(newBuf[n:], this.buf[:this.tail])
	}
	this.head = 0
	this.tail = this.count
	this.buf = newBuf
}

// Add将一个元素放在队列的末尾。
func (this *Queue) Add(elem interface{}) {
	if this.count == len(this.buf) {
		this.resize()
	}
	this.buf[this.tail] = elem
	// 按位模数
	this.tail = (this.tail + 1) & (len(this.buf) - 1)
	this.count++
}

// Peek返回队列头的元素。这里恐慌
// 如果队列为空。
func (this *Queue) Peek() interface{} {
	if this.count <= 0 {
		panic("queue: Peek() called no empty queue")
	}
	return this.buf[this.head]
}

// Get返回队列中索引I处的元素。如果索引为
// 无效，调用会死机。此方法同时接受正数和负数
// 负索引值。索引0引用第一个元素，并且
// index -1指最后一个。
func (this *Queue) Get(i int) interface{} {
	//如果向后索引，则转换为正索引。
	if i < 0 {
		i += this.count
	}
	if i < 0 || i >= this.count {
		panic("queue: Get() called with index out of range")
	}
	// 按位模数
	return this.buf[(this.head+i)&(len(this.buf)-1)]
}

// Remove从队列的前面移除并返回元素。如果
// 队列为空，调用会死机。
func (this *Queue) Remove() interface{} {
	if this.count <= 0 {
		panic("queue: Remove() called on empty queue")
	}
	ret := this.buf[this.head]
	this.buf[this.head] = nil
	//按位模数
	this.head = (this.head + 1) & (len(this.buf) - 1)
	this.count--
	//如果缓冲区已满1/4，则向下调整大小。
	if len(this.buf) > minQueueLen && (this.count<<2) == len(this.buf) {
		this.resize()
	}
	return ret
}
