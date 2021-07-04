package workpool

import (
	"errors"
	"fmt"
	"testing"
	"time"
)

// 支持最大任务数, 放到工作池里面 并等待全部完成
// 支持超时退出
func TestWorkerPoolStart(t *testing.T) {
	wp := New(10)
	wp.SetTimeout(time.Millisecond)
	for i := 0; i < 20; i++ {
		ii := i
		wp.Do(func() error {
			for j := 0; j < 10; j++ {
				fmt.Println(fmt.Sprintf("%v->\t%v", ii, j))
				time.Sleep(1 * time.Millisecond)
			}
			return nil
		})
	}

	err := wp.Wait()
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println("down")
}

// 支持错误返回
func TestWorkerPoolError(t *testing.T) {
	wp := New(10)
	for i := 0; i < 20; i++ {
		ii := i
		wp.Do(func() error {
			for j := 0; j < 10; j++ {
				fmt.Println(fmt.Sprintf("%v->\t%v", ii, j))
				if ii == 1 {
					return errors.New("my test err")
				}
				time.Sleep(1 * time.Millisecond)
			}
			return nil
		})
	}
	err := wp.Wait()
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println("down")
}

// 确定完成(非阻塞)是否放置在工作工具中，并等待执行结果
// 支持同步等待结果
func TestWorkerPoolDoWait(t *testing.T) {
	wp := New(5)
	for i := 0; i < 10; i++ {
		ii := i
		wp.DoWait(func() error {
			for j := 0; j < 5; j++ {
				fmt.Println(fmt.Sprintf("%v -> \t%v", ii, j))
				time.Sleep(1 * time.Millisecond)
			}
			return nil
		})
	}

	err := wp.Wait()
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println("down")
}

// 支持判断是否完成 (非阻塞)
func TestWorkerPoolIsDone(t *testing.T) {
	wp := New(5)
	for i := 0; i < 10; i++ {
		ii := i
		wp.Do(func() error {
			for j := 0; j < 5; j++ {
				fmt.Println(fmt.Sprintf("%v->\t%v", ii, j))
				time.Sleep(1 * time.Millisecond)
			}
			return nil
		})
		fmt.Println(wp.IsDone())
	}
	wp.Wait()
	fmt.Println(wp.IsDone())
	fmt.Println("down")
}
