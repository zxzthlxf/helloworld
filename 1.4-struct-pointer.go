package main

import "fmt"

type Books2 struct {
	title   string
	author  string
	subject string
	press   string
}

func main() {
	var bookGo Books2
	var bookPython Books2

	bookGo.title = "Go Web编程实战派——从入门到精通"
	bookGo.author = "廖显东"
	bookGo.subject = "Go语言教程"
	bookGo.press = "电子工业出版社"

	bookPython.title = "Python教程xxx"
	bookPython.author = "张三"
	bookPython.subject = "Python语言教程"
	bookPython.press = "xxx出版社"

	printBook1(&bookGo)

	printBook1(&bookPython)

}

func printBook1(book *Books2) {
	fmt.Printf("Book title : %s\n", book.title)
	fmt.Printf("Book author : %s\n", book.author)
	fmt.Printf("Book subject : %s\n", book.subject)
	fmt.Printf("Book press : %s\n", book.press)
}
