package main

import (
	"fmt"

	"github.com/LindsayBradford/go-dbf/godbf"
)

func main() {
	dbfTable, err := godbf.NewFromFile("exampleFile.dbf", "UTF8")
	if err != nil {
		fmt.Println("Error creating DBF file:", err)
		return
	}
	exampleList := make(ExampleList, dbfTable.NumberOfRecords())

	for i := 0; i < dbfTable.NumberOfRecords(); i++ {
		exampleList[i] = new(ExampleListEntry)

		exampleList[i].someColumnId, err = dbfTable.FieldValueByName(i, "SOME_COLUMN_ID")
	}
}
