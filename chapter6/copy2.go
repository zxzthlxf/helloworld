package main

import (
	"fmt"
	"io"
	"log"
	"os"
)

func DoCopy(srcFileName string, dstFileName string) {
	srcFile, err := os.Open(srcFileName)
	if err != nil {
		log.Fatalf("源文件读取失败，err:%v\n", err)
	}
	defer func() {
		err = srcFile.Close()
		if err != nil {
			log.Fatalf("源文件关闭失败,err:%v\n", err)
		}
	}()

	distFile, err := os.Create(dstFileName)
	if err != nil {
		log.Fatalf("目标文件创建失败,err:%v\n", err)
	}
	defer func() {
		err = distFile.Close()
		if err != nil {
			log.Fatalf("目标文件关闭失败,err%v\n", err)
		}
	}()

	var tmp = make([]byte, 1024*4)
	for {
		n, err := srcFile.Read(tmp)
		n, _ = distFile.Write(tmp[:n])
		if err != nil {
			if err == io.EOF {
				return
			} else {
				log.Fatalf("复制过程中发生错误，错误err:%v\n", err)
			}
		}
	}
}

func main() {
	_, err := os.Create("./test.zip")
	if err != nil {
		fmt.Println(err)
	}
	DoCopy("./test.zip", "./test2.zip")
}
