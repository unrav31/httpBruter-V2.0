package database

import (
	"context"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	options2 "httpBruter/pkg/options"
	"httpBruter/pkg/structs"
	"log"
	"time"
)

//const mongoURI = "mongodb://meetsec-root:ab5036c95988c6bf179ffff62adf3a82@175.24.16.184:27018"

const mongoURI = "mongodb://localhost:27017"
const databaseName = "testRefactor"
const collectName = "testRefactor"

func Collect() (data *mongo.Database, client *mongo.Client) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(mongoURI))

	if err != nil {
		log.Fatal(err)
	}
	data = client.Database(databaseName)

	return data, client
}

// CloseConnection s
func CloseConnection(client *mongo.Client) {
	err := client.Disconnect(context.TODO())
	if err != nil {
		log.Fatal(err)
	}
}

func InsertOne(data *structs.Database, arg *options2.Args) {

	collection := arg.MongoDB.Collection(collectName)

	_, err := collection.InsertOne(context.TODO(), *data)
	if err != nil {
		log.Fatal(err)
	}

}
