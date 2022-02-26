package main

import "fmt"

type Square struct {
	sideLen float32
}

type Triangle1 struct {
	Bottom float32
	Height float32
}

func (t *Triangle1) Area() float32 {
	return (t.Bottom * t.Height) / 2
}

type Shape interface {
	Area() float32
}

func (sq *Square) Area() float32 {
	return sq.sideLen * sq.sideLen
}

func main() {
	t := &Triangle1{6, 8}
	s := &Square{8}
	shapes := []Shape{t, s}
	for n, _ := range shapes {
		fmt.Println("图形数据：", shapes[n])
		fmt.Println("它的面积是：", shapes[n].Area())
	}
}
