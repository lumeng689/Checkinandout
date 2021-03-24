package tests

import (
	"context"
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
var instFormMemberTest = svc.InstitutionForm{
	Type:          string(svc.InstTypeHospital),
	MemberType:    string(svc.MemberTypeStandard),
	WorkflowType:  string(svc.WorkflowTypeCC),
	Name:          "MEMBER_CC_TEST",
	Address:       "001 Test Drive",
	State:         "AZ",
	ZipCode:       "09999",
	RequireSurvey: false,
}

var memberFormMemberTest = svc.MemberRegForm{
	Email:     "example1@123.com",
	FirstName: "John",
	LastName:  "Doe",
	Group:     "Level 1",
	PhoneNum:  "654-321-0987",
}

func initTestMemberCC(s controllers.CCServer) {

	// Create Institution for Tag-CC Test
	instToCreate := instFormMemberTest

	res, err := svc.CreateInst(instToCreate)
	if err != nil {
		panic(err)
	}

	// Create Member for Member-CC Test
	instID := res.InsertedID.(primitive.ObjectID).Hex()

	memberToCreate := memberFormMemberTest
	memberToCreate.InstID = instID

	_, err = svc.CreateMember(memberToCreate)
	if err != nil {
		panic(err)
	}
}

func TestMemberCCScan(t *testing.T) {
	instName := instFormMemberTest.Name

	// Get Institution
	var inst svc.Institution
	if err := svc.GetInstByName(instName).Decode(&inst); err != nil {
		if err == mongo.ErrNoDocuments {
			// Set-Up testing data if not already
			initTestMemberCC(testCCServer)
			svc.GetInstByName(instName).Decode(&inst)
		} else {
			panic(err)
		}
	}
	instID := inst.ID.Hex()

	// Get Member
	var mParams svc.GetMemberParams
	mParams.InstID = instID
	cursor, _ := svc.GetManyMembers(&mParams)

	var members []svc.Admin
	if err := cursor.All(context.TODO(), &members); err != nil {
		panic(err)
	}
	memberID := members[0].ID.Hex()

	var stage string
	// Test Scan-1 - <memberID>|checkin|<timestamp>
	stage = "checkin"
	postCCSync(t, getSyncRequestMember(instID, memberID))
	data := makeGateKeeperPost(testTemperatureNormal, testDeviceIMEI,
		getMemberUniqueID(memberID, stage))
	postCCScanTestCase(t, data, getExpectedResponseCaseTempNormal(stage))
	checkCCRecordMember(t, getExpectedRecordMember(memberID, svc.CCrCheckInComplete))
	// Test Scan-2 - <memberID>|checkout|<timestamp>
	stage = "checkout"
	postCCSync(t, getSyncRequestMember(instID, memberID))
	data = makeGateKeeperPost(testTemperatureNormal, testDeviceIMEI,
		getMemberUniqueID(memberID, stage))
	postCCScanTestCase(t, data, getExpectedResponseCaseTempNormal(stage))
	checkCCRecordMember(t, getExpectedRecordMember(memberID, svc.CCrCheckOutComplete))
	// Test Scan-3 & 4 <memberID>|checkin|<timestamp> + HighTemperature
	// Test Scan-3 (Posting Check-In with HighTemp)
	stage = "checkin"
	postCCSync(t, getSyncRequestMember(instID, memberID))
	data = makeGateKeeperPost(testTemperatureHigh, testDeviceIMEI,
		getMemberUniqueID(memberID, stage))
	postCCScanTestCase(t, data, getExpectedResponseCaseTempHigh(stage))
	checkCCRecordMember(t, getExpectedRecordMember(memberID, svc.CCrFailed))
	// Test Scan-4 (Posting Check-In again, so to make sure failed record will not require checkout stage)
	stage = "checkin"
	postCCSync(t, getSyncRequestMember(instID, memberID))
	data = makeGateKeeperPost(testTemperatureHigh, testDeviceIMEI,
		getMemberUniqueID(memberID, stage))
	postCCScanTestCase(t, data, getExpectedResponseCaseTempHigh(stage))
	checkCCRecordMember(t, getExpectedRecordMember(memberID, svc.CCrFailed))
}

func checkCCRecordMember(t *testing.T, expectedRecord svc.CCRecord) {
	var ccRecord svc.CCRecord
	ccParams := svc.GetCCRecordParams{
		MemberTagID: expectedRecord.MT.Info.ID,
		Status:      -1, // set Status to "-1" to disable status filter
		GetLatest:   true,
	}
	if err := svc.GetCCRecord(&ccParams).Decode(&ccRecord); err != nil {
		panic(err)
	}

	assert.Equal(t, expectedRecord.Status, ccRecord.Status)

	assert.NotEmpty(t, ccRecord.MT)
	assert.Equal(t, expectedRecord.MT.Info.ID, ccRecord.MT.Info.ID)
	assert.Equal(t, expectedRecord.MT.Info.Name, strings.TrimSpace(ccRecord.MT.Info.Name))
	assert.Equal(t, expectedRecord.MT.Info.Group, strings.TrimSpace(ccRecord.MT.Info.Group))
	assert.Equal(t, expectedRecord.MT.Info.PhoneNum, strings.TrimSpace(ccRecord.MT.Info.PhoneNum))
}

func getMemberUniqueID(memberID string, stage string) string {
	timestamp := time.Now().Unix() * 1000
	return strings.Join([]string{memberID, stage, strconv.FormatInt(timestamp, 10)}, "|")
}

func getExpectedRecordMember(memberID string, status svc.CCRecordStatus) svc.CCRecord {
	expectedMT := svc.MT{
		Info: svc.MemberTagInfo{
			ID:       memberID,
			Name:     "John Doe",
			Group:    "Level 1",
			PhoneNum: "654-321-0987",
		},
	}
	expectedRecord := svc.CCRecord{
		MT:     &expectedMT,
		Status: status,
	}
	return expectedRecord
}

func getSyncRequestMember(instID string, memberID string) svc.CCSyncPostingForm {
	return svc.CCSyncPostingForm{
		InstID:   instID,
		MemberID: &memberID,
	}
}
