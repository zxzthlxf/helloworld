package main

import "fmt"

func Sum(s []int, ch chan int) {
	sum := 0
	for _, v := range s {
		sum += v
	}
	ch <- sum
}

func main() {
	s := []int{6, 7, 8, -9, 1, 8}
	ch := make(chan int)
	go Sum(s[:len(s)/2], ch)
	go Sum(s[len(s)/2:], ch)
	a, b := <-ch, <-ch
	fmt.Println(a, b, a+b)
}
