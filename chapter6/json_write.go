package main

import (
	"encoding/json"
	"fmt"
	"os"
)

type User struct {
	UserName string
	NickName string `json:"nickname"`
	Email    string
}

func main() {
	user := &User{
		UserName: "Jack",
		NickName: "Ma",
		Email:    "zxzthlxf@163.com",
	}

	data, err := json.Marshal(user)
	if err != nil {
		fmt.Printf("json.Marshal failed,err:", err)
		return
	}

	fmt.Printf("%s\n", string(data))

	file, _ := os.Create("json_write.json")
	defer file.Close()
	file.Write(data)
}
