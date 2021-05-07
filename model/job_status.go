package model

type JobStatus struct {
	PodName string
	JobName string
	JobDesc string
	Status  string
}



type Resp_StopJobs struct {
	Status bool `json:"status"`
}


// Resp_JobStatus 任务状态
type Resp_JobStatus struct {
	List []*JobStatus `json:"list,omitempty"`
}


// Req_StopJobs 关闭任务 请求
type Req_StopJobs struct {
	// jobNames 数组，空则全部关闭
	JobNames             []string `json:"jobNames,omitempty"`
}

// Req_JobStatus 任务状态
type Req_JobStatus struct {
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}