package dao

import (
	gredis "github.com/go-redis/redis"
	"twist/config"
)

type Dao struct {
	Config                 *config.Config
	JobRedis               *gredis.Client
}

func New(c *config.Config) (*Dao, error) {
	d := &Dao{
		Config: c,
	}
	d.JobRedis = gredis.NewClient(c.JobRedis)
	return d, nil
}

func (d *Dao) Close()  {
	d.JobRedis.Close()
}