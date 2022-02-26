package main

import "fmt"

func main() {
	f := func(data int) {
		fmt.Println("hi, this is a closure", data)
	}
	f(6)

	func(data int) {
		fmt.Println("hi, this is a closure, directly", data)
	}(8)
}
