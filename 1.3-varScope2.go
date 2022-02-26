package main

import "fmt"

var global int

func main() {
	var local1, local2 int
	local1 = 8
	local2 = 10
	global = local1 + local2

	fmt.Printf("local1 = %d, local2 = %d and g = %d\n", local1, local2, global)
}
