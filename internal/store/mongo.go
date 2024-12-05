package store

import (
	"context"
	"mc/internal/models"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type Mongo struct {
	client *mongo.Client
}

func NewMongo(connectionString string) (DB, error) {
	client, err := mongo.Connect(options.Client().ApplyURI(connectionString))
	if err != nil {
		return nil, err
	}

	return &Mongo{
		client: client,
	}, nil
}

func (db *Mongo) GetWorkersByStatus(ctx context.Context, status string) ([]models.Worker, error) {
	collection := db.client.Database("mc").Collection("workers")
	var workers []models.Worker
	filter := bson.D{{"status", status}}
	cursor, err := collection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)
	for cursor.Next(ctx) {
		var worker models.Worker
		err := cursor.Decode(&worker)
		if err != nil {
			return nil, err
		}
		workers = append(workers, worker)
	}
	return workers, nil
}

func (db *Mongo) CreateWorker(ctx context.Context, id string) error {
	collection := db.client.Database("mc").Collection("workers")

	worker := models.Worker{
		ID:     id,
		Status: "Waiting",
	}
	_, err := collection.InsertOne(ctx, worker)
	return err
}

func (db *Mongo) UpdateWorkerStatus(ctx context.Context, id, status string) error {
	collection := db.client.Database("mc").Collection("workers")
	filter := bson.D{{Key: "_id", Value: id}}
	update := bson.D{{Key: "$set", Value: bson.D{{Key: "status", Value: status}}}}
	_, err := collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}

	return nil
}

func (db *Mongo) DeleteWorker(ctx context.Context, podName string) error {
	collection := db.client.Database("mc").Collection("workers")
	_, err := collection.DeleteMany(ctx, bson.D{{Key: "_id", Value: podName}})
	return err
}

func (db *Mongo) SetCallback(callback func()) {
}
