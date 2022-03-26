package main

import "fmt"

func IntegerGenerator() chan int {
	var ch chan int = make(chan int)
	go func() {
		for i := 0; ; i++ {
			ch <- i
		}
	}()
	return ch
}

func main() {
	generator := IntegerGenerator()
	for i := 0; i < 100; i++ {
		fmt.Println(<-generator)
	}
}
