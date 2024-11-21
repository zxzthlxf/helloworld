package main

import (
	"fmt"

	"github.com/san-pang/godbf"
)

func main() {
	dbf, err := godbf.LoadFrom("./testdata/ZRTBDQXFL.DBF", "gbk")
	if err != nil {
		panic(err)
	}
	defer dbf.Close()
	for !dbf.EOF() {
		if err := dbf.Next(); err != nil {
			panic(err)
		}
		// read record
		var d1 = dbf.StringValueByNameX("jllx")
		var d2 = dbf.StringValueByNameX("scdm")
		var d3 = dbf.StringValueByNameX("zqdm")
		var d4 = dbf.StringValueByNameX("qx")
		var d5 = dbf.StringValueByNameX("rrfl")
		var d6 = dbf.StringValueByNameX("rcfl")
		var d7 = dbf.StringValueByNameX("jyrq")
		// update record, use Post() method to post changes to file
		if err := dbf.SetFieldValue("jllx", "2"); err != nil {
			panic(err)
		}
		if err := dbf.SetFieldValue("rrfl", "0.0130000"); err != nil {
			panic(err)
		}
		if err := dbf.Post(); err != nil {
			panic(err)
		}
		fmt.Printf("jllx:%s,scdm:%s,zqdm:%s,qx:%s,rrfl:%s,rcfl:%s,jyrq:%s\n", d1, d2, d3, d4, d5, d6, d7)
	}
}
