package main

import (
	"fmt"
	"os"
)

func main() {
	file, err := os.OpenFile("write1.txt", os.O_CREATE|os.O_RDWR, 0666)
	if err != nil {
		fmt.Println(err)
	}
	defer file.Close()

	content := []byte("你好世界！")
	if _, err = file.Write(content); err != nil {
		fmt.Println(err)
	}
	fmt.Println("写入成功！")
}
