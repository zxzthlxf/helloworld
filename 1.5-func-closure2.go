package main

import "fmt"

func visitPrint(list []int, f func(int)) {
	for _, value := range list {
		f(value)
	}
}

func main() {
	sli := []int{1, 6, 8}
	visitPrint(sli, func(value int) {
		fmt.Println(value)
	})
}
