package main

import "fmt"

func main() {
	var score int = 100
	var name string = "Barry"
	fmt.Printf("%p %p", &score, &name)
}
