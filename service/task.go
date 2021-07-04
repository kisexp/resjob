package service

import (
	"log"
	"sync"
	"time"
)

const (
	JobtaskStatusRunning = "running"
	JobtaskStatusFinish  = "finish"
)

type JobTask struct {
	JobName        string      //必填
	TaskFunc       JobTaskFunc //必填
	Spec           string
	Desc           string
	ThreadNum      int
	RunStatus      string        //任务运行状态
	StopChan       chan struct{} //任务终止信号
	stopSignRecved bool          //是否已经接收到任务终止信号
}

type JobTaskFunc interface {
	RunTask(chan struct{})
}

func (jf *JobTask) recordStart() {
	jf.RunStatus = JobtaskStatusRunning
	log.Printf("【%v】【%s】【%s】【%s】!\n", time.Now(), jf.JobName, jf.Desc, jf.status())
}
func (jf *JobTask) recordFinish() {
	jf.RunStatus = JobtaskStatusFinish
	log.Printf("【%v】【%s】【%s】【%s】!\n", time.Now(), jf.JobName, jf.Desc, jf.status())
}

func (jf *JobTask) status() string {
	return jf.RunStatus
}

func (jf *JobTask) Run() {
	if JobtaskStatusRunning == jf.status() {
		//防止crontab 任务处理不过来一直创建任务
		return
	}
	jf.recordStart()
	defer jf.recordFinish()

	if jf.TaskFunc == nil {
		return
	}

	wg := sync.WaitGroup{}
	wg.Add(jf.ThreadNum)
	for i := 0; i < jf.ThreadNum; i++ {
		go func() {
			defer wg.Done()
			jf.TaskFunc.RunTask(jf.StopChan)
		}()
	}
	wg.Wait()
}

//dispatchStopSign 下发停止信号
func (jf *JobTask) dispatchStopSign() {
	if jf.stopSignRecved {
		return
	}
	jf.stopSignRecved = true

	if jf.RunStatus != JobtaskStatusRunning {
		return
	}

	for i := 0; i < jf.ThreadNum; i++ {
		jf.StopChan <- struct{}{}
	}

}
