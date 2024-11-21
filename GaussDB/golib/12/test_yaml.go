package main

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v2"
)

type Conf struct {
	ApiVersion string   `yaml:"apiVersion" json:"apiVersion"`
	Kind       string   `yaml:"kind" json:"kind"`
	Metadata   Metadata `yaml:"metadata" json:"metadata"`
	Spec       Spec     `yaml:"spec" json:"spec"`
}
type Metadata struct {
	Name string `yaml:"name" json:"name"`
}

type Spec struct {
	Containers []Containers `yaml:"containers" json:"containers"`
}
type A struct {
	Abc string `yaml:"abc" json:"abc"`
}
type Containers struct {
	Name            string              `yaml:"name" json:"name"`
	Image           string              `yaml:"image" json:"image"`
	ImagePullPolicy string              `yaml:"imagePullPolicy" json:"imagePullPolicy"`
	Stdin           string              `yaml:"stdin" json:"stdin"`
	Ports           []map[string]string `yaml:"ports" json:"ports"`
}

func main() {
	//读取全文
	file, err := os.ReadFile("config.yaml")
	if err != nil {
		fmt.Println("打开文件失败", err.Error())
		os.Exit(1)
	}
	//序列化其中大部分 可以不全，也可以多
	conf := Conf{}
	err = yaml.Unmarshal(file, &conf)
	if err != nil {
		fmt.Println(err.Error())
	}
	fmt.Printf("%+v\n", conf)
	fmt.Println(conf.Spec.Containers[1].Ports[1]["hostPort"])
	//单独序列化其中一个选项
	a := A{}
	yaml.Unmarshal(file, &a)
	fmt.Println(a.Abc)
}
