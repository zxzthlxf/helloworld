package main

import (
	"log"
)

func main() {
	/* 	//创建一个新的DBF文件
	   	f, err := os.Create("example.dbf")
	   	if err != nil {
	   		log.Fatal(err)
	   	}
	   	defer f.Close()

	   	//创建一个新的DBF表
	   	header := dbf.NewHeader()
	   	header.Fields = []dbf.Field{} */

	dbfTable, err := godbf.NewFromFile("exampleFile.dbf", "UTF8")
	if err != nil {
		log.Fatal(err)
	}

	exampleList := make(ExampleList, dbfTable.NumberOfRecords())

	for i := 0; i < dbfTable.NumberOfRecords(); i++ {
		exampleList[i] = new(ExampleListEntry)

		exampleList[i].someColumnId, err = dbfTable.FieldValueByName(i, "SOME_COLUMN_ID")
	}

}
