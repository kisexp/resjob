package main

import (
	"fmt"
	"time"
)

func main() {
	ch1 := make(chan string, 1)
	ch2 := make(chan string, 2)
	go func() {
		time.Sleep(time.Second * 2)
		ch1 <- "one"
	}()

	go func() {
		time.Sleep(time.Second)
		ch2 <- "two"
	}()

	select {
	case msg1 := <-ch1:
		fmt.Println(msg1)
	case msg2 := <-ch2:
		fmt.Println(msg2)

	}

	time.Sleep(10 * time.Second)
}
