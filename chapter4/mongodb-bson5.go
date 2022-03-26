package main

import (
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"helloworld/chapter4/model"
	"helloworld/chapter4/mongodb"
	"log"
)

func main() {
	var (
		client     = mongodb.MgoCli()
		collection *mongo.Collection
		err        error
		cursor     *mongo.Cursor
	)
	collection = client.Database("mydb").Collection("cols1")

	groupStage := []model.Group{}
	groupStage = append(groupStage, model.Group{
		Group: bson.D{
			{"_id", "$jobName"},
			{"countJob", model.Sum{Sum: 1}},
		},
	})

	if cursor, err = collection.Aggregate(context.TODO(), groupStage); err != nil {
		log.Fatal(err)
	}
	defer func() {
		if err = cursor.Close(context.TODO()); err != nil {
			log.Fatal(err)
		}
	}()
	var results []bson.M
	if err = cursor.All(context.TODO(), &results); err != nil {
		log.Fatal(err)
	}
	for _, result := range results {
		fmt.Println(result)
	}
}
