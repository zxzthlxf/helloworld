package main

import "fmt"

func main() {
	var a interface{} = func(a int) string {
		return fmt.Sprintf("d:%d", a)
	}
	switch b := a.(type) {
	case nil:
		println("nil")
	case *int:
		println(*b)
	case func(int) string:
		println(b(66))
	case fmt.Stringer:
		fmt.Println(b)
	default:
		println("unknown")
	}
}
