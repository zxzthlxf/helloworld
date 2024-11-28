package main

import "fmt"

// 声明或创建变量
var name string
var age int = get_age()

// 定义函数，用于声明或创建变量
func get_age() int {
	fmt.Printf("这是声明或创建变量\n")
	return 10
}

// 定义初始化函数init()
func init() {
	// 变量赋值操作
	name = "Tom"
	fmt.Printf("这是第一个初始化函数init()\n")
	// 输出变量值
	fmt.Printf("变量name和age的值：%v，%v\n", name, age)
}

// 定义初始化函数init()
func init() {
	// 变量赋值操作
	name = "Tim"
	fmt.Printf("这是第二个初始化函数init()\n")
	// 输出变量值
	fmt.Printf("变量name和age的值：%v，%v\n", name, age)
}

// 主函数main()
func main() {
	fmt.Printf("这是主函数main()\n")
	// 输出变量值
	fmt.Printf("变量name和age的值：%v，%v\n", name, age)
}
