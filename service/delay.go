package service

import (
	"encoding/json"
	"fmt"
	"github.com/go-redis/redis"
	"log"
	"sync"
	"time"
	core "twist/core/json"
)

const (
	keyPre  = "delay:queue"
	zScript = `local a=redis.call('ZRANGEBYSCORE',KEYS[1],ARGV[1],ARGV[2],ARGV[3],ARGV[4],ARGV[5]) for k,v in ipairs(a) do redis.call('ZREM',KEYS[1],v) end return a`
)

type Delay struct {
	Group string
	Key   string
	Rds   *redis.Client
}

type Item struct {
	Ts    int64           `json:"ts"`
	Event string          `json:"event"`
	Args  json.RawMessage `json:"args"`
}

type Args interface {
	Marshal() []byte
	Unmarshal([]byte) error
}

func NewDelay(group string, client *redis.Client) *Delay {
	return &Delay{
		Group: group,
		Rds:   client,
		Key:   fmt.Sprintf("%s:%s", keyPre, group),
	}
}

func (d *Delay) Add(runAt time.Time, event string, args Args) (err error) {
	param := Item{
		Event: event,
		Args:  args.Marshal(),
		Ts:    time.Now().UnixNano(),
	}
	vals, err := core.JSON.MarshalToString(param)
	if err != nil {
		log.Printf("add event err %v", err)
		return
	}

	members := redis.Z{
		Score:  float64(runAt.Unix()),
		Member: vals,
	}
	return d.Rds.ZAdd(d.Key, members).Err()

}

func (d *Delay) Run(fun func(it Item), size int, force bool) {
	ch := make(chan Item, 0)
	if size < 1 {
		size = 1
	}
	if !force {
		go func() {
			for {
				select {
				case it := <-ch:
					fun(it)
				}
			}
		}()
	}
	t := time.NewTicker(1 * time.Second)
	for {
		select {
		case <-t.C:
			ts := time.Now().Unix()
			for {
				items := make([]string, 0)
				cmd := d.Rds.Eval(zScript, []string{d.Key}, "0", ts, "LIMIT", 0, size)
				ret, err := cmd.Result()
				if err != nil {
					log.Printf("cmd err: %v", err)
					continue
				}
				buf, err := core.JSON.Marshal(ret)
				if err != nil || string(buf) == "[]" {
					log.Printf("%s is empty", d.Key)
					break
				}
				err = core.JSON.Unmarshal(buf, &items)
				if err != nil {
					log.Printf("json unmarshal: %v", err)
					continue
				}
				if !force {
					for _, js := range items {
						item := Item{}
						err = core.JSON.UnmarshalFromString(js, &item)
						if err != nil {
							log.Printf("item err %v", err)
						} else {
							ch <- item
						}

					}
					continue
				}
				wg := sync.WaitGroup{}
				wg.Add(len(items))
				for _, js := range items {
					go func(_js string) {
						item := Item{}
						err = core.JSON.UnmarshalFromString(_js, &item)
						if err != nil {
							log.Printf("item err: %v", err)
						} else {
							fun(item)
						}
						wg.Done()
					}(js)
				}
				wg.Wait()
			}

		}
	}
}
