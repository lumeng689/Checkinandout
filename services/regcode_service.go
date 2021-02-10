package services

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// RegCode - DB Model for Guardian Account Activation Code
type RegCode struct {
	ID         primitive.ObjectID `bson:"_id" json:"_id"`
	GuardianID string             `bson:"guardian_id" json:"guardian_id"`
	RegCode    string             `bson:"reg_code" json:"reg_code"`
}

// GetRegCodeParams - QueryString params for RegCode
type GetRegCodeParams struct {
	GuardianID string `json:"guardian_id"`
}

var regCodeCollection *mongo.Collection

// RegCodeCollection returns reference to DB collection
func RegCodeCollection(c *mongo.Database) {
	regCodeCollection = c.Collection("regCodes")
}

// GetManyRegCodes as name suggests
func GetManyRegCodes() (*mongo.Cursor, error) {
	// TODO: not sending status: "2 - deleted"
	return regCodeCollection.Find(context.TODO(), bson.M{})
}

// GetRegCodeByID
func GetRegCodeByID(id string) *mongo.SingleResult {
	oid, _ := primitive.ObjectIDFromHex(id)
	return regCodeCollection.FindOne(context.TODO(), bson.M{
		"_id": oid},
	)
}

// GetRegCodeByGuardianID as name suggests
func GetRegCodeByGuardianID(guardianID string) *mongo.SingleResult {
	// TODO: err handling for ID Parsing
	return regCodeCollection.FindOne(context.TODO(), bson.D{
		primitive.E{Key: "guardian_id", Value: guardianID},
	})
}

// CreateRegCodeByGuardianID as name suggests
func CreateRegCodeByGuardianID(guardianID string) (*mongo.InsertOneResult, error) {

	newRegCode := RegCode{
		ID:         primitive.NewObjectID(),
		GuardianID: guardianID,
		RegCode:    getNewRegCode(),
	}

	return regCodeCollection.InsertOne(context.TODO(), newRegCode)
}

// DeleteRegCodeByGuardianID as name suggests
func DeleteRegCodeByGuardianID(guardianID string) (*mongo.DeleteResult, error) {
	return regCodeCollection.DeleteOne(context.TODO(), bson.D{
		primitive.E{Key: "guardian_id", Value: guardianID},
	})
}

// generate a regCode of length 6, all digits
func getNewRegCode() string {
	rand.Seed(time.Now().UnixNano())
	s := fmt.Sprintf("%06d", rand.Intn(1000000))
	return s
}
