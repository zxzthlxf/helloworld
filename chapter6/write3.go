package main

import (
	"fmt"
	"os"
)

func main() {
	file, err := os.Create("WriteString.txt")
	if err != nil {
		panic(err)
	}
	defer file.Close()
	_, err1 := file.WriteString("Go Web编程实战派——从入门到精通！")
	if err1 != nil {
		panic(err1)
	} else {
		fmt.Println("写入成功！")
	}
}
