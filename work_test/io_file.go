package main

func main() {
	//proverbs := []string{
	//	"Channels orchestrate mutexes serialize\n",
	//	"Cgo is not Go\n",
	//	"Errors are values\n",
	//	"Don't panic\n",
	//}
	//file, err := os.Create("./proverbs.txt")
	//if err != nil {
	//	fmt.Println(err)
	//	os.Exit(1)
	//}
	//defer file.Close()
	//
	//for _, p := range proverbs {
	//	// file 类型实现了io.writer
	//	n, err := file.Write([]byte(p))
	//	if err != nil {
	//		fmt.Println(err)
	//		os.Exit(1)
	//	}
	//	if n != len(p) {
	//		fmt.Println("failed to write data")
	//		os.Exit(1)
	//	}
	//}
	//fmt.Println("file write done")
	//file, err := os.Open("./proverbs.txt")
	//if err != nil {
	//	fmt.Println(err)
	//	os.Exit(1)
	//}
	//defer file.Close()
	//p := make([]byte, 4)
	//for {
	//	n, err := file.Read(p)
	//	if err == io.EOF {
	//		break
	//	}
	//	fmt.Println(string(p[:n]))
	//}

	//for _, p := range proverbs {
	//	// 因为 os.Stdout 也实现了 io.Writer
	//	n, err := os.Stdout.Write([]byte(p))
	//	if err != nil {
	//		fmt.Println(err)
	//		os.Exit(1)
	//	}
	//	if n != len(p) {
	//		fmt.Println("failed to write data")
	//		os.Exit(1)
	//	}
	//}

	//proverbs := new(bytes.Buffer)
	//proverbs.WriteString("Channels orchestrate mutexes serialize\n")
	//proverbs.WriteString("Cgo is not Go\n")
	//proverbs.WriteString("Errors are values\n")
	//proverbs.WriteString("Don't panic\n")
	//
	//file, err := os.Create("./proverbs.txt")
	//if err != nil {
	//	fmt.Println(err)
	//	os.Exit(1)
	//}
	//defer file.Close()
	//
	//// io.Copy 完成了从proverbs 读取数据并写入 file 的流程
	//if _, err := io.Copy(file, proverbs); err != nil {
	//	fmt.Println(err)
	//	os.Exit(1)
	//}
	//fmt.Println("file created")

	//file, err := os.Open("./proverbs.txt")
	//if err != nil {
	//	fmt.Println(err)
	//	os.Exit(1)
	//}
	//defer file.Close()
	//if _, err := io.Copy(os.Stdout, file); err != nil {
	//	fmt.Println(err)
	//	os.Exit(1)
	//}

	//file, err := os.Create("./magic_msg.txt")
	//if err != nil {
	//	fmt.Println(err)
	//	os.Exit(1)
	//}
	//defer file.Close()
	//if _, err := io.WriteString(file, "Go is fun!"); err != nil {
	//	fmt.Println(err)
	//	os.Exit(1)
	//}

	//proverbs := new(bytes.Buffer)
	//proverbs.WriteString("Channels orchestrate mutexes serialize\n")
	//proverbs.WriteString("Cgo is not Go\n")
	//proverbs.WriteString("Errors are values\n")
	//proverbs.WriteString("Don't panic\n")
	//piper, pipew := io.Pipe()
	//
	//// 将proverbs写入pipew这一端
	//go func() {
	//	defer pipew.Close()
	//	io.Copy(pipew, proverbs)
	//}()
	//
	//// 从另一端piper中读取数据并拷贝到标准输出
	//io.Copy(os.Stdout, piper)
	//piper.Close()

	//file, err := os.Open("./proverbs.txt")
	//if err != nil {
	//	fmt.Println(err)
	//	os.Exit(1)
	//}
	//
	//defer file.Close()
	//reader := bufio.NewReader(file)
	//for {
	//	line, err := reader.ReadString('\n')
	//	if err != nil {
	//		if err == io.EOF {
	//			break
	//		} else {
	//			fmt.Println(err)
	//			os.Exit(1)
	//		}
	//	}
	//	fmt.Println(line)
	//}

	//bytes, err := ioutil.ReadFile("./proverbs.txt")
	//if err != nil {
	//	fmt.Println(err)
	//	os.Exit(1)
	//}
	//
	//fmt.Printf("%s", bytes)
}
