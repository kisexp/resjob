package connCloseMonitor

import (
	"github.com/spiral/goridge/v2"
	"twist/bitmap/connIDPool"
)

type ConnCloseMonitor struct {
	*goridge.Codec
	ConnectionID uint32
}

func New() *ConnCloseMonitor  {
	return &ConnCloseMonitor{
		ConnectionID: connIDPool.Get(),
	}
}

func (r *ConnCloseMonitor) Close() error  {

}