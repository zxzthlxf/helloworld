package model

import (
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
)

type User struct {
	Uid   int
	Name  string
	Phone string
}

var db *sql.DB

func init() {
	db, _ = sql.Open("mysql", "root:Jzyz.8888@tcp(127.0.0.1:3306)/chapter3")
}

func GetUser(uid int) (u User) {
	err := db.QueryRow("select uid,name,phone from `user` where uid=?", uid).Scan(&u.Uid, &u.Name, &u.Phone)
	if err != nil {
		fmt.Printf("scan failed,err:%v\n", err)
		return
	}
	return u
}
