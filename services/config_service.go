package services

import (
	"context"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// Config - for hot-reloading config from DB
type Config struct {
	Name         string `json:"name"`
	SMSAuthToken string `bson:"sms_auth_token" json:"sms_auth_token"`
}

var configCollection *mongo.Collection

// ConfigCollection returns reference to DB collection
func ConfigCollection(c *mongo.Database) {
	configCollection = c.Collection("configs")
}

// GetConfigByName - as is
func GetConfigByName(configName string) *mongo.SingleResult {
	return configCollection.FindOne(context.TODO(), bson.D{
		primitive.E{Key: "name", Value: configName},
	})
}
