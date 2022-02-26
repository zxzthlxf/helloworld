package main

import "fmt"

func exchange(c, d *int) {
	t := *c
	*c = *d
	*d = t
}

func main() {
	a, b := 6, 8
	exchange(&a, &b)
	fmt.Println(a, b)
}
