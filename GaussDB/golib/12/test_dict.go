package main

import (
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v2"
)

type Node struct {
	NODE_SETTGROUP   string `yaml:"NODE_SETTGROUP" json:"NODE_SETTGROUP"`
	NODE_SETTUNIT    string `yaml:"NODE_SETTUNIT" json:"NODE_SETTUNIT"`
	NODE_FUNDACCT    string `yaml:"NODE_FUNDACCT" json:"NODE_FUNDACCT"`
	NODE_FUNDUNIT    string `yaml:"NODE_FUNDUNIT" json:"NODE_FUNDUNIT"`
	NODE_STKHOLDUNIT string `yaml:"NODE_STKHOLDUNIT" json:"NODE_STKHOLDUNIT"`
	NODE_SECUID      string `yaml:"NODE_SECUID" json:"NODE_SECUID"`
}

type Dbf struct {
	SH_JSMX_DBF_FIELDS   string `yaml:"SH_JSMX_DBF_FIELDS" json:"SH_JSMX_DBF_FIELDS"`
	SZ_SJSMX1_DBF_FIELDS string `yaml:"SZ_SJSMX1_DBF_FIELDS" json:"SZ_SJSMX1_DBF_FIELDS"`
}

type Trade struct {
	NODE_TRADE_SH_B string `yaml:"NODE_TRADE_SH_B" json:"NODE_TRADE_SH_B"`
	NODE_TRADE_SZ_B string `yaml:"NODE_TRADE_SZ_B" json:"NODE_TRADE_SZ_B"`
	NODE_TRADE_SH_S string `yaml:"NODE_TRADE_SH_S" json:"NODE_TRADE_SH_S"`
	NODE_TRADE_SZ_S string `yaml:"NODE_TRADE_SZ_S" json:"NODE_TRADE_SZ_S"`
}

type Jsmx struct {
	JSMX_BUY    string `yaml:"JSMX_BUY" json:"JSMX_BUY"`
	SJSMX1_BUY  string `yaml:"SJSMX1_BUY" json:"SJSMX1_BUY"`
	JSMX_SALE   string `yaml:"JSMX_SALE" json:"JSMX_SALE"`
	SJSMX1_SALE string `yaml:"SJSMX1_SALE" json:"SJSMX1_SALE"`
}

type Hold struct {
	STKHOLDBOOKKEEPING_SH string `yaml:"STKHOLDBOOKKEEPING_SH" json:"STKHOLDBOOKKEEPING_SH"`
	STKHOLDBOOKKEEPING_SZ string `yaml:"STKHOLDBOOKKEEPING_SZ" json:"STKHOLDBOOKKEEPING_SZ"`
}

func main() {
	//读取全文
	file, err := os.ReadFile("dict.yaml")
	if err != nil {
		fmt.Println("打开文件失败", err.Error())
		os.Exit(1)
	}
	//序列化其中大部分 可以不全，也可以多
	node := Node{}
	err = yaml.Unmarshal(file, &node)
	if err != nil {
		fmt.Println(err.Error())
	}
	fmt.Printf("%+v\n", node)
	// fmt.Println(conf.Spec.Containers[1].Ports[1]["hostPort"])

	dbf := Dbf{}
	err = yaml.Unmarshal(file, &dbf)
	if err != nil {
		fmt.Println(err.Error())
	}
	fmt.Printf("%+v\n", dbf)

	trade := Trade{}
	err = yaml.Unmarshal(file, &trade)
	if err != nil {
		fmt.Println(err.Error())
	}
	fmt.Printf("%+v\n", trade)

	jsmx := Jsmx{}
	err = yaml.Unmarshal(file, &jsmx)
	if err != nil {
		fmt.Println(err.Error())
	}
	fmt.Printf("%+v\n", jsmx)

	hold := Hold{}
	err = yaml.Unmarshal(file, &hold)
	if err != nil {
		fmt.Println(err.Error())
	}
	fmt.Printf("%+v\n", hold)
	fmt.Println(hold.STKHOLDBOOKKEEPING_SZ)
	fmt.Println(hold.STKHOLDBOOKKEEPING_SZ[10])

	result_hold_sh := strings.Split(hold.STKHOLDBOOKKEEPING_SH, ",")
	for _, shod_sh := range result_hold_sh {
		fmt.Println(shod_sh)
	}

}
