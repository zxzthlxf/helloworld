package main

import (
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
)

func main() {
	testM := bson.M{
		"jobName": "job multil",
	}
	var raw bson.Raw
	tmp, _ := bson.Marshal(testM)
	bson.Unmarshal(tmp, &raw)

	fmt.Println(testM)
	fmt.Println(raw)
}
