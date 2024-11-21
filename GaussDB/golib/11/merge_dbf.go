package main

import (
	"fmt"

	"github.com/gogap/godbf"
)

func mergeDBFFiles(filenames []string, outputFilename string) error {
	// 创建输出文件
	outputTable, err := godbf.SaveFile(outputFilename)
	if err != nil {
		return err
	}

	for _, filename := range filenames {
		table, err := godbf.NewFromFile(filename)
		if err != nil {
			return err
		}

		numRecords := table.NumberOfRecords()
		fields := table.Fields()

		// 将输入文件的字段添加到输出文件
		for _, field := range fields {
			err = outputTable.AddField(field.Name, field.Type, field.Length, field.Decimals)
			if err != nil {
				return err
			}
		}

		// 将输入文件的记录添加到输出文件
		for i := 0; i < numRecords; i++ {
			record, err := table.GetRowAsSlice(i)
			if err != nil {
				return err
			}
			err = outputTable.Append(record)
			if err != nil {
				return err
			}
		}
	}

	return outputTable.Save()
}

func main() {
	inputFilenames := []string{"sjsqs00819_500w_1.dbf", "sjsqs00819_500w_2.dbf"}
	outputFilename := "sjsqs00819_500w.dbf"
	err := mergeDBFFiles(inputFilenames, outputFilename)
	if err != nil {
		fmt.Println("Error:", err)
	} else {
		fmt.Println("Files merged successfully.")
	}
}
