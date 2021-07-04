package myqueue

import (
	"runtime"
	"sync"
	"sync/atomic"
	"time"
	"twist/workpool/queue"
)

// 队列结构
type MyQueue struct {
	sync.Mutex
	popable *sync.Cond
	buffer  *queue.Queue
	closed  bool
	count   int32
	cc      chan interface{}
	once    sync.Once
}

// 创建队列
func New() *MyQueue {
	ch := &MyQueue{
		buffer: queue.New(),
	}
	ch.popable = sync.NewCond(&ch.Mutex)
	return ch
}

// 队列长度
func (this *MyQueue) Len() int {
	return (int)(atomic.LoadInt32(&this.count))
}

// Pop 取出队列，（阻塞模式）
func (this *MyQueue) Pop() (v interface{}) {
	c := this.popable
	this.Mutex.Lock()
	defer this.Mutex.Unlock()

	for this.Len() == 0 && !this.closed {
		c.Wait()
	}
	if this.closed { // 已关闭
		return
	}
	if this.Len() > 0 {
		buffer := this.buffer
		v = buffer.Peek()
		buffer.Remove()
		atomic.AddInt32(&this.count, -1)
	}
	return
}

// TryPop 试着取出队列（非阻塞模式）返回ok == false 表示空
func (this *MyQueue) TryPop() (v interface{}, ok bool) {
	buffer := this.buffer

	this.Mutex.Lock()
	defer this.Mutex.Unlock()

	if this.Len() > 0 {
		v = buffer.Peek()
		buffer.Remove()
		atomic.AddInt32(&this.count, -1)
		ok = true
	} else if this.closed {
		ok = true
	}
	return
}

// Pop 取出队列（阻塞模式）
func (this *MyQueue) popChan(v *chan interface{}) {
	c := this.popable
	this.Mutex.Lock()
	defer this.Mutex.Unlock()

	for this.Len() == 0 && !this.closed {
		c.Wait()
	}
	if this.closed { // 已关闭
		*v <- nil
		return
	}
	if this.Len() > 0 {
		buffer := this.buffer
		tmp := buffer.Peek()
		buffer.Remove()
		atomic.AddInt32(&this.count, -1)
		*v <- tmp
	} else {
		*v <- nil
	}
}

// TryPopTimeout 试着取出队列（塞模式+timeout）返回ok == false 表示超时
func (this *MyQueue) TryPopTimeout(tm time.Duration) (v interface{}, ok bool) {
	this.once.Do(func() {
		this.cc = make(chan interface{}, 1)
	})
	go func() {
		this.popChan(&this.cc)
	}()

	ok = true
	timeout := time.After(tm)
	select {
	case v = <-this.cc:
	case <-timeout:
		if !this.closed {
			this.popable.Signal()
		}
		ok = false
	}
	return
}

// Push插入列队，非阻塞
func (this *MyQueue) Push(v interface{}) {
	this.Mutex.Lock()
	defer this.Mutex.Unlock()
	if !this.closed {
		this.buffer.Add(v)
		atomic.AddInt32(&this.count, 1)
		this.popable.Signal()
	}

}

//关闭我的队列
//关闭后，Pop返回无阻塞0，TryPop返回v =0，ok=True
func (this *MyQueue) Close() {
	this.Mutex.Lock()
	defer this.Mutex.Unlock()
	if !this.closed {
		this.closed = true
		atomic.StoreInt32(&this.count, 0)
		this.popable.Broadcast() // 广播
	}
}

func (this *MyQueue) IsClose() bool {
	return this.closed
}

func (this *MyQueue) Wait() {
	for {
		if this.closed || this.Len() == 0 {
			break
		}
		runtime.Gosched() // 出让时间片
	}
}
