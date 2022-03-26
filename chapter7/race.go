package main

import "fmt"

func main() {
	c := make(chan bool)
	m := make(map[string]string)
	go func() {
		m["a"] = "one"
		c <- true
	}()
	m["b"] = "two"
	<-c
	for k, v := range m {
		fmt.Println(k, v)
	}
}
