package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/sirupsen/logrus"
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

func parseJsonConfig(filepath string) (conf Config) {
	//打开文件
	file, _ := os.Open(filepath)

	//关闭文件
	defer file.Close()

	conf = Config{}
	//NewDecoder创建一个从file读取并解码json对象的*Decoder，解码器有自己的缓冲，并可能超前读取部分json数据。
	//Decode从输入流读取下一个json编码值并保存在v指向的值里
	err := json.NewDecoder(file).Decode(&conf)
	if err != nil {
		log.Panicln("Error:", err)
	}
	return
}

func main() {
	pdd := parseJsonConfig("config.json")
	fmt.Println(pdd)
	logrus.Println(pdd)
	fmt.Println(pdd.FromDb.Database, "-to-", pdd.ToDb.Database)
	logrus.Infoln(pdd.FromDb.Database, "-to-", pdd.ToDb.Database)
}
