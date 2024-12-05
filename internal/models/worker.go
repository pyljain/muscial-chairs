package models

type Worker struct {
	ID     string `json:"id" bson:"_id"`
	Status string `json:"status" bson:"status"`
}
