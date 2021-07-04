package limiter

import (
	"fmt"
	"testing"
	"time"
)

func TestAfter(t *testing.T) {
	duration := time.Duration(6) * time.Second
	ch := make(chan int)

	_ = time.AfterFunc(duration, func() {
		go func() {
			fmt.Println("6 seconds over......")
			ch <- 30
		}()
	})

	fmt.Println("ing....")

	for {
		select {
		case n := <-ch:
			fmt.Println(n, "is arriving")
			fmt.Println("Done!")
			return
		default:
			fmt.Println("time to wait")
			time.Sleep(3 * time.Second)
		}
	}

}
