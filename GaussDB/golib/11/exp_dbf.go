package main

import (
	"database/sql"
	"fmt"

	_ "github.com/mattn/go-odbc"
)

func main() {
	db, err := sql.Open("odbc", "DSN=your_dsn;UID=your_username;PWD=your_password")
	if err != nil {
		fmt.Println("Error connecting to database:", err)
		return
	}
	defer db.Close()

	rows, err := db.Query("SELECT * FROM your_table")
	if err != nil {
		fmt.Println("Error querying database:", err)
		return
	}
	defer rows.Close()

	// create a new DBF file
	file, err := dbf.Create("output.dbf", []dbf.Field{
		{Name: "field1", Type: dbf.TypeString, Size: 50},
		{Name: "field2", Type: dbf.TypeNumeric, Size: 10},
	})
	if err != nil {
		fmt.Println("Error creating DBF file:", err)
		return
	}
	defer file.Close()

	// iterate over query results and write to DBF file
	for rows.Next() {
		var field1 string
		var field2 int
		err := rows.Scan(&field1, &field2)
		if err != nil {
			fmt.Println("Error scanning row:", err)
			continue
		}

		err = file.Write([]interface{}{field1, field2})
		if err != nil {
			fmt.Println("Error writing to DBF file:", err)
			continue
		}
	}

	fmt.Println("DBF file exported successfully")
}
