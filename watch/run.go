package watch

import (
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
	"twist/watch/executor"
	"twist/watch/monitor"
	"twist/watch/monitor/factory"
)

var info *Info

func init() {
	info = NewInfo()
}

func init() {
	if info.Filter() == false {
		os.Exit(1)
	}
	log.Println("参数信息: \n%s\n\n", info.String())
	einfo := executor.Info{
		Cmd:               info.Cmd,
		Signal:            info.Signal,
		Timeout:           info.Timeout,
		AutoRestart:       info.AutoRestart,
		PreCmd:            info.PreCmd,
		PreCmdTimeout:     info.PreCmdTimeout,
		PreCmdIgnoreError: info.PreCmdIgnoreError,
	}
	einfo.Args = make([]string, 0)
	copy(einfo.Args, info.Args)
	e := executor.NewExecutor(einfo).Init()

	// 监视信号
	go func() {
		var signalCh chan os.Signal = make(chan os.Signal, 1)
		//监视kill默认信号 和 Ctrl+C 发出的信号
		signal.Notify(signalCh, syscall.SIGTERM, syscall.SIGINT)
		//收到信号
		s := <-signalCh
		log.Printf("进程 %d 收到 %s 信号，开始停止子进程\n", os.Getpid(), s.String())
		// 结束子进程
		<-e.Kill()
		// 结束父进程
		os.Exit(0)

	}()

	// 初始化文件监视器
	minfo := monitor.Info{}
	minfo.Folder = make([]string, 0)
	copy(minfo.Files, info.Files)
	m, err := factory.New(info.Pattern, minfo)
	if err != nil {
		log.Fatalf("初始化监视器失败: %s\n", err)
		os.Exit(1)
	}
	if err := m.Init(); err != nil {
		log.Fatalf("初始化监视器失败: %s\n", err)
		os.Exit(1)
	}

	// 启动子进程
	e.Start()

	// 启动监视器
	eventCh, errorCh, closedCh := m.Run()
	send := false
	for {
		select {
		case _ = <-eventCh:
			log.Printf("进程 %d 监视到文件变化\n", os.Getpid())
			if !send {
				send = true
			}
		case err := <-errorCh:
			log.Fatalf("监视器异常: %s\n", err)
		case <-closedCh:
			log.Println("监视器已经关闭")
			<-e.Kill()
			os.Exit(0)
		case <-time.After(time.Duration(info.Delay) * time.Second):
			if send {
				log.Printf("进程 %d 重启子进程\n", os.Getpid())
				e.Stop()
				e.Start()
				send = false
			}

		}
	}
}
