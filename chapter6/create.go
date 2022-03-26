package main

import (
	"fmt"
	"os"
)

func main() {
	fp, err := os.Create("./demo.txt")
	fmt.Println(fp, err)
	fmt.Printf("%T", fp)

	if err != nil {
		fmt.Println("文件创建失败。")
		return
	}
	defer fp.Close()
}
