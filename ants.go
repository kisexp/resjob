package main

import (
	"errors"
	"log"
	"runtime"
	"sync"
	"sync/atomic"
	"time"
)

type Worker struct {
	pool      *Pool       // worker 所属的pool
	task      chan func() // 任务队列
	releaseTs time.Time   // 回收时间，即该worker的最后一次结束的时间
}

var workerChanCap = func() int {
	if runtime.GOMAXPROCS(0) == 1 {
		return 0
	}
	return 1
}

type Pool struct {
	cap          int32             // 协程池的容量（groutine数量的上限）
	running      int32             // 正在执行中的groutine
	Ttl          time.Duration     // 过期清理间隔时间
	workers      []*Worker         // 当前可用空闲的groutine
	release      int32             // 表示pool是否关闭
	cond         *sync.Cond        // 用于控制pool等待可用的groutine
	once         *sync.Once        // 确保pool只被关闭一次
	workerCache  sync.Pool         // worker 临时对象池，在复用worker时减少新对象的创建并加速worker从pool中的获取速度
	PanicHandler func(interface{}) // pool引发panic时的执行函数
	lock         sync.Locker
}

func (p *Pool) incRunning() {
	atomic.AddInt32(&p.running, 1)
}

func (p *Pool) decRunning() {
	atomic.AddInt32(&p.running, -1)
}

func (p *Pool) Running() int {
	return int(atomic.LoadInt32(&p.running))
}

func (p *Pool) Cap() int {
	return int(atomic.LoadInt32(&p.cap))
}

func (w *Worker) run() {
	// pool中正在执行的worker数+1
	w.pool.incRunning()
	go func() {
		defer func() {
			if p := recover(); p != nil {
				// 若worker因各种问题引发panic,
				// pool中正在执行的worker数-1,
				// 如果设置了pool中的PanicHandler,此时会被调用
				w.pool.decRunning()
				if w.pool.PanicHandler != nil {
					w.pool.PanicHandler(p)
				} else {
					log.Printf("worker exits from a panic: %v", p)
				}
			}
		}()

		// worker执行任务队列
		for fn := range w.task {
			// 任务队列中的函数全部被执行完后
			// pool中正在执行的worker数 -1，
			// 将worker 放回对象池
			if fn == nil {
				w.pool.decRunning()
				w.pool.workerCache.Put(w)
			}
			fn()
			// worker 执行完任务后放回pool
			// 使得其余正在阻塞的任务可以获取worker
			w.pool.revertWorker(w)
		}
	}()
}

// 释放worker回pool
func (p *Pool) revertWorker(worker *Worker) {
	worker.releaseTs = time.Now()
	p.lock.Lock()
	p.workers = append(p.workers, worker)
	// 通知pool中已经获取锁的groutine，有一个worker已完成任务
	p.cond.Signal()
	p.lock.Unlock()
}

// 向pool提交任务
func (p *Pool) Submit(task func()) error {
	if 1 == atomic.LoadInt32(&p.release) {
		return errors.New("pool closed")
	}
	// 获取pool中的可用worker并向其他任务队列中写入任务
	p.retrieveWorker().task <- task
	return nil
}

// 获取可用worker
func (p *Pool) retrieveWorker() *Worker {
	var w *Worker
	p.lock.Lock()
	idleWorkers := p.workers
	n := len(idleWorkers) - 1
	// 当前pool中有可用worker，取出(队尾)worker并执行
	if n >= 0 {
		w = idleWorkers[n]
		idleWorkers[n] = nil
		p.workers = idleWorkers[:n]
		p.lock.Unlock()
	} else if p.Running() < p.Cap() {
		p.lock.Unlock()
		// 当前pool中无空闲worker,且pool数量未达到上线
		// pool会先从临时对象池中寻找是否有已完成任务的worker,
		// 若临时对象池中不存在，则重新创建一个worker并将其启动
		if cacheWorker := p.workerCache.Get(); cacheWorker != nil {
			w = cacheWorker.(*Worker)
		} else {
			w = &Worker{
				pool: p,
				task: make(chan func(), workerChanCap()),
			}
		}
		w.run()
	} else {
		// pool中没有空余worker且达到并发上限
		// 任务会阻塞等待当前运行的worker完成任务释放会pool
		for {
			p.cond.Wait() // 等待通知，暂时阻塞
			l := len(p.workers) - 1
			if l < 0 {
				continue
			}
			// 当有可用worker释放回pool之后，取出
			w = p.workers[l]
			p.workers[l] = nil
			p.workers = p.workers[:l]
			break
		}
		p.lock.Unlock()
	}
	return w
}

func NewPoolAnts(size int) (*Pool, error)  {
	if size < 1 {
		size = 1
	}

	var lc *sync.Mutex

	p := &Pool{
		cap: int32(size),
		lock: lc,
	}
	p.workerCache.New = func() interface{} {
		return &Worker{
			pool: p,
			task: make(chan func(), workerChanCap()),
		}
	}
	p.cond = sync.NewCond(p.lock)
	return p, nil

}
func main() {
}
