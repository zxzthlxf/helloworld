package main

import (
	"fmt"
	"net/http"
)

func Home(w http.ResponseWriter, r *http.Request) {
	html := `<html>
	<head>
		<title>用Write()方法返回HTML文档</title>
	</head>
	<body>
		<h1>你好~，欢迎一起学习《Go Web编程实战派——从入门到精通》
	</body>
	</html>`
	w.Write([]byte(html))
}

func main() {
	http.HandleFunc("/", Home)
	err := http.ListenAndServe(":8086", nil)
	if err != nil {
		fmt.Println(err)
	}
}
