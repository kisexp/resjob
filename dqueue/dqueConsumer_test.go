package dqueue

import (
	"fmt"
	"strconv"
	"sync"
	"testing"
	"time"
)

func DqueConsumer() {
	syncTicker := time.NewTicker(time.Second * 10000000)
	/*
		dq := disk.New("deque", "/opt/dque", 1024, 4, 1<<10, 2500, 2*time.Second,
			func(lvl disk.LogLevel, f string, args ...interface{}) {
				//fmt.Println(fmt.Sprintf(lvl.String()+": "+f, args...))
			})
	*/
	dqName := "test"
	tmpDir := "tmp"
	dq := New(dqName, tmpDir, 10, 4, 1<<10, 2500, 2*time.Second, 1*time.Second,
		func(lvl LogLevel, f string, args ...interface{}) {
			fmt.Println(fmt.Sprintf(lvl.String()+": "+f, args...))
		})

	go func() {
		for {
			select {
			case ms := <-dq.ReadChan():
				fmt.Println(">>>>>>>>>>>>>>>>>   " + string(ms))

				if ms == nil {
					dq.Close()
					dq = New(dqName, tmpDir, 10, 4, 1<<10, 2500, 2*time.Second, 1*time.Second,
						func(lvl LogLevel, f string, args ...interface{}) {
							fmt.Println(fmt.Sprintf(lvl.String()+": "+f, args...))
						})

					time.Sleep(time.Millisecond * 1000)
					//fmt.Println("read over !!!! exit!!!!!!!!!!!!!!!!!!")
					//defer wg.Done()
					//break
				}
			case <-syncTicker.C:
				dq.Close()
				time.Sleep(time.Millisecond * 2000)
				dq = New(dqName, tmpDir, 10, 4, 1<<10, 2500, 5*time.Second, 1*time.Second,
					func(lvl LogLevel, f string, args ...interface{}) {
						fmt.Println(fmt.Sprintf(lvl.String()+": "+f, args...))
					})

			}
		}
	}()
}

var wg sync.WaitGroup

func TestDqueueConsumer(t *testing.T) {

	dqName := "test"
	tmpDir := "tmp"
	dq := New(dqName, tmpDir, 10, 4, 1<<10, 2500, 2*time.Second, 1*time.Second,
		func(lvl LogLevel, f string, args ...interface{}) {
			fmt.Println(fmt.Sprintf(lvl.String()+": "+f, args...))
		})


	go func() {
		wg.Add(1)
		for i := 0; i < 1000; i++ {
			dq.Put([]byte("hello worker," + strconv.Itoa(i) + "\n"))
			time.Sleep(time.Millisecond * 10)
		}
		defer wg.Done()
	}()
	time.Sleep(time.Millisecond * 1000)

	//cnt := 0

	///*
	go func() {
		wg.Add(1)
		cnt := 0
		for {
			select {
			case ms := <-dq.ReadChan():
				cnt++
				fmt.Println("<<<<<<<<<<<<<<<<<<<<<< " + string(ms))
			}
		}
		defer wg.Done()
	}()
	//*/

	//wg.Add(1)
	//go DqueConsumer()
	time.Sleep(time.Millisecond * 1000)
	wg.Wait()
	fmt.Print("end")
}