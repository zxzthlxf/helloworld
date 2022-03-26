package main

import (
	"github.com/julienschmidt/httprouter"
	"log"
	"net/http"
)

func main() {
	router := httprouter.New()
	router.ServeFiles("/static/*filepath", http.Dir("./files"))
	log.Fatal(http.ListenAndServe(":8086", router))
}
