package services

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// TagRegForm - Input Form for Admin
type TagRegForm struct {
	InstID    string `bson:"institution_id" json:"institution_id"`
	TagString string `bson:"tag_string" json:"tag_string" validate:"required,tag_string"`
	PhoneNum  string `bson:"phone_num" json:"phone_num"`
	Email     string `json:"email"`
	FirstName string `bson:"first_name" json:"first_name"`
	LastName  string `bson:"last_name" json:"last_name"`
	Group     string `json:"group"`
}

// TagEditForm - Input Form for Admin
type TagEditForm struct {
	PhoneNum  string `bson:"phone_num" json:"phone_num"`
	Email     string `json:"email"`
	FirstName string `bson:"first_name" json:"first_name"`
	LastName  string `bson:"last_name" json:"last_name"`
	Group     string `json:"group"`
}

// Tag - DB Model for Tag
type Tag struct {
	ID         primitive.ObjectID `bson:"_id" json:"_id"`
	InstID     string             `bson:"institution_id" json:"institution_id"`
	TagString  string             `bson:"tag_string" json:"tag_string"`
	PhoneNum   string             `bson:"phone_num" json:"phone_num"`
	Email      string             `json:"email"`
	FirstName  string             `bson:"first_name" json:"first_name"`
	LastName   string             `bson:"last_name" json:"last_name"`
	Group      string             `json:"group"`
	ModifiedAt time.Time          `bson:"modified_at" json:"modified_at"`
}

// GetTagParams - QueryString Params for GetTag
type GetTagParams struct {
	InstID    string `json:"inst_id"`
	TagString string `json:"tag_string"`
}

// CountTagParams - Searching Params for CountTag
type CountTagParams struct {
	InstID    string `json:"inst_id"`
	TagString string `json:"tag_string"`
}

var tagCollection *mongo.Collection

// TagCollection returns reference to DB collection
func TagCollection(c *mongo.Database) {
	tagCollection = c.Collection("tags")
}

// GetManyTags as is
func GetManyTags(params *GetTagParams) (*mongo.Cursor, error) {
	var filters bson.D
	filters = append(filters, primitive.E{Key: "institution_id", Value: params.InstID})

	return tagCollection.Find(context.TODO(), filters)
}

// GetTagByID as is
func GetTagByID(id string) *mongo.SingleResult {
	oid, _ := primitive.ObjectIDFromHex(id)
	return tagCollection.FindOne(context.TODO(), bson.M{
		"_id": oid,
	})
}

// GetTagByTagString as is
func GetTag(params *GetTagParams) *mongo.SingleResult {

	var filters bson.D
	if len(params.InstID) > 0 {
		filters = append(filters, primitive.E{Key: "institution_id", Value: params.InstID})
	}
	if len(params.TagString) > 0 {
		filters = append(filters, primitive.E{Key: "tag_string", Value: params.TagString})
	}

	return tagCollection.FindOne(context.TODO(), filters)
}

// CountTag as is
func CountTag(countTagParams CountTagParams) (int64, error) {
	return tagCollection.CountDocuments(context.TODO(), bson.D{
		primitive.E{Key: "tag_string", Value: countTagParams.TagString},
		primitive.E{Key: "institution_id", Value: countTagParams.InstID},
	})
}

// CreateTag as is
func CreateTag(t TagRegForm) (*mongo.InsertOneResult, error) {
	newTag := Tag{
		ID:         primitive.NewObjectID(),
		InstID:     t.InstID,
		TagString:  t.TagString,
		PhoneNum:   t.PhoneNum,
		Email:      t.Email,
		FirstName:  t.FirstName,
		LastName:   t.LastName,
		Group:      t.Group,
		ModifiedAt: time.Now(),
	}

	return tagCollection.InsertOne(context.TODO(), newTag)
}

// UpdateTagByID as is
func UpdateTagByID(t TagEditForm, idToUpdate string) (*mongo.UpdateResult, error) {
	oid, _ := primitive.ObjectIDFromHex(idToUpdate)
	return tagCollection.UpdateOne(context.TODO(), bson.M{"_id": oid}, bson.D{
		primitive.E{Key: "$set", Value: bson.D{
			primitive.E{Key: "phone_num", Value: t.PhoneNum},
			primitive.E{Key: "email", Value: t.Email},
			primitive.E{Key: "first_name", Value: t.FirstName},
			primitive.E{Key: "last_name", Value: t.LastName},
			primitive.E{Key: "group", Value: t.Group},
		}},
		primitive.E{Key: "$currentDate", Value: bson.D{
			primitive.E{Key: "modified_at", Value: true},
		}},
	})
}

// DeleteTagByID as is
func DeleteTagByID(idToDelete string) (*mongo.DeleteResult, error) {
	oid, _ := primitive.ObjectIDFromHex(idToDelete)
	return tagCollection.DeleteOne(context.TODO(), bson.M{"_id": oid})
}
