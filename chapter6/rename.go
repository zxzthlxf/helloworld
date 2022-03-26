package main

import (
	"fmt"
	"log"
	"os"
)

func main() {
	err := os.Mkdir("dir_name1", 0777)
	if err != nil {
		fmt.Println(err)
	}
	oldName := "dir_name1"
	newName := "dir_name2"
	err = os.Rename(oldName, newName)
	if err != nil {
		log.Fatal(err)
	}
}
