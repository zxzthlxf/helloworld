package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
)

type MongoConfig struct {
	MongoAddr       string
	MongoPoolLimit  int
	MongoDb         string
	MongoCollection string
}

type Config struct {
	Port  string
	Mongo MongoConfig
}

type JsonStruct struct{}

func NewJsonStruct() *JsonStruct {
	return &JsonStruct{}
}

func (js *JsonStruct) Load(filename string, v interface{}) {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return
	}

	err = json.Unmarshal(data, v)
	if err != nil {
		return
	}

}

func main() {
	JsonParse := NewJsonStruct()
	v := Config{}
	JsonParse.Load("./json_parse.json", &v)
	fmt.Println(v.Port)
	fmt.Println(v.Mongo.MongoDb)
}
