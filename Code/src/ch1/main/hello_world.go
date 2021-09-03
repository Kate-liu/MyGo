package main

import (
	"fmt"
	"os"
)

func main() {
	// go run hello_world.go ming
	// os.Args: 获得函数的输入参数
	fmt.Println(os.Args)
	if len(os.Args) > 1 {
		fmt.Println("hello world", os.Args)
	}

	// 打印内容
	fmt.Println("hello world!")

	// 函数的返回值，退出状态，exit status 255
	os.Exit(-1)
}
