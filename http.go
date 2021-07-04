package main

import (
	"fmt"
	"sync"
	"time"
)

type SimplePool struct {
	wg   sync.WaitGroup
	work chan func() // 任务队列
}

func NewPool(workers int) *SimplePool {
	p := &SimplePool{
		wg:   sync.WaitGroup{},
		work: make(chan func()),
	}
	p.wg.Add(workers)
	// 根据指定的并发量去读取管道并执行
	for i := 0; i < workers; i++ {
		go func() {
			defer func() {
				if err := recover(); err != nil {
					fmt.Println(err)
					p.wg.Done()
				}
			}()
			// 从workChannel中取出任务执行
			for fn := range p.work {
				fn()
			}
			p.wg.Done()
		}()
	}
	return p
}

// 添加任务
func (p *SimplePool) Add(fn func()) {
	p.work <- fn
}

// 执行
func (p *SimplePool) Run() {
	close(p.work)
	p.wg.Wait()
}

func parseTask(i int) func() {
	return func() {
		// 模拟抓取数据的过程
		time.Sleep(time.Second * 1)
		fmt.Println("finish parse ", i)
	}
}
func main() {
	p := NewPool(20)
	for i := 0; i < 100; i++ {
		p.Add(parseTask(i))
	}
	p.Run()
}
