package main

import (
	"fmt"
	"log"
	"net/http"
)

func helloWeb(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello Go Web!")
}

func main() {
	http.HandleFunc("/hello", helloWeb)
	if err := http.ListenAndServe(":8081", nil); err != nil {
		log.Fatal(err)
	}
}
