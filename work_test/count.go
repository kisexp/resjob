package main

import (
	"fmt"
	"time"
)

func main() {
	ticker := time.NewTicker(time.Second)
	go func() {
		for t := range ticker.C {
			fmt.Println("Tick at ", t)
		}
	}()

	time.Sleep(time.Second * 5)
	ticker.Stop()
	fmt.Println("ticker stopped")
}
