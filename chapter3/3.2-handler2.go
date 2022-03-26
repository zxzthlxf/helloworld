package main

import (
	"fmt"
	"log"
	"net/http"
)

type WelcomeHandler struct {
	Language string
}

func (h WelcomeHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "%s", h.Language)
}

func main() {
	mux := http.NewServeMux()
	mux.Handle("/cn", WelcomeHandler{"欢迎一起来学Go Web"})
	mux.Handle("/en", WelcomeHandler{"Welcome you,let's learn Go Web!"})
	server := &http.Server{
		Addr:    ":8082",
		Handler: mux,
	}

	if err := server.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}
