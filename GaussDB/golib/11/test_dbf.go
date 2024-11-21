package main

import (
	"fmt"

	"github.com/LindsayBradford/go-dbf/godbf"
)

func main() {
	// 创建一个DBF文件
	dbf, err := godbf.NewFromFile("output.dbf", "GBK")
	if err != nil {
		fmt.Println("Error creating DBF file:", err)
		return
	}

	// 添加数据到DBF文件
	dbf.AddRow([]interface{}{"Alice", 25})
	dbf.AddRow([]interface{}{"Bob", 30})

	// 保存DBF文件
	err = dbf.WriteFile()
	if err != nil {
		fmt.Println("Error writing DBF file:", err)
		return
	}

	fmt.Println("Data exported to DBF file successfully")
}
