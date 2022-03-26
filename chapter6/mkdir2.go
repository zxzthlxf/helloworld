package main

import (
	"fmt"
	"os"
)

func main() {
	err := os.MkdirAll("test/test/test", 0777)
	if err != nil {
		fmt.Println(err)
	}
}
