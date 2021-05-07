package interceptor

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// Reload Reload
func Reload(cls func()) {
	env := os.Getenv("ENV_DEVELOPMENT")
	// signal
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT, syscall.SIGKILL)
	for {
		s := <-c
		fmt.Println(fmt.Sprintf("service get a signal %s, %v", s.String(), s))
		switch s {
		case syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGSTOP, syscall.SIGINT, syscall.SIGHUP:
			if env != "debug" {
				time.Sleep(3 * time.Second)
			}
			cls()
			fmt.Println("service Closed")
			if env != "debug" {
				time.Sleep(2 * time.Second)
			}
			return
		default:
			return
		}
	}
}
