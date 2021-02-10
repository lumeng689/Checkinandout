package services

import (
	"context"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// AdminRegForm - Input Form for Admin
type AdminRegForm struct {
	FrasUsername string `json:"fras_username"`
	InstID       string `json:"institution_id"`
}

// AdminLoginForm - Incoming Request for Login
type AdminLoginForm struct {
	FrasUsername string `json:"fras_username"`
}

// Admin - DB Model for Admin
type Admin struct {
	ID           primitive.ObjectID `bson:"_id" json:"_id"`
	FrasUsername string             `bson:"fras_username" json:"fras_username"`
	InstID       string             `bson:"institution_id" json:"institution_id"`
	LastLoginAt  time.Time          `bson:"last_login_at" json:"last_login_at"`
	CreatedAt    time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt    time.Time          `bson:"updated_at" json:"updated_at"`
}

// GetAdminParams - QueryString Params for GetAdmin
type GetAdminParams struct {
	FrasUsername string `json:"fras_username"`
}

var adminCollection *mongo.Collection

// AdminCollection returns reference to DB collection
func AdminCollection(c *mongo.Database) {
	adminCollection = c.Collection("admins")
}

// GetManyAdmins - as name suggests;
func GetManyAdmins() (*mongo.Cursor, error) {
	return adminCollection.Find(context.TODO(), bson.M{})
}

// CountAdminsByInstID - as name suggests; Invoked when deleting institution
func CountAdminsByInstID(instID string) (int64, error) {
	// oid, _ := primitive.ObjectIDFromHex(instID)
	// log.Printf("GetManyAdminsByInstID: Decoded InstID - %v\n", oid.String())
	return adminCollection.CountDocuments(context.TODO(), bson.D{primitive.E{
		Key: "institution_id", Value: instID,
	}})
}

// GetManyAdminsByInstID - as name suggests
func GetManyAdminsByInstID(instID string) (*mongo.Cursor, error) {
	// oid, _ := primitive.ObjectIDFromHex(instID)
	// log.Printf("GetManyAdminsByInstID: Decoded InstID - %v\n", oid.String())
	return adminCollection.Find(context.TODO(), bson.D{primitive.E{
		Key: "institution_id", Value: instID,
	}})
}

// GetAdminByFrasUsername - as name suggests; Invoked when logging in from FRAS
func GetAdminByFrasUsername(frasUsername string) *mongo.SingleResult {
	log.Printf("Getting Admin with username - %v\n", frasUsername)
	return adminCollection.FindOne(context.TODO(), bson.M{
		"fras_username": frasUsername,
	})
}

// CreateAdmin - as name suggests;
func CreateAdmin(a AdminRegForm) (*mongo.InsertOneResult, error) {
	newAdmin := Admin{
		ID:           primitive.NewObjectID(),
		FrasUsername: a.FrasUsername,
		InstID:       a.InstID,
		CreatedAt:    time.Now(),
	}
	return adminCollection.InsertOne(context.TODO(), newAdmin)
}

// UpdateAdminByID as name suggests
func UpdateAdminByID(i AdminRegForm, idToUpdate string) (*mongo.UpdateResult, error) {
	oid, _ := primitive.ObjectIDFromHex(idToUpdate)
	return adminCollection.UpdateOne(context.TODO(), bson.M{"_id": oid}, bson.D{
		primitive.E{Key: "$set", Value: bson.D{
			primitive.E{Key: "fras_username", Value: i.FrasUsername},
			primitive.E{Key: "institution_id", Value: i.InstID},
		}},
		primitive.E{Key: "$currentDate", Value: bson.D{
			primitive.E{Key: "updated_at", Value: true},
		}},
	})
}

// UpdateAdminLoginTime gets the Admin to be logged in, update the login time, and return Admin
func UpdateAdminLoginTime(adminID string, admin *Admin) error {
	oid, _ := primitive.ObjectIDFromHex(adminID)
	_, err := adminCollection.UpdateOne(context.TODO(), bson.M{"_id": oid}, bson.D{
		primitive.E{Key: "$currentDate", Value: bson.D{
			primitive.E{Key: "last_login_at", Value: true},
		}},
	})
	getAdminByID(oid).Decode(&admin)
	return err
}

// DeleteAdminByID as name suggests
func DeleteAdminByID(idToDelete string) (*mongo.DeleteResult, error) {
	// TODO: err handling for ID Parsing
	oid, _ := primitive.ObjectIDFromHex(idToDelete)
	return adminCollection.DeleteOne(context.TODO(), bson.M{"_id": oid})
}

// GetAdminByFrasUsername - as name suggests; Invoked when logging in from FRAS
func getAdminByID(oid primitive.ObjectID) *mongo.SingleResult {
	return adminCollection.FindOne(context.TODO(), bson.M{
		"_id": oid,
	})
}
