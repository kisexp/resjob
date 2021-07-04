package bash

import (
	"bytes"
	"errors"
	"os/exec"
	"strings"
	"syscall"
	"time"
)

type Bash struct {
	cmd         string
	timeout     time.Duration
	terminateCh chan bool
	closed      chan struct{}
	command     *exec.Cmd
	stdout      bytes.Buffer
	stderr      bytes.Buffer
}

func NewBash(cmd string, timeout time.Duration) *Bash {
	b := new(Bash)
	b.cmd = cmd
	b.timeout = timeout
	if b.timeout <= 0*time.Second {
		b.timeout = 3600 * time.Second
	}
	b.terminateCh = make(chan bool)
	b.closed = make(chan struct{})
	b.command = exec.Command("/bin/bash", "-c", b.cmd)
	b.command.Stderr = &b.stderr
	b.command.Stdout = &b.stdout
	return b
}

func (b *Bash) Start() error {
	if err := b.command.Start(); err != nil {
		return err
	}
	errCH := make(chan error)
	go func() {
		defer func() {
			close(errCH)
		}()
		err := b.command.Wait()
		select {
		case <-b.closed:
			return
		default:
			errCH <- err
		}
	}()

	var err error
	select {
	case err = <-errCH:
	case <-time.After(b.timeout):
		err = b.terminate()
		if err == nil {
			err = errors.New("cmd run timeout")
		}
	case <-b.terminateCh:
		err = b.terminate()
		if err == nil {
			err = errors.New("cmd is terminated")
		}
	}
	close(b.closed)
	return err
}

func (b *Bash) terminate() error {
	return b.command.Process.Signal(syscall.SIGKILL)
}

func (b *Bash) Stop() {
	select {
	case <-b.closed:
		return
	default:
		b.terminateCh <- true

	}
}

func (b *Bash) HasErr() bool {
	if b.command.ProcessState != nil {
		if b.command.ProcessState.Success() {
			return false
		}
		return true
	}
	return b.stderr.Len() != 0
}

func (b *Bash) StdErr() string {
	return strings.TrimSpace(b.stderr.String())
}

func (b *Bash) StdOut() string {
	return strings.TrimSpace(b.stdout.String())
}
