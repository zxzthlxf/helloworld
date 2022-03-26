package main

import (
	"go.mongodb.org/mongo-driver/mongo"
	"helloworld/chapter4/mongodb"
)

func main() {
	var (
		client     = mongodb.MgoCli()
		db         *mongo.Database
		collection *mongo.Collection
	)

	//选择数据库mydb
	db = client.Database("mydb")

	//选择集合my_collection
	collection = db.Collection("my_collection")
	collection = collection
}
