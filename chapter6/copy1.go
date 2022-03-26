package main

import (
	"fmt"
	"io"
	"os"
)

func main() {
	fc, err := os.Create("./test_copy1.zip")
	if err != nil {
		fmt.Println(err)
	}
	fc.Close()

	srcFile, err := os.Open("./test_copy1.zip")
	if err != nil {
		fmt.Printf("open file err = %v\n", err)
		return
	}
	defer srcFile.Close()

	dstFile, err := os.OpenFile("./test_copy2.zip", os.O_WRONLY|os.O_CREATE, 0755)

	if err != nil {
		fmt.Printf("open file err = %v\n", err)
		return
	}

	defer dstFile.Close()

	result, err := io.Copy(dstFile, srcFile)

	if err == nil {
		fmt.Println("复制成功，复制的字节数为：", result)
	}

}
