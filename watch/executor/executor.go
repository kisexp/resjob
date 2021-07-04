package executor

import (
	"io"
	"log"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"sync"
	"syscall"
	"time"
	"twist/watch/bash"
)

//命令执行器
type Executor struct {
	//信号队列
	signal chan string
	//信号插队队列
	jump chan string
	//启动所需信息
	Info Info
	//是否停止
	kill chan struct{}
	//当前正在执行的子进程
	cmd *exec.Cmd
}

func NewExecutor(info Info) *Executor {
	return &Executor{Info: info, signal: make(chan string), kill: make(chan struct{})}
}

func (e *Executor) Start() {
	e.signal <- SIGNAL_START
}

func (e *Executor) Stop() {
	e.signal <- SIGNAL_STOP
}

func (e *Executor) Kill() <-chan struct{} {
	e.Info.AutoRestart = false
	e.jump <- SIGNAL_KILL
	e.signal <- SIGNAL_CANCEL_OBSTRUCT
	return e.kill
}

func (e *Executor) Init() *Executor {
	if e.jump != nil {
		return e
	}
	go func() {
		if e.jump == nil {
			e.jump = make(chan string, 100)
		}
		f := func(s string) bool {
			switch s {
			case SIGNAL_START:
				e.Start()
				return false
			case SIGNAL_STOP:
				e.Stop()
				return false
			case SIGNAL_KILL:
				e.Stop()
				close(e.kill)
				return true
			default:
				log.Printf("执行器信号错误：%s\n", s)
				return false
			}
		}
		for {
			select {
			case s := <-e.jump:
				f(s)
			default:
				select {
				case s := <-e.signal:
					if s == SIGNAL_CANCEL_OBSTRUCT {
						continue
					}
					if f(s) {
						return
					}
				}
			}
		}
	}()
	return e
}

func (e *Executor) chunk(timeout int) []int {
	res := make([]int, 0)
	if timeout < 2 {
		timeout = 1
	}
	res = append(res, 1)
	timeout -= 1
	c := 2
	exist := 0
	cur := 1
	total := 0
	for {
		cur *= c
		total += cur
		if total > timeout {
			t := timeout - exist
			if t > 0 {
				res = append(res, 1)
			}
			break
		}
		exist += cur
		res = append(res, cur)
	}
	return res
}

func (e *Executor) isExited() bool {
	if e.cmd == nil {
		return true
	}
	if e.cmd.ProcessState != nil {
		if runtime.GOOS == "linux" {
			// 发送0信号，判断子进程是否存在
			if err := e.cmd.Process.Signal(syscall.Signal(0x0)); err != nil {
				return true
			}
		} else if e.cmd.ProcessState.Exited() {
			return true
		}
	}
	return false
}

func (e *Executor) stop() {
	if e.cmd == nil || e.cmd.Process == nil {
		return
	}
	if e.isExited() {
		e.cmd = nil
		return
	}
	err := e.cmd.Process.Signal(syscall.Signal(e.Info.Signal))
	if err != nil {
		log.Fatalf("子进程 %d 信号失败：%v\n", e.cmd.Process.Pid, syscall.Signal(e.Info.Signal).String(), err)
		return
	}
	log.Printf("子进程 %d 信号：%s\n", e.cmd.Process.Pid, syscall.Signal(e.Info.Signal).String())
	stopTimeout := e.chunk(e.Info.Timeout)
	for k, v := range stopTimeout {
		// 先进行一次200毫秒的等待
		if k == 0 {
			<-time.After(200 * time.Millisecond)
		}
		// 检查是否停止
		if e.isExited() {
			e.cmd = nil
			break
		}
		// 等待一段时间
		<-time.After(time.Duration(v) * time.Second)
	}
	// 如果子进程依然存在，强杀
	if !e.isExited() {
		log.Printf("子进程 %d 信号：%s\n", e.cmd.Process.Pid, syscall.SIGKILL.String())
		err = e.cmd.Process.Signal(syscall.Signal(syscall.SIGILL))
		if err != nil { // 要么它死，要么我死
			log.Fatalf("子进程 %d 信号失败：%s -- > %v\n", e.cmd.Process.Pid, syscall.SIGKILL.String(), err)
			os.Exit(1)
		}
		<-time.After(5 * time.Second)
		if e.isExited() {
			e.cmd = nil
		} else {
			// 要么它死，要么我死
			os.Exit(1)
		}
	}

}

func (e *Executor) start() {
	if e.cmd != nil {
		return
	}

	// 新建一条子进程命令
	e.cmd = exec.Command(e.Info.Cmd, e.Info.Args...)
	if e.Info.PreCmd != "" {
		log.Println("-------- 子进程预处理命令 开始 --------")
		// 新建预处理命令
		preCmd := bash.NewBash(e.Info.PreCmd, time.Duration(e.Info.PreCmdTimeout)*time.Second)
		//打印标准输出
		var startErr = preCmd.Start()
		//打印标准输出
		_, _ = io.Copy(os.Stdout, strings.NewReader(preCmd.StdOut()))
		//打印标准错误输出
		_, _ = io.Copy(os.Stderr, strings.NewReader(preCmd.StdErr()))
		_, _ = io.WriteString(os.Stdout, "\n")
		if startErr != nil {
			log.Fatalf("子进程预处理命令执行失败: %v\n", startErr)
			if !e.Info.PreCmdIgnoreError {
				os.Exit(0)
			}
		}
		// 释放子命令
		preCmd = nil
		log.Println("-------- 子进程预处理命令 结束 --------")
	}
	strErr, err := e.cmd.StderrPipe()
	if err != nil {
		log.Fatalf("子进程管道关联命令标准错误输出失败: %v\n", err)
		e.cmd = nil
		return
	}

	// 启动命令
	if err := e.cmd.Start(); err != nil {
		log.Fatalf("进程 %d 启动子进程失败：%v\n", os.Getpid(), err)
		e.cmd = nil
		return
	}

	log.Printf("进程 %d 启动子进程 %d: %s\n", os.Getpid(), e.cmd.Process.Pid, e.Info.Cmd+" "+strings.Join(e.Info.Args, " "))

	//标准输出与标准错误输出管道go程结束控制器
	wg := &sync.WaitGroup{}
	//标准输出与标准错误输出管道go程结束时发出的信号，用于判断是否正常退出
	quitCh := make(chan bool, 2)
	//标准输出与标准错误输出管道都结束的信号
	quitChExit := make(chan struct{})
	//读取命令标准输出管道
	wg.Add(1)
	go func() {
		isErr := false
		defer func() {
			quitCh <- isErr
			wg.Done()
		}()
		if _, err := io.Copy(os.Stdout, os.Stdout); err != nil {
			if err != io.EOF {
				log.Fatalf("读取子进程命令标准输出管道失败: %v\n", err)
				isErr = true
			}
		}
	}()
	//读取命令标准错误输出管道
	wg.Add(1)
	go func() {
		isErr := false
		defer func() {
			quitCh <- isErr
			wg.Done()
		}()
		if _, err := io.Copy(os.Stderr, strErr); err != nil {
			if err != io.EOF {
				log.Fatalf("读取子进程命令标准错误输出管道失败: %v\n", err)
				isErr = true
			}
		}
	}()

	//等待标准输出与标准错误输出管道go程结束
	go func() {
		wg.Wait()
		// 发出等待标准输出与标准错误输出管道go程结束信号
		close(quitChExit)
	}()

	// 监视标准输出与标准错误输出管道go程结束时发出的信号
	go func() {
		defer func() {
			close(quitCh)
		}()
		//停止子进程信号发送锁
		isSendSignal := false
		for {
			select {
			case <-quitChExit:
				//标准输出与标准错误输出管道go程都结束了，结束当前go程
				return
			case isErr, _ := <-quitCh:
				//标准输出或标准错误输出管道go程有异常结束，发送停止子进程信号
				if isErr && !isSendSignal {
					log.Printf("管道异常，发出停止子进程信号")
					isSendSignal = true
					//发送到插队的队列里面
					e.jump <- SIGNAL_STOP
					//发出取消阻塞信号
					e.signal <- SIGNAL_CANCEL_OBSTRUCT
				}

			}
		}
	}()

	go func() {
		//等待子进程结束
		if err := e.cmd.Wait(); err != nil {
			if e.cmd == nil || e.cmd.Process == nil {
				log.Fatalf("子进程停止异常: %v\n", err)
			} else {
				log.Fatalf("子进程 %d 停止异常: %v\n", e.cmd.Process.Pid, err)
			}
		} else {
			log.Printf("子进程 %d 停止正常\n", e.cmd.Process.Pid)
		}
		// 判断是否自动重启
		if e.Info.AutoRestart {
			log.Printf("进程 %d 发出重启子进程信号\n", os.Getpid())
			// 发出重启信号
			e.jump <- SIGNAL_STOP
			e.jump <- SIGNAL_START
			// 发送取消阻塞信号
			e.signal <- SIGNAL_CANCEL_OBSTRUCT
		}
	}()
}
