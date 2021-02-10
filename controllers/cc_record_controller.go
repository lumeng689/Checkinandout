package controllers

import (
	"context"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"

	svc "cloudminds.com/harix/cc-server/services"
)

var tempThrd float32 = 99.2
var needCheckOutTemperature = false
var serverAddr = "http://192.168.86.101:8000"
var surveyBaseAddr = "/surveys/"

//GetCCRecords - as is
func (s *CCServer) GetCCRecords(c *gin.Context) {
	// TODO - handle error when parsing time
	var queryParams svc.GetCCRecordParams
	err := extractCCRecordParams(c, &queryParams)

	if err != nil {
		log.Printf("Error while parsing CCEvents Query Parameters - %v\n", err)
		c.JSON(http.StatusOK, gin.H{
			"status":  http.StatusOK,
			"message": "Bad CCEvents Query Parameters",
		})
		return
	}
	log.Printf("cc-event query params - %v\n", queryParams)

	cursor, err := svc.GetManyCCRecords(&queryParams)

	if err != nil {
		log.Printf("Error while getting all CCEvents - %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  http.StatusInternalServerError,
			"message": "Something went wrong",
		})
		return
	}

	// Iterate through the returned cursor
	ccRecords := []svc.CCRecord{}
	for cursor.Next(context.TODO()) {
		var ccRecord svc.CCRecord
		cursor.Decode(&ccRecord)
		ccRecords = append(ccRecords, ccRecord)
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "All CCEvents",
		"data":    ccRecords,
	})

	return
}

// func CreateCCRecord(c *gin.Context) {
// 	var CCRecordsForm svc.CCAppPostingForm
// 	c.BindJSON(&CCRecordsForm)

// 	log.Printf("CCRecordForm is - %v\n", CCRecordsForm)
// 	for _, wardID := range CCRecordsForm.WardIDs {
// 		// TODO - Check Existing CC Records with wardID. If so, Abort
// 		createCCRecordByWardID(c, wardID, &svc.CCRecord{})
// 	}

// 	c.JSON(http.StatusCreated, gin.H{
// 		"message": "CCRecord created Successfully",
// 	})
// 	return
// }

// GetOrCreateManyCCRecordsByManyWardIDs - For "syncing" on Mobile CC App
func (s *CCServer) GetOrCreateManyCCRecordsByManyWardIDs(c *gin.Context) {
	var CCRecordsForm svc.CCAppPostingForm
	c.BindJSON(&CCRecordsForm)
	log.Printf("CCRecordForm is - %v\n", CCRecordsForm)
	var ccRecords []svc.CCRecord
	for _, wardID := range CCRecordsForm.WardIDs {

		getParams := svc.GetCCRecordParams{
			WardID:        wardID,
			Status:        -1, // set Status to "-1" to disable status filter
			ExcludeStatus: svc.CC_CheckOutComplete,
		}
		ccRecord := svc.CCRecord{}
		err := svc.GetCCRecord(&getParams).Decode(&ccRecord)
		if err != nil {
			// When no CCRecord Found, Create a New One and return
			if err == mongo.ErrNoDocuments {
				createCCRecordByWardID(c, wardID, &ccRecord)
			} else {
				log.Printf("Error while finding CCRecords - %v\n", err)
				c.JSON(http.StatusInternalServerError, gin.H{
					"status":  http.StatusInternalServerError,
					"message": "Something went wrong",
				})
				return
			}
		}
		// Append the obtained CC Record to List
		ccRecords = append(ccRecords, ccRecord)
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "All CCEvents",
		"data":    ccRecords,
	})

	return
}

// HandleCCScanEvent - API for Checkin/Checkout Scanning
func (s *CCServer) HandleCCScanEvent(c *gin.Context) {
	var sPostingForm svc.ScanPostingForm
	// c.BindJSON(&sPostingForm)
	sPostingForm.ScanResult = c.PostForm("unique_transaction_id")
	temperature, _ := strconv.ParseFloat(c.PostForm("temperature"), 32)
	sPostingForm.Temperature = float32(temperature)
	scanType, _ := strconv.ParseInt(c.PostForm("scan_type"), 10, 0)
	sPostingForm.ScanType = svc.CCScanType(int(scanType))
	sPostingForm.DeviceID = c.PostForm("device_id")
	log.Printf("CCRecordForm is - %v\n", sPostingForm)
	scanResultContent := parseScanResult(sPostingForm.ScanResult)

	// Get Guardian
	familyToUpdate := svc.Family{}
	err := svc.GetFamilyByGuardianID(scanResultContent.GuardianID).Decode(&familyToUpdate)
	if err != nil {
		log.Printf("Error while Getting Family By GuardianID - %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  http.StatusInternalServerError,
			"message": "Something went wrong",
		})
		return
	}
	guardian := getGuardianInFamilyByID(familyToUpdate, scanResultContent.GuardianID)

	stage := scanResultContent.Stage
	var statusParam int = -1
	if stage == "checkin" {
		statusParam = int(svc.CC_Init)
	} else if stage == "checkout" {
		statusParam = int(svc.CC_ScheduleComplete)
	}

	// Update CCRecord with Scan Event
	gInfo := svc.GuardianInfo{
		Name:     guardian.FirstName + " " + guardian.LastName,
		PhoneNum: guardian.PhoneNum,
	}

	ccEventToAdd := svc.CCEvent{
		IsSingleEvent: scanResultContent.isSingleEvent,
		Temperature:   sPostingForm.Temperature,
		ScanType:      sPostingForm.ScanType,
		DeviceID:      sPostingForm.DeviceID,
		GuardianID:    scanResultContent.GuardianID,
		GuardianInfo:  gInfo,
		Time:          scanResultContent.Time,
	}

	// Get and Update CCRecord
	if scanResultContent.isSingleEvent {
		// Single Scan Event
		params := svc.GetCCRecordParams{
			WardID: scanResultContent.WardID,
			Status: statusParam,
		}
		updateCCRecordWithEvent(c, params, scanResultContent.Stage, ccEventToAdd)

	} else {
		// Family Scan Event
		for _, ward := range familyToUpdate.Wards {
			params := svc.GetCCRecordParams{
				WardID: ward.ID.Hex(),
				Status: statusParam,
			}
			updateCCRecordWithEvent(c, params, scanResultContent.Stage, ccEventToAdd)
		}
	}

	// Reponse to Temp Scanner
	//// make url+query for the survey
	surveyURL, err := http.NewRequest("GET", serverAddr+surveyBaseAddr+"check-in-survey.html", nil)
	if err != nil {
		log.Printf("HandleCCEvent - error occurs when making survey url %v\n", err)
	}
	q := surveyURL.URL.Query()
	q.Add("instID", familyToUpdate.InstID)
	q.Add("guardianID", ccEventToAdd.GuardianID)
	surveyURL.URL.RawQuery = q.Encode()
	if stage == "checkin" {
		if sPostingForm.Temperature < tempThrd {
			// TODO - generate a url with guardianID
			log.Println("Checkin Scan Received, returning Success & Survey URL")
			c.JSON(http.StatusOK, gin.H{
				"success": true,
				// "data":    surveyURL.URL.String(),
			})
		}
		if sPostingForm.Temperature >= tempThrd {
			log.Println("Checkin Scan Received, returning Failed")
			c.JSON(http.StatusOK, gin.H{
				"success": false,
			})
			return
		}
	}
	if stage == "checkout" {
		if !needCheckOutTemperature || sPostingForm.Temperature < tempThrd {
			log.Println("CheckOut Scan Received, returning Success")
			c.JSON(http.StatusOK, gin.H{
				"success": true,
			})
			return
		}
		if sPostingForm.Temperature > tempThrd {
			log.Println("CheckOut Scan Received, returning Failed")
			c.JSON(http.StatusOK, gin.H{
				"success": false,
			})
			return
		}
	}

	return
}

// HandleCheckoutScheduleEvent - API for Checkin/Checkout Scanning
func (s *CCServer) HandleCheckoutScheduleEvent(c *gin.Context) {
	var sPostingForm svc.SchedulePostingForm
	c.BindJSON(&sPostingForm)

	log.Printf("Schedule Form is - %v\n", sPostingForm)
	for _, wardID := range sPostingForm.WardIDs {
		// Get CCRecord
		params := svc.GetCCRecordParams{
			WardID: wardID,
			Status: svc.CC_CheckInComplete,
		}
		ccRecord := svc.CCRecord{}
		err := svc.GetCCRecord(&params).Decode(&ccRecord)
		if err != nil {
			log.Printf("CCRecord with requested WardID and Status not Exist - %v\n", err)
			c.JSON(http.StatusMethodNotAllowed, gin.H{
				"status":  http.StatusMethodNotAllowed,
				"message": "CCRecord with requested WardID and Status not Exist",
			})
			return
		}
		// Update CCRecord with Scheduled Time
		scheduledTime := time.Unix(int64(sPostingForm.TimeStamp), 0)
		svc.UpdateCCRecordScheduleTime(ccRecord.ID.Hex(), scheduledTime)

	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Checkout Scheduled Successfully",
	})
	return

}

// DeleteCCRecordByID - as is
func (s *CCServer) DeleteCCRecordByID(c *gin.Context) {
	// TODO: Set CC Records to "expired"
	idToDelete := c.Param("id")

	res, err := svc.DeleteCCRecordByID(idToDelete)
	if err != nil {
		log.Printf("Error while deleting CCRecord from DB - %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Something went wrong",
		})
		return
	}

	if res.DeletedCount == 0 {
		c.JSON(http.StatusForbidden, gin.H{
			"message": "ID To Delete CCRecord Not Found",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "CCRecord Deleted Successfully",
	})
	return
}

func updateCCRecordWithEvent(c *gin.Context, params svc.GetCCRecordParams, stage string, ccEventToAdd svc.CCEvent) {

	ccRecord := svc.CCRecord{}
	err := svc.GetCCRecord(&params).Decode(&ccRecord)
	if err != nil {
		log.Printf("CCRecord with requested WardID and Status not Exist - %v\n", err)
		c.JSON(http.StatusMethodNotAllowed, gin.H{
			"message": "CCRecord with requested WardID and Status not Exist",
		})
		return
	}
	_, err = svc.UpdateCCRecordWithEvent(ccRecord, ccEventToAdd, stage)
	if err != nil {
		log.Printf("Error when updating CCRecord with Event - %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Somthing went wrong",
		})
		return
	}
}

func createCCRecordByWardID(c *gin.Context, wardID string, newCCRecord *svc.CCRecord) {
	// Get Family By WardID
	familyToProcess := svc.Family{}
	err := svc.GetFamilyByWardID(wardID).Decode(&familyToProcess)
	log.Printf("Decorded familyToProcess is - %v\n", familyToProcess)
	if err != nil {
		log.Printf("Error while Getting Family by WardID into DB - %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  http.StatusInternalServerError,
			"message": "Something went wrong",
		})
		return
	}
	// Get Ward And Create CCRecord
	wardToProcess := getWardInFamilyByID(familyToProcess, wardID)
	log.Printf("wardToProcess is - %v\n", wardToProcess)
	if wardToProcess == nil {
		log.Printf("Error when getting Ward In Family")
		return
	}
	res, err := svc.CreateCCRecord(familyToProcess, *wardToProcess)
	if err != nil {
		log.Printf("Error while inserting new CCRecord into DB - %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  http.StatusInternalServerError,
			"message": "Something went wrong",
		})
		return
	}

	err = svc.GetCCRecordByID(res.InsertedID.(primitive.ObjectID).Hex()).Decode(newCCRecord)

	if err != nil {
		log.Printf("Error while Getting new CCRecord By ID - %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  http.StatusInternalServerError,
			"message": "Something went wrong",
		})
		return
	}
	return
}

func extractCCRecordParams(c *gin.Context, params *svc.GetCCRecordParams) error {
	// const shortDateForm = "2006-01-02"

	params.InstID = c.DefaultQuery("instID", "000000000000000000000000")
	params.WardID = c.DefaultQuery("wardID", "")

	param, ok := c.GetQuery("startDate")
	if ok {
		param, err := time.Parse(time.RFC3339, param)
		if err != nil {
			return err
		}
		params.StartDate = param
	}

	param, ok = c.GetQuery("endDate")
	if ok {
		param, err := time.Parse(time.RFC3339, param)
		if err != nil {
			return err
		}
		params.EndDate = param
	}

	param, ok = c.GetQuery("tempThrd")
	if ok {
		param, err := strconv.ParseFloat(param, 32)
		if err != nil {
			return err
		}
		params.TemperatureThrd = float32(param)
	}

	param, ok = c.GetQuery("status")
	if ok {
		param, _ := strconv.ParseInt(param, 10, 0)
		params.Status = int(param)
	} else {
		params.Status = -1
	}
	param, ok = c.GetQuery("excludeStatus")
	if ok {
		param, _ := strconv.ParseInt(param, 10, 0)
		params.ExcludeStatus = int(param)
	} else {
		params.ExcludeStatus = -1
	}
	return nil
}

type parsedScanResult struct {
	GuardianID    string
	WardID        string
	Stage         string
	isSingleEvent bool
	Time          time.Time
}

func parseScanResult(s string) parsedScanResult {
	contents := strings.Split(s, "|")
	isSingleEvent := true
	var scanTime time.Time
	if len(contents) > 4 {
		if contents[3] == "all" {
			isSingleEvent = false
		}
		timestamp, _ := strconv.ParseInt(contents[4], 10, 64)
		scanTime = time.Unix(timestamp, 0)
	}
	return parsedScanResult{
		contents[0],
		contents[1],
		contents[2],
		isSingleEvent,
		scanTime,
	}
}
