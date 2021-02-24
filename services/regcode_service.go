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
	ID       primitive.ObjectID `bson:"_id" json:"_id"`
	MemberID string             `bson:"member_id" json:"member_id"`
	RegCode  string             `bson:"reg_code" json:"reg_code"`
}

// GetRegCodeParams - QueryString params for RegCode
type GetRegCodeParams struct {
	MemberID string `json:"member_id"`
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

// GetRegCodeByID - as is
func GetRegCodeByID(id string) *mongo.SingleResult {
	oid, _ := primitive.ObjectIDFromHex(id)
	return regCodeCollection.FindOne(context.TODO(), bson.M{
		"_id": oid},
	)
}

// GetRegCodeByMemberID as name suggests
func GetRegCodeByMemberID(memberID string) *mongo.SingleResult {
	// TODO: err handling for ID Parsing
	return regCodeCollection.FindOne(context.TODO(), bson.D{
		primitive.E{Key: "member_id", Value: memberID},
	})
}

// CreateRegCodeByMemberID as name suggests
func CreateRegCodeByMemberID(memberID string) (*mongo.InsertOneResult, error) {

	newRegCode := RegCode{
		ID:       primitive.NewObjectID(),
		MemberID: memberID,
		RegCode:  getNewRegCode(),
	}

	return regCodeCollection.InsertOne(context.TODO(), newRegCode)
}

// DeleteRegCodeByMemberID as name suggests
func DeleteRegCodeByMemberID(memberID string) (*mongo.DeleteResult, error) {
	return regCodeCollection.DeleteOne(context.TODO(), bson.D{
		primitive.E{Key: "member_id", Value: memberID},
	})
}

// generate a regCode of length 6, all digits
func getNewRegCode() string {
	rand.Seed(time.Now().UnixNano())
	s := fmt.Sprintf("%06d", rand.Intn(1000000))
	return s
}
