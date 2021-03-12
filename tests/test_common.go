package tests

import (
	"encoding/json"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"cloudminds.com/harix/cc-server/controllers"
	svc "cloudminds.com/harix/cc-server/services"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

var appName = "cc-server"
var testRouter *gin.Engine
var testCCServer controllers.CCServer
var testTemperatureNormal = "98.1"
var testTemperatureHigh = "100.1"
var testDeviceIMEI = "1111222233334444"

type ScanResponse struct {
	Data    string `json:"data"`
	Stage   string `json:"stage"`
	Success bool   `json:"success"`
}

type SyncResponse struct {
	Data svc.CCRecord `json:"data"`
}

func postCCScanTestCase(t *testing.T, data url.Values, expectedResponse ScanResponse) {
	log.Printf("Data for TestCase Scan-1-1: %v\n", data.Encode())
	req, _ := http.NewRequest("POST", "/api/cc-record/scan", strings.NewReader(data.Encode()))
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()
	testRouter.ServeHTTP(w, req)

	var respData ScanResponse
	if err := json.Unmarshal(w.Body.Bytes(), &respData); err != nil {
		panic(err)
	}
	assert.Equal(t, 200, w.Code)
	assert.Equal(t, expectedResponse.Success, respData.Success)
	assert.Equal(t, expectedResponse.Stage, respData.Stage)
	if respData.Success && respData.Stage == "checkin" {
		t.Log("postTestCase - asserting non-empty url(data) field:")
		assert.NotEmpty(t, strings.TrimSpace(respData.Data))
	}
}

func getExpectedResponseCaseTempNormal(stage string) ScanResponse {
	return ScanResponse{
		Success: true,
		Stage:   stage,
	}
}

func getExpectedResponseCaseTempHigh(stage string) ScanResponse {
	return ScanResponse{
		Success: false,
		Stage:   stage,
	}
}
