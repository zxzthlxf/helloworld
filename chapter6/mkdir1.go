package main

import (
	"fmt"
	"os"
)

func main() {
	err := os.Mkdir("test", 0777)
	if err != nil {
		fmt.Println(err)
	}
}
