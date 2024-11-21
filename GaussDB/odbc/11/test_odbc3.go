package main

import (
	"encoding/csv"
	"fmt"
	"log"
	"os"

	_ "github.com/alexbrainman/odbc" // google's odbc driver
	// "github.com/axgle/mahonia"
	"github.com/go-xorm/xorm"
	"xorm.io/core"
)

// 定义全局变量
var (
	engine *xorm.Engine
	dsn    = "Driver={PostgreSQL Unicode};Server=10.203.60.62;Port=30100;UID=comm;PWD=SZtest@123;Database=macbs_db01;Pooling=true;Min Pool Size=1"
)

/* type Address struct {
	Addressid  int64  `xorm:"addressid"`
	Address1   string `xorm:"address1"`
	Address2   string `xorm:"address2"`
	City       string `xorm:"city"`
	Postalcode string `xorm:"postalcode"`
} */

/* type FlowChart struct {
	Stepno   int
	Stepname string
	Begtime  int
	Endtime  int
} */

// 导出csv文件
func ExportCsv(filePath string, data [][]string) {
	fp, err := os.Create(filePath) //创建文件句柄
	if err != nil {
		log.Fatalf("创建文件["+filePath+"]句柄失败,%v", err)
		return
	}
	defer fp.Close()
	fp.WriteString("\xEF\xBB\xBF") //写入UTF-8 BOM
	w := csv.NewWriter(fp)         //创建一个新的写入文件流
	w.WriteAll(data)
	w.Flush()
	// fmt.Print("导出CSV完成！")
	log.Print("导出CSV完成！")
}

func main() {
	//连接数据库
	engine, err := xorm.NewEngine("odbc", dsn)
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
	//定义导出的文件名
	filename := "./exportUsers.csv"

	//定义一个二维数组
	column := [][]string{{"流程步骤号", "步骤名称", "执行开始时间", "执行结束时间"}}

	// 1) sql查询
	results, err := engine.Query("select stepno, stepname, begtime, endtime from comm.comm_flowchart")
	if err != nil {
		fmt.Println("查询出错:", err)
		return
	}
	// var flowcharts []FlowChart
	for _, row := range results {
		/* 		stepNo, _ := strconv.Atoi(string(row["stepno"]))
		   		stepName := string(row["stepname"])
		   		// 		begTime, _ := time.Parse(time.RFC3339, string(row["begtime"]))
		   		//   		endTime, _ := time.Parse(time.RFC3339, string(row["endtime"]))
		   		begTime, _ := strconv.Atoi(string(row["begtime"]))
		   		endTime, _ := strconv.Atoi(string(row["endtime"]))
		   		flowchart := FlowChart{
		   			Stepno:   stepNo,
		   			Stepname: stepName,
		   			Begtime:  begTime,
		   			Endtime:  endTime,
		   		}
		   		flowcharts = append(flowcharts, flowchart) */
		stepNo := string(row["stepno"])
		stepName := string(row["stepname"])
		// 		begTime, _ := time.Parse(time.RFC3339, string(row["begtime"]))
		//   	endTime, _ := time.Parse(time.RFC3339, string(row["endtime"]))
		begTime := string(row["begtime"])
		endTime := string(row["endtime"])

		str := []string{}
		str = append(str, stepNo)
		str = append(str, stepName)
		str = append(str, begTime)
		str = append(str, endTime)
		column = append(column, str)

		// column = append(column, flowcharts)
	}
	log.Printf("查询结果：", column)
	fmt.Println("*********************************")

	//导出
	ExportCsv(filename, column)

}
