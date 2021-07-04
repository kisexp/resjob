package main

import (
	"fmt"
	"time"
)

func worker(id int, jobs <-chan int, results chan<- int) {
	for j := range jobs {
		fmt.Println("worker", id, "process job", j)
		time.Sleep(time.Second * 2)
		results <- j
	}
}

func main() {
	jobs := make(chan int, 100)
	results := make(chan int, 100)

	// 开启5个进程
	for w := 1; w <= 5; w++ {
		go worker(w, jobs, results)
	}

	// 向通道push任务
	for j := 1; j <= 9; j++ {
		jobs <- j
	}

	close(jobs)

	// result作用是告知主进程执行结束，当所有的执行结束后，主进程结束退出任务，如果没有result可能会导致子进程还没有结束，主进程就退出了
	for r := 1; r <= 9; r++ {
		<-results
	}

}
