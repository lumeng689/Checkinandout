package services

import (
	"context"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

func UpdateGuardianLoginTime(familyId string, guardianID string, family *Family) error {
	foid, _ := primitive.ObjectIDFromHex(familyId)
	goid, _ := primitive.ObjectIDFromHex(guardianID)
	_, err := familyCollection.UpdateOne(context.TODO(), bson.D{
		primitive.E{Key: "_id", Value: foid},
		primitive.E{Key: "guardians._id", Value: goid},
	}, bson.D{
		primitive.E{Key: "$currentDate", Value: bson.D{
			primitive.E{Key: "guardians.$.last_login_at", Value: true},
		}},
	})
	getFamilyByGuardianID(goid).Decode(&family)
	return err
}

func GetNewGuardian(g GuardianAddForm) Guardian {
	return Guardian{
		ID:        primitive.NewObjectID(),
		PhoneNum:  g.PhoneNum,
		Email:     g.Email,
		DeviceID:  "",
		FirstName: g.FirstName,
		LastName:  g.LastName,
		Relation:  g.Relation,
		Status:    GAssigned,
	}
}

func getGuardianInFamilyByPhoneNum(f Family, phoneNum string) *Guardian {
	for _, g := range f.Guardians {
		if g.PhoneNum == phoneNum {
			return &g
		}
	}
	return nil
}

func getFamilyByGuardianID(goid primitive.ObjectID) *mongo.SingleResult {
	return familyCollection.FindOne(context.TODO(), bson.D{
		primitive.E{Key: "guardians._id", Value: goid},
	})
}
