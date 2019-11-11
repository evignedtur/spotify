package main

import (
	"context"
	"fmt"
	"log"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var client *mongo.Client
var collection mongo.Collection

func connectToDb() {
	clientOptions := options.Client().ApplyURI(config.Databaseurl)
	clientOptions.SetRetryWrites(false)
	client, err := mongo.Connect(context.TODO(), clientOptions)

	if err != nil {
		log.Fatal(err)
	}

	err = client.Ping(context.TODO(), nil)

	if err != nil {
		log.Fatal(err)
	}

	collection = *client.Database("evignedtur").Collection("tokens")

	fmt.Println("Connected to MongoDB!")

}

func getTokensFromDb() {

	cur, err := collection.Find(context.TODO(), bson.D{})
	if err != nil {
		log.Println(err)
	}

	for cur.Next(context.TODO()) {
		var elem Token
		err := cur.Decode(&elem)
		Tokens = append(Tokens, &elem)
		if err != nil {
			log.Println(err)
		}
	}
}

func insertTokenToDb(token Token) {

	insertResult, err := collection.InsertOne(context.TODO(), token)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Inserted a Single Document: ", insertResult.InsertedID)

}

func updateTokenFromDb(token *Token) {

	filter := bson.D{{"uuid", token.UUID}}

	update := bson.D{
		{"$set", bson.D{
			{"token", token.Token},
			{"expiry", token.Expiry},
		}},
	}

	_, err := collection.UpdateOne(context.TODO(), filter, update)
	if err != nil {
		log.Fatal(err)
	}
}
