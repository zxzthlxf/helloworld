package main

import (
	"fmt"
	"net"
)

func main() {
	udpAddr, err := net.ResolveUDPAddr("udp", "127.0.0.1:8012")
	if err != nil {
		fmt.Println("ResolveUDPAddr err:", err)
		return
	}
	conn, err := net.ListenUDP("udp", udpAddr)
	if err != nil {
		fmt.Println("ListenUDP err:", err)
		return
	}
	defer conn.Close()

	buf := make([]byte, 1024)
	n, raddr, err := conn.ReadFromUDP(buf)
	if err != nil {
		return
	}
	fmt.Println("客户端发送：", string(buf[:n]))
	_, err = conn.WriteToUDP([]byte("你好，客户端，我是服务器端"), raddr)
	if err != nil {
		fmt.Println("WriteToUDP err:", err)
		return
	}
}
