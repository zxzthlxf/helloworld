package main

import "fmt"

// 将string类型取一个别名叫meString
type meString = string

// 将myString类型定义为string
type myString string

func main() {
	// 将s1声明为meString类型
	var s1 meString
	// 查看s1的类型名
	fmt.Printf("s1的数据类型为: %T\n", s1)

	// 将s2声明为myString类型
	var s2 myString
	// 查看s2的类型名
	fmt.Printf("s2的数据类型为: %T\n", s2)
}
