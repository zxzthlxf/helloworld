package main

import "fmt"

func main() {
	var arr [6]int
	var i, j int
	for i = 0; i < 6; i++ {
		arr[i] = i + 66
	}
	for j = 0; j < 6; j++ {
		fmt.Printf("Array[%d] = %d\n", j, arr[j])
	}
}
