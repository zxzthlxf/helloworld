package main

import "fmt"

type Triangle struct {
	Bottom float32
	Height float32
}

func (t *Triangle) Area() float32 {
	return (t.Bottom * t.Height) / 2
}

func main() {
	r := Triangle{6, 8}
	fmt.Println(r.Area())
}
