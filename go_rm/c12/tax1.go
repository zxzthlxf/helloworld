package main

import (
	"fmt"
)

func main() {
	var distance, cost float64
	// 定义匿名函数
	// 使用defer将匿名函数在程序结束之前执行
	defer func() {
		// 使用recover()捕捉异常
		if err := recover(); err != nil {
			// err不为空值，说明主动抛出异常
			fmt.Printf("捕捉异常：%v\n", err)
		} else {
			// err为空值，说明程序没有抛出异常
			// 输出当前公里数所付车费
			fmt.Printf("当前路程数：%v，车费：%v\n", distance, cost)
		}
	}()
	// 输出操作提示
	fmt.Printf("输入公里数km：\n")
	// 存储用户输入的数据计算车费
	if distance <= 0 {
		panic("公里数小于等于0，无法计算车费")
	} else if distance <= 3 {
		cost = 13.0
	} else if distance <= 10 {
		cost = 13.0 + (distance-3)*2.3
	} else {
		cost = 13.0 + (10-3)*2.3 + (distance-10)*3.2
	}
}
