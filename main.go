package main

import (
	"context"
	"fmt"
	"github.com/json-iterator/go/extra"
	"twist/config"
	"twist/core/middleware/interceptor"
	"twist/service"
)

const addr = ":8012"

var (
	JobComsumer string
	JobCrontab  string
)

func init() {
	extra.RegisterFuzzyDecoders()

}
func main() {
	fmt.Println("JobComsumer:", JobComsumer)
	fmt.Println("JobCrontab:", JobCrontab)

	jobSvr := service.NewJobSvr(config.GlobalConfig, JobComsumer == "open", JobCrontab == "open")

	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	interceptor.Reload(jobSvr.Close)

}
