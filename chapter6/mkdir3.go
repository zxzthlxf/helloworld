package main

import (
	"fmt"
	"os"
	"time"
)

func main() {
	uploadDir := "static/upload/" + time.Now().Format("2022/03/23/")
	err := os.MkdirAll(uploadDir, 0777)
	if err != nil {
		fmt.Println(err)
	}
}
