package main

import (
	"fmt"
	"net/http"
	"net/rpc"
)

type Algorithm int

//参数结构体
type Args struct {
	X, Y int
}

type Response int

func (t *Algorithm) Sum(args *Args, reply *int) error {
	*reply = args.X + args.Y
	fmt.Println("Exec Sum ", reply)
	return nil
}

func main() {
	//实例化
	algorithm := new(Algorithm)
	fmt.Println("Algorithm start", algorithm)
	//注册服务
	rpc.Register(algorithm)
	rpc.HandleHTTP()
	err := http.ListenAndServe(":8808", nil)
	if err != nil {
		fmt.Println("err=====", err.Error())
	}
}
