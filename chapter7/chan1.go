package main

import (
	"fmt"
	"time"
)

func main() {
	ch := make(chan string)
	go func() {
		fmt.Println("开始goroutine")
		ch <- "signal"
		fmt.Println("退出goroutine")
	}()
	fmt.Println("等待goroutine")
	<-ch
	fmt.Println("完成")

	timeout := make(chan bool, 1)

	go func() {
		time.Sleep(6)
		timeout <- true
	}()

	select {
	case <-ch:
	case <-timeout:
	}

}
