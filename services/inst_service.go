package services

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// InstitutionForm - Input Form for Institution
type InstitutionForm struct {
	Type    int    `json:"type"`
	Name    string `json:"name" validate:"required,min=2,max=100"`
	Address string `json:"address" validate:"required,min=5,max=100"`
	State   string `json:"state" validate:"required,state"`
	ZipCode string `bson:"zip_code" json:"zip_code" validate:"required,zip_code"`
}

// Institution - DB Model for Institution
type Institution struct {
	ID        primitive.ObjectID `bson:"_id" json:"_id"`
	Type      int                `json:"type"`
	Name      string             `json:"name"`
	Address   string             `json:"address"`
	State     string             `json:"state"`
	ZipCode   string             `bson:"zip_code" json:"zip_code"`
	CreatedAt time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt time.Time          `bson:"updated_at" json:"updated_at"`
}

var instCollection *mongo.Collection

// InstCollection returns reference to DB collection
func InstCollection(c *mongo.Database) {
	instCollection = c.Collection("institutions")
}

// GetManyInsts as name suggests
func GetManyInsts() (*mongo.Cursor, error) {
	// TODO: not sending status: "2 - deleted"
	return instCollection.Find(context.TODO(), bson.M{})
}

// GetInstByID as name suggests
func GetInstByID(id string) *mongo.SingleResult {
	// TODO: err handling for ID Parsing
	oid, _ := primitive.ObjectIDFromHex(id)
	return instCollection.FindOne(context.TODO(), bson.M{"_id": oid})
}

// CreateInst as name suggests
func CreateInst(i InstitutionForm) (*mongo.InsertOneResult, error) {

	newInst := Institution{
		ID:        primitive.NewObjectID(),
		Type:      i.Type,
		Name:      i.Name,
		Address:   i.Address,
		State:     i.State,
		ZipCode:   i.ZipCode,
		CreatedAt: time.Now(),
	}

	return instCollection.InsertOne(context.TODO(), newInst)
}

// UpdateInstByID as name suggests
func UpdateInstByID(i InstitutionForm, idToUpdate string) (*mongo.UpdateResult, error) {

	oid, _ := primitive.ObjectIDFromHex(idToUpdate)

	return instCollection.UpdateOne(context.TODO(), bson.M{"_id": oid}, bson.D{
		primitive.E{Key: "$set", Value: bson.D{
			primitive.E{Key: "type", Value: i.Type},
			primitive.E{Key: "name", Value: i.Name},
			primitive.E{Key: "address", Value: i.Address},
			primitive.E{Key: "state", Value: i.State},
			primitive.E{Key: "zip_code", Value: i.ZipCode},
		}},
		primitive.E{Key: "$currentDate", Value: bson.D{
			primitive.E{Key: "updated_at", Value: true},
		}},
	})
}

// DeleteInstByID as name suggests
func DeleteInstByID(idToDelete string) (*mongo.DeleteResult, error) {
	// TODO: err handling for ID Parsing
	oid, _ := primitive.ObjectIDFromHex(idToDelete)
	return instCollection.DeleteOne(context.TODO(), bson.M{"_id": oid})
}
