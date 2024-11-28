package main

import "fmt"

func main() {
	/* 值类型变量 */
	s := "hello"
	fmt.Printf("变量s的内存地址：%p，变量值为：%v\n", s, s)
	fmt.Printf("变量s的内存地址：%p，变量值为：%v\n", &s, &s)
	// 将变量赋值给另一个变量，执行深拷贝方式
	ss := s
	fmt.Printf("变量ss的内存地址：%p，变量值为：%v\n", ss, ss)
	fmt.Printf("变量ss的内存地址：%p，变量值为：%v\n", &ss, &ss)

	/* 引用类型变量 */
	m := make(map[string]interface{})
	m["name"] = "Tom"
	fmt.Printf("变量m的内存地址：%p，变量值为：%v\n", m, m)
	fmt.Printf("变量m的内存地址：%p，变量值为：%v\n", &m, &m)
	// 将变量赋值给另一个变量，执行浅拷贝方式
	mm := m
	fmt.Printf("变量mm的内存地址：%p，变量值为：%v\n", mm, mm)
	fmt.Printf("变量mm的内存地址：%p，变量值为：%v\n", &mm, &mm)
	//修改某个变量的值，另一个变量随之变化
	mm["name"] = "Tim"
	fmt.Printf("变量m的内存地址：%p，变量值为：%v\n", m, m)
	fmt.Printf("变量m的内存地址：%p，变量值为：%v\n", &m, &m)
	fmt.Printf("变量mm的内存地址：%p，变量值为：%v\n", mm, mm)
	fmt.Printf("变量mm的内存地址：%p，变量值为：%v\n", &mm, &mm)
}
