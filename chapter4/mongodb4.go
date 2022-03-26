package main

import (
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"helloworld/chapter4/model"
	"helloworld/chapter4/mongodb"
	"log"
	"time"
)

func main() {
	var (
		client     = mongodb.MgoCli()
		err        error
		collection *mongo.Collection
		result     *mongo.InsertManyResult
		id         primitive.ObjectID
	)
	collection = client.Database("mydb").Collection("test")

	//批量插入
	result, err = collection.InsertMany(context.TODO(), []interface{}{
		model.LogRecord{
			JobName: "job multi",
			Command: "echo multi",
			Err:     "",
			Content: "1",
			Tp: model.ExecTime{
				StartTime: time.Now().Unix(),
				EndTime:   time.Now().Unix() + 10,
			},
		},
	})
	if err != nil {
		log.Fatal(err)
	}
	if result == nil {
		log.Fatal("result nil")
	}
	for _, v := range result.InsertedIDs {
		id = v.(primitive.ObjectID)
		fmt.Println("自增ID", id.Hex())
	}
}
