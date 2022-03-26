package main

import (
	"fmt"
	"log"
	"net/rpc/jsonrpc"
)

type Send struct {
	Java, Go string
}

func main() {
	fmt.Println("client start......")
	client, err := jsonrpc.Dial("tcp", "127.0.0.1:8085")
	if err != nil {
		log.Fatal("Dial err=", err)
	}
	send := Send{"Java", "Go"}
	var receive string
	err = client.Call("Programmer.GetSkill", send, &receive)
	if err != nil {
		fmt.Println("Call err=", err)
	}
	fmt.Println("receive", receive)
}
