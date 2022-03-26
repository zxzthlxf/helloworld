package main

import (
	"helloworld/chapter3/controller"
	"log"
	"net/http"
)

func main() {
	http.HandleFunc("/getUser", controller.UserController{}.GetUser)
	if err := http.ListenAndServe(":8088", nil); err != nil {
		log.Fatal(err)
	}
}
