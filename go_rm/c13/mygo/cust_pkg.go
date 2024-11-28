package main

import (
	"fmt"

	"mpb"
)

func main() {
	// 输出自定义包mpb的MyVariable
	fmt.Println(mpb.MyVariable)
	// 调用自定义包mpb的Get_data()
	mpb.Get_data()
}
