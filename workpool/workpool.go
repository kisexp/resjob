package workpool

import (
	"context"
	"runtime"
	"sync"
	"sync/atomic"
	"time"
	"twist/workpool/myqueue"
)

type TaskHandler func() error

type WorkPool struct {
	closed       int32
	isQueTask    int32 // 标记是否队列取出任务
	errChan      chan error
	timeout      time.Duration
	wg           sync.WaitGroup
	task         chan TaskHandler
	waitingQueue *myqueue.MyQueue
}

func New(max int) *WorkPool {
	if max < 1 {
		max = 1
	}

	this := &WorkPool{
		task:         make(chan TaskHandler, 2*max),
		errChan:      make(chan error, 1),
		waitingQueue: myqueue.New(),
	}

	go this.loop(max)
	return this
}

func (this *WorkPool) SetTimeout(timeout time.Duration) { // 设置超时时间
	this.timeout = timeout
}

func (this *WorkPool) Do(fn TaskHandler) { // 添加到工作池，并立即返回
	if this.IsClosed() {                   // 已关闭
		return
	}
	this.waitingQueue.Push(fn)

}

func (this *WorkPool) DoWait(task TaskHandler) { // 添加到工作池，并等待执行完成之后再返回
	if this.IsClosed() {
		return
	}
	doneChan := make(chan struct{})
	this.waitingQueue.Push(TaskHandler(func() error {
		defer close(doneChan)
		return task()
	}))
	<-doneChan
}

func (this *WorkPool) Wait() error { // 等待工作线程执行结束
	this.waitingQueue.Wait() // 等待队列结束
	this.waitingQueue.Close()
	this.waitTask()
	close(this.task)
	this.wg.Wait() // 等待结束
	select {
	case err := <-this.errChan:
		return err
	default:
		return nil
	}
}

func (this *WorkPool) IsClosed() bool {
	if atomic.LoadInt32(&this.closed) == 1 {
		return true
	}
	return false
}

func (this *WorkPool) startQueue() {
	this.isQueTask = 1
	for {
		tmp := this.waitingQueue.Pop()
		if this.IsClosed() { // closed
			this.waitingQueue.Close()
			break
		}
		if tmp != nil {
			fn := tmp.(TaskHandler)
			if fn != nil {
				this.task <- fn
			}
		} else {
			break
		}
	}
	atomic.StoreInt32(&this.isQueTask, 0)
}

func (this *WorkPool) IsDone() bool { // 判断是否完成（非阻塞）
	if this == nil || this.task == nil {
		return true
	}
	return this.waitingQueue.Len() == 0 && len(this.task) == 0
}

func (this *WorkPool) waitTask() {
	for {
		runtime.Gosched()
		if this.IsDone() {
			if atomic.LoadInt32(&this.isQueTask) == 0 {
				break
			}
		}
	}
}

func (this *WorkPool) loop(maxWorkersCount int) {
	go this.startQueue()         //  启动队列
	this.wg.Add(maxWorkersCount) // 最大的工作协程数

	for i := 0; i < maxWorkersCount; i++ {
		go func() {
			defer this.wg.Done()
			// 开始干活
			for fn := range this.task {
				if fn == nil || atomic.LoadInt32(&this.closed) == 1 { // 有err 立即返回
					continue // 需要先消费完了之后再返回
				}
				closed := make(chan struct{}, 1)
				if this.timeout > 0 { // 有设置超时,优先task 的超时
					ct, cancel := context.WithTimeout(context.Background(), this.timeout)
					go func() {
						select {
						case <-ct.Done():
							this.errChan <- ct.Err()
							atomic.StoreInt32(&this.closed, 1)
							cancel()
						case <-closed:

						}
					}()
				}

				err := fn()
				close(closed)
				if err != nil {
					select {
					case this.errChan <- err:
						atomic.StoreInt32(&this.closed, 1)
					default:

					}
				}
			}
		}()
	}
}
