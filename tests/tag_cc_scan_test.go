package tests

import (
	"strconv"
	"strings"
	"testing"
	"time"

	"cloudminds.com/harix/cc-server/controllers"
	svc "cloudminds.com/harix/cc-server/services"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var instFormTagTest = svc.InstitutionForm{
	Type:                 string(svc.InstTypeHospital),
	MemberType:           string(svc.MemberTypeStandard),
	WorkflowType:         string(svc.WorkflowTypeCC),
	Name:                 "TAG_CC_TEST",
	Address:              "001 Test Drive",
	State:                "AZ",
	ZipCode:              "09999",
	Identifier:           "TCCT",
	CustomTagStringRegex: "\\d{4}",
}

var tagFormTagTest = svc.TagRegForm{
	TagString: "1230",
	FirstName: "John",
	LastName:  "Doe",
	Group:     "Level 1",
	PhoneNum:  "654-321-0987",
}

func initTestTagCC(s controllers.CCServer) {

	// Create Institution for Tag-CC Test
	instToCreate := instFormTagTest

	res, err := svc.CreateInst(instToCreate)
	if err != nil {
		panic(err)
	}

	// var inst svc.Institution
	// err = svc.GetInstByID(res.InsertedID.(primitive.ObjectID).Hex()).Decode(&inst)
	instID := res.InsertedID.(primitive.ObjectID).Hex()
	// Create Member for Tag-CC Test
	tagToCreate := tagFormTagTest
	tagToCreate.InstID = instID

	_, err = svc.CreateTag(tagToCreate)
	if err != nil {
		panic(err)
	}
}

func TestTagCCScan(t *testing.T) {
	instName := instFormTagTest.Name
	identifier := instFormTagTest.Identifier

	// Set-Up testing data if not already
	if count, _ := svc.CountInstByName(instName); count == 0 {
		initTestTagCC(testCCServer)
	}

	var tagString string
	var stage string
	// Test Scan-1-1 - TCCT|1230|checkin|<timestamp>
	tagString = tagFormTagTest.TagString
	stage = "checkin"
	data := makeGateKeeperPost(testTemperatureNormal, testDeviceIMEI,
		getTagUniqueID(identifier, tagString, stage))
	postCCScanTestCase(t, data, getExpectedResponseCaseTempNormal(stage))
	checkCCRecordTag(t, getExpectedRecordTagCaseExisted(svc.CCrCheckInComplete))

	// Test Scan-1-2 - TCCT|1230|checkout|<timestamp>
	stage = "checkout"
	data = makeGateKeeperPost(testTemperatureNormal, testDeviceIMEI,
		getTagUniqueID(identifier, tagString, stage))
	postCCScanTestCase(t, data, getExpectedResponseCaseTempNormal(stage))
	checkCCRecordTag(t, getExpectedRecordTagCaseExisted(svc.CCrCheckOutComplete))

	// Test Scan-1-3 & 1-4 - TCCT|1230|checkin|<timestamp> + HighTemperature
	// Test Scan-1-3 (Posting Check-In with HighTemp)
	stage = "checkin"
	data = makeGateKeeperPost(testTemperatureHigh, testDeviceIMEI,
		getTagUniqueID(identifier, tagString, stage))
	postCCScanTestCase(t, data, getExpectedResponseCaseTempHigh(stage))
	checkCCRecordTag(t, getExpectedRecordTagCaseExisted(svc.CCrFailed))
	// Test Scan-1-4 (Posting Check-In again, so to make sure failed record will not require checkout stage)
	stage = "checkin"
	data = makeGateKeeperPost(testTemperatureHigh, testDeviceIMEI,
		getTagUniqueID(identifier, tagString, stage))
	postCCScanTestCase(t, data, getExpectedResponseCaseTempHigh(stage))
	checkCCRecordTag(t, getExpectedRecordTagCaseExisted(svc.CCrFailed))

	// Test Scan-2-1 - TCCT|1231|checkin|<timestamp>
	tagString = "1231"
	stage = "checkin"
	data = makeGateKeeperPost(testTemperatureNormal, testDeviceIMEI,
		getTagUniqueID(identifier, tagString, stage))
	postCCScanTestCase(t, data, getExpectedResponseCaseTempNormal(stage))
	checkCCRecordTag(t, getExpectedRecordTagCaseNew(svc.CCrCheckInComplete))
	// Test Scan-2-2 - TCCT|1231|checkout|<timestamp>
	stage = "checkout"
	data = makeGateKeeperPost(testTemperatureNormal, testDeviceIMEI,
		getTagUniqueID(identifier, tagString, stage))
	postCCScanTestCase(t, data, getExpectedResponseCaseTempNormal(stage))
	checkCCRecordTag(t, getExpectedRecordTagCaseNew(svc.CCrCheckOutComplete))
}

func checkCCRecordTag(t *testing.T, expectedRecord svc.CCRecord) {
	var ccRecord svc.CCRecord
	ccParams := svc.GetCCRecordParams{
		MemberTagID: expectedRecord.MT.Info.ID,
		Status:      -1, // set Status to "-1" to disable status filter
		GetLatest:   true,
	}
	if err := svc.GetCCRecord(&ccParams).Decode(&ccRecord); err != nil {
		panic(err)
	}
	// t.Logf("checkCCRecord - ccRecord found: %v\n", ccRecord)

	assert.Equal(t, expectedRecord.Status, ccRecord.Status)
	assert.NotEmpty(t, ccRecord.MT)
	// t.Logf("checkCCRecord - expected info: %v, ccRecord info: %v\n", expectedRecord.MT.Info, ccRecord.MT.Info)
	assert.Equal(t, expectedRecord.MT.Info.ID, ccRecord.MT.Info.ID)
	assert.Equal(t, expectedRecord.MT.Info.Name, strings.TrimSpace(ccRecord.MT.Info.Name))
	assert.Equal(t, expectedRecord.MT.Info.Group, strings.TrimSpace(ccRecord.MT.Info.Group))
	assert.Equal(t, expectedRecord.MT.Info.PhoneNum, strings.TrimSpace(ccRecord.MT.Info.PhoneNum))

}

func getTagUniqueID(identifier string, tagString string, stage string) string {
	timestamp := time.Now().Unix() * 1000
	return strings.Join([]string{identifier, tagString, stage, strconv.FormatInt(timestamp, 10)}, "|")
}

func getExpectedRecordTagCaseExisted(status svc.CCRecordStatus) svc.CCRecord {
	expectedMT := svc.MT{
		Info: svc.MemberTagInfo{
			ID:       tagFormTagTest.TagString,
			Name:     tagFormTagTest.FirstName + " " + tagFormTagTest.LastName,
			Group:    tagFormTagTest.Group,
			PhoneNum: tagFormTagTest.PhoneNum,
		},
	}
	expectedRecord := svc.CCRecord{
		MT:     &expectedMT,
		Status: status,
	}
	return expectedRecord
}

func getExpectedRecordTagCaseNew(status svc.CCRecordStatus) svc.CCRecord {
	expectedMT := svc.MT{
		Info: svc.MemberTagInfo{
			ID:       "1231",
			Name:     "",
			Group:    "",
			PhoneNum: "",
		},
	}
	expectedRecord := svc.CCRecord{
		MT:     &expectedMT,
		Status: status,
	}
	return expectedRecord
}
