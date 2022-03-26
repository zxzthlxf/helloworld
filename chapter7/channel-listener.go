package main

import "fmt"

func foo(i int) chan int {
	ch := make(chan int)
	go func() { ch <- i }()
	return ch
}

func main() {
	ch1, ch2, ch3 := foo(3), foo(6), foo(9)

	ch := make(chan int)

	go func() {
		for {
			select {
			case v1 := <-ch1:
				ch <- v1
			case v2 := <-ch2:
				ch <- v2
			case v3 := <-ch3:
				ch <- v3
			}
		}
	}()

	for i := 0; i < 3; i++ {
		fmt.Println(<-ch)
	}
}
