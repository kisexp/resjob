package notify

import (
	"bytes"
	"errors"
	"github.com/fsnotify/fsnotify"
	"io/fs"
	"log"
	"path/filepath"
	"twist/watch/monitor"
)

type Notify struct {
	Info    monitor.Info
	watcher *fsnotify.Watcher
}

func New(info monitor.Info) *Notify {
	return &Notify{Info: info}
}

func (n *Notify) Init() error {
	var e error
	if n.watcher, e = fsnotify.NewWatcher(); e != nil {
		return e
	}

	//错误
	errs := bytes.Buffer{}
	//递归监视文件夹
	for _, v := range n.Info.Folder {
		if err := filepath.Walk(v, func(path string, info fs.FileInfo, err error) error {
			return n.watcher.Add(path)
		}); err != nil {
			errs.WriteString(err.Error())
			errs.WriteByte('\n')
		}
	}

	//监视文件
	for _, v := range n.Info.Files {
		if err := n.watcher.Add(v); err != nil {
			errs.WriteString(err.Error())
			errs.WriteByte('\n')
		}
	}

	//返回错误
	if errs.Len() > 0 {
		return errors.New(errs.String())
	}
	return nil
}

func (n *Notify) Run() (<-chan monitor.Event, <-chan error, chan struct{}) {
	eventCh := make(chan monitor.Event)
	errorCh := make(chan error)
	closed := make(chan struct{})
	go func() {
		for {
			select {
			case e := <-n.watcher.Events:
				if e.Op&fsnotify.Create == fsnotify.Create {
					if err := n.watcher.Add(e.Name); err != nil {
						log.Fatalf("监视器添加新的监视对象失败: %s\n", err.Error())
					}
				}
				event := monitor.Event{}
				event.Name = e.Name
				if e.Op&fsnotify.Create == fsnotify.Create {
					event.Op = monitor.Create
				} else if e.Op&fsnotify.Remove == fsnotify.Remove {
					event.Op = monitor.Remove
				} else if e.Op&fsnotify.Write == fsnotify.Write {
					event.Op = monitor.Write
				} else if e.Op&fsnotify.Rename == fsnotify.Rename {
					event.Op = monitor.Rename
				} else if e.Op&fsnotify.Chmod == fsnotify.Chmod {
					event.Op = monitor.Chmod
				} else {
					break
				}
				eventCh <- event
			case err := <-n.watcher.Errors:
				errorCh <- err
			}
		}
	}()
	return eventCh, errorCh, closed
}
