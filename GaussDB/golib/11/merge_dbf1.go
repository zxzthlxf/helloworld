package main

import (
	"bytes"
	"io/ioutil"
)

func main() {
	files := []string{"sjsqs00819_500w_1.dbf", "sjsqs00819_500w_2.dbf", "sjsqs00819_500w_3.dbf"} // 这里填入你的dbf文件列表
	var mergedContent bytes.Buffer

	for _, file := range files {
		data, err := ioutil.ReadFile(file)
		if err != nil {
			panic(err)
		}
		mergedContent.Write(data)
	}

	if err := ioutil.WriteFile("sjsqs00819_500w.dbf", mergedContent.Bytes(), 0644); err != nil {
		panic(err)
	}
}
