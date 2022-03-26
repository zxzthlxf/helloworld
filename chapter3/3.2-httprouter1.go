package main

import (
	"github.com/julienschmidt/httprouter"
	"net/http"
)

func main() {
	router := httprouter.New()
	router.GET("/default", func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		w.Write([]byte("default get"))
	})
	router.POST("/default", func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		w.Write([]byte("default post"))
	})
	router.GET("/user/:name", func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		w.Write([]byte("user name:" + p.ByName("name")))
	})
	/*
		router.GET("/user/*name", func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
			w.Write([]byte("user name:" + p.ByName("name")))
		})
	*/
	http.ListenAndServe(":8083", router)
}
