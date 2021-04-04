package tests

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

	"cloudminds.com/harix/cc-server/controllers"
	svc "cloudminds.com/harix/cc-server/services"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// TODO - test RegCode
var instFormFamilyTest = svc.InstitutionForm{
	Type:          string(svc.InstTypeSchool),
	MemberType:    string(svc.MemberTypeGuardian),
	WorkflowType:  string(svc.WorkflowTypeCC),
	Name:          "FAMILY_CC_TEST",
	Address:       "001 Test Drive",
	State:         "AZ",
	ZipCode:       "09999",
	RequireSurvey: false,
}

var guardianFormFamilyTest = svc.MemberInFamilyRegForm{
	PhoneNum:  "194-508-0608",
	Email:     "example1@123.com",
	FirstName: "John",
	LastName:  "Doe",
	Relation:  "Brother",
}

var w1 = svc.WardForm{
	FirstName: "Ward 1",
	LastName:  "Brown",
	Group:     "Level 1",
}

var w2 = svc.WardForm{
	FirstName: "Ward 2",
	LastName:  "Brown",
	Group:     "Level 2",
}

var familyFormFamilyTest = svc.FamilyRegForm{
	Wards: []svc.WardForm{w1, w2},
}

func initTestFamilyCC(t *testing.T, s controllers.CCServer) {
	// Create Institution for Tag-CC Test
	instToCreate := instFormFamilyTest
	iRes, err := svc.CreateInst(instToCreate)
	if err != nil {
		t.Logf("Error when creating Testing Institution - %v\n", err)
	}
	// Create Family for Family-CC Test
	instID := iRes.InsertedID.(primitive.ObjectID).Hex()
	familyFormFamilyTest.InstID = instID

	// Preparing for Creating Family
	wardsToCreate := []svc.Ward{}
	for _, wForm := range familyFormFamilyTest.Wards {
		newWard := svc.GetNewWard(wForm)
		wardsToCreate = append(wardsToCreate, newWard)
	}

	fRes, err := svc.CreateFamily(familyFormFamilyTest, guardianFormFamilyTest, wardsToCreate, []svc.Vehicle{})
	if err != nil {
		t.Logf("Error when creating Testing Family - %v\n", err)
	}

	// Create Member In Family
	insertedFamilyID := fRes.InsertedID.(primitive.ObjectID).Hex()
	fInfo := svc.FamilyInfo{
		ID:       insertedFamilyID,
		Relation: guardianFormFamilyTest.Relation,
	}
	mRegForm := svc.MemberRegForm{
		InstID:     instID,
		FamilyInfo: &fInfo,
		PhoneNum:   guardianFormFamilyTest.PhoneNum,
		Email:      guardianFormFamilyTest.Email,
		FirstName:  guardianFormFamilyTest.FirstName,
		LastName:   guardianFormFamilyTest.LastName,
	}
	mRes, err := svc.CreateMember(mRegForm)

	if err != nil {
		t.Logf("Error when creating Testing Member - %v\n", err)
	}

	// Set Contact Member ID
	insertedMemberID := mRes.InsertedID.(primitive.ObjectID).Hex()

	_, err = svc.SetFamilyContactMemberID(insertedFamilyID, insertedMemberID)
	if err != nil {
		t.Logf("Error when Setting ContactID to testing Family - %v\n", err)
	}
}

func TestFamilyCCScan(t *testing.T) {
	instName := instFormFamilyTest.Name

	// Get Institution
	var inst svc.Institution
	if err := svc.GetInstByName(instName).Decode(&inst); err != nil {
		if err == mongo.ErrNoDocuments {
			initTestFamilyCC(t, testCCServer)
			svc.GetInstByName(instName).Decode(&inst)
		} else {
			panic(err)
		}
	}

	instID := inst.ID.Hex()

	var fParams svc.GetFamilyParams
	fParams.InstID = instID
	cursor, _ := svc.GetManyFamilies(&fParams)

	var families []svc.Family
	if err := cursor.All(context.TODO(), &families); err != nil {
		panic(err)
	}
	family := families[0]

	var stage string
	wardIDList := getWardIDList(family)
	// Test Scan-1 & 2 Normal Temp + Single
	// Test Scan-1 - <guardianID>|<wardID>|checkin|single|<timestamp>
	wardID := wardIDList[0]
	stage = "checkin"
	postCCSync(t, getSyncRequestFamily(instID, []string{wardID}))
	data := makeGateKeeperPost(testTemperatureNormal, testDeviceIMEI,
		getFamilyUniqueIDSingle(family.ContactGuardianInfo.ID, wardID, stage))
	postCCScanTestCase(t, data, getExpectedResponseCaseTempNormal(stage))
	checkCCRecordFamily(t, getExpectedRecordList(family, []string{wardID}, svc.CCrCheckInComplete))
	// Mock Schedule Check-out
	postScheduleCheckOut(t, getScheduleCheckOutRequest([]string{wardID}))
	checkCCRecordFamily(t, getExpectedRecordList(family, []string{wardID}, svc.CCrScheduleComplete))
	// Test Scan-2 - <guardianID>|<wardID>|checkout|single|<timestamp>
	stage = "checkout"
	data = makeGateKeeperPost(testTemperatureNormal, testDeviceIMEI,
		getFamilyUniqueIDSingle(family.ContactGuardianInfo.ID, wardID, stage))
	postCCScanTestCase(t, data, getExpectedResponseCaseTempNormal(stage))
	checkCCRecordFamily(t, getExpectedRecordList(family, []string{wardID}, svc.CCrCheckOutComplete))

	// Test Scan-3 & 4 Normal Temp + All
	// Test Scan-3 - <guardianID>|<wardID>|checkin|all|<timestamp>
	stage = "checkin"
	postCCSync(t, getSyncRequestFamily(instID, wardIDList))
	data = makeGateKeeperPost(testTemperatureNormal, testDeviceIMEI,
		getFamilyUniqueIDAll(family.ContactGuardianInfo.ID, stage))
	postCCScanTestCase(t, data, getExpectedResponseCaseTempNormal(stage))
	checkCCRecordFamily(t, getExpectedRecordList(family, wardIDList, svc.CCrCheckInComplete))
	// Mock Schedule Check-out
	postScheduleCheckOut(t, getScheduleCheckOutRequest(wardIDList))
	checkCCRecordFamily(t, getExpectedRecordList(family, wardIDList, svc.CCrScheduleComplete))
	// Test Scan-4 - <guardianID>|<wardID>|checkout|all|<timestamp>
	stage = "checkout"
	data = makeGateKeeperPost(testTemperatureNormal, testDeviceIMEI,
		getFamilyUniqueIDAll(family.ContactGuardianInfo.ID, stage))
	postCCScanTestCase(t, data, getExpectedResponseCaseTempNormal(stage))
	checkCCRecordFamily(t, getExpectedRecordList(family, wardIDList, svc.CCrCheckOutComplete))

	// Test Scan-5 & 6 <guardianID>|<wardID>|checkin|single|<timestamp> + HighTemperature
	// Test Scan-5 - (Posting Check-In with HighTemp)
	wardID = wardIDList[0]
	stage = "checkin"
	postCCSync(t, getSyncRequestFamily(instID, []string{wardID}))
	data = makeGateKeeperPost(testTemperatureHigh, testDeviceIMEI,
		getFamilyUniqueIDSingle(family.ContactGuardianInfo.ID, wardID, stage))
	postCCScanTestCase(t, data, getExpectedResponseCaseTempHigh(stage))
	checkCCRecordFamily(t, getExpectedRecordList(family, []string{wardID}, svc.CCrFailed))
	// Test Scan-6 - (Posting Check-In again, to make sure failed record will not require checkout stage)
	wardID = wardIDList[0]
	stage = "checkin"
	postCCSync(t, getSyncRequestFamily(instID, []string{wardID}))
	data = makeGateKeeperPost(testTemperatureHigh, testDeviceIMEI,
		getFamilyUniqueIDSingle(family.ContactGuardianInfo.ID, wardID, stage))
	postCCScanTestCase(t, data, getExpectedResponseCaseTempHigh(stage))
	checkCCRecordFamily(t, getExpectedRecordList(family, []string{wardID}, svc.CCrFailed))
}

func checkCCRecordFamily(t *testing.T, expectedRecordList []svc.CCRecord) {
	// var ccRecordList []svc.CCRecord
	for _, expectedRecord := range expectedRecordList {
		ccRecord := svc.CCRecord{}
		ccParams := svc.GetCCRecordParams{
			WardID:    expectedRecord.GW.WardInfo.ID,
			Status:    -1, // set Status to "-1" to disable status filter
			GetLatest: true,
		}
		if err := svc.GetCCRecord(&ccParams).Decode(&ccRecord); err != nil {
			panic(err)
		}
		assert.Equal(t, expectedRecord.Status, ccRecord.Status)

		assert.NotEmpty(t, ccRecord.GW)
		assert.Equal(t, expectedRecord.GW.WardInfo.ID, ccRecord.GW.WardInfo.ID)
		assert.Equal(t, expectedRecord.GW.WardInfo.Name, ccRecord.GW.WardInfo.Name)
		assert.Equal(t, expectedRecord.GW.WardInfo.Group, ccRecord.GW.WardInfo.Group)

		var expectedEvent svc.GuardianEvent
		var actualEvent svc.GuardianEvent
		if expectedRecord.Status == svc.CCrCheckInComplete {
			expectedEvent = expectedRecord.GW.CheckInEvent
			actualEvent = ccRecord.GW.CheckInEvent
		}
		if expectedRecord.Status == svc.CCrCheckOutComplete {
			expectedEvent = expectedRecord.GW.CheckOutEvent
			actualEvent = ccRecord.GW.CheckOutEvent
		}
		assert.Equal(t, expectedEvent.IsSingleEvent, actualEvent.IsSingleEvent)
		assert.Equal(t, expectedEvent.GuardianInfo.ID, actualEvent.GuardianInfo.ID)
		assert.Equal(t, expectedEvent.GuardianInfo.Name, actualEvent.GuardianInfo.Name)
		assert.Equal(t, expectedEvent.GuardianInfo.PhoneNum, actualEvent.GuardianInfo.PhoneNum)
		assert.Equal(t, expectedEvent.GuardianInfo.Relation, actualEvent.GuardianInfo.Relation)
		assert.Equal(t, expectedEvent.DeviceID, actualEvent.DeviceID)
		// assert.Equal(t, expectedEvent.Temperature, actualEvent.Temperature)
	}

}

func postScheduleCheckOut(t *testing.T, scheduleRequest svc.SchedulePostingForm) {

	postRequestString, _ := json.Marshal(scheduleRequest)
	req, _ := http.NewRequest("POST", "/api/cc-record/schedule", strings.NewReader(string(postRequestString)))
	w := httptest.NewRecorder()
	testRouter.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
}

func getExpectedRecordList(family svc.Family, wardIDList []string, status svc.CCRecordStatus) []svc.CCRecord {
	expectedRecordList := []svc.CCRecord{}
	isSingleEvent := len(wardIDList) == 1
	for index, wardID := range wardIDList {
		ward := family.Wards[index]
		expectedGW := svc.GW{
			WardInfo: svc.WardInfo{
				ID:    wardID,
				Name:  ward.FirstName + " " + ward.LastName,
				Group: ward.Group,
			},
		}
		cgInfo := family.ContactGuardianInfo
		gwEvent := svc.GuardianEvent{
			IsSingleEvent: isSingleEvent,
			GuardianInfo: svc.MemberTagInfo{
				ID:       cgInfo.ID,
				Name:     cgInfo.Name,
				PhoneNum: cgInfo.PhoneNum,
				Relation: cgInfo.Relation,
			},
			DeviceID: testDeviceIMEI,
			// Temperature: testTemperatureNormal,
		}
		if status == svc.CCrCheckInComplete || status == svc.CCrScheduleComplete {
			expectedGW.CheckInEvent = gwEvent
		}
		if status == svc.CCrCheckOutComplete {
			expectedGW.CheckInEvent = gwEvent
			expectedGW.CheckOutEvent = gwEvent
		}

		expectedRecord := svc.CCRecord{
			GW:     &expectedGW,
			Status: status,
		}
		expectedRecordList = append(expectedRecordList, expectedRecord)
	}

	return expectedRecordList
}

func getFamilyUniqueIDSingle(guardianID string, wardID string, stage string) string {
	timestamp := time.Now().Unix() * 1000
	return strings.Join([]string{guardianID, wardID, stage, "single", strconv.FormatInt(timestamp, 10)}, "|")
}

func getFamilyUniqueIDAll(guardianID string, stage string) string {
	timestamp := time.Now().Unix() * 1000
	return strings.Join([]string{guardianID, stage, "all", strconv.FormatInt(timestamp, 10)}, "|")
}

func getSyncRequestFamily(instID string, wardIDs []string) svc.CCSyncPostingForm {
	return svc.CCSyncPostingForm{
		InstID:     instID,
		WardIDList: &wardIDs,
	}
}

func getScheduleCheckOutRequest(wardIDs []string) svc.SchedulePostingForm {
	return svc.SchedulePostingForm{
		WardIDs:   wardIDs,
		TimeStamp: int(time.Now().Unix()),
	}
}

func getWardIDList(family svc.Family) []string {

	wardIDList := []string{}
	for _, ward := range family.Wards {
		wardIDList = append(wardIDList, ward.ID.Hex())
	}
	return wardIDList
}
