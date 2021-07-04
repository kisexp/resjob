package service

import (
	"context"
	"log"
	"os"
	"time"
	"twist/model"

	"twist/config"
	"twist/core/cron"
)

var (
	cronObj    *cron.Cron
	RedpJobSvr *Job
)

type Job struct {
	jobComsumerEnable bool
	jobCrontabEnable  bool
	*Service
	Comsumers []*JobTask
	Crontabs  []*JobTask
}

func NewJobSvr(conf *config.Config, jobComsumerEnable, jobCrontabEnable bool) *Job {
	srv, err := NewService(conf)
	if err != nil {
		panic(err)
	}
	RedpJobSvr = &Job{
		Service:           srv,
		jobComsumerEnable: jobComsumerEnable,
		jobCrontabEnable:  jobCrontabEnable,
		Crontabs:          CronList,
		Comsumers:         ComsumerList,
	}
	RedpJobSvr.Start()
	return RedpJobSvr
}

func (job *Job) Start() {

	//第一优先级
	job.resetAllStatus()

	go job.syncStopFlags()
	go job.syncJobTaskStatus()

	if job.jobComsumerEnable {
		for _, task := range job.Comsumers {
			if task.JobName == "" || task.TaskFunc == nil {
				continue
			}
			log.Printf("运行 消费任务:【%s】,协程数:【%d】,描述:【%s】\n", task.JobName, task.ThreadNum, task.Desc)
			go task.Run()
		}
	}
	if job.jobCrontabEnable {
		cronObj = cron.New()
		for _, task := range job.Crontabs {
			if task.JobName == "" || task.Spec == "" || task.TaskFunc == nil {
				continue
			}
			if err := cronObj.AddJob(task.Spec, task); err != nil {
				log.Printf("AddJob %v", err)

			}
			log.Printf("运行 crontab任务:【%s】,协程数:【%d】,描述:【%s】\n", task.JobName, task.ThreadNum, task.Desc)
		}
		cronObj.Start()
	}

}

func (job *Job) Close() {
	job.dispatchAllStopSign()
	job.Dao.ClearAllJobTaskStatus()
	if cronObj != nil {
		cronObj.Stop()
	}
	time.Sleep(time.Second * 3)
	// TODO job service
}

//resetAllStatus 程序启动，清理任务停止状态
func (job *Job) resetAllStatus() {
	job.Dao.ClearStopFlags(context.Background())
}

// syncStopFlags 定时同步任务结束状态
func (job *Job) syncStopFlags() {
	ticker := time.NewTicker(time.Second)
	for {
		<-ticker.C
		tmp := job.Dao.GetStopFlags()
		for jobName, v := range tmp {
			if v != "stop" {
				continue
			}
			for _, jf := range job.Comsumers {
				if jf.JobName == jobName {
					go jf.dispatchStopSign()
				}
			}
			for _, jf := range job.Crontabs {
				if jf.JobName == jobName {
					go jf.dispatchStopSign()
				}
			}
		}
	}
}

// syncJobTaskStatus 定时同步任务状态
func (job *Job) syncJobTaskStatus() {
	ticker := time.NewTicker(time.Second)
	for {
		<-ticker.C
		jobTaskList := make([]*JobTask, 0)

		if job.jobCrontabEnable {
			jobTaskList = append(jobTaskList, job.Crontabs...)
		}
		if job.jobComsumerEnable {
			jobTaskList = append(jobTaskList, job.Comsumers...)
		}

		podName, _ := os.Hostname()
		list := make([]*model.JobStatus, 0)
		for _, v := range jobTaskList {
			list = append(list, &model.JobStatus{
				PodName: podName,
				JobName: v.JobName,
				Status:  v.RunStatus,
			})
		}
		job.Dao.UpdateJobTaskStatus(context.Background(), list)
	}
}

// dispatchAllStopSign 下发任务全部关闭信号
func (job *Job) dispatchAllStopSign() {
	list := append(job.Comsumers, job.Crontabs...)
	for _, task := range list {
		go task.dispatchStopSign()
	}
}
