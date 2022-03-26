package main

import (
	"fmt"
	"os"
)

func main() {
	fp, err := os.Create("./link1.txt")
	defer fp.Close()
	if err != nil {
		fmt.Println("文件创建失败。")
	}
	err = os.Link("link1.txt", "link2.txt")
	if err != nil {
		fmt.Println("err:", err)
	}
}
