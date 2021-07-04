package main

import (
	"fmt"
	"sync"
	"time"
)

func main() {

	ls := []string{
		"1111",
		"2222",
		"3333",
		"4444",
		"5555",
		"6666",
		"7777",
		"8888",
		"9999",
		"10101010",
		"12121212",
		"1313131313",
		"1414141414",
		"1515151151515",
		"161616161616",
		"1717171717",
		"1818181818",
		"191919191919",
		"202020202020",
		"212121212121",
		"222222222222",
	}

	wg := sync.WaitGroup{}
	wg.Add(len(ls))
	sem := make(chan int, 2)
	// 每次2个goroutine一起做，做完继续做下一组
	for _, v := range ls {
		s := v
		sem <- 1
		go func() {
			syncProd(s)
			<-sem
			wg.Done()
		}()

	}
	wg.Wait()
	fmt.Println("end")

}

func syncProd(v string) {
	time.Sleep(10 * time.Second)
	fmt.Println(v)
}
