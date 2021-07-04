package main

import (
	"fmt"
	"io/ioutil"
	"os/exec"
)

func main() {
	// 执行时间命令
	dateCmd := exec.Command("date")
	dateOut, err := dateCmd.Output()
	if err != nil {
		panic(err)
	}
	fmt.Println(string(dateOut))

	// 执行ls命令
	lsCmd := exec.Command("bash", "-c", "ls -a -l -h")
	lsOut, err := lsCmd.Output()
	if err != nil {
		panic(err)
	}
	fmt.Println(string(lsOut))

	// 执行grep命令
	grepCmd := exec.Command("grep", "hello")

	// 获取输入和输出对象
	grepIn, _ := grepCmd.StdinPipe()
	grepOut, _ := grepCmd.StdoutPipe()

	// 开始执行
	grepCmd.Start()

	// 输入字符
	grepIn.Write([]byte("hello grep\ngoodbye grep"))
	grepIn.Close()

	// 读取输出
	grepBytes, _ := ioutil.ReadAll(grepOut)
	grepCmd.Wait()
	fmt.Println("> grep hello")
	fmt.Println(string(grepBytes))
}
