package zero

import (
	"fmt"
	"io"
	"log"
	"net"
	"strings"
	"sync/atomic"
	"time"
)

type Server struct {
	concurrency uint32
	Concurrency int
	open        int32 // 正在打开连接数
	stop        int32 // 暂停连接数
	//默认情况下，并发连接数不受限制
	//可以从单个IP地址建立到服务器。
	MaxConnsPerIP    int
	perIPConnCounter perIPConnCounter
	//超过并发限制（默认值为[when为0]：请勿入睡）
	//并立即接受新的连接）
	SleepWhenConcurrencyLimitsExceeded time.Duration
	IdleTimeout                        time.Duration //启用保持活动状态时的下一个请求。 空闲超时
	//读取第一个字节后的保持活动连接。
	//默认情况下，请求读取超时是无限的。
	ReadTimeout time.Duration
	//默认情况下，响应写超时是无限的。
	WriteTimeout time.Duration
	//最大请求主体大小。
	//服务器拒绝正文超过此限制的请求。
	//默认情况下，请求正文大小受DefaultMaxRequestBodySize限制。
	MaxRequestBodySize int
}

const DefaultConcurrency = 256 * 1024
const DefaultMaxRequestBodySize = 4 * 1024 * 1024

func (s *Server) getConcurrency() int {
	n := s.Concurrency
	if n <= 0 {
		n = DefaultConcurrency
	}
	return n
}

func wrapPerIPConn(s *Server, c net.Conn) net.Conn {
	ip := getUinit32IP(c)
	if ip == 0 {
		return c
	}
	n := s.perIPConnCounter.Register(ip)
	if n > s.MaxConnsPerIP {
		s.perIPConnCounter.Unregister(ip)
		log.Println("The number of connections from your ip exceeds MaxConnsPerIP")
		c.Close()
		return nil
	}
	return acquirePerIPConn(c, ip, &s.perIPConnCounter)
}

func acceptConn(s *Server, ln net.Listener, lastPerIPErrorTime *time.Time) (net.Conn, error) {
	for {
		c, err := ln.Accept()
		if err != nil {
			if c != nil {
				panic("net.Listener returned non-nil conn and non-nil error")
			}
			if netErr, ok := err.(net.Error); ok && netErr.Temporary() {
				log.Printf("Temporary error when accepting new connections: %s", netErr)
				time.Sleep(time.Second)
				continue
			}
			// 使用封闭的网络连接
			if err != io.EOF && !strings.Contains(err.Error(), "use of closed network connection") {
				// 接受新连接时永久性错误
				return nil, err
			}
			return nil, io.EOF
		}
		if c == nil {
			panic("net.Listener returned (nil, nil)")
		}
		if s.MaxConnsPerIP > 0 {
			ipc := wrapPerIPConn(s, c)
			if ipc == nil {
				if time.Since(*lastPerIPErrorTime) > time.Minute {
					log.Printf("The number of connections from %s exceeds MaxConnsPerIP=%d",
						getConnIP4(c), s.MaxConnsPerIP)
					*lastPerIPErrorTime = time.Now()
				}
				continue
			}
			c = ipc
		}
		return c, nil
	}
}

func (s *Server) Serve(ln net.Listener) error {
	maxWorkersCount := s.getConcurrency()
	var lastPerIPErrorTime time.Time
	var lastOverflowErrorTime time.Time
	var err error
	var c net.Conn
	w := &workerPool{
		WorkerFunc:      s.serveConn,
		MaxWorkersCount: maxWorkersCount,
	}
	w.Start()

	atomic.AddInt32(&s.open, 1)
	defer atomic.AddInt32(&s.open, -1)

	for {
		if c, err = acceptConn(s, ln, &lastPerIPErrorTime); err != nil {
			w.Stop()
			if err == io.EOF {
				return nil
			}
			return err
		}
		atomic.AddInt32(&s.open, 1)
		if !w.Serve(c) {
			atomic.AddInt32(&s.open, -1)
			log.Println("The connection cannot be served because Server.Concurrency limit exceeded")
			c.Close()
			if time.Since(lastOverflowErrorTime) > time.Minute {
				log.Printf("The incoming connection cannot be served, because %d concurrent connections are served. "+
					"Try increasing Server.Concurrency", maxWorkersCount)
				lastOverflowErrorTime = time.Now()
			}
			if s.SleepWhenConcurrencyLimitsExceeded > 0 {
				time.Sleep(s.SleepWhenConcurrencyLimitsExceeded)
			}
		}
		c = nil
	}

}

func (s *Server) serveConnCleanup() {
	atomic.AddInt32(&s.open, -1)
	atomic.AddUint32(&s.concurrency, ^uint32(0))
}

func (s *Server) idleTimeout() time.Duration {
	if s.IdleTimeout != 0 {
		return s.IdleTimeout
	}
	return s.ReadTimeout

}

var globalConnID uint64

func nextConnID() uint64 {
	return atomic.AddUint64(&globalConnID, 1)
}

func (s *Server) serveConn(c net.Conn) (err error) {
	defer s.serveConnCleanup()
	atomic.AddUint32(&s.concurrency, 1)

	connID := nextConnID()
	connTime := time.Now()

	maxRequestBodySize := s.MaxRequestBodySize
	if maxRequestBodySize <= 0 {
		maxRequestBodySize = DefaultMaxRequestBodySize
	}
	writeTimeout := s.WriteTimeout

	fmt.Println(connID, connTime, writeTimeout)
	return
}
