package main

import (
	"fmt"
	"log"
	"net/http"
)

func hi1(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hi Web")
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/", hi1)

	server := &http.Server{
		Addr:    ":8081",
		Handler: mux,
	}

	if err := server.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}
