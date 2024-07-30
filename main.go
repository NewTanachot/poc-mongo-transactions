package main

import (
	"context"
	"fmt"
	"strconv"

	"go.mongodb.org/mongo-driver/bson"
	"poc-mongo-transactions/mongodb"

	"go.mongodb.org/mongo-driver/mongo"
)

func main() {
	mongoDb := mongodb.NewMongoDbClient()
	var transactions []func(mongo.SessionContext) (interface{}, error)

	count := 5
	for i := 1; i <= count; i++ {
		transactions = append(transactions, getSuccessMongoTransaction(mongoDb.DbClient, "jai_"+strconv.Itoa(i)))
	}

	_, err := mongoDb.Execute(context.Background(), transactions...)

	if err != nil {
		fmt.Println(err.Error())
	} else {
		fmt.Println("Success")
	}
}

func getSuccessMongoTransaction(db *mongo.Database, name string) func(mongo.SessionContext) (interface{}, error) {
	return func(sessCtx mongo.SessionContext) (interface{}, error) {
		data := bson.M{
			"name": name,
		}

		_, err := db.Collection("dummy").InsertOne(sessCtx, data)
		return nil, err
	}
}

// func getFailMongoTransaction() func(mongo.SessionContext) (interface{}, error) {
// 	return func(sessCtx mongo.SessionContext) (interface{}, error) {
// 		return nil, errors.New("some error")
// 	}
// }
