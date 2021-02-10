package services

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type CCRecordStatus int

const (
	CC_Init             CCRecordStatus = 0
	CC_CheckInComplete                 = 1
	CC_ScheduleComplete                = 2
	CC_CheckOutComplete                = 3
	CC_Failed                          = 4
)

type CCScanType int

const (
	CC_QRCode CCScanType = 0
)

type GuardianInfo struct {
	Name     string `json:"name"`
	PhoneNum string `json:"phone_num"`
}

type WardInfo struct {
	Name  string `json:"name"`
	Group string `json:"group"`
}

type FamilyInfo struct {
	FamilyID string `bson:"family_id" json:"family_id"`
}

// CCAppPostingForm - Can Create multiple CCRecords
type CCAppPostingForm struct {
	WardIDs []string `json:"ward_ids"`
}

// ScanPostingForm - As Name Suggests
type ScanPostingForm struct {
	// ScanResult = GaurdianID + WardID + "checkin/checkOut"+"single/all" + Timestamp
	ScanResult  string     `json:"scan_result"`
	Temperature float32    `json:"temperature"`
	ScanType    CCScanType `bson:"scan_type" json:"scan_type"`
	DeviceID    string     `bson:"device_id" json:"device_id"`
}

// SchedulePostingForm - As Name Suggests
type SchedulePostingForm struct {
	WardIDs   []string `json:"ward_ids"`
	TimeStamp int      `json:"timestamp"`
}

// CCEvent - Can Update multiple CCRecords of Wards under a Family
type CCEvent struct {
	IsSingleEvent bool         `json:"is_single_event"`
	Temperature   float32      `json:"temperature"`
	ScanType      CCScanType   `bson:"scan_type" json:"scan_type"`
	DeviceID      string       `bson:"device_id" json:"device_id"`
	GuardianID    string       `bson:"guardian_id" json:"guardian_id"`
	GuardianInfo  GuardianInfo `bson:"guardian_info" json:"guardian_info"`
	Time          time.Time    `json:"time"`
}

// CCRecord - Single Checkin/Checkout Event
type CCRecord struct {
	ID                  primitive.ObjectID `bson:"_id" json:"_id"`
	HasExpired          bool               `bson:"has_expired" json:"has_expired"`
	Temperature         float32            `json:"temperature"`
	InstitutionID       string             `bson:"institution_id" json:"institution_id"`
	WardID              string             `bson:"ward_id" json:"ward_id"`
	WardInfo            WardInfo           `bson:"ward_info" json:"ward_info"`
	CheckInEvent        CCEvent            `bson:"check_in_event" json:"check_in_event"`
	CheckOutScheduledAt time.Time          `bson:"check_out_scheduled_at" json:"check_out_scheduled_at"`
	CheckOutEvent       CCEvent            `bson:"check_out_event" json:"check_out_event"`
	Status              CCRecordStatus     `json:"status"`
}

type GetCCRecordParams struct {
	InstID          string
	WardID          string
	StartDate       time.Time
	EndDate         time.Time
	TemperatureThrd float32
	Status          int
	ExcludeStatus   int
}

type MarkCCRecordAsExpiredParams struct {
	GuardianID string
	WardID     string
	InstID     string
}

var ccRecordCollection *mongo.Collection

// CCRecordCollection returns reference to DB collection
func CCRecordCollection(c *mongo.Database) {
	ccRecordCollection = c.Collection("CCRecords")
}

// GetManyCCRecords returns Cursor to all CCEvents in the Database
func GetManyCCRecords(params *GetCCRecordParams) (*mongo.Cursor, error) {
	// TODO: not sending status: "4 - deleted"
	var filters bson.D
	filters = append(filters, primitive.E{Key: "institution_id", Value: params.InstID})
	if !params.StartDate.IsZero() {
		filters = append(filters, primitive.E{
			Key: "check_in_event.time", Value: bson.D{primitive.E{
				Key: "$gt", Value: params.StartDate,
			}},
		})
	}
	if !params.EndDate.IsZero() {
		filters = append(filters, primitive.E{
			Key: "check_in_event.time", Value: bson.D{primitive.E{
				Key: "$lt", Value: params.EndDate,
			}},
		})
	}
	if params.TemperatureThrd != 0 {
		filters = append(filters, primitive.E{
			Key: "temperature", Value: bson.D{primitive.E{
				Key: "$gt", Value: params.TemperatureThrd,
			}},
		})
	}
	// log.Printf("GetCCEvents: filters - %f\n", filters)
	return ccRecordCollection.Find(context.TODO(), filters)
}

// GetCCRecord - find a Record by "WardID" and "Status"
func GetCCRecord(params *GetCCRecordParams) *mongo.SingleResult {

	var filters bson.D
	if len(params.WardID) > 0 {
		filters = append(filters, primitive.E{Key: "ward_id", Value: params.WardID})
	}
	if params.Status != -1 {
		filters = append(filters, primitive.E{Key: "status", Value: params.Status})
	} else if params.ExcludeStatus != -1 {
		filters = append(filters, primitive.E{Key: "status", Value: bson.D{
			primitive.E{Key: "$ne", Value: params.ExcludeStatus},
		}})
	}

	return ccRecordCollection.FindOne(context.TODO(), filters)
}

// GetCCRecordByID - find a Record by "WardID" and "Status"
func GetCCRecordByID(id string) *mongo.SingleResult {
	oid, _ := primitive.ObjectIDFromHex(id)
	return ccRecordCollection.FindOne(context.TODO(), bson.M{
		"_id": oid})
}

// CreateCCRecord register a new CCEvent in the DB
func CreateCCRecord(f Family, w Ward) (*mongo.InsertOneResult, error) {

	wardInfo := WardInfo{
		Name:  w.FirstName + " " + w.LastName,
		Group: w.Group,
	}

	newCCRecord := CCRecord{
		ID:            primitive.NewObjectID(),
		InstitutionID: f.InstID,
		WardID:        w.ID.Hex(),
		WardInfo:      wardInfo,
		Status:        CC_Init,
	}

	return ccRecordCollection.InsertOne(context.TODO(), newCCRecord)
}

// UpdateCCRecordWithEvent - as is
func UpdateCCRecordWithEvent(ccr CCRecord, ccEventToAdd CCEvent, stage string) (*mongo.UpdateResult, error) {
	updatedCCR := getUpdatedCCRecordWithEvent(ccr, ccEventToAdd, stage)
	return ccRecordCollection.ReplaceOne(context.TODO(), bson.M{
		"_id": updatedCCR.ID}, updatedCCR)
}

// UpdateCCRecordScheduleTime - as is
func UpdateCCRecordScheduleTime(id string, time time.Time) (*mongo.UpdateResult, error) {
	oid, _ := primitive.ObjectIDFromHex(id)
	return ccRecordCollection.UpdateOne(context.TODO(), bson.M{"_id": oid}, bson.D{
		primitive.E{Key: "$set", Value: bson.D{
			primitive.E{Key: "check_out_scheduled_at", Value: time},
			primitive.E{Key: "status", Value: CC_ScheduleComplete},
		}},
	})
}

// MarkCCRecordAsExpired - as is
func MarkCCRecordAsExpired(params MarkCCRecordAsExpiredParams) (*mongo.UpdateResult, error) {
	// Mark Guardian ID
	if len(params.GuardianID) > 0 {
		filter := bson.D{
			primitive.E{Key: "check_in_event.guardian_id", Value: bson.D{
				primitive.E{Key: "$eq", Value: params.GuardianID},
			}},
		}
		res, err := ccRecordCollection.UpdateMany(context.TODO(), filter, bson.D{
			primitive.E{Key: "$set", Value: bson.D{
				primitive.E{Key: "has_expired", Value: true},
			}},
		})
		if err != nil {
			return res, err
		}
		filter = bson.D{
			primitive.E{Key: "check_out_event.guardian_id", Value: bson.D{
				primitive.E{Key: "$eq", Value: params.GuardianID},
			}},
		}
		return ccRecordCollection.UpdateMany(context.TODO(), filter, bson.D{
			primitive.E{Key: "$set", Value: bson.D{
				primitive.E{Key: "has_expired", Value: true},
			}},
		})
	}
	// Mark Ward ID
	if len(params.WardID) > 0 {
		filter := bson.D{
			primitive.E{Key: "ward_id", Value: bson.D{
				primitive.E{Key: "$eq", Value: params.WardID},
			}},
		}
		return ccRecordCollection.UpdateMany(context.TODO(), filter, bson.D{
			primitive.E{Key: "$set", Value: bson.D{
				primitive.E{Key: "has_expired", Value: true},
			}},
		})
	}
	// Mark Inst ID
	if len(params.InstID) > 0 {
		filter := bson.D{
			primitive.E{Key: "institution_id", Value: bson.D{
				primitive.E{Key: "$eq", Value: params.InstID},
			}},
		}
		return ccRecordCollection.UpdateMany(context.TODO(), filter, bson.D{
			primitive.E{Key: "$set", Value: bson.D{
				primitive.E{Key: "has_expired", Value: true},
			}},
		})
	}
	return nil, nil
}

// UpdateManyCCRecordsGuardianInfoByGuardianID - as is
func UpdateManyCCRecordsGuardianInfoByGuardianID(gID string, gInfo GuardianInfo) (*mongo.UpdateResult, error) {

	res, err := ccRecordCollection.UpdateMany(context.TODO(), bson.M{
		"check_in_event.guardian_id": gID,
	}, bson.D{
		primitive.E{Key: "$set", Value: bson.D{
			primitive.E{Key: "check_in_event.guardian_info.name", Value: gInfo.Name},
			primitive.E{Key: "check_in_event.guardian_info.phone_num", Value: gInfo.PhoneNum},
		}},
	})
	if err != nil {
		return res, err
	}
	return ccRecordCollection.UpdateMany(context.TODO(), bson.M{
		"check_out_event.guardian_id": gID,
	}, bson.D{
		primitive.E{Key: "$set", Value: bson.D{
			primitive.E{Key: "check_out_event.guardian_info.name", Value: gInfo.Name},
			primitive.E{Key: "check_out_event.guardian_info.phone_num", Value: gInfo.PhoneNum},
		}},
	})
}

// UpdateManyCCRecordsWardInfoByWardID - as is
func UpdateManyCCRecordsWardInfoByWardID(wID string, wInfo WardInfo) (*mongo.UpdateResult, error) {

	return ccRecordCollection.UpdateMany(context.TODO(), bson.M{
		"ward_id": wID,
	}, bson.D{
		primitive.E{Key: "$set", Value: bson.D{
			primitive.E{Key: "ward_info.name", Value: wInfo.Name},
			primitive.E{Key: "ward_info.group", Value: wInfo.Group},
		}},
	})
}

// DeleteCCRecordByID - as is
func DeleteCCRecordByID(idToDelete string) (*mongo.DeleteResult, error) {
	// TODO: err handling for ID Parsing
	oid, _ := primitive.ObjectIDFromHex(idToDelete)
	return ccRecordCollection.DeleteOne(context.TODO(), bson.M{"_id": oid})
}

// Update (maximum) Temperature
func getUpdatedCCRecordWithEvent(ccr CCRecord, ccEventToAdd CCEvent, stage string) CCRecord {

	maxTemp := ccr.Temperature
	newTemp := ccEventToAdd.Temperature
	if newTemp > maxTemp {
		maxTemp = newTemp
	}
	newCCR := CCRecord{
		ID:            ccr.ID,
		Temperature:   maxTemp,
		InstitutionID: ccr.InstitutionID,
		WardID:        ccr.WardID,
		WardInfo:      ccr.WardInfo,
	}
	if stage == "checkin" {
		newCCR.CheckInEvent = ccEventToAdd
		newCCR.Status = CC_CheckInComplete
	} else if stage == "checkout" {
		newCCR.CheckInEvent = ccr.CheckInEvent
		newCCR.CheckOutScheduledAt = ccr.CheckOutScheduledAt
		newCCR.CheckOutEvent = ccEventToAdd
		newCCR.Status = CC_CheckOutComplete
	}
	return newCCR
}
