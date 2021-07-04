package vlog

import (
	"errors"
	"fmt"
	"log"
	"math/rand"
	"os"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"
)

type Logger struct {
	LogID string
}

const (
	LOG_LEVEL_FATAL   = 1
	LOG_LEVEL_WARNING = 2
	LOG_LEVEL_NOTICE  = 4
	LOG_LEVEL_TRACE   = 8
	LOG_LEVEL_DEBUG   = 16
)

const (
	LOG_LEVEL_FATAL_STR   = "FATAL"
	LOG_LEVEL_WARNING_STR = "WARNING"
	LOG_LEVEL_NOTICE_STR  = "NOTICE"
	LOG_LEVEL_TRACE_STR   = "TRACE"
	LOG_LEVEL_DEBUG_STR   = "DEBUG"
)

var logLevelMap = map[int]string{
	LOG_LEVEL_FATAL:   LOG_LEVEL_FATAL_STR,
	LOG_LEVEL_WARNING: LOG_LEVEL_WARNING_STR,
	LOG_LEVEL_NOTICE:  LOG_LEVEL_NOTICE_STR,
	LOG_LEVEL_TRACE:   LOG_LEVEL_TRACE_STR,
	LOG_LEVEL_DEBUG:   LOG_LEVEL_DEBUG_STR,
}

func New(logID string) *Logger {
	return &Logger{
		LogID: logID,
	}
}

func (l *Logger) Notice(msg string) {
	l.syncMsg(LOG_LEVEL_NOTICE, msg)
}

func (l *Logger) Trace(msg string) {
	l.syncMsg(LOG_LEVEL_TRACE, msg)
}

func (l *Logger) Debug(msg string) {
	l.syncMsg(LOG_LEVEL_DEBUG, msg)
}

func (l *Logger) Fatal(msg string) {
	l.syncMsg(LOG_LEVEL_FATAL, msg)
}

func (l *Logger) Warning(msg string) {
	l.syncMsg(LOG_LEVEL_WARNING, msg)
}

func (l *Logger) syncMsg(level int, msg string) (err error) {
	if (level & vLog.LogLevel) != level {
		return
	}
	if level <= 0 || msg == "" {
		err = errors.New("level or msg param is empty")
		return
	}
	// 消息内容
	ret, err := l.padMsg(level, msg)
	if err != nil {
		return
	}
	// 消息格式
	data := Msg{
		Level: level,
		Data:  ret,
	}
	vLog.LogChan <- data

	// 判断当前整个channel的buffer大小是否超过90%的阀值，超过就直接发送刷盘信号
	curChanLen := len(vLog.LogChan)
	ratio := float64(curChanLen) / float64(vLog.LogChanBuffSize)
	if ratio >= 0.9 && !flushLogFlag {
		flushLock.Lock()
		flushLogFlag = true
		flushLock.Unlock()
		vLog.FlushLogChan <- true
		if isDebugOpen() {
			debugmsg := fmt.Sprintf(
				"out ratio!! "+
					"current vLog.LogChan: %v;"+
					" vLog.LogChanBuffSize: %v",
				curChanLen,
				vLog.LogChanBuffSize,
			)
			debugPrint(debugmsg, nil)
		}
	}
	return

}

func (l *Logger) padMsg(level int, msg string) (ret string, err error) {
	//获取调用的 函数/文件名/行号 等信息
	pc, file, line, ok := runtime.Caller(3)
	if !ok {
		err = errors.New("call runtime.Caller() fail")
		return
	}
	//判断当前操作系统路径分割符，获取调用文件最后两组路径信息
	callfunc := runtime.FuncForPC(pc).Name()
	dirSep := dirSep(file)
	callPath := strings.Split(file, dirSep)
	if len := len(callPath); len > 2 {
		file = strings.Join(callPath[len-2:], dirSep)
	}
	ymdHis := time.Now().Format("2006-01-02 15:04:05")
	logID := l.getLogID()
	//日志类型
	levelstr, ok := logLevelMap[level]
	if !ok {
		err = errors.New("log_type is invalid")
		return
	}
	format := "%s: %s [logid=%s file=%s, no=%d call=%s] %s\n"
	ret = fmt.Sprintf(format, levelstr, ymdHis, logID, file, line, callfunc, msg)
	return
}

func (l *Logger) getLogID() string {
	if l.LogID != "" {
		return l.LogID
	}
	return l.genLogID()
}

func (l *Logger) genLogID() string {
	microtime := time.Now().UnixNano()
	r := rand.New(rand.NewSource(microtime))
	randNum := r.Intn(100000)
	logID := strconv.FormatInt(microtime, 10) + strconv.Itoa(randNum)
	return logID
}

func dirSep(path string) string {
	var sep = "/"
	if strings.ContainsAny(path, "\\") {
		sep = "\\"
	}
	return sep
}

func isDebugOpen() bool {
	return vLog.LogDebugOpen
}

func debugPrint(msg string, v interface{}) {
	pc, file, line, ok := runtime.Caller(1)
	if !ok {
		log.Fatal("call runtime.Caller() fail")
		return
	}
	callfunc := runtime.FuncForPC(pc).Name()
	dirSep := dirSep(file)
	path := strings.Split(file, dirSep)
	if len := len(path); len > 2 {
		file = strings.Join(path[len-2:], dirSep)
	}

	fmt.Println("\n=======================Log Debug Info Start=======================")
	fmt.Printf("[ call=%v  file=%v  line=%v ]\n", callfunc, file, line)
	if msg != "" {
		fmt.Println(msg)
	}
	fmt.Println(v)
	fmt.Println("=======================Log Debug Info End=======================")
}

// 日志结构
type Msg struct {
	Level int
	Data  string
}

var vLog *VLog

var flushLogFlag = false

var flushLock *sync.Mutex

var vOnce sync.Once

// log主chan队列配置
type VLog struct {
	LogChan      chan Msg
	FlushLogChan chan bool
	LogFilePath  map[int]string

	LogChanBuffSize int
	LogCronTime     string
	LogFlushTimer   int
	LogLevel        int
	LogDebugOpen    bool

	LogConf *VlogConf

	LogNoticeFilePath string
	LogErrorFilePath  string

	//去重的日志文件名和fd (实际需需要物理写入文件名和句柄)
	MergeLogFile map[string]string
	MergeLogFd   map[string]*os.File
}

type VlogConf struct {
	LogNoticeFilePath  string
	LogDebugFilePath   string
	LogTraceFilePath   string
	LogFatalFilePath   string
	LogWarningFilePath string
	LogCronTime        string
	LogChanBuffSize    int
	LogFlushTimer      int
	LogDebugOpen       int
	LogLevel           int
}

const (
	NOTICE_FILE_PATH  = "log_notice_file_path"
	DEBUG_FILE_PATH   = "log_debug_file_path"
	TRACE_FILE_PATH   = "log_trace_file_path"
	FATAL_FILE_PATH   = "log_fatal_file_path"
	WARNING_FILE_PATH = "log_warning_file_path"
)

//日志文件名与日志类型的映射
var LogFileLevelMap = map[string]int{
	NOTICE_FILE_PATH:  LOG_LEVEL_NOTICE,
	DEBUG_FILE_PATH:   LOG_LEVEL_DEBUG,
	TRACE_FILE_PATH:   LOG_LEVEL_TRACE,
	FATAL_FILE_PATH:   LOG_LEVEL_FATAL,
	WARNING_FILE_PATH: LOG_LEVEL_WARNING,
}

func Run(conf *VlogConf) {
	if vLog == nil {
		vLog = new(VLog)
	}
	vLog.LogConf = conf
	// 初始化，全局只运行一次
	vOnce.Do(LogInit)

	go func() {
		var logMsg Msg
		for {
			select {
			case logMsg = <-vLog.LogChan:
				if isDebugOpen() {
					debugPrint("In select{ logMsg = <- vLog.LogChan, logWriteFile() } vLog.LogChan length:", len(vLog.LogChan))
				}
				logWriteFile(logMsg)
			default:
				if isDebugOpen() {
					debugPrint("In select{ default}, vLog.LogChan length:", len(vLog.LogChan))
				}
				time.Sleep(time.Duration(vLog.LogFlushTimer) * time.Millisecond)
			}
			//监控刷盘timer
			select {
			//如果收到刷盘channel的信号则刷盘且全局标志状态为
			case <-vLog.FlushLogChan:
				if isDebugOpen() {
					debugPrint("In select{ FlushLogChan}, vLog.LogChan Length:", len(vLog.LogChan))
				}
				flushLock.Lock()
				flushLogFlag = false
				flushLock.Unlock()
				break
			default:
				break
			}
		}
	}()

}

// 写日志
func logWriteFile(logMsg Msg) {
	// 打开文件
	logOpenFile()
	logMap := make(map[string][]string, len(LogFileLevelMap))
	for fileName, _ := range vLog.MergeLogFile {
		logMap[fileName] = make([]string, 0)
	}
	fileName := vLog.LogFilePath[logMsg.Level]
	logMap[fileName] = []string{logMsg.Data}

	select {
	case logItem := <-vLog.LogChan:
		fileName = vLog.LogFilePath[logItem.Level]
		logMap[fileName] = append(logMap[fileName], logItem.Data)
	default:
		break
	}

	if isDebugOpen() {
		debugPrint("LogMap:", logMap)
	}

	// 写入所有日志
	for fileName, _ := range vLog.MergeLogFile {
		if len(logMap[fileName]) == 0 {
			continue
		}
		var writeBuff, line string
		for _, line = range logMap[fileName] {
			writeBuff += line
		}
		_, _ = vLog.MergeLogFd[fileName].WriteString(writeBuff)
		_ = vLog.MergeLogFd[fileName].Sync()
		if isDebugOpen() {
			debugPrint("Log String: ", writeBuff)
		}
	}
}

// 打开/切割日志文件
func logOpenFile() (err error) {
	fileSuffix := logSuffix()
	for confFileName, runFileName := range vLog.MergeLogFile {
		newLogFileName := fmt.Sprintf("%s.%s", confFileName, fileSuffix)
		if newLogFileName == runFileName {
			continue
		}
		if vLog.MergeLogFd[confFileName] != nil {
			err = vLog.MergeLogFd[confFileName].Close()
			if err != nil {
				log.Fatalf("Close log file %s fail", runFileName)
				continue
			}
		}
		vLog.MergeLogFile[confFileName] = newLogFileName
		vLog.MergeLogFd[confFileName] = nil
		//创建&打开新日志文件
		newLogFileFd, err := os.OpenFile(newLogFileName, os.O_WRONLY|os.O_CREATE, 0644)

		if err != nil {
			log.Fatalf("Open log file %s fail", newLogFileName)
			continue
		}
		newLogFileFd.Seek(0, os.SEEK_END)
		vLog.MergeLogFile[confFileName] = newLogFileName
		vLog.MergeLogFd[confFileName] = newLogFileFd
	}
	return nil
}

func logSuffix() string {
	var fileSuffix string
	now := time.Now()
	switch vLog.LogCronTime {
	case "day":
		fileSuffix = now.Format("20060102")
	case "hour":
		fileSuffix = now.Format("20060102_15")
	case "ten":
		fileSuffix = fmt.Sprintf("%s%d0", now.Format("20060102_15"), now.Minute()/10)
	default:
		fileSuffix = now.Format("20060102_15")
	}
	return fileSuffix
}

func LogInit() {
	if vLog.LogConf == nil {
		log.Fatal("LogInit fail: LogConf data is nil")
		return
	}
	vLog.LogFilePath = make(map[int]string, len(logLevelMap))

	for _, level := range LogFileLevelMap {
		switch level {
		case LOG_LEVEL_FATAL:
			vLog.LogFilePath[level] = vLog.LogConf.LogFatalFilePath
		case LOG_LEVEL_WARNING:
			vLog.LogFilePath[level] = vLog.LogConf.LogWarningFilePath
		case LOG_LEVEL_NOTICE:
			vLog.LogFilePath[level] = vLog.LogConf.LogNoticeFilePath
		case LOG_LEVEL_TRACE:
			vLog.LogFilePath[level] = vLog.LogConf.LogTraceFilePath
		case LOG_LEVEL_DEBUG:
			vLog.LogFilePath[level] = vLog.LogConf.LogDebugFilePath
		}
	}

	vLog.LogCronTime = vLog.LogConf.LogCronTime
	vLog.LogChanBuffSize = vLog.LogConf.LogChanBuffSize
	vLog.LogFlushTimer = vLog.LogConf.LogFlushTimer
	vLog.LogLevel = vLog.LogConf.LogLevel
	if vLog.LogConf.LogDebugOpen == 1 {
		vLog.LogDebugOpen = true
	} else {
		vLog.LogDebugOpen = false
	}
	//设置日志channel buffer
	if vLog.LogChanBuffSize <= 0 {
		vLog.LogChanBuffSize = 1024
	}
	vLog.LogChan = make(chan Msg, vLog.LogChanBuffSize)
	vLog.MergeLogFile = make(map[string]string, len(logLevelMap))
	vLog.MergeLogFd = make(map[string]*os.File, len(logLevelMap))

	for _, logFilePath := range vLog.LogFilePath {
		vLog.MergeLogFile[logFilePath] = ""
		vLog.MergeLogFd[logFilePath] = nil
	}

	if isDebugOpen() {
		debugPrint("[ vLog data ]", vLog)
	}
	return

}
