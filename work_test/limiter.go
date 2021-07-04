package main

import (
	"fmt"
	"time"
)

func main() {
	// 限制每2秒执行一次请求
	ticker := time.Tick(time.Second * 2)
	for i := 1; i <= 5; i++ {
		<-ticker
		fmt.Println("request:", i, time.Now())
	}

	// 先向burstylimiter push 3个值
	limiter := make(chan time.Time, 3)
	for i := 0; i < 3; i++ {
		limiter <- time.Now()
	}

	// 然后开启另外一个线程，每2秒向burstylimiter push 一个值
	go func() {
		for t := range time.Tick(time.Second * 2) {
			limiter <- t
		}
	}()

	// 最后实现效果，前三次没有限速，最后两次每2秒执行一次
	for i := 1; i <= 5; i++ {
		<-limiter
		fmt.Println("request", i, time.Now())
	}
}
