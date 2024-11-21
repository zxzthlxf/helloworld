package main

import (
	"database/sql"
	"fmt"

	// _ "odbc/driver"
	_ "github.com/alexbrainman/odbc" // google's odbc driver
)

func main() {
	fmt.Printf("%s\n", "数据库链接测试")
	conn, err := sql.Open("odbc", "DSN=go_test;UID=godb;PWD=SZtest898")
	if err != nil {
		// fmt.Println("链接错误")
		fmt.Println("Error:", err)
		return
	} else {
		fmt.Printf("%s\n", "数据库链接成功！")
	}
	defer conn.Close()
	fmt.Printf("%s\n", "构建查询")
	stmt, err := conn.Prepare("select 666;")
	if err != nil {
		fmt.Println("查询异常: ", err)
		return
	}
	defer stmt.Close()
	row, err := stmt.Query()
	if err != nil {
		fmt.Println("查询错误：", err)
	}
	defer row.Close()
	fmt.Printf("%s\n", "数据集显示")
	for row.Next() {
		var id int
		if err := row.Scan(&id); err == nil {
			fmt.Println(id)
		}
	}

}
