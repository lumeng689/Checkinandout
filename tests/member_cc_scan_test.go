package tests

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
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

func initTestMemberCC(s controllers.CCServer) {

	// Create Institution for Tag-CC Test
	instToCreate := svc.InstitutionForm{
		Type:          string(svc.InstTypeHospital),
		MemberType:    string(svc.MemberTypeStandard),
		WorkflowType:  string(svc.WorkflowTypeCC),
		Name:          "MEMBER_CC_TEST",
		Address:       "001 Test Drive",
		State:         "AZ",
		ZipCode:       "09999",
		RequireSurvey: false,
	}

	res, err := svc.CreateInst(instToCreate)
	if err != nil {
		panic(err)
	}

	// Create Member for Member-CC Test
	instID := res.InsertedID.(primitive.ObjectID).Hex()

	memberToCreate := svc.MemberRegForm{
		InstID:    instID,
		Email:     "example1@123.com",
		FirstName: "John",
		LastName:  "Doe",
		Group:     "Level 1",
		PhoneNum:  "654-321-0987",
	}

	_, err = svc.CreateMember(memberToCreate)
	if err != nil {
		panic(err)
	}
}

func TestMemberCCScan(t *testing.T) {
	instName := "MEMBER_CC_TEST"

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
	postCCSyncTestCase(t, instID, memberID)
	data := makeGateKeeperPostMemberCC(t, memberID, stage, testTemperatureNormal, testDeviceIMEI)
	postCCScanTestCase(t, data, getExpectedResponseCaseTempNormal(stage))
	checkCCRecordMember(t, memberID, getExpectedRecordMember(memberID, svc.CCrCheckInComplete))
	// Test Scan-2 - <memberID>|checkout|<timestamp>
	stage = "checkout"
	postCCSyncTestCase(t, instID, memberID)
	data = makeGateKeeperPostMemberCC(t, memberID, stage, testTemperatureNormal, testDeviceIMEI)
	postCCScanTestCase(t, data, getExpectedResponseCaseTempNormal(stage))
	checkCCRecordMember(t, memberID, getExpectedRecordMember(memberID, svc.CCrCheckOutComplete))
	// Test Scan-3 & 4 <memberID>|checkout|<timestamp> + HighTemperature
	// Test Scan-3 (Posting Check-In with HighTemp)
	stage = "checkin"
	postCCSyncTestCase(t, instID, memberID)
	data = makeGateKeeperPostMemberCC(t, memberID, stage, testTemperatureHigh, testDeviceIMEI)
	postCCScanTestCase(t, data, getExpectedResponseCaseTempHigh(stage))
	checkCCRecordMember(t, memberID, getExpectedRecordMember(memberID, svc.CCrFailed))
	// Test Scan-4 (Posting Check-In again, so to make sure failed record will not require checkout stage)
	stage = "checkin"
	postCCSyncTestCase(t, instID, memberID)
	data = makeGateKeeperPostMemberCC(t, memberID, stage, testTemperatureHigh, testDeviceIMEI)
	postCCScanTestCase(t, data, getExpectedResponseCaseTempHigh(stage))
	checkCCRecordMember(t, memberID, getExpectedRecordMember(memberID, svc.CCrFailed))
}

func checkCCRecordMember(t *testing.T, memberID string, expectedRecord svc.CCRecord) {
	var ccRecord svc.CCRecord
	ccParams := svc.GetCCRecordParams{
		MemberTagID: memberID,
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

func makeGateKeeperPostMemberCC(t *testing.T, memberID string, stage string, temperature string, imei string) url.Values {
	timestamp := time.Now().Unix() * 1000
	uniqueID := strings.Join([]string{memberID, stage, strconv.FormatInt(timestamp, 10)}, "|")
	t.Logf("makeGateKeeperPostMemberCC - uniqueID is %v\n", uniqueID)

	data := url.Values{}
	data.Set("unique_transaction_id", uniqueID)
	data.Set("temperature", temperature)
	data.Set("scan_type", "0")
	data.Set("device_id", imei)

	return data
}

func postCCSyncTestCase(t *testing.T, instID string, memberID string) {
	// make sync posting request
	syncRequest := svc.CCSyncPostingForm{
		InstID:   instID,
		MemberID: &memberID,
	}

	syncRequestString, _ := json.Marshal(syncRequest)
	req, _ := http.NewRequest("POST", "/api/cc-record/sync", strings.NewReader(string(syncRequestString)))
	w := httptest.NewRecorder()
	testRouter.ServeHTTP(w, req)

	var respData SyncResponse
	if err := json.Unmarshal(w.Body.Bytes(), &respData); err != nil {
		panic(err)
	}

	t.Logf("postCCSyncTestCase - returned CCRecord: %v\n", respData.Data)

	assert.Equal(t, 200, w.Code)
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
