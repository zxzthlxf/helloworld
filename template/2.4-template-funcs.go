package main

import (
	"fmt"
	"html/template"
	"io/ioutil"
	"net/http"
)

func Welcome() string {
	return "Welcome"
}

func Doing(name string) string {
	return name + ", Learning Go Web template "
}

func sayHello(w http.ResponseWriter, r *http.Request) {
	htmlByte, err := ioutil.ReadFile("template/funcs.html")
	if err != nil {
		fmt.Println("read html failed, err:", err)
		return
	}

	loveGo := func() string {
		return "欢迎一起学习《Go Web编程实战派——从入门到精通》"
	}
	tmpl1, err := template.New("funcs").Funcs(template.FuncMap{"loveGo": loveGo}).Parse(string(htmlByte))
	if err != nil {
		fmt.Println("create template failed, err:", err)
		return
	}
	funcMap := template.FuncMap{
		"Welcome": Welcome,
		"Doing":   Doing,
	}
	name := "Shirdon"
	tmpl2, err := template.New("test").Funcs(funcMap).Parse("{{Welcome}}\n{{Doing .}}\n")
	if err != nil {
		panic(err)
	}
	tmpl1.Execute(w, name)
	tmpl2.Execute(w, name)
}

func main() {
	http.HandleFunc("/", sayHello)
	http.ListenAndServe(":8087", nil)
}
