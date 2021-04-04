package services

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// WardForm - Input Form for Ward
type WardForm struct {
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Group     string `json:"group"`
}

// VehicleForm - Input Form for Vehicle
type VehicleForm struct {
	Make     string `json:"make"`
	Model    string `json:"model"`
	Color    string `json:"color"`
	PlateNum string `json:"plate_num"`
}

// FamilyRegForm - Input(Registration) Form for Family
type FamilyRegForm struct {
	InstID   string                  `json:"institution_id"`
	Members  []MemberInFamilyRegForm `json:"members"`
	Wards    []WardForm              `json:"wards"`
	Vehicles []VehicleForm           `json:"vehicles"`
}

// Ward - DB Model for Ward
type Ward struct {
	ID        primitive.ObjectID `bson:"_id" json:"_id"`
	FirstName string             `bson:"first_name" json:"first_name"`
	LastName  string             `bson:"last_name" json:"last_name"`
	Group     string             `json:"group"`
}

// Vehicle - DB Model for Vehicle
type Vehicle struct {
	ID       primitive.ObjectID `bson:"_id" json:"_id"`
	Make     string             `json:"make"`
	Model    string             `json:"model"`
	Color    string             `json:"color"`
	PlateNum string             `bson:"plate_num" json:"plate_num"`
}

// Family is the top-level struct, consisting of Guardians and Wards
type Family struct {
	ID                  primitive.ObjectID `bson:"_id" json:"_id"`
	InstID              string             `bson:"institution_id" json:"institution_id"`
	AllRegCodeSent      bool               `json:"all_reg_code_sent"`
	ContactGuardianInfo MemberTagInfo      `bson:"contact_member_info" json:"contact_member_info"`
	Wards               []Ward             `json:"wards"`
	Vehicles            []Vehicle          `json:"vehicles"`
	ModifiedAt          time.Time          `bson:"modified_at" json:"modified_at"`
}

// GetFamilyParams - QueryString Params for GetFamily
type GetFamilyParams struct {
	InstID string `json:"inst_id"`
}

// AddGuardianParams - QueryString Params for AddGuardian
type AddMemberInFamilyParams struct {
	FamilyID string `json:"family_id"`
}

// AddWardParams - QueryString Params for AddWard
type AddWardParams struct {
	FamilyID string `json:"family_id"`
}

// AddVehicleParams - QueryString Params for AddWard
type AddVehicleParams struct {
	FamilyID string `json:"family_id"`
}

var familyCollection *mongo.Collection

// FamilyCollection returns reference to DB collection
func FamilyCollection(c *mongo.Database) {
	familyCollection = c.Collection("families")
}

// GetManyFamilies returns Cursor to all Families in the Database
func GetManyFamilies(params *GetFamilyParams) (*mongo.Cursor, error) {
	var filters bson.D

	filters = append(filters, primitive.E{Key: "institution_id", Value: params.InstID})

	return familyCollection.Find(context.TODO(), filters)
}

// CreateFamily register a new Family in the DB
func CreateFamily(f FamilyRegForm, cMember MemberInFamilyRegForm, ws []Ward, vs []Vehicle) (*mongo.InsertOneResult, error) {
	// log.Printf("family to be created - %v\n", family.Guardians)
	cMemberInfo := MemberTagInfo{
		Name:     cMember.FirstName + " " + cMember.LastName,
		PhoneNum: cMember.PhoneNum,
		Relation: cMember.Relation,
	}

	newFamily := Family{
		ID:                  primitive.NewObjectID(),
		InstID:              f.InstID,
		AllRegCodeSent:      false,
		ContactGuardianInfo: cMemberInfo,
		Wards:               ws,
		Vehicles:            vs,
		ModifiedAt:          time.Now(),
	}

	return familyCollection.InsertOne(context.TODO(), newFamily)
}

// GetFamilyByID searches & returns a Family with Guardian matching the phone number
func GetFamilyByID(id string) *mongo.SingleResult {
	// TODO: err handling for ID Parsing
	oid, _ := primitive.ObjectIDFromHex(id)
	return familyCollection.FindOne(context.TODO(), bson.M{
		"_id": oid})
}

func GetFamilyByMemberID(memberID string) *mongo.SingleResult {
	// TODO: err handling for ID Parsing
	return familyCollection.FindOne(context.TODO(), bson.D{
		primitive.E{Key: "contact_member_info.id", Value: bson.D{
			primitive.E{Key: "$eq", Value: memberID},
		}},
	})
}

// GetFamilyByWardID  searches & returns a Family with Guardian matching GuardianID
func GetFamilyByWardID(wardID string) *mongo.SingleResult {
	// TODO: err handling for ID Parsing
	oid, _ := primitive.ObjectIDFromHex(wardID)
	return familyCollection.FindOne(context.TODO(), bson.D{
		primitive.E{Key: "wards._id", Value: bson.D{
			primitive.E{Key: "$eq", Value: oid},
		}},
	})
}

// GetFamilyByVehicleID  searches & returns a Family with Vehicle matching VehicleID
func GetFamilyByVehicleID(vehicleID string) *mongo.SingleResult {
	// TODO: err handling for ID Parsing
	oid, _ := primitive.ObjectIDFromHex(vehicleID)
	return familyCollection.FindOne(context.TODO(), bson.D{
		primitive.E{Key: "vehicles._id", Value: bson.D{
			primitive.E{Key: "$eq", Value: oid},
		}},
	})
}

// ReplaceFamily - Made a Family with updated "ContactMemberInfo", "Wards" and "Vehicles", and Replace the original
func ReplaceFamily(f Family, cMemberInfo MemberTagInfo, ws []Ward, vs []Vehicle) (*mongo.UpdateResult, error) {
	familyToReplace := getFamilyToReplace(f, cMemberInfo, ws, vs)
	return familyCollection.ReplaceOne(context.TODO(), bson.M{
		"_id": f.ID}, familyToReplace)
}

func SetFamilyContactMemberID(id string, cMemberID string) (*mongo.UpdateResult, error) {
	oid, _ := primitive.ObjectIDFromHex(id)
	return familyCollection.UpdateOne(context.TODO(), bson.M{"_id": oid}, bson.D{
		primitive.E{Key: "$set", Value: bson.D{
			primitive.E{Key: "contact_member_info.id", Value: cMemberID},
		}},
	})

}

// DeleteFamilyByID deletes a Family Object from DB
func DeleteFamilyByID(idToDelete string) (*mongo.DeleteResult, error) {
	// TODO: err handling for ID Parsing
	oid, _ := primitive.ObjectIDFromHex(idToDelete)

	return familyCollection.DeleteOne(context.TODO(), bson.M{
		"_id": oid})
}

func getFamilyToReplace(f Family, cMemberInfo MemberTagInfo, ws []Ward, vs []Vehicle) Family {
	return Family{
		ID:                  f.ID,
		InstID:              f.InstID,
		AllRegCodeSent:      f.AllRegCodeSent,
		ContactGuardianInfo: cMemberInfo,
		Wards:               ws,
		Vehicles:            vs,
		ModifiedAt:          time.Now(),
	}
}
