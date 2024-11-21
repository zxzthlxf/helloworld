package main

import (
	"bufio"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"os"

	_ "github.com/go-sql-driver/mysql"
)

type Config struct {
	SqlType string `json:"sqlType"`
	FromDb  struct {
		Host     string `json:"host"`
		Port     int    `json:"port"`
		UserName string `json:"userName"`
		Password string `json:"password"`
		Database string `json:"database"`
		Charset  string `json:"charset"`
	} `json:"fromDb`
	ToDb struct {
		Host     string `json:"host"`
		Port     int    `json:"port"`
		UserName string `json:"userName"`
		Password string `json:"password"`
		Database string `json:"database"`
		Charset  string `json:"charset"`
	} `json:"toDb`
}

var cfg Config
var fromDb *sql.DB
var toDb *sql.DB

func readConfig() {
	bts, err := os.ReadFile("./config.json")
	if err != nil {
		log.Panicln("ReadFile error:", err)
	}
	err = json.Unmarshal(bts, &cfg)
	if err != nil {
		log.Panicln("Unmarshal config file error:", err)
	}
}

func connectSQL() {
	fromDSN := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s", cfg.FromDb.UserName, cfg.FromDb.Password, cfg.FromDb.Host, cfg.FromDb.Port, cfg.FromDb.Database, cfg.FromDb.Charset)
	// fmt.Println("fromDNS:", fromDSN)
	//sql.Open("mysql", "root:Poc123123@tcp(10.80.30.9:3306)/chapter4")
	fDb, err := sql.Open(cfg.SqlType, fromDSN)
	if err != nil {
		log.Panicln("sql Open error:", err)
	}

	toDSN := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s", cfg.ToDb.UserName, cfg.ToDb.Password, cfg.ToDb.Host, cfg.ToDb.Port, cfg.ToDb.Database, cfg.ToDb.Charset)

	tDb, err := sql.Open(cfg.SqlType, toDSN)
	if err != nil {
		fDb.Close()
		log.Panicln("sql Open error:", err)
	}

	fromDb = fDb
	toDb = tDb
}

func getTables() []string {
	/* 	sf, err := os.Open("./success.txt")
	   	if err != nil {
	   		log.Panicln("Open success file error:", err)
	   	}
	   	defer sf.Close()

	   	tableMap := make(map[string]bool)
	   	scanner1 := bufio.NewScanner(sf)
	   	for scanner1.Scan() {
	   		line := scanner1.Text()
	   		if len(line) < 1 {
	   			continue
	   		}
	   		tableMap[line] = true
	   	} */

	f, err := os.Open("./tableList.txt")
	if err != nil {
		log.Panicln("Opentable list error:", err)
	}
	defer f.Close()
	tableMap := make(map[string]bool)
	scanner := bufio.NewScanner(f)
	var tables []string
	for scanner.Scan() {
		line := scanner.Text()
		if len(line) < 1 {
			continue
		}
		if tableMap[line] {
			continue
		}

		tables = append(tables, line)
	}
	return tables
}

func importData(table string) {
	querySql := "select * from " + table + ";"
	stmt, err := fromDb.Prepare(querySql)
	if err != nil {
		log.Panicln("import data from db Prepare error:", err)
	}

	rows, err := stmt.Query()
	if err != nil {
		log.Panicln("import data from db Query error:", err)
	}
	cols, err := rows.Columns()
	if err != nil {
		log.Panicln("import data from db Columns error:", err)
	}
	vals := make([][]byte, len(cols))
	scans := make([]interface{}, len(cols))
	for i := range vals {
		scans[i] = &vals[i]
	}

	//先删除再插入
	truncateSql := "truncate " + table + ";"
	//先清空表再插入
	truncateStmt, err := toDb.Prepare(truncateSql)
	if err != nil {
		log.Panicln("import data to db Prepare truncate error:", err)
	}
	_, err = truncateStmt.Exec()
	if err != nil {
		log.Panicln("import data to db Exec truncate sql error:", err)
	}

	//根据字段数量拼接insert语句
	insertSql := "insert into " + table + " ("
	insertTail := ""
	for i := range cols {
		//使用``将字段括起来防止出现表的字段和关键字一样时报1064 sql错误
		insertSql += "`" + cols[i] + "`"
		insertTail += "?"
		if i+1 == len(cols) {
			break
		}
		insertSql += ","
		insertTail += ","
	}
	insertSql += ") values (" + insertTail + ");"
	insertStmt, err := toDb.Prepare(insertSql)
	if err != nil {
		log.Panicln("import data to db Prepare error:", err)
	}

	for rows.Next() {
		//动态获取字段信息
		err = rows.Scan(scans...)
		if err != nil {
			log.Panicln("import data from db Scan error:", err)
		}
		//动态插入数据
		_, err = insertStmt.Exec(scans...)
		if err != nil {
			log.Panicln("import data to db Exec error:", err)
		}
	}
}

func main() {
	readConfig()

	connectSQL()
	defer fromDb.Close()
	defer toDb.Close()

	tables := getTables()
	f, err := os.OpenFile("./success.txt", os.O_CREATE, 0666)
	if err != nil {
		log.Panicln("Open success file error:", err, "Creating file")
		// os.Create("./success.txt")
		// fmt.Println("Create file success!")
	}
	defer f.Close()

	for index := range tables {
		log.Println("import data to table:", tables[index])
		importData(tables[index])
		log.Println("import data to tables:", tables[index], " Success")

		f.WriteString(tables[index] + "\r\n")
	}
}
