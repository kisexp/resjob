package main

import (
	"twist/vlog"
)

func main() {
	path := "/Users/huliangliang/Documents/data/go/src/twist/vlog/log/"
	conf := &vlog.VlogConf{
		LogNoticeFilePath:  path + "huliangliang.log",
		LogDebugFilePath:   path + "huliangliang.log",
		LogTraceFilePath:   path + "huliangliang.log",
		LogFatalFilePath:   path + "huliangliang.log.wf",
		LogWarningFilePath: path + "huliangliang.log.wf",
		LogCronTime:        "day",
		LogChanBuffSize:    1024,
		LogFlushTimer:      1000,
		LogDebugOpen:       1,
		LogLevel:           31,
	}
	vlog.Run(conf)

	go LogTest()

	select {}
}

func LogTest() {
	logHandle := vlog.New("987654321")
	logHandle.Notice("[logger=logHandle msg='The notice message is test']")
	logHandle.Warning("[logger=logHandle msg='The warning message is test']")
	logHandle.Fatal("[logger=logHandle msg='The fatal message is test']")
	logHandle.Debug("[logger=logHandle msg='The debug message is test']")
	logHandle.Trace("[logger=logHandle msg='The trace message is test']")

}
