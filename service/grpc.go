package service

import (
	"context"
	"twist/model"


)

// 停止任务
func (job *Job) StopJob(ctx context.Context, in *model.Req_StopJobs) (*model.Resp_StopJobs, error) {
	if in == nil {
		in = &model.Req_StopJobs{}
	}
	if len(in.JobNames) == 0 {
		for _, v := range job.Comsumers {
			in.JobNames = append(in.JobNames, v.JobName)
		}
		for _, v := range job.Crontabs {
			in.JobNames = append(in.JobNames, v.JobName)
		}
	}
	ret := job.Dao.StopJobs(in.JobNames)
	return &model.Resp_StopJobs{Status: ret}, nil
}

// 任务状态
func (job *Job) JobStatus(context.Context, *model.Req_JobStatus) (*model.Resp_JobStatus, error) {
	tmp := make(map[string]string)
	for _, v := range job.Crontabs {
		tmp[v.JobName] = v.Desc
	}
	for _, v := range job.Comsumers {
		tmp[v.JobName] = v.Desc
	}

	data := job.Dao.GetJobTasksStatus()
	list := make([]*model.JobStatus, 0)
	for podName, v := range data {
		for _, vv := range v {
			vv.PodName = podName
			vv.JobDesc = tmp[vv.JobName]
			list = append(list, vv)
		}
	}

	return &model.Resp_JobStatus{
		List: list,
	}, nil
}
