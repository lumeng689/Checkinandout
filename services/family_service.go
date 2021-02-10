package services

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// GuardianStatus - ENUM for Guardian Status
type GuardianStatus int

const (
	// GAssigned - as is
	GAssigned GuardianStatus = 0
	// GRegCodeSent - as is
	GRegCodeSent = 1
	// GActivated - as is
	GActivated = 2
)

// GuardianAddForm - Input(Add) Form for Guardian
type GuardianAddForm struct {
	PhoneNum  string `json:"phone_num"`
	Email     string `json:"email"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Relation  string `json:"relation"`
}

// GuardianActivateForm - Input(Activate) Form for Guardian
type GuardianActivateForm struct {
	PhoneNum string `json:"phone_num"`
	RegCode  string `json:"reg_code"`
}

// GuardianLoginForm - User should be able to login using one of "PhoneNum" or "DeviceID"
type GuardianLoginForm struct {
	PhoneNum string `json:"phone_num"`
	DeviceID string `json:"device_id"`
}

// GuardianEditForm - Input(Edit) Form for Guardian
type GuardianEditForm struct {
	PhoneNum  string `json:"phone_num"`
	Email     string `json:"email"`
	DeviceID  string `json:"device_id"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Relation  string `json:"relation"`
}

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
	InstID    string            `json:"institution_id"`
	Guardians []GuardianAddForm `json:"guardians"`
	Wards     []WardForm        `json:"wards"`
	Vehicles  []VehicleForm     `json:"vehicles"`
}

// FamilyEditForm - Input(Edit) Form for Family
type FamilyEditForm struct {
	RegCodeSent bool   `json:"reg_code_sent"`
	InstID      string `json:"institution_id"`
	ContactID   string `json:"contact_id"`
}

// Guardian - DB Model for Guardian
type Guardian struct {
	ID          primitive.ObjectID `bson:"_id" json:"_id"`
	PhoneNum    string             `bson:"phone_num" json:"phone_num"`
	Email       string             `json:"email"`
	DeviceID    string             `json:"device_id"`
	FirstName   string             `bson:"first_name" json:"first_name"`
	LastName    string             `bson:"last_name" json:"last_name"`
	Relation    string             `json:"relation"`
	LastLoginAt time.Time          `bson:"last_login_at" json:"last_login_at"`
	Status      GuardianStatus     `json:"status"`
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
	ID             primitive.ObjectID `bson:"_id" json:"_id"`
	AllRegCodeSent bool               `json:"all_reg_code_sent"`
	InstID         string             `bson:"institution_id" json:"institution_id"`
	ContactID      string             `bson:"contact_id" json:"contact_id"`
	Guardians      []Guardian         `json:"guardians"`
	Wards          []Ward             `json:"wards"`
	Vehicles       []Vehicle          `json:"vehicles"`
	CreatedAt      time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt      time.Time          `bson:"updated_at" json:"updated_at"`
}

// GetFamilyParams - QueryString Params for GetFamily
type GetFamilyParams struct {
	InstID string `json:"inst_id"`
}

// AddGuardianParams - QueryString Params for AddGuardian
type AddGuardianParams struct {
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
func CreateFamily(f FamilyRegForm, gs []Guardian, ws []Ward, vs []Vehicle) (*mongo.InsertOneResult, error) {
	// log.Printf("family to be created - %v\n", family.Guardians)

	newFamily := Family{
		ID:             primitive.NewObjectID(),
		AllRegCodeSent: false,
		InstID:         f.InstID,
		ContactID:      gs[0].ID.Hex(),
		Guardians:      gs,
		Wards:          ws,
		Vehicles:       vs,
		CreatedAt:      time.Now(),
	}

	return familyCollection.InsertOne(context.TODO(), newFamily)
}

// GetFamilyByPhoneNum searches & returns a Family with Guardian matching the phone number
func GetFamilyByPhoneNum(phoneNum string) *mongo.SingleResult {
	return familyCollection.FindOne(context.TODO(), bson.D{
		primitive.E{Key: "guardians.phone_num", Value: bson.D{
			primitive.E{Key: "$eq", Value: phoneNum},
		}},
	})
}

// GetFamilyByID searches & returns a Family with Guardian matching the phone number
func GetFamilyByID(id string) *mongo.SingleResult {
	// TODO: err handling for ID Parsing
	oid, _ := primitive.ObjectIDFromHex(id)
	return familyCollection.FindOne(context.TODO(), bson.M{
		"_id": oid})
}

// GetFamilyByGuardianID  searches & returns a Family with Guardian matching GuardianID
func GetFamilyByGuardianID(id string) *mongo.SingleResult {
	// TODO: err handling for ID Parsing
	oid, _ := primitive.ObjectIDFromHex(id)
	return familyCollection.FindOne(context.TODO(), bson.D{
		primitive.E{Key: "guardians._id", Value: bson.D{
			primitive.E{Key: "$eq", Value: oid},
		}},
	})
}

// GetFamilyByWardID  searches & returns a Family with Guardian matching GuardianID
func GetFamilyByWardID(id string) *mongo.SingleResult {
	// TODO: err handling for ID Parsing
	oid, _ := primitive.ObjectIDFromHex(id)
	return familyCollection.FindOne(context.TODO(), bson.D{
		primitive.E{Key: "wards._id", Value: bson.D{
			primitive.E{Key: "$eq", Value: oid},
		}},
	})
}

// GetFamilyByVehicleID  searches & returns a Family with Vehicle matching VehicleID
func GetFamilyByVehicleID(id string) *mongo.SingleResult {
	// TODO: err handling for ID Parsing
	oid, _ := primitive.ObjectIDFromHex(id)
	return familyCollection.FindOne(context.TODO(), bson.D{
		primitive.E{Key: "vehicles._id", Value: bson.D{
			primitive.E{Key: "$eq", Value: oid},
		}},
	})
}

// CountFamiliesByInstID as name implies
func CountFamiliesByInstID(instID string) (int64, error) {
	// oid, _ := primitive.ObjectIDFromHex(instID)
	// log.Printf("GetManyAdminsByInstID: Decoded InstID - %v\n", oid.String())
	return familyCollection.CountDocuments(context.TODO(), bson.D{primitive.E{
		Key: "institution_id", Value: instID,
	}})
}

// UpdateFamilyInfoByID - update only "RegCodeSent", "InstID" and "ContactID" of the family
func UpdateFamilyInfoByID(f FamilyEditForm, idToUpdate string) (*mongo.UpdateResult, error) {
	oid, _ := primitive.ObjectIDFromHex(idToUpdate)
	return familyCollection.UpdateOne(context.TODO(), bson.M{"_id": oid}, bson.D{
		primitive.E{Key: "$set", Value: bson.D{
			primitive.E{Key: "reg_code_sent", Value: f.RegCodeSent},
			primitive.E{Key: "institution_id", Value: f.InstID},
			primitive.E{Key: "contact_id", Value: f.ContactID},
		}},
		primitive.E{Key: "$currentDate", Value: bson.D{
			primitive.E{Key: "updated_at", Value: true},
		}},
	})
}

// ReplaceFamily - Made a Family with updated "Gaurdians", "Wards" and "Vehicles", and Replace the original
func ReplaceFamily(f Family, gs []Guardian, ws []Ward, vs []Vehicle) (*mongo.UpdateResult, error) {

	familyToReplace := getFamilyToReplace(f, gs, ws, vs)
	return familyCollection.ReplaceOne(context.TODO(), bson.M{
		"_id": f.ID}, familyToReplace)
}

// DeleteFamilyByID deletes a Family Object from DB
func DeleteFamilyByID(idToDelete string) (*mongo.DeleteResult, error) {
	// TODO: err handling for ID Parsing
	oid, _ := primitive.ObjectIDFromHex(idToDelete)

	return familyCollection.DeleteOne(context.TODO(), bson.M{
		"_id": oid})
}

func getFamilyToReplace(f Family, gs []Guardian, ws []Ward, vs []Vehicle) Family {
	var allRegCodeSent bool
	if len(gs) > len(f.Guardians) {
		// set "regCodeSent" to false if new Guardians were added
		allRegCodeSent = false
	} else {
		// set "regCodeSent" to false if any of Guardian Status is "0 - assigned"
		allRegCodeSent = isAllRegCodeSent(gs)
	}

	return Family{
		ID:             f.ID,
		AllRegCodeSent: allRegCodeSent,
		ContactID:      f.ContactID,
		InstID:         f.InstID,
		Guardians:      gs,
		Wards:          ws,
		Vehicles:       vs,
		CreatedAt:      f.CreatedAt,
		UpdatedAt:      time.Now(),
	}
}

func isAllRegCodeSent(gs []Guardian) bool {
	// set "regCodeSent" to false if any of Guardian Status is "0 - assigned"
	for _, v := range gs {
		if v.Status == GAssigned {
			return false
		}
	}
	return true
}

// func getGuardianIDByGuardianName(name string) primitive.ObjectID {

// 	projectStage := bson.D{
// 		primitive.E{"$project", bson.D{
// 			primitive.E{"name", bson.D{
// 				primitive.E{"$concat", []string{
// 					"$first_name", " ", "$last_name",
// 				}},
// 			}},
// 		}},
// 	}

// 	matchStage := bson.D{
// 		primitive.E{"$match", bson.D{
// 			primitive.E{"name", name},
// 		}},
// 	}

// 	cursor, err := familyCollection.Aggregate(context.TODO(), mongo.Pipeline{projectStage, matchStage})

// 	if err != nil {
// 		panic(err)
// 	}
// 	var families []Family
// 	if err = cursor.All(context.TODO(), &families); err != nil {
// 		panic(err)
// 	}
// 	log.Println(families)

// 	return primitive.NewObjectID()
// }
