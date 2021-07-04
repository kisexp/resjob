package factory

import (
	"errors"
	"strings"
	"twist/watch/monitor"
	"twist/watch/monitor/notify"
	"twist/watch/monitor/poll"
)

type Monitor interface {
	Init() error
	Run() (<-chan monitor.Event, <-chan error, chan struct{})
}

func New(pattern string, info monitor.Info) (Monitor, error)  {
	if strings.EqualFold(pattern, "poll") {
		return poll.New(info), nil
	} else if strings.EqualFold(pattern, "notify") {
		return notify.New(info), nil
	}
	return nil, errors.New("监视模式必须是 poll 或 notify")
}