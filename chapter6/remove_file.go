package main

import (
	"fmt"
	"os"
)

func main() {
	err := os.Mkdir("test_remove", 0777)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println("created dir:test_remove")
	fc, err1 := os.Create("./test_remove/test_remove1.txt")
	if err1 != nil {
		fmt.Println(err1)
	}
	fc.Close()
	fmt.Println("created file:test_remove1.txt")
	fc, err1 = os.Create("./test_remove/test_remove2.txt")
	if err1 != nil {
		fmt.Println(err1)
	}
	fc.Close()
	fmt.Println("created file:test_remove3.txt")
	err = os.Remove("./test_remove/test_remove1.txt")
	if err != nil {
		fmt.Printf("removed ./test_remove/test_remove1.txt err : %v\n", err)
	}
	fmt.Println("removed file:./test_remove/test_remove1.txt")
	err = os.RemoveAll("./test_remove")
	if err != nil {
		fmt.Printf("remove all ./test_remove err : %v\n", err)
	}
	fmt.Println("removed all files:./test_remove")
}
