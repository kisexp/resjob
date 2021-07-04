package main

import (
	"fmt"
	"math/rand"
	"sync"
	"time"
)

func main() {
	var state = make(map[int]int)
	var mutex = &sync.Mutex{}

	for w := 0; w < 10; w++ {
		go func() {
			for {
				key := rand.Intn(5)
				val := rand.Intn(100)
				mutex.Lock() //  加锁
				state[key] = val
				mutex.Unlock() // 解锁
			}
		}()
	}

	time.Sleep(time.Second)
	fmt.Println("state:", state)
}
