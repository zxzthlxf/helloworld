package main

import (
	"fmt"
	"os"
)

func main() {
	fout, err := os.Create("./write4.txt")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer fout.Close()
	for i := 0; i < 5; i++ {
		outstr := fmt.Sprintf("%s: %d\r\n", "Hello Go", i)
		fout.WriteString(outstr)
		fout.Write([]byte("I love go\r\n"))
	}
}
