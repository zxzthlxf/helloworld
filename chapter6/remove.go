package main

import (
	"log"
	"os"
)

func main() {
	err := os.RemoveAll("test")
	if err != nil {
		log.Fatal(err)
	}
}
