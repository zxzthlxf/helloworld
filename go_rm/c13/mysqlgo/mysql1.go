package main

import (
	"database/sql"
	"fmt"

	_ "github.com/go-sql-driver/mysql"
)

func main() {
	db, err := sql.Open("mysql", "root:Poc123123@tcp(10.80.30.9:3306)/test")
	if err != nil {
		panic(err.Error())
	}
	defer db.Close()

	fmt.Println(db.Ping())
}
