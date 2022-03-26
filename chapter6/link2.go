package main

import (
	"fmt"
	"os"
)

func main() {
	fp, err := os.Create("./link2.txt")
	defer fp.Close()
	if err != nil {
		fmt.Println("文件创建失败。")
	}
	err = os.Symlink("link2.txt", "link3.txt")
	if err != nil {
		fmt.Println("err:", err)
	}
}
