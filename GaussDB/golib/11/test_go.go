package main

import (
	"database/sql"
	"fmt"
	"log"
	"strings"
	"time"

	_ "huawei.com/openGauss-go"
)

func main() {
	str := "host=10.80.30.41 port=30100 user=godb password=SZtest898 dbname=go_test sslmode=disable"
	db, err := sql.Open("opengauss", str)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	err = db.Ping()
	if err != nil {
		log.Fatal(err)
	}

	sqls := []string{
		"drop table if exists testExec",
		"create table testExec(f1 int,f2 varchar(20),f3 number,f4 timestamptz,f5 boolean)",
		"insert into testExec values(1,'abcdefg',123.3,'2022-02-08 10:30:43.31 +08',true)",
		"insert into testExec values(:f1,:f2,:f3,:f4,:f5)",
	}
	inF1 := []int{2, 3, 4, 5, 6}
	intF2 := []string{"hello world", "华为", "北京", "nanjing", "研究所"}
	intF3 := []float64{641.43, 431.54, 5423.52, 665537.63, 6503.1}
	intF4 := []time.Time{
		time.Date(2022, 2, 8, 10, 35, 43, 623431, time.Local),
		time.Date(2022, 2, 10, 10, 35, 43, 623431, time.Local),
		time.Date(2022, 2, 12, 10, 35, 43, 623431, time.Local),
		time.Date(2022, 2, 14, 10, 35, 43, 623431, time.Local),
		time.Date(2022, 2, 16, 10, 35, 43, 623431, time.Local),
	}
	intF5 := []bool{false, true, false, true, true}

	for _, s := range sqls {
		if strings.Contains(s, ":f") {
			for i, _ := range inF1 {
				_, err := db.Exec(s, inF1[i], intF2[i], intF3[i], intF4[i], intF5[i])
				if err != nil {
					log.Fatal(err)
				}
			}
		} else {
			_, err = db.Exec(s)
			if err != nil {
				log.Fatal(err)
			}
		}
	}
	var f1 int
	var f2 string
	var f3 float64
	var f4 time.Time
	var f5 bool
	err = db.QueryRow("select * from testExec").Scan(&f1, &f2, &f3, &f4, &f5)
	if err != nil {
		log.Fatal(err)
	} else {
		fmt.Printf("f1:%v, f2:%v, f3:%v, f4:%v, f5:%v\n", f1, f2, f3, f4, f5)
	}

	row, err := db.Query("select * from testExec where f1 > :1", 1)
	if err != nil {
		log.Fatal(err)
	}
	defer row.Close()

	for row.Next() {
		err = row.Scan(&f1, &f2, &f3, &f4, &f5)
		if err != nil {
			log.Fatal(err)
		} else {
			fmt.Printf("f1:%v, f2:%v, f3:%v, f4:%v, f5:%v\n", f1, f2, f3, f4, f5)
		}
	}
}
