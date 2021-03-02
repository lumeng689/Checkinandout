package services

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type CCRecordStatus int
type CCScanType int

// CCRecordStatus Enum Defs
const (
	CCrInit             CCRecordStatus = 0
	CCrCheckInComplete                 = 1
	CCrScheduleComplete                = 2
	CCrCheckOutComplete                = 3
	CCrFailed                          = 4
)

// CCScanType Enum Defs
const (
	CC_QRCode CCScanType = 0
)

type WardInfo struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Group string `json:"group"`
}

type MemberTagInfo struct {
	ID       string `bson:"id" json:"id"`
	Name     string `bson:"name" json:"name"`
	PhoneNum string `bson:"phone_num" json:"phone_num"`
	Relation string `bson:"relation" json:"relation"`
	Group    string `json:"group"`
}

// CCAppPostingForm - Can Create multiple CCRecords
type CCAppPostingForm struct {
	InstID   string    `json:"institution_id"`
	MemberID *string   `json:"member_id"`
	WardIDs  *[]string `json:"ward_ids"`
}

// ScanPostingForm - As Name Suggests
type ScanPostingForm struct {
	// ScanResult = GaurdianID + WardID + "checkin/checkOut"+"single/all" + Timestamp
	ScanResult  string     `json:"scan_result"`
	Temperature float32    `json:"temperature"`
	Mask        bool       `json:"mask"`
	ScanType    CCScanType `bson:"scan_type" json:"scan_type"`
	DeviceID    string     `bson:"device_id" json:"device_id"`
}

// SchedulePostingForm - As Name Suggests
type SchedulePostingForm struct {
	WardIDs   []string `json:"ward_ids"`
	TimeStamp int      `json:"timestamp"`
}

// GuardianEvent - Can Update multiple CCRecords of Wards under a Family
type GuardianEvent struct {
	IsSingleEvent bool          `json:"is_single_event"`
	GuardianInfo  MemberTagInfo `bson:"guardian_info" json:"guardian_info"`
	ScanType      CCScanType    `bson:"scan_type" json:"scan_type"`
	Temperature   float32       `json:"temperature"`
	Mask          bool          `json:"mask"`
	Time          time.Time     `json:"time"`
}

// MemberTagEvent - Can Update multiple CCRecords of Wards under a Family
type MemberTagEvent struct {
	ScanType    CCScanType `bson:"scan_type" json:"scan_type"`
	Temperature float32    `json:"temperature"`
	Mask        bool       `json:"mask"`
	Time        time.Time  `json:"time"`
}

type GW struct {
	WardInfo      WardInfo      `bson:"ward_info" json:"ward_info"`
	CheckInEvent  GuardianEvent `bson:"check_in_event" json:"check_in_event"`
	CheckOutEvent GuardianEvent `bson:"check_out_event" json:"check_out_event"`
}

type MT struct {
	Info          MemberTagInfo  `bson:"info" json:"info"`
	CheckInEvent  MemberTagEvent `bson:"check_in_event" json:"check_in_event"`
	CheckOutEvent MemberTagEvent `bson:"check_out_event" json:"check_out_event"`
}

// CCRecord - Single Checkin/Checkout Event
type CCRecord struct {
	ID                  primitive.ObjectID `bson:"_id" json:"_id"`
	InstID              string             `bson:"institution_id" json:"institution_id"`
	HasExpired          bool               `bson:"has_expired" json:"has_expired"`
	Temperature         float32            `json:"temperature"`
	GW                  *GW                `json:"gw"`
	MT                  *MT                `json:"mt"`
	CheckOutScheduledAt time.Time          `bson:"check_out_scheduled_at" json:"check_out_scheduled_at"`
	Status              CCRecordStatus     `json:"status"`
}

type GetCCRecordParams struct {
	InstID            string
	WardID            string
	MemberTagID       string
	StartDate         time.Time
	EndDate           time.Time
	TemperatureThrd   float32
	Status            int
	ExcludeStatusList []int
}

type MarkCCRecordAsExpiredParams struct {
	MemberType MemberType
	WardID     string
	MemberID   string
	TagID      string
	InstID     string
}

// CreateCCRecordData - used in "CreateCCRecord" to init Subject Data
type CreateCCRecordData struct {
	Ward   *Ward
	Member *Member
	Tag    *Tag
}

type NewEventData struct {
	GuardianEvent  *GuardianEvent
	MemberTagEvent *MemberTagEvent
}

var ccRecordCollection *mongo.Collection

// CCRecordCollection returns reference to DB collection
func CCRecordCollection(c *mongo.Database) {
	ccRecordCollection = c.Collection("CCRecords")
}

func GetManyCCRecords(params *GetCCRecordParams, mType MemberType) (*mongo.Cursor, error) {
	// Determine Field on which to filter time
	timeFilterKey := getFilterKeyRoot(mType) + ".check_in_event.time"

	// TODO: not sending status: "4 - deleted"
	var filters bson.D
	filters = append(filters, primitive.E{Key: "institution_id", Value: params.InstID})
	if !params.StartDate.IsZero() {
		filters = append(filters, primitive.E{
			Key: timeFilterKey, Value: bson.D{primitive.E{
				Key: "$gt", Value: params.StartDate,
			}},
		})
	}
	if !params.EndDate.IsZero() {
		filters = append(filters, primitive.E{
			Key: timeFilterKey, Value: bson.D{primitive.E{
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
	if len(params.InstID) > 0 {
		filters = append(filters, primitive.E{Key: "institution_id", Value: params.InstID})
	}
	if len(params.MemberTagID) > 0 {
		filters = append(filters, primitive.E{Key: "mt.info.id", Value: params.MemberTagID})
	}
	if len(params.WardID) > 0 {
		filters = append(filters, primitive.E{Key: "gw.ward_info.id", Value: params.WardID})
	}
	if params.Status != -1 {
		filters = append(filters, primitive.E{Key: "status", Value: params.Status})
	} else if len(params.ExcludeStatusList) > 0 {
		filters = append(filters, primitive.E{Key: "status", Value: bson.D{
			primitive.E{Key: "$nin", Value: params.ExcludeStatusList},
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

func CreateCCRecord(instID string, initData CreateCCRecordData) (*mongo.InsertOneResult, error) {
	newCCRecord := CCRecord{
		ID:         primitive.NewObjectID(),
		InstID:     instID,
		HasExpired: false,
	}
	if initData.Ward != nil {
		// Case 1 - Initialize Ward data (in GW)
		w := initData.Ward
		wInfo := WardInfo{
			ID:    w.ID.Hex(),
			Name:  w.FirstName + " " + w.LastName,
			Group: w.Group,
		}

		GW := GW{
			WardInfo: wInfo,
		}
		newCCRecord.GW = &GW

	} else if initData.Member != nil {
		// Case 2 - Initialize Member data (in MT)
		m := initData.Member
		mInfo := MemberTagInfo{
			ID:       m.ID.Hex(),
			Name:     m.FirstName + " " + m.LastName,
			PhoneNum: m.PhoneNum,
			Group:    m.Group,
		}

		M := MT{
			Info: mInfo,
		}
		newCCRecord.MT = &M

	} else if initData.Tag != nil {
		// Case 3 - Initialize Tag data (in MT)
		t := initData.Tag
		tInfo := MemberTagInfo{
			ID:       t.TagString,
			Name:     t.FirstName + " " + t.LastName,
			PhoneNum: t.PhoneNum,
			Group:    t.Group,
		}
		T := MT{
			Info: tInfo,
		}
		newCCRecord.MT = &T
	}
	return ccRecordCollection.InsertOne(context.TODO(), newCCRecord)
}

func UpdateCCRecordWithEvent(ccr CCRecord, eventData NewEventData, mType MemberType, stage string, scanFailed bool) (*mongo.UpdateResult, error) {
	updatedCCR := getUpdatedCCRecordWithEvent(ccr, eventData, mType, stage, scanFailed)
	return ccRecordCollection.ReplaceOne(context.TODO(), bson.M{
		"_id": updatedCCR.ID}, updatedCCR)
}

// UpdateCCRecordScheduleTime - as is
func UpdateCCRecordScheduleTime(id string, time time.Time) (*mongo.UpdateResult, error) {
	oid, _ := primitive.ObjectIDFromHex(id)
	return ccRecordCollection.UpdateOne(context.TODO(), bson.M{"_id": oid}, bson.D{
		primitive.E{Key: "$set", Value: bson.D{
			primitive.E{Key: "check_out_scheduled_at", Value: time},
			primitive.E{Key: "status", Value: CCrScheduleComplete},
		}},
	})
}

func MarkCCRecordAsExpired(params MarkCCRecordAsExpiredParams) (*mongo.UpdateResult, error) {

	// Mark Inst ID
	if len(params.InstID) > 0 {
		filter := bson.D{
			primitive.E{Key: "institution_id", Value: params.InstID},
		}
		return ccRecordCollection.UpdateMany(context.TODO(), filter, bson.D{
			primitive.E{Key: "$set", Value: bson.D{
				primitive.E{Key: "has_expired", Value: true},
			}},
		})
	}

	// Mark by Info
	// Case 2 & 3
	if !(params.MemberType == MemberTypeGuardian) {
		var mtID string
		// Case 2 - Check MemberID in Info
		if len(params.MemberID) > 0 {
			mtID = params.MemberID
		}
		// Case 3 - Check TagID in Info
		if len(params.TagID) > 0 {
			mtID = params.TagID
		}
		filter := bson.D{
			primitive.E{Key: "mt.info.id", Value: mtID},
		}
		return ccRecordCollection.UpdateMany(context.TODO(), filter, bson.D{
			primitive.E{Key: "$set", Value: bson.D{
				primitive.E{Key: "has_expired", Value: true},
			}},
		})
	}
	// Mark by Info or Events
	// Case 1 - Check MemberID in Events & Check WardID in Info
	if len(params.MemberID) > 0 {
		filter := bson.D{
			primitive.E{Key: "gw.check_in_event.guardian_info.id", Value: params.MemberID},
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
			primitive.E{Key: "gw.check_out_event.guardian_info.id", Value: params.MemberID},
		}
		return ccRecordCollection.UpdateMany(context.TODO(), filter, bson.D{
			primitive.E{Key: "$set", Value: bson.D{
				primitive.E{Key: "has_expired", Value: true},
			}},
		})
	}
	if len(params.WardID) > 0 {
		filter := bson.D{
			primitive.E{Key: "gw.ward_info.id", Value: params.WardID},
		}
		return ccRecordCollection.UpdateMany(context.TODO(), filter, bson.D{
			primitive.E{Key: "$set", Value: bson.D{
				primitive.E{Key: "has_expired", Value: true},
			}},
		})
	}
	return nil, nil
}

// UpdateManyCCRecordsMTInfoByMTID - as is
func UpdateManyCCRecordsMTInfoByMTID(mtID string, mtInfo MemberTagInfo, mType MemberType) (*mongo.UpdateResult, error) {
	// Case 2 & 3
	if !(mType == MemberTypeGuardian) {
		return handleUpdateMTInfo("mt.info", mtID, mtInfo)
	}
	// Case 1
	res, err := handleUpdateMTInfo("gw.check_in_event.guardian_info", mtID, mtInfo)
	if err != nil {
		return res, err
	}
	return handleUpdateMTInfo("gw.check_out_event.guardian_info", mtID, mtInfo)
}

// UpdateManyCCRecordsWardInfoByWardID - as is
func UpdateManyCCRecordsWardInfoByWardID(wID string, wInfo WardInfo) (*mongo.UpdateResult, error) {

	return ccRecordCollection.UpdateMany(context.TODO(), bson.D{
		primitive.E{Key: "gw.ward_info.id", Value: wID},
	}, bson.D{
		primitive.E{Key: "$set", Value: bson.D{
			primitive.E{Key: "gw.ward_info.name", Value: wInfo.Name},
			primitive.E{Key: "gw.ward_info.group", Value: wInfo.Group},
		}},
	})
}

// DeleteCCRecordByID - as is
func DeleteCCRecordByID(idToDelete string) (*mongo.DeleteResult, error) {
	// TODO: err handling for ID Parsing
	oid, _ := primitive.ObjectIDFromHex(idToDelete)
	return ccRecordCollection.DeleteOne(context.TODO(), bson.M{"_id": oid})
}

func getUpdatedCCRecordWithEvent(ccr CCRecord, eventData NewEventData, mType MemberType, stage string, scanFailed bool) CCRecord {

	newCCR := CCRecord{
		ID:     ccr.ID,
		InstID: ccr.InstID,
	}
	if scanFailed {
		newCCR.Status = CCrFailed
	} else if stage == "checkin" {
		// if mType == MemberTypeGuardian {
		// 	newCCR.Status = CCrCheckInComplete
		// } else {
		// 	newCCR.Status = CCrScheduleComplete
		// }
		newCCR.Status = CCrCheckInComplete
	} else if stage == "checkout" {
		newCCR.CheckOutScheduledAt = ccr.CheckOutScheduledAt
		newCCR.Status = CCrCheckOutComplete
	}

	// Update Events & get New Temperature
	var newTemp float32
	if eventData.GuardianEvent != nil {
		// Case 1 - Add Guardian Event
		gEvent := eventData.GuardianEvent
		newTemp = gEvent.Temperature
		if stage == "checkin" {
			newCCR.GW = &GW{
				WardInfo:     ccr.GW.WardInfo,
				CheckInEvent: *gEvent,
			}
		} else if stage == "checkout" {
			newCCR.GW = &GW{
				WardInfo:      ccr.GW.WardInfo,
				CheckInEvent:  ccr.GW.CheckInEvent,
				CheckOutEvent: *gEvent,
			}
		}
	} else if eventData.MemberTagEvent != nil {
		// Case 2/3 - Add Member/Tag Event
		mtEvent := eventData.MemberTagEvent
		newTemp = mtEvent.Temperature
		if stage == "checkin" {
			newCCR.MT = &MT{
				Info:         ccr.MT.Info,
				CheckInEvent: *mtEvent,
			}
		} else if stage == "checkout" {
			newCCR.MT = &MT{
				Info:          ccr.MT.Info,
				CheckInEvent:  ccr.MT.CheckInEvent,
				CheckOutEvent: *mtEvent,
			}
		}
	}
	// Update Temperature
	maxTemp := ccr.Temperature
	if newTemp > maxTemp {
		maxTemp = newTemp
	}
	newCCR.Temperature = maxTemp

	return newCCR

}

func getFilterKeyRoot(mType MemberType) string {
	if mType == MemberTypeGuardian {
		return "gw"
	}
	return "mt"
}

func handleUpdateMTInfo(keyRoot string, mtID string, mtInfo MemberTagInfo) (*mongo.UpdateResult, error) {
	return ccRecordCollection.UpdateMany(context.TODO(), bson.D{
		primitive.E{Key: keyRoot + ".id", Value: mtID},
	}, bson.D{
		primitive.E{Key: "$set", Value: getMTInfoBson(keyRoot, mtInfo)},
	})
}

func getMTInfoBson(keyRoot string, mtInfo MemberTagInfo) bson.D {
	return bson.D{
		primitive.E{Key: keyRoot + ".name", Value: mtInfo.Name},
		primitive.E{Key: keyRoot + ".phone_num", Value: mtInfo.PhoneNum},
		primitive.E{Key: keyRoot + ".relation", Value: mtInfo.Relation},
		primitive.E{Key: keyRoot + ".group", Value: mtInfo.Group},
	}
}
