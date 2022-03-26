package main

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type Greeting1 struct {
	Message string `json:"message"`
}

func Hello1(w http.ResponseWriter, r *http.Request) {
	greeting := Greeting1{
		"欢迎一起学习《Go Web编程实战派——从入门到精通》",
	}
	message, _ := json.Marshal(greeting)
	w.Header().Set("Content-Type", "application/json")
	w.Write(message)
}

func main() {
	http.HandleFunc("/", Hello1)
	err := http.ListenAndServe(":8086", nil)
	if err != nil {
		fmt.Println(err)
	}
}
