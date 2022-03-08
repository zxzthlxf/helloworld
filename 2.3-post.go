package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
)

func main() {
	url := "https://www.shirdon.com/comment/add"
	body := "{\"userId\":1,\"articleId\":1,\"comment\":\"这是一条评论\"}"
	response, err := http.Post(url, "*application/x-www-form-urlencoded", bytes.NewBuffer([]byte(body)))
	if err != nil {
		fmt.Println("err", err)
	}
	b, err := ioutil.ReadAll(response.Body)
	fmt.Println(string(b))
}
