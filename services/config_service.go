package services

import (
	"context"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// ConfigForm - for adding config to DB
type ConfigForm struct {
	Name         string `json:"name"`
	SMSAuthToken string `bson:"sms_auth_token" json:"sms_auth_token"`
	ServerAddr   string `bson:"server_address" json:"server_address"`
}

// ConfigEditForm - for adding config to DB
type ConfigEditForm struct {
	SMSAuthToken string `bson:"sms_auth_token" json:"sms_auth_token"`
	ServerAddr   string `bson:"server_address" json:"server_address"`
}

// Config - for hot-reloading config from DB
type Config struct {
	ID           primitive.ObjectID `bson:"_id" json:"_id"`
	Name         string             `json:"name"`
	SMSAuthToken string             `bson:"sms_auth_token" json:"sms_auth_token"`
	ServerAddr   string             `bson:"server_address" json:"server_address"`
}

var configCollection *mongo.Collection

// ConfigCollection returns reference to DB collection
func ConfigCollection(c *mongo.Database) {
	configCollection = c.Collection("configs")
}

// GetManyConfigs - as is
func GetManyConfigs() (*mongo.Cursor, error) {
	return configCollection.Find(context.TODO(), bson.M{})
}

// GetConfigByName - as is
func GetConfigByName(configName string) *mongo.SingleResult {
	return configCollection.FindOne(context.TODO(), bson.D{
		primitive.E{Key: "name", Value: configName},
	})
}

// CountConfigByName - as is
func CountConfigByName(configName string) (int64, error) {
	return configCollection.CountDocuments(context.TODO(), bson.D{
		primitive.E{Key: "name", Value: configName},
	})
}

// CreateConfig - as is
func CreateConfig(configForm ConfigForm) (*mongo.InsertOneResult, error) {
	newConfig := Config{
		ID:           primitive.NewObjectID(),
		Name:         configForm.Name,
		SMSAuthToken: configForm.SMSAuthToken,
		ServerAddr:   configForm.ServerAddr,
	}
	return configCollection.InsertOne(context.TODO(), newConfig)
}

// UpdateConfigByID - as is
func UpdateConfigByID(c ConfigEditForm, idToUpdate string) (*mongo.UpdateResult, error) {
	oid, _ := primitive.ObjectIDFromHex(idToUpdate)
	return configCollection.UpdateOne(context.TODO(), bson.M{"_id": oid}, bson.D{
		primitive.E{Key: "$set", Value: bson.D{
			primitive.E{Key: "sms_auth_token", Value: c.SMSAuthToken},
			primitive.E{Key: "server_address", Value: c.ServerAddr},
		}},
	})
}
