package main

import "fmt"

func main() {
	//声明局部变量
	var local1, local2, local3 int

	local1 = 8
	local2 = 10
	local3 = local1 + local2

	fmt.Printf("local1 = %d, local2 = %d and local3 = %d\n", local1, local2, local3)
}
