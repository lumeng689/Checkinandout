package tests

import (
	"net/url"
	"strconv"
	"strings"
	"testing"
	"time"

	"cloudminds.com/harix/cc-server/controllers"
	svc "cloudminds.com/harix/cc-server/services"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func initTestTagCC(s controllers.CCServer) {

	// Create Institution for Tag-CC Test
	instToCreate := svc.InstitutionForm{
		Type:                 string(svc.InstTypeCorporate),
		MemberType:           string(svc.MemberTypeTag),
		WorkflowType:         string(svc.WorkflowTypeCC),
		Name:                 "TAG_CC_TEST",
		Address:              "001 Test Drive",
		State:                "AZ",
		ZipCode:              "09999",
		Identifier:           "TCCT",
		CustomTagStringRegex: "\\d{4}",
	}

	res, err := svc.CreateInst(instToCreate)
	if err != nil {
		panic(err)
	}

	// var inst svc.Institution
	// err = svc.GetInstByID(res.InsertedID.(primitive.ObjectID).Hex()).Decode(&inst)
	instID := res.InsertedID.(primitive.ObjectID).Hex()
	// Create Member for Tag-CC Test
	tagToCreate := svc.TagRegForm{
		InstID:    instID,
		TagString: "1230",
		FirstName: "John",
		LastName:  "Doe",
		Group:     "Level 1",
		PhoneNum:  "654-321-0987",
	}

	_, err = svc.CreateTag(tagToCreate)
	if err != nil {
		panic(err)
	}
}

func TestTagCCScan(t *testing.T) {
	instName := "TAG_CC_TEST"
	identifier := "TCCT"

	// Set-Up testing data if not already
	if count, _ := svc.CountInstByName(instName); count == 0 {
		initTestTagCC(testCCServer)
	}

	var tagString string
	var stage string
	// Test Scan-1-1 - TCCT|1230|checkin|<timestamp>
	tagString = "1230"
	stage = "checkin"
	data := makeGateKeeperPostTagCC(t, identifier, tagString, stage, testTemperatureNormal, testDeviceIMEI)
	postCCScanTestCase(t, data, getExpectedResponseCaseTempNormal(stage))
	checkCCRecordTag(t, tagString, getExpectedRecordTagCaseExisted(svc.CCrCheckInComplete))

	// Test Scan-1-2 - TCCT|1230|checkout|<timestamp>
	stage = "checkout"
	data = makeGateKeeperPostTagCC(t, identifier, tagString, stage, testTemperatureNormal, testDeviceIMEI)
	postCCScanTestCase(t, data, getExpectedResponseCaseTempNormal(stage))
	checkCCRecordTag(t, tagString, getExpectedRecordTagCaseExisted(svc.CCrCheckOutComplete))

	// Test Scan-1-3 & 1-4 - TCCT|1230|checkin|<timestamp> + HighTemperature
	// Test Scan-1-3 (Posting Check-In with HighTemp)
	stage = "checkin"
	data = makeGateKeeperPostTagCC(t, identifier, tagString, stage, testTemperatureHigh, testDeviceIMEI)
	postCCScanTestCase(t, data, getExpectedResponseCaseTempHigh(stage))
	checkCCRecordTag(t, tagString, getExpectedRecordTagCaseExisted(svc.CCrFailed))
	// Test Scan-1-4 (Posting Check-In again, so to make sure failed record will not require checkout stage)
	stage = "checkin"
	data = makeGateKeeperPostTagCC(t, identifier, tagString, stage, testTemperatureHigh, testDeviceIMEI)
	postCCScanTestCase(t, data, getExpectedResponseCaseTempHigh(stage))
	checkCCRecordTag(t, tagString, getExpectedRecordTagCaseExisted(svc.CCrFailed))

	// Test Scan-2-1 - TCCT|1231|checkin|<timestamp>
	tagString = "1231"
	stage = "checkin"
	data = makeGateKeeperPostTagCC(t, identifier, tagString, stage, testTemperatureNormal, testDeviceIMEI)
	postCCScanTestCase(t, data, getExpectedResponseCaseTempNormal(stage))
	checkCCRecordTag(t, tagString, getExpectedRecordTagCaseNew(svc.CCrCheckInComplete))
	// Test Scan-2-2 - TCCT|1231|checkout|<timestamp>
	stage = "checkout"
	data = makeGateKeeperPostTagCC(t, identifier, tagString, stage, testTemperatureNormal, testDeviceIMEI)
	postCCScanTestCase(t, data, getExpectedResponseCaseTempNormal(stage))
	checkCCRecordTag(t, tagString, getExpectedRecordTagCaseNew(svc.CCrCheckOutComplete))
}

func checkCCRecordTag(t *testing.T, tagString string, expectedRecord svc.CCRecord) {
	var ccRecord svc.CCRecord
	ccParams := svc.GetCCRecordParams{
		MemberTagID: tagString,
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

func makeGateKeeperPostTagCC(t *testing.T, identifier string, tagString string, stage string, temperature string, imei string) url.Values {
	timestamp := time.Now().Unix() * 1000
	uniqueID := strings.Join([]string{identifier, tagString, stage, strconv.FormatInt(timestamp, 10)}, "|")
	t.Logf("makeGateKeeperPostTagCC - uniqueID is %v\n", uniqueID)

	data := url.Values{}
	data.Set("unique_transaction_id", uniqueID)
	data.Set("temperature", temperature)
	data.Set("scan_type", "0")
	data.Set("device_id", imei)

	return data
}

func getExpectedRecordTagCaseExisted(status svc.CCRecordStatus) svc.CCRecord {
	expectedMT := svc.MT{
		Info: svc.MemberTagInfo{
			ID:       "1230",
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
