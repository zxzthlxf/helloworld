package main

import (
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"helloworld/chapter4/model"
	"helloworld/chapter4/mongodb"
	"time"
)

func main() {
	var (
		client     = mongodb.MgoCli()
		err        error
		collection *mongo.Collection
		iResult    *mongo.InsertOneResult
		id         primitive.ObjectID
	)
	//选择数据库mydb中的某个集合
	collection = client.Database("mydb").Collection("my_collection")

	//插入某一条数据
	logRecord := model.LogRecord{
		JobName: "job1",
		Command: "echo 1",
		Err:     "",
		Content: "1",
		Tp: model.ExecTime{
			StartTime: time.Now().Unix(),
			EndTime:   time.Now().Unix() + 10,
		},
	}
	if iResult, err = collection.InsertOne(context.TODO(), logRecord); err != nil {
		fmt.Print(err)
		return
	}
	//_id:默认生成一个全局唯一ID
	id = iResult.InsertedID.(primitive.ObjectID)
	fmt.Println("自增ID", id.Hex())

}
