package main

import (
	"fmt"

	_ "github.com/alexbrainman/odbc" // google's odbc driver
	"github.com/axgle/mahonia"
	"github.com/go-xorm/xorm"
	"xorm.io/core"
)

/* type Address struct {
	Addressid  int64  `xorm:"addressid"`
	Address1   string `xorm:"address1"`
	Address2   string `xorm:"address2"`
	City       string `xorm:"city"`
	Postalcode string `xorm:"postalcode"`
} */

// 字符串解码函数，处理中文乱码
func ConvertToString(src string, srcCode string, tagCode string) string {
	srcCoder := mahonia.NewDecoder(srcCode)
	srcResult := srcCoder.ConvertString(src)
	tagCoder := mahonia.NewDecoder(tagCode)
	_, cdata, _ := tagCoder.Translate([]byte(srcResult), true)
	result := string(cdata)
	return result
}

func main() {
	engine, err := xorm.NewEngine("odbc", "Driver={PostgreSQL Unicode};Server=10.80.30.41,10.80.30.42,10.80.30.43;Port=30100;UID=godb;PWD=SZtest898;Database=go_test;Pooling=true;Min Pool Size=1")
	if err != nil {
		fmt.Println("new engine got error:", err)
		return
	}
	engine.ShowSQL(true) //控制台打印出生成的SQL语句；
	engine.Logger().SetLevel(core.LOG_DEBUG)
	if err := engine.Ping(); err != nil {
		fmt.Println("ping got error:", err)
		return
	}

	// 1) sql查询
	results, err := engine.Query("select * from test")
	if err != nil {
		fmt.Println("查询出错:", err)
		return
	}
	for i, e := range results {
		fmt.Printf("%v\t", i)
		for k, v := range e {
			fmt.Printf("%v=%v\t", k, ConvertToString(string(v), "utf-8", "utf-8"))
		}
		fmt.Printf("\n")
	}
	fmt.Println("*********************************")
}
