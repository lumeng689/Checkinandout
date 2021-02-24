package services

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type MemberStatus int

const (
	// MAssigned - as is
	MAssigned MemberStatus = 0
	// MRegCodeSent - as is
	MRegCodeSent = 1
	// MActivated - as is
	MActivated = 2
)

type MemberInFamilyRegForm struct {
	PhoneNum  string `json:"phone_num" validate:"required,phone_num"`
	Email     string `json:"email"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Relation  string `json:"relation"`
}

type MemberRegForm struct {
	InstID     string      `json:"institution_id"`
	FamilyInfo *FamilyInfo `json:"family_info" validate:"omitempty"`
	PhoneNum   string      `json:"phone_num" validate:"required,phone_num"`
	Email      string      `json:"email"`
	FirstName  string      `json:"first_name"`
	LastName   string      `json:"last_name"`
	Group      string      `json:"group"`
}

type MemberActivateForm struct {
	PhoneNum string `json:"phone_num" validate:"required,phone_num"`
	RegCode  string `json:"reg_code"`
}

type MemberLoginForm struct {
	PhoneNum string `json:"phone_num" validate:"omitempty,phone_num"`
	DeviceID string `json:"device_id"`
}

type FamilyInfo struct {
	ID       string `bson:"id" json:"id"`
	Relation string `json:"relation"`
}

type Member struct {
	ID          primitive.ObjectID `bson:"_id" json:"_id"`
	InstID      string             `bson:"institution_id" json:"institution_id"`
	FamilyInfo  *FamilyInfo        `bson:"family_info" json:"family_info"`
	PhoneNum    string             `bson:"phone_num" json:"phone_num"`
	Email       string             `json:"email"`
	DeviceID    string             `bson:"device_id" json:"device_id"`
	FirstName   string             `bson:"first_name" json:"first_name"`
	LastName    string             `bson:"last_name" json:"last_name"`
	Group       string             `json:"group"`
	CreatedAt   time.Time          `bson:"created_at" json:"created_at"`
	LastLoginAt time.Time          `bson:"last_login_at" json:"last_login_at"`
	Status      MemberStatus       `json:"status"`
}

type GetMemberParams struct {
	InstID   string `json:"inst_id"`
	FamilyID string `json:"family_id"`
}

type MemberEditForm struct {
	FamilyInfo *FamilyInfo `json:"family_info" validate:"omitempty"`
	PhoneNum   string      `json:"phone_num" validate:"required,phone_num"`
	Email      string      `json:"email"`
	DeviceID   string      `json:"device_id"`
	FirstName  string      `json:"first_name"`
	LastName   string      `json:"last_name"`
	Group      string      `json:"group"`
}

var memberCollection *mongo.Collection

func MemberCollection(c *mongo.Database) {
	memberCollection = c.Collection("members")
}

func GetManyMembers(params *GetMemberParams) (*mongo.Cursor, error) {
	var filters bson.D
	if len(params.InstID) > 0 {
		filters = append(filters, primitive.E{Key: "institution_id", Value: params.InstID})
	} else if len(params.FamilyID) > 0 {
		filters = append(filters, primitive.E{Key: "family_info.id", Value: params.FamilyID})

	}
	return memberCollection.Find(context.TODO(), filters)
}

func GetMemberByID(id string) *mongo.SingleResult {
	oid, _ := primitive.ObjectIDFromHex(id)
	return memberCollection.FindOne(context.TODO(), bson.M{
		"_id": oid,
	})
}

func GetMemberByPhoneNum(phoneNum string) *mongo.SingleResult {
	return memberCollection.FindOne(context.TODO(), bson.D{
		primitive.E{Key: "phone_num", Value: phoneNum},
	})
}

func CountMembersByPhoneNum(phoneNum string) (int64, error) {
	return memberCollection.CountDocuments(context.TODO(), bson.D{
		primitive.E{Key: "phone_num", Value: phoneNum},
	})
}

func CreateMember(m MemberRegForm) (*mongo.InsertOneResult, error) {
	newMember := Member{
		ID:         primitive.NewObjectID(),
		InstID:     m.InstID,
		FamilyInfo: m.FamilyInfo,
		PhoneNum:   m.PhoneNum,
		Email:      m.Email,
		FirstName:  m.FirstName,
		LastName:   m.LastName,
		Group:      m.Group,
		CreatedAt:  time.Now(),
		Status:     MAssigned,
	}

	return memberCollection.InsertOne(context.TODO(), newMember)
}

// UpdateMemberLoginTimeByID - as is
func UpdateMemberLoginTimeByID(id string) error {
	oid, _ := primitive.ObjectIDFromHex(id)
	_, err := memberCollection.UpdateOne(context.TODO(), bson.M{"_id": oid}, bson.D{
		primitive.E{Key: "$currentDate", Value: bson.D{
			primitive.E{Key: "last_login_at", Value: true},
		}},
	})
	return err
}

// ActivateMemberByID - as is
func ActivateMemberByID(id string) error {
	oid, _ := primitive.ObjectIDFromHex(id)
	_, err := memberCollection.UpdateOne(context.TODO(), bson.M{"_id": oid}, bson.D{
		primitive.E{Key: "$set", Value: bson.D{
			primitive.E{Key: "status", Value: MActivated},
		}},
	})
	return err
}

// SetMemberRegCodeSentByPhoneNum - as is
func SetMemberRegCodeSentByPhoneNum(phoneNum string) error {
	_, err := memberCollection.UpdateOne(context.TODO(), bson.D{
		primitive.E{Key: "phone_num", Value: phoneNum}}, bson.D{
		primitive.E{Key: "$set", Value: bson.D{
			primitive.E{Key: "status", Value: MRegCodeSent},
		}},
	})
	return err
}

// UpdateMemberByID - as is
func UpdateMemberByID(m MemberEditForm, idToUpdate string) (*mongo.UpdateResult, error) {
	oid, _ := primitive.ObjectIDFromHex(idToUpdate)
	return memberCollection.UpdateOne(context.TODO(), bson.M{"_id": oid}, bson.D{
		primitive.E{Key: "$set", Value: bson.D{
			primitive.E{Key: "family_info", Value: m.FamilyInfo},
			primitive.E{Key: "phone_num", Value: m.PhoneNum},
			primitive.E{Key: "email", Value: m.Email},
			primitive.E{Key: "device_id", Value: m.DeviceID},
			primitive.E{Key: "first_name", Value: m.FirstName},
			primitive.E{Key: "last_name", Value: m.LastName},
			primitive.E{Key: "group", Value: m.Group},
		}},
	})
}

// DeleteMemberByID - as is
func DeleteMemberByID(idToDelete string) (*mongo.DeleteResult, error) {
	oid, _ := primitive.ObjectIDFromHex(idToDelete)
	return memberCollection.DeleteOne(context.TODO(), bson.M{"_id": oid})
}
