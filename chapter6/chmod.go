package main

import (
	"fmt"
	"os"
)

func main() {
	fp, err := os.Create("./chmod1.txt")
	defer fp.Close()
	if err != nil {
		fmt.Println("文件创建失败。")
	}
	fileInfo, err := os.Stat("./chmod1.txt")
	fileMode := fileInfo.Mode()
	fmt.Println(fileMode)
	os.Chmod("./chmod1.txt", 0777)
	fileInfo, err = os.Stat("./chmod1.txt")
	fileMode = fileInfo.Mode()
	fmt.Println(fileMode)
}
