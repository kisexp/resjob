package dao

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"
	"twist/model"

	"twist/core/json"
)

//ClearStopFlags 清理任务停止状态
func (dao *Dao) ClearStopFlags(ctx context.Context) {
	dao.JobRedis.Del("redp:job:stop:map")
}

//StopJobs 关闭多个任务
func (dao *Dao) StopJobs(jobNames []string) bool {
	if len(jobNames) == 0 {
		return false
	}
	tmp := make(map[string]interface{})
	for _, v := range jobNames {
		tmp[v] = "stop"
	}
	sc := dao.JobRedis.HMSet("redp:job:stop:map", tmp)
	dao.JobRedis.Expire("redp:job:stop:map", time.Hour*4)
	return "OK" == strings.ToUpper(sc.Val())
}

//GetStopFlags 获取关闭标志
func (dao *Dao) GetStopFlags() map[string]string {
	ssc := dao.JobRedis.HGetAll("redp:job:stop:map")
	return ssc.Val()
}

// GetJobTaskStatus 获取所有节点任务的运行状态 map[pod名称]map[任务名称]数据
func (dao *Dao) GetJobTasksStatus() map[string]map[string]*model.JobStatus {
	ssmc := dao.JobRedis.HGetAll("redp:job:run:status:keys")
	resp := make(map[string]map[string]*model.JobStatus)
	for podName, _ := range ssmc.Val() {
		resp[podName] = make(map[string]*model.JobStatus)
		key := fmt.Sprintf("redp:job:run:status:%s", podName)
		vv := dao.JobRedis.HGetAll(key).Val()
		for jobName, jsonStr := range vv {
			jobStatus := &model.JobStatus{}
			if err := core.JSON.Unmarshal([]byte(jsonStr), jobStatus); err == nil {
				resp[podName][jobName] = jobStatus
			}
		}
	}
	return resp
}

//UpdateJobTaskStatus 更新当前pod任务状态
func (dao *Dao) UpdateJobTaskStatus(ctx context.Context, list []*model.JobStatus) {
	if len(list) == 0 {
		return
	}
	podName, _ := os.Hostname()
	key := fmt.Sprintf("redp:job:run:status:%s", podName)
	fields := make(map[string]interface{})
	for _, v := range list {
		bs, _ := core.JSON.Marshal(v)
		fields[v.JobName] = string(bs)
	}

	dao.JobRedis.HMSet(key, fields)
	dao.JobRedis.Expire(key, time.Hour)

	key2 := "redp:job:run:status:keys"
	dao.JobRedis.HSet(key2, podName, time.Now().Unix())
	dao.JobRedis.Expire(key2, time.Hour)
}

// ClearAllJobTaskStatus pod 销毁时，清理状态
func (dao *Dao) ClearAllJobTaskStatus() {
	podName, _ := os.Hostname()
	key := fmt.Sprintf("redp:job:run:status:%s", podName)
	dao.JobRedis.Del(key)

	key2 := "redp:job:run:status:keys"
	dao.JobRedis.HDel(key2, podName)
}
