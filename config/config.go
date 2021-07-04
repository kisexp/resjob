package config

import (
	gredis "github.com/go-redis/redis"
	"time"
)

// Config Config
type Config struct {
	Env         string
	ServiceName string
	// 脚本redis
	JobRedis *gredis.Options


}

// NewConfig new config
func NewConfig() *Config {
	return &Config{}
}

// GlobalConfig
var (
	GlobalConfig = &Config{
		Env: "dev",
		ServiceName: "twist",
		JobRedis: &gredis.Options{
			Network:            "tcp",
			Addr:               "r-uf6hypdzb6k7bdnni1.redis.rds.aliyuncs.com:6379",
			Password:           "3VWY80TtM6",
			DB:                 1,
			DialTimeout:        time.Second,
			ReadTimeout:        time.Second,
			WriteTimeout:       time.Second,
			PoolSize:           256,
			MinIdleConns:       32,
			MaxRetries:         1,
			MaxConnAge:         time.Minute * 3,
			PoolTimeout:        time.Second,
			IdleTimeout:        time.Minute * 3,
			IdleCheckFrequency: time.Minute,
		},
	}
)

// init（小写）方法中初始化conf并绑定config变量
func init() {

}
