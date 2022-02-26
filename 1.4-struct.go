package main

import "fmt"

type Book struct {
	title   string
	author  string
	subject string
	press   string
}

func main() {
	fmt.Println(Book{"Go Web编程实战派——从入门到精通", "廖显东", "Go语言教程", "电子工业出版社"})
	fmt.Println(Book{title: "Go Web编程实战派——从入门到精通", author: "廖显东", subject: "Go语言教程", press: "电子工业出版社"})
	fmt.Println(Book{title: "Go Web编程实战派——从入门到精通", author: "廖显东"})
}
