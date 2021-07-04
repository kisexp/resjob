package zero

import (
	"net"
	"runtime"
	"sync"
	"time"
)

type ServeHandler func(c net.Conn) error

type workerChan struct {
	lastTime time.Time // 最后使用时间
	ch       chan net.Conn
}

type workerPool struct {
	WorkerFunc            ServeHandler  // 自定义处理的方法
	MaxWorkersCount       int           // 最大worker数量
	workersCount          int           // 当前worker数量
	MaxIdleWorkerDuration time.Duration // worker最大空闲时长，超过就被释放回收
	lock                  *sync.Mutex   // pool修改对象时需要用到的互斥锁
	Ready                 []*workerChan // 当前可用的workerchan 的对象池，避免频繁创建
	stopCh                chan struct{} // worker pool 停止信号
	mustStop              bool
	workerChanPool        sync.Pool // 缓存workerChan的对象池，避免频繁创建

}

var workerChanCap = func() int {
	if runtime.GOMAXPROCS(0) == 1 {
		return 0
	}
	return 1
}()

func (w *workerPool) MaxIdleDuration() time.Duration {
	if w.MaxIdleWorkerDuration <= 0 {
		return 10 * time.Second
	}
	return w.MaxIdleWorkerDuration
}

func (w *workerPool) Start() {
	if w.stopCh != nil {
		panic("workerPool already started")
	}
	w.stopCh = make(chan struct{})
	stopCh := w.stopCh
	w.workerChanPool.New = func() interface{} {
		return &workerChan{
			ch: make(chan net.Conn, workerChanCap),
		}
	}

	go func() {
		var chans = make([]*workerChan, 0)
		// 每隔一段时间清理一次
		ticker := time.NewTicker(w.MaxIdleDuration())
		for {
			select {
			case <-stopCh:
				return
			case <-ticker.C:
				w.Clean(&chans) // 清理

			}
		}
	}()
}

func (w *workerPool) Stop() {
	if w.stopCh == nil {
		panic("workerPool wasn't started")
	}
	close(w.stopCh)
	w.stopCh = nil
	w.lock.Lock()
	ready := w.Ready
	// 停止所有worker等待传入的连接
	// 不要等待忙碌的worker-他们将在工作后停下来
	// 服务连接并注意到wp.mustStop = true。
	for i := range ready {
		ready[i].ch <- nil
		ready[i] = nil
	}
	w.Ready = ready[:0]
	w.mustStop = true
	w.lock.Unlock()
}

func (w *workerPool) Clean(chans *[]*workerChan) {
	maxIdleDuration := w.MaxIdleDuration()
	beforeTime := time.Now().Add(-maxIdleDuration)
	w.lock.Lock()
	ready := w.Ready
	n := len(ready)
	l, r, mid := 0, n-1, 0
	for l <= r {
		mid = (l + r) / 2
		if beforeTime.After(w.Ready[mid].lastTime) {
			l = mid + 1
		} else {
			r = mid - 1
		}
	}
	i := r
	if i == -1 {
		w.lock.Unlock()
		return
	}
	*chans = append((*chans)[:0], ready[:i+1]...)
	m := copy(ready, ready[i+1:])
	for i = m; i < n; i++ {
		ready[i] = nil
	}
	w.Ready = ready[:m]
	w.lock.Unlock()
	clears := *chans
	for l := range clears {
		clears[l].ch <- nil
		clears[l] = nil
	}

}

// 处理http请求的入口
func (w *workerPool) Serve(c net.Conn) bool {
	// 获取workerChan，并将连接发送到chan中
	ch := w.getWorkerChan()
	if ch == nil {
		return false
	}
	ch.ch <- c
	return true
}

// 获取一个workerChan
func (w *workerPool) getWorkerChan() (ch *workerChan) {
	createWorker := false
	w.lock.Lock()
	ready := w.Ready
	if n := len(ready) - 1; n > 0 { // 如果ready队列长度大于0，取最后一个
		ch = ready[n]
		ready[n] = nil
		w.Ready = ready[:n] // 取出后将最后一个pop掉
	} else {
		if w.workersCount < w.MaxWorkersCount {
			createWorker = true // 允许新创建worker
			w.workersCount++
		}
	}
	w.lock.Unlock()

	if ch == nil {
		if !createWorker {
			return
		}
		wch := w.workerChanPool.Get() // 从对象池中取一个，避免重复创建
		ch = wch.(*workerChan)
		go func() {
			//ch和协程绑定
			w.workerFunc(ch)
			// 用完了返回去
			w.workerChanPool.Put(wch)
		}()
	}
	return
}

func (w *workerPool) workerFunc(ch *workerChan) {
	var (
		c   net.Conn
		err error
	)
	for c = range ch.ch {
		if c == nil {
			break
		}
		err = w.WorkerFunc(c)
		if err != nil {
			_ = c.Close()
		}
		c = nil

		if !w.Release(ch) { // 最终 release 这个channel，将其重新放进ready中
			break
		}
	}
	w.lock.Lock()
	w.workersCount--
	w.lock.Unlock()
}

func (w *workerPool) Release(ch *workerChan) bool {
	ch.lastTime = time.Now()
	w.lock.Lock()
	if w.mustStop {
		w.lock.Unlock()
		return false
	}
	// 往尾部追加, 头部最早过期待清理
	// 当协程完成工作后，就会把workerChan放回Slice尾部，以待其他请求使用
	w.Ready = append(w.Ready, ch) // 重新放入ready队列
	w.lock.Unlock()
	return true
}
