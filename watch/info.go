package watch

import (
	"fmt"
	"log"
	"os"
	"runtime"
	"strings"
	"syscall"
)

type Info struct {
	Cmd               string
	Args              []string
	Folder            []string
	Files             []string
	Delay             uint
	Signal            int
	Timeout           int
	AutoRestart       bool
	PreCmd            string
	PreCmdTimeout     int
	PreCmdIgnoreError bool
	Pattern           string
}

func NewInfo() *Info {
	info := new(Info)
	info.Args = make([]string, 0)
	info.Folder = make([]string, 0)
	info.Files = make([]string, 0)
	info.Delay = 2
	info.Signal = int(syscall.SIGTERM)
	info.Timeout = 5
	info.Pattern = "poll"
	info.PreCmdTimeout = 10

	if runtime.GOOS == "windows" {
		info.Signal = int(syscall.SIGKILL)
	}
	return info
}

func (i *Info) String() string {
	f := "Cmd: %s\n"
	f += "Folder: %+v\n"
	f += "Files: %+v\n"
	f += "Delay: %d\n"
	f += "Signal: %s (%d) \n"
	f += "Timeout: %d \n"
	f += "AutoRestart: %v \n"
	f += "PreCmd: %s \n"
	f += "PreCmdTimeout: %d \n"
	f += "PreCmdIgnoreError: %v \n"
	f += "Pattern: %s \n"

	return fmt.Sprintf(f,
		i.Cmd+" "+strings.Join(i.Args, " "),
		i.Folder,
		i.Files,
		i.Delay,
		syscall.Signal(i.Signal).String(), i.Signal,
		i.Timeout,
		i.AutoRestart,
		i.PreCmd,
		i.PreCmdTimeout,
		i.PreCmdIgnoreError,
		i.Pattern,
	)
}

func (i *Info) Filter() bool {
	//过滤命令
	i.Cmd = strings.TrimSpace(i.Cmd)
	if i.Cmd == "" {
		log.Fatal("参数缺失：--cmd")
		return false
	}
	args := make([]string, 0)
	for _, v := range i.Args {
		v = strings.TrimSpace(v)
		if len(v) > 0 {
			args = append(args, v)
		}
	}
	i.Args = args

	// 过滤文件夹
	folders := make([]string, 0)
	for _, v := range i.Folder {
		v = strings.TrimSpace(strings.Replace(v, "\\", "/", -1))
		if fi, err := os.Stat(v); err == nil && fi.IsDir() {
			folders = append(folders, v)
		} else {
			log.Fatal("忽略无效的文件夹：%s\n", v)
		}
	}
	i.Folder = folders

	// 过滤文件
	files := make([]string, 0)
	for _, v := range i.Files {
		v = strings.TrimSpace(strings.Replace(v, "\\", "/", -1))
		if fi, err := os.Stat(v); err == nil && !fi.IsDir() {
			files = append(files, v)
		} else {
			log.Fatal("忽略无效的文件：%s\n", v)
		}
	}
	i.Files = files

	if len(i.Folder) == 0 && len(i.Files) == 0 {
		log.Fatal("参数缺失：--folder or --files")
		return false
	}

	if strings.EqualFold(i.Pattern, "poll") && !strings.EqualFold(i.Pattern, "notify") {
		log.Fatal("监视模式必须是 poll 或 notify")
		return false
	}
	i.PreCmd = strings.TrimSpace(i.PreCmd)
	if i.PreCmdTimeout <= 0 {
		i.PreCmdTimeout = 1
	}
	return true
}
