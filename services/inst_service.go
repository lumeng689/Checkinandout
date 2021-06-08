package services

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// TODO - throw error if input is outside const values
type InstType string
type WorkflowType string
type MemberType string

// InstType Enum Defs
const (
	InstTypeSchool    InstType = "school"
	InstTypeHospital  InstType = "hospital"
	InstTypeCorporate InstType = "corporate"
)

// RecordType Enum Defs
const (
	WorkflowTypeCC      WorkflowType = "cc"
	WorkflowTypeCheckIn WorkflowType = "checkin"
)

// MemberType Enum Defs
const (
	MemberTypeGuardian MemberType = "guardian"
	MemberTypeStandard MemberType = "standard"
	MemberTypeTag      MemberType = "tag"
)

// InstitutionForm - Input Form for Institution
type InstitutionForm struct {
	Type                 string `json:"type"`
	WorkflowType         string `json:"workflow_type"`
	MemberType           string `json:"member_type"`
	Identifier           string `json:"identifier"`
	CustomTagStringRegex string `json:"custom_tag_string_regex"`
	Name                 string `json:"name" validate:"required,min=2,max=100"`
	Address              string `json:"address" validate:"required,min=5,max=100"`
	State                string `json:"state" validate:"required,state"`
	ZipCode              string `json:"zip_code" validate:"required,zip_code"`
	RequireSurvey        bool   `json:"require_survey"`
	SurveyFile           string `json:"survey_file"`
}

// Institution - DB Model for Institution
type Institution struct {
	ID                   primitive.ObjectID `bson:"_id" json:"_id"`
	Type                 InstType           `json:"type"`
	WorkflowType         WorkflowType       `bson:"workflow_type" json:"workflow_type"`
	MemberType           MemberType         `bson:"member_type" json:"member_type"`
	Identifier           string             `bson:"identifier" json:"identifier"`
	CustomTagStringRegex string             `bson:"custom_tag_string_regex" json:"custom_tag_string_regex"`
	Name                 string             `json:"name"`
	Address              string             `json:"address"`
	State                string             `json:"state"`
	ZipCode              string             `bson:"zip_code" json:"zip_code"`
	RequireSurvey        bool               `bson:"require_survey" json:"require_survey"`
	SurveyFile           string             `bson:"survey_file" json:"survey_file"`
	CreatedAt            time.Time          `bson:"created_at" json:"created_at"`
	ModifiedAt           time.Time          `bson:"modified_at" json:"modified_at"`
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

// GetInstByIdentifier as name suggests
func GetInstByIdentifier(identifier string) *mongo.SingleResult {
	// TODO: err handling for ID Parsing
	return instCollection.FindOne(context.TODO(), bson.D{
		primitive.E{Key: "identifier", Value: identifier},
	})
}

// GetInstByName as name suggests
func GetInstByName(name string) *mongo.SingleResult {
	// TODO: err handling for ID Parsing
	return instCollection.FindOne(context.TODO(), bson.D{
		primitive.E{Key: "name", Value: name},
	})
}

// CountInstByName - as is
func CountInstByName(name string) (int64, error) {
	// TODO: err handling for ID Parsing
	return instCollection.CountDocuments(context.TODO(), bson.D{
		primitive.E{Key: "name", Value: name},
	})
}

// CreateInst as name suggests
func CreateInst(i InstitutionForm) (*mongo.InsertOneResult, error) {
	// requireSurvey, _ := strconv.ParseBool(i.RequireSurvey)
	newInst := Institution{
		ID:                   primitive.NewObjectID(),
		Type:                 InstType(i.Type),
		WorkflowType:         WorkflowType(i.WorkflowType),
		MemberType:           MemberType(i.MemberType),
		Identifier:           i.Identifier,
		CustomTagStringRegex: i.CustomTagStringRegex,
		Name:                 i.Name,
		Address:              i.Address,
		State:                i.State,
		ZipCode:              i.ZipCode,
		RequireSurvey:        i.RequireSurvey,
		SurveyFile:           i.SurveyFile,
		CreatedAt:            time.Now(),
		ModifiedAt:           time.Now(),
	}
	return instCollection.InsertOne(context.TODO(), newInst)
}

// UpdateInstByID as name suggests
func UpdateInstByID(i InstitutionForm, idToUpdate string) (*mongo.UpdateResult, error) {
	oid, _ := primitive.ObjectIDFromHex(idToUpdate)
	return instCollection.UpdateOne(context.TODO(), bson.M{"_id": oid}, bson.D{
		primitive.E{Key: "$set", Value: bson.D{
			primitive.E{Key: "type", Value: InstType(i.Type)},
			primitive.E{Key: "workflow_type", Value: WorkflowType(i.WorkflowType)},
			primitive.E{Key: "member_type", Value: MemberType(i.MemberType)},
			primitive.E{Key: "identifier", Value: i.Identifier},
			primitive.E{Key: "custom_tag_string_regex", Value: i.CustomTagStringRegex},
			primitive.E{Key: "name", Value: i.Name},
			primitive.E{Key: "address", Value: i.Address},
			primitive.E{Key: "state", Value: i.State},
			primitive.E{Key: "zip_code", Value: i.ZipCode},
			primitive.E{Key: "require_survey", Value: i.RequireSurvey},
			primitive.E{Key: "survey_file", Value: i.SurveyFile},
		}},
		primitive.E{Key: "$currentDate", Value: bson.D{
			primitive.E{Key: "modified_at", Value: true},
		}},
	})
}

// DeleteInstByID as name suggests
func DeleteInstByID(idToDelete string) (*mongo.DeleteResult, error) {
	// TODO: err handling for ID Parsing
	oid, _ := primitive.ObjectIDFromHex(idToDelete)
	return instCollection.DeleteOne(context.TODO(), bson.M{"_id": oid})
}
