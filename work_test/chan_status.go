package main

import (
	"fmt"
	"math/rand"
	"time"
)

type readOp struct {
	key  int
	resp chan int
}

type writeOp struct {
	key  int
	val  int
	resp chan bool
}
// go状态协程

func main() {
	reads := make(chan *readOp)
	writes := make(chan *writeOp)

	// 通过select选择来保证同时只能读或写操作
	go func() {
		var state = make(map[int]int)
		for {
			select {
			case read := <-reads:
				//read.resp <- state[read.key]
				read.resp <- 1
				fmt.Println("read.key -> ", read.key)
			case writes := <-writes:
				fmt.Println("writes.key -> ", writes.key, "writes.val -> ", writes.val)
				state[writes.key] = writes.val
				writes.resp <- true

			}
		}
	}()

	for r := 0; r < 100; r++ {
		go func() {
			for true {
				read := &readOp{
					key:  rand.Intn(5),
					resp: make(chan int),
				}
				reads <- read
				fmt.Println(read)
				<-read.resp
			}
		}()
	}

	for w := 0; w < 10; w++ {
		go func() {
			for {
				write := &writeOp{
					key:  rand.Intn(5),
					val:  rand.Intn(100),
					resp: make(chan bool),
				}
				writes <- write
				fmt.Println(write)
				<-write.resp
			}
		}()
	}
	// 通过select选择来保证同时只有一个读或写操作，这样比通过互斥锁复杂
	time.Sleep(time.Second*100)
}
