package main

import (
	"fmt"
	"os"
)

func main() {
	fp, err := os.OpenFile("./open.txt", os.O_CREATE|os.O_APPEND, 0666)

	if err != nil {
		fmt.Println("文件打开失败。")
		return
	}

	defer fp.Close()
}
