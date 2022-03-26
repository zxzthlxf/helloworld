package main

import (
	"fmt"
	"time"
)

func Echo(s string) {
	for i := 0; i < 3; i++ {
		time.Sleep(100 * time.Millisecond)
		fmt.Println(s)
	}
}

func main() {
	go Echo("go")
	Echo("web program")

	//	ch := make(chan string)
	//	go func(){ch <- "sleep"}()
	//   <-ch
	//	fmt.Println("通道正常结束！")
	ch := make(chan string, 1)
	ch <- "sleep"
	fmt.Println(<-ch, "通道正常结束！")

}
