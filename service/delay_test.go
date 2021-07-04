package service

import (
	"fmt"
	"github.com/go-redis/redis"
	"log"
	"testing"
	"time"
	core "twist/core/json"
)

type MyArgs struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

var rds = redis.NewClient(&redis.Options{
	Addr:     "127.0.0.1:6379",
	Password: "123456",
})

func (d *MyArgs) Marshal() []byte {
	buf, _ := core.JSON.Marshal(d)
	return buf
}

func (d *MyArgs) Unmarshal(raw []byte) (err error) {
	return core.JSON.Unmarshal(raw, d)
}

func TestDelay_Run(t *testing.T) {
	d := NewDelay("test", rds)
	at := time.Now().Add(10 * time.Second)
	for i := 0; i < 10; i++ {
		p := &MyArgs{
			ID:   i + 1,
			Name: fmt.Sprintf("test%d", i),
		}
		d.Add(at, "test", p)
	}
	d.Run(func(it Item) {
		switch it.Event {
		case "test":
			args := MyArgs{}
			err := args.Unmarshal(it.Args)
			if err != nil {
				log.Printf("event unmarshal %v", err)
				return
			}
			fmt.Println(args)
		default:
			fmt.Println("error [default]", it.Event, it.Args)
		}
	}, 4, true)

	select {}
}

var timeLine = 1

func TestTraffic(t *testing.T) {
	traffic("count")
	timeLine = 1
	traffic("leak")

}

//流程模拟
func traffic(trafficType string) {
	fmt.Println(trafficType + ":")
	request(trafficType)
	sleep(4)
	for i := 0; i < 4; i++ {
		request(trafficType)
	}
	sleep(1)
	for i := 0; i < 4; i++ {
		request(trafficType)
	}
	sleep(4)
	request(trafficType)

}

//功能访问
func request(trafficType string) {
	var msg string
	// 5秒内访问3次
	rule := Rule{rds: rds, ttl: 5, count: 3}
	key := fmt.Sprintf("%s:127.0.0.1", trafficType)
	switch trafficType {
	case "count":
		if isOut, _ := rule.CountBrowse(key); isOut {
			msg = fmt.Sprintf("%ds:  deny", timeLine)
		} else {
			msg = fmt.Sprintf("%ds:  access", timeLine)
			rule.CountUpdate(key)
		}
	default:
		if isOut, _ := rule.LeakZsetBrowse(key); isOut {
			msg = fmt.Sprintf("%ds: deny", timeLine)
		} else {
			rule.LeakZsetUpdate(key)
			msg = fmt.Sprintf("%ds: access", timeLine)
		}
	}
	fmt.Println(msg)
}

func sleep(i int) {
	time.Sleep(time.Second * time.Duration(i))
	timeLine += i
}
