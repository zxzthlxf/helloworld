package main

import (
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
)

func process(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		t, _ := template.ParseFiles("chapter3/form.html")
		t.Execute(w, nil)
	} else {
		r.ParseForm()
		fmt.Fprintln(w, "表单键值对和URL键值对：", r.Form)
		fmt.Fprintln(w, "表单键值对：", r.PostForm)
	}
}

func multiProcess(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		t, _ := template.ParseFiles("chapter3/form.html")
		t.Execute(w, nil)
	} else {
		r.ParseMultipartForm(1024) //从表单里提取多少字节的数据
		//multipartform是包含2个映射的结构
		fmt.Fprintln(w, "表单键值对：", r.MultipartForm)
	}
}

func upload(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		t, _ := template.ParseFiles("chapter3/upload.html")
		t.Execute(w, nil)
	} else {
		r.ParseMultipartForm(4096)
		fileHeader := r.MultipartForm.File["uploaded"][0]
		file, err := fileHeader.Open()
		if err != nil {
			fmt.Println("error")
			return
		}
		data, err := ioutil.ReadAll(file)
		if err != nil {
			fmt.Println("error!")
			return
		}
		fmt.Fprintln(w, string(data))
	}
}

func main() {
	//	http.HandleFunc("/",upload)
	http.HandleFunc("/", multiProcess)
	//	http.HandleFunc("/",process)
	err := http.ListenAndServe(":8089", nil)
	if err != nil {
		log.Fatal("ListenAndServe:", err)
	}
}
