package poll

import (
	"bytes"
	"errors"
	"github.com/radovskyb/watcher"
	"log"
	"time"
	"twist/watch/monitor"
)

type Poll struct {
	Info    monitor.Info
	watcher *watcher.Watcher
}

func New(info monitor.Info) *Poll {
	return &Poll{Info: info}
}

func (p *Poll) Init() error {
	p.watcher = watcher.New()
	//错误
	errs := bytes.Buffer{}
	//递归监视文件夹
	for _, v := range p.Info.Folder {
		if e := p.watcher.AddRecursive(v); e != nil {
			errs.WriteString(e.Error())
			errs.WriteByte('\n')
		}
	}
	//监视文件
	for _, v := range p.Info.Files {
		if e := p.watcher.Add(v); e != nil {
			errs.WriteString(e.Error())
			errs.WriteByte('\n')
		}
	}
	//返回错误
	if errs.Len() > 0 {
		return errors.New(errs.String())
	}
	return nil
}

func (p *Poll) Run() (<-chan monitor.Event, <-chan error, chan struct{}) {
	eventCh := make(chan monitor.Event)
	errorCh := make(chan error)
	closed := make(chan struct{})

	go func() {
		if err := p.watcher.Start(time.Millisecond * 100); err != nil {
			log.Fatalf("启动监视器失败: %s\n", err)
			p.watcher.Close()
			close(closed)
		}
	}()

	go func() {
		for {
			select {
			case e := <-p.watcher.Event:
				event := monitor.Event{}
				event.Name = e.Name()
				if e.Op == watcher.Create {
					event.Op = monitor.Create
				} else if e.Op == watcher.Remove {
					event.Op = monitor.Remove
				} else if e.Op == watcher.Write {
					event.Op = monitor.Write
				} else if e.Op == watcher.Rename {
					event.Op = monitor.Rename
				} else if e.Op == watcher.Chmod {
					event.Op = monitor.Chmod
				} else {
					break
				}
				eventCh <- event
			case err := <-p.watcher.Error:
				errorCh <- err
			case <-p.watcher.Closed:
				close(closed)
				return

			}
		}
	}()
	return eventCh, errorCh, closed
}
