package main

import (
	"fmt"
	"time"
)

func Timer(duration time.Duration) chan bool {
	ch := make(chan bool)

	go func() {
		time.Sleep(duration)
		ch <- true
	}()
	return ch
}

func main() {
	timeout := Timer(5 * time.Second)
	for {
		select {
		case <-timeout:
			fmt.Println("already 5s!")
			return
		}
	}
}
