package main

import (
	"fmt"
	"github.com/julienschmidt/httprouter"
	"log"
	"net/http"
)

func Index1(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	panic("error")
}

func main() {
	router := httprouter.New()
	router.GET("/", Index1)
	router.PanicHandler = func(w http.ResponseWriter, r *http.Request, v interface{}) {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "error:%s", v)
	}
	log.Fatal(http.ListenAndServe(":8085", router))
}
