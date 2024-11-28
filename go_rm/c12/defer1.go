package main

import "fmt"

func myFunc() int {
	defer fmt.Printf("这是defer\n")
	fmt.Printf("这是函数的业务逻辑\n")
	return 1
}

func main() {
	fmt.Printf("这是函数返回值：%v\n", myFunc())
}
