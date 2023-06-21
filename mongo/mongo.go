package mongodb

import (
	"context"
	"encoding/json"
	"fmt"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func CreateMongoSession(uri string) (*mongo.Client, error) {
	clientOptions := options.Client().ApplyURI(uri)
	client, err := mongo.Connect(context.Background(), clientOptions)
	if err != nil {
		return nil, err
	}

	return client, nil
}

func CheckMongoConnection(client *mongo.Client) error {
	var result bson.M
	fmt.Println("Checking mongo connection...")
	if err := client.Database("admin").RunCommand(context.Background(), bson.D{{"ping", 1}}).Decode(&result); err != nil {
		return err
	}
	return nil
}

func InsertLogLine(doc []byte, collection *mongo.Collection) error {
	var jsonDoc interface{}
	err := json.Unmarshal(doc, &jsonDoc)
	if err != nil {
		return err
	}
	println("jsonDoc ", jsonDoc)
	_, insertErr := collection.InsertOne(context.Background(), jsonDoc)
	if insertErr != nil {
		return insertErr
	}

	return nil
}

// func (d *Driver) Log(msg *logger.Message) error {
// 	doc := bson.M{
// 		"timestamp": msg.Timestamp,
// 		"message":   string(msg.Line),
// 	}
// 	_, err := d.collection.InsertOne(context.Background(), doc)
// 	return err
// }
