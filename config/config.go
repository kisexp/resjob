package config

import (
	gredis "github.com/go-redis/redis"

)

// Config Config
type Config struct {
	Env         string
	ServiceName string
	HTTPAddr    string
	// 脚本redis
	JobRedis *gredis.Options


}

// NewConfig new config
func NewConfig() *Config {
	return &Config{}
}

// GlobalConfig
var (
	GlobalConfig = &Config{}
)

// init（小写）方法中初始化conf并绑定config变量
func init() {

}
