package main

import (
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"helloworld/chapter4/model"
	"helloworld/chapter4/mongodb"
	"log"
)

func main() {
	var (
		client     = mongodb.MgoCli()
		err        error
		collection *mongo.Collection
		cursor     *mongo.Cursor
	)
	//选择数据库mydb中的某个集合
	/*
	db.cols1.insert({jobName: 'Java 教程',
	description: 'Java 是由Sun Microsystems公司于1995年5月推出的高级程序设计语言。',
	by: '菜鸟教程',
	url: 'http://www.runoob.com',
	tags: ['java'],
	likes: 150
	})
	db.cols1.insert({jobName: 'job multil',
	description: 'Go 是由Sun Microsystems公司于1995年5月推出的高级程序设计语言。',
	by: '菜鸟教程',
	url: 'http://www.runoob.com',
	tags: ['go'],
	likes: 150
	})
	*/
	collection = client.Database("mydb").Collection("cols1")
	cond := model.FindByJobName{JobName: "job multil"}
	if cursor, err = collection.Find(
		context.TODO(),
		cond,
		options.Find().SetSkip(0),
		options.Find().SetLimit(2)); err != nil {
		fmt.Println(err)
		return
	}
	defer func() {
		if err = cursor.Close(context.TODO()); err != nil {
			log.Fatal(err)
		}
	}()
	//遍历游标获取结果数据
	for cursor.Next(context.TODO()) {
		var lr model.LogRecord
		if cursor.Decode(&lr) != nil {
			fmt.Print(err)
			return
		}
		fmt.Println(lr)
	}

	var results []model.LogRecord
	if err = cursor.All(context.TODO(), &results); err != nil {
		log.Fatal(err)
	}
	for _, result := range results {
		fmt.Println(result)
	}
}
