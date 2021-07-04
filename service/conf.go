package service

//CronList 任务列表
var CronList = []*JobTask{
}

//ComsumerList 消费者列表
var ComsumerList = []*JobTask{
}

func genJF(jobName string, desc, spec string, taskFunc JobTaskFunc, threadNum int) *JobTask {
	if threadNum <= 0 {
		threadNum = 1
	}
	return &JobTask{
		JobName:   jobName,
		Desc:      desc,
		Spec:      spec,
		TaskFunc:  taskFunc,
		ThreadNum: threadNum,
		StopChan:  make(chan struct{}, threadNum),
	}
}
