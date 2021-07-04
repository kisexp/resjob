package main

import "fmt"

func main() {
	jobs := make(chan int, 5)
	done := make(chan bool)

	go func() {
		for {
			j, ok := <-jobs
			if ok {
				fmt.Println("receive job ", j)
			} else {
				fmt.Println("receive all jobs")
				done <- true // 通知主程序，已经接受全部任务
				return
			}
		}
	}()

	for i := 1; i < 3; i++ {
		jobs <- i
		fmt.Println("send job", i)
	}
	close(jobs)
	fmt.Println("send all jobs")

	<-done // 等待通知
}
