package main

import (
	"fmt"
	"time"
)

func main() {
	ch1 := make(chan string, 1)

	go func() {
		time.Sleep(time.Second * 1)
		ch1 <- "1"
	}()

	select {
	case res := <-ch1:
		fmt.Println(res)
	case <-time.After(time.Second * 1):
		fmt.Println("time out 1")
	}

	ch2 := make(chan string, 1)
	go func() {
		time.Sleep(time.Second * 2)
		ch2 <- "2"
	}()

	select {
	case res := <-ch2:
		fmt.Println(res)
	case <-time.After(time.Second * 2):
		fmt.Println("time out 2")


	}

	time.Sleep(time.Second * 3)
}
