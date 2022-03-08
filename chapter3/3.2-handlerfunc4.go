package main

import (
	"fmt"
	"log"
	"net/http"
)

func hiHandler1(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hi, Go HandleFunc")
}

type welcomeHandler struct {
	Name string
}

func (h welcomeHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "hi, %s", h.Name)
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/hi", hiHandler1)

	mux.Handle("/welcome/goweb", welcomeHandler{Name: "Hi, Go Handle"})

	server := &http.Server{
		Addr:    ":8085",
		Handler: mux,
	}
	if err := server.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}
