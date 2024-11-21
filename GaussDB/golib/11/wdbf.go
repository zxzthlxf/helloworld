package main

import (
	"log"

	"github.com/san-pang/godbf"
)

func main() {
	dbf := godbf.NewFile("./testdata/test_newfile.DBF", "gbk")
	defer dbf.Close()
	dbf.AddStringField("BEGIN_DATE", 10)
	dbf.AddStringField("END_DATE", 10)
	dbf.AddFloatField("PRICE", 12, 2)
	dbf.AddNumericField("QTY", 8, 2)
	dbf.AddBooleanField("FINISHED")
	dbf.AddStringField("STOCK_CODE", 20)
	// append record, use Post() method to post changes to file
	dbf.Append()
	if err := dbf.SetFieldValue("BEGIN_DATE", "101616"); err != nil {
		panic(err)
	}
	if err := dbf.SetFieldValue("END_DATE", "101634"); err != nil {
		panic(err)
	}
	if err := dbf.SetFieldValue("QTY", "2"); err != nil {
		panic(err)
	}
	if err := dbf.SetFieldValue("PRICE", "12.13"); err != nil {
		panic(err)
	}
	if err := dbf.Post(); err != nil {
		panic(err)
	}
	log.Fatal("Write success!")
}
