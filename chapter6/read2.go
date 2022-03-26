package main

import (
	"fmt"
	"io/ioutil"
)

func main() {
	filePath := "read2.txt"
	content, err := ioutil.ReadFile(filePath)
	if err != nil {
		fmt.Printf("读取文件出错：%v", err)
	}
	fmt.Printf("%v\n", content)
	fmt.Printf("%v\n", string(content))
}
