package main

import (
	"fmt"
)

func myFunc() {
	defer func() {
		if err := recover(); err != nil {
			fmt.Printf("捕捉异常：%v\n", err)
		} else {
			fmt.Println("程序没有异常")
		}
	}()
	fmt.Println("程序正常运行")
	panic("这是自定义异常")
}

func main() {
	myFunc()
}
