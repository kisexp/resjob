package main

import (
	"fmt"
	"io"
	"os"
)

type alphaReader2 struct {
	// alphaReader2 里组合了标准库的 io.Reader
	reader io.Reader
}

func newAlphaReader2(reader io.Reader) *alphaReader2 {
	return &alphaReader2{reader: reader}
}

func alpha2(r byte) byte {
	if (r >= 'A' && r <= 'Z') || (r >= 'a' && r <= 'z') {
		return r
	}
	return 0
}

func (this *alphaReader2) Read(p []byte) (int, error) {
	// 这行代码调用的就是 io.Reader
	n, err := this.reader.Read(p)
	if err != nil {
		return n, err
	}
	buf := make([]byte, n)
	for i := 0; i < n; i++ {
		if char := alpha2(p[i]); char != 0 {
			buf[i] = char
		}
	}
	copy(p, buf)
	return n, nil
}

func main() {
	file, err := os.Open("./io_str.go")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	defer file.Close()
	reader := newAlphaReader2(file)
	p := make([]byte, 4)
	for {
		n, err := reader.Read(p)
		if err == io.EOF {
			break
		}
		fmt.Println(string(p[:n]))
	}
	fmt.Println()
}
