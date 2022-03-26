package main

import "fmt"

func randGenerator() chan int {
	ch := make(chan int)
	go func() {
		for {
			select {
			case ch <- 0:
			case ch <- 1:
			}
		}
	}()
	return ch
}

func main() {
	generator := randGenerator()
	for i := 0; i < 10; i++ {
		fmt.Println(<-generator)
	}
}
