package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
	"strings"
)

func Client() {
	conn, err := net.Dial("tcp", "127.0.0.1:8088")
	if err != nil {
		log.Fatal(err)
	}

	reader := bufio.NewReader(os.Stdin)
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			log.Fatal(err)
		}
		line = strings.Trim(line, "\r\n")

		if line == "exit" {
			fmt.Println("用户退出客户端")
			break
		}
		conent, err := conn.Write([]byte(line + "\n"))
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("客户端发送了 %d 字节的数据到服务器端\n", conent)
	}
}

func main() {
	Client()
}
