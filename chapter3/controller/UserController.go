package controller

import (
	"fmt"
	"helloworld/chapter3/model"
	"html/template"
	"net/http"
	"strconv"
)

type UserController struct {
}

func (c UserController) GetUser(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	uid, _ := strconv.Atoi(query["uid"][0])

	user := model.GetUser(uid)
	fmt.Println(user)

	t, _ := template.ParseFiles("chapter3/view/t3.html")
	userInfo := []string{strconv.Itoa(user.Uid), user.Name, user.Phone}
	t.Execute(w, userInfo)
}
