package main

import "net/http"

func SayHello(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Hello"))
}

func main() {
	http.HandleFunc("/hello", SayHello)
	http.ListenAndServe(":8080", nil)
}
