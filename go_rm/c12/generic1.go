package main

import "fmt"

// 定义泛型函数
func sum[K string, V int | float64](m map[K]V) V {
	var s V
	for _, v := range m {
		s += v
	}
	return s
}

func main() {
	// 定义变量
	myints := map[string]int{
		"first":  34,
		"second": 12,
	}

	// 定义变量
	myfloats := map[string]float64{
		"first":  35.98,
		"second": 26.99,
	}

	// 输出计算结果
	fmt.Printf("泛型函数的int：%v\n", sum[string, int](myints))
	fmt.Printf("泛型函数的float64：%v\n", sum[string, float64](myfloats))
}
