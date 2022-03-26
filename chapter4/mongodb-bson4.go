package main

import (
	"context"
	"go.mongodb.org/mongo-driver/mongo"
	"helloworld/chapter4/model"
	"helloworld/chapter4/mongodb"
	"log"
	"time"
)

func main() {
	var (
		client     = mongodb.MgoCli()
		collection *mongo.Collection
		err        error
		uResult    *mongo.DeleteResult
		delCond    *model.DeleteCond
	)
	collection = client.Database("mydb").Collection("cols1")

	delCond = &model.DeleteCond{
		BeforeCond: model.TimeBeforeCond{
			BeforeTime: time.Now().Unix(),
		},
	}
	if uResult, err = collection.DeleteMany(context.TODO(),
		delCond); err != nil {
		log.Fatal(err)
	}
	log.Println(uResult.DeletedCount)
}
