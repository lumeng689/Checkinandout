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

// var tempThrd float32 = 99.2
// var requireCheckOutTemperature = false
// var serverAddr = "http://192.168.86.101:8000"
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

	// Get Institution
	inst := svc.Institution{}
	err = svc.GetInstByID(queryParams.InstID).Decode(&inst)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			log.Printf("Institution does not exist, need to create a new one")
			c.JSON(http.StatusForbidden, gin.H{
				"message": "Institution does not exist, need to create a new one",
			})
			return
		}
		log.Printf("Error while getting institution by ID - %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Something went wrong",
		})
		return
	}

	// Get CCRecords
	cursor, err := svc.GetManyCCRecords(&queryParams, inst.MemberType)

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

// GetOrCreateManyCCRecords - For "syncing" on Mobile CC App
func (s *CCServer) GetOrCreateManyCCRecords(c *gin.Context) {
	var CCRecordsForm svc.CCAppPostingForm
	c.BindJSON(&CCRecordsForm)
	log.Printf("CCRecordForm is - %v\n", CCRecordsForm)

	// Get Institution
	inst := svc.Institution{}
	err := svc.GetInstByID(CCRecordsForm.InstID).Decode(&inst)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			log.Printf("Institution not found, Getting CCRecords Failed!")
			c.JSON(http.StatusForbidden, gin.H{
				"message": "Institution not found, Getting CCRecords Failed!",
			})
			return
		}
		log.Printf("Error while getting institution by ID - %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Something went wrong",
		})
		return
	}

	var excludeStatus svc.CCRecordStatus
	if inst.WorkflowType == svc.WorkflowTypeCC {
		excludeStatus = svc.CCrCheckOutComplete
	} else if inst.WorkflowType == svc.WorkflowTypeCheckIn {
		excludeStatus = svc.CCrCheckInComplete
	}

	// Case 2 - Member
	if CCRecordsForm.MemberID != nil {
		memberID := *CCRecordsForm.MemberID
		getParams := svc.GetCCRecordParams{
			MemberTagID:   memberID,
			Status:        -1, // set Status to "-1" to disable status filter
			ExcludeStatus: int(excludeStatus),
		}
		ccRecord := svc.CCRecord{}
		err := svc.GetCCRecord(&getParams).Decode(&ccRecord)
		if err != nil {
			// When no CCRecord Found, Create a New One and return
			if err == mongo.ErrNoDocuments {
				if ok := createCCRecordMByMemberID(c, memberID, &ccRecord); !ok {
					return
				}
			} else {
				log.Printf("Error while finding CCRecords - %v\n", err)
				c.JSON(http.StatusInternalServerError, gin.H{
					"message": "Something went wrong",
				})
				return
			}
		}
		c.JSON(http.StatusOK, gin.H{
			"message": "CCEvent",
			"data":    ccRecord,
		})
		return
	}
	// Case 1 - Gaurdian-Ward
	if CCRecordsForm.WardIDs != nil {
		var ccRecords []svc.CCRecord
		for _, wardID := range *CCRecordsForm.WardIDs {

			getParams := svc.GetCCRecordParams{
				WardID:        wardID,
				Status:        -1, // set Status to "-1" to disable status filter
				ExcludeStatus: int(excludeStatus),
			}
			ccRecord := svc.CCRecord{}
			err := svc.GetCCRecord(&getParams).Decode(&ccRecord)
			if err != nil {
				// When no CCRecord Found, Create a New One and return
				if err == mongo.ErrNoDocuments {
					if ok := createCCRecordGWByWardID(c, wardID, &ccRecord); !ok {
						return
					}
				} else {
					log.Printf("Error while finding CCRecords - %v\n", err)
					c.JSON(http.StatusInternalServerError, gin.H{
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
}

func (s *CCServer) HandleCCScanEvent(c *gin.Context) {

	var tempThrd = s.Config.TempThrd
	// Parse Posting Form
	var sPostingForm svc.ScanPostingForm
	// c.BindJSON(&sPostingForm)
	sPostingForm.ScanResult = c.PostForm("unique_transaction_id")
	temperature, _ := strconv.ParseFloat(c.PostForm("temperature"), 32)
	sPostingForm.Temperature = float32(temperature)
	scanType, _ := strconv.ParseInt(c.PostForm("scan_type"), 10, 0)
	sPostingForm.ScanType = svc.CCScanType(int(scanType))
	mask, _ := strconv.ParseBool(c.PostForm("mask"))
	sPostingForm.Mask = mask
	sPostingForm.DeviceID = c.PostForm("device_id")
	log.Printf("CCRecordForm is - %v\n", sPostingForm)
	scanResultContent := parseScanResult(sPostingForm.ScanResult)

	// Get Stage Param
	stage := scanResultContent.Stage
	var statusParam int = -1
	if stage == "checkin" {
		statusParam = int(svc.CCrInit)
	} else if stage == "checkout" {
		if scanResultContent.Type == ScanResultGWType {
			statusParam = int(svc.CCrScheduleComplete)
		} else {
			statusParam = int(svc.CCrCheckInComplete)
		}
	}

	var ok bool
	if scanResultContent.Type == ScanResultGWType {
		ok = s.handleCCScanGuardianEvent(c, sPostingForm, scanResultContent, statusParam)
	} else if scanResultContent.Type == ScanResultMemberType {
		ok = s.handleCCScanMemberEvent(c, sPostingForm, scanResultContent, statusParam)
	} else if scanResultContent.Type == ScanResultTagType {
		log.Println("scantype - tag")
		ok = s.handleCCScanTagEvent(c, sPostingForm, scanResultContent)
	}
	if !ok {
		return
	}

	// Response to Temp Scanner
	if scanResultContent.Type == ScanResultTagType {
		c.JSON(http.StatusOK, gin.H{
			"success": true,
		})
		return
	}
	// surveyURL, err := http.NewRequest("GET", s.Config.ServerAddr+surveyBaseAddr+"check-in-survey.html", nil)
	// if err != nil {
	// 	log.Printf("HandleCCEvent - error occurs when making survey url %v\n", err)
	// }
	// q := surveyURL.URL.Query()
	// q.Add("memberID", scanResultContent.MemberTagID)
	// surveyURL.URL.RawQuery = q.Encode()
	// log.Printf("SurveyURL: %v\n", surveyURL.URL.String())
	if stage == "checkin" {
		if sPostingForm.Temperature < tempThrd {
			// TODO - generate a url with guardianID
			log.Println("Checkin Scan Received, returning Success & Survey URL")
			c.JSON(http.StatusOK, gin.H{
				"success": true,
				"data":    s.Config.ServerAddr + surveyBaseAddr + "succeed-page.html",
				"stage":   "checkin",
			})
		}
		if sPostingForm.Temperature >= tempThrd {
			log.Println("Checkin Scan Received, returning Failed")
			c.JSON(http.StatusOK, gin.H{
				"success": false,
				"data":    s.Config.ServerAddr + surveyBaseAddr + "failed-page.html",
				"stage":   "checkin",
			})
			return
		}
	}
	if stage == "checkout" {
		if !s.Config.RequireCheckOutTemp || sPostingForm.Temperature < tempThrd {
			log.Println("CheckOut Scan Received, returning Success")
			c.JSON(http.StatusOK, gin.H{
				"success": true,
				"stage":   "checkout",
			})
			return
		}
		if sPostingForm.Temperature > tempThrd {
			log.Println("CheckOut Scan Received, returning Failed")
			c.JSON(http.StatusOK, gin.H{
				"success": false,
				"stage":   "checkout",
			})
			return
		}
	}

	return
}

func (s *CCServer) HandleCheckoutScheduleEvent(c *gin.Context) {
	var sPostingForm svc.SchedulePostingForm
	c.BindJSON(&sPostingForm)

	log.Printf("Schedule Form is - %v\n", sPostingForm)
	for _, wardID := range sPostingForm.WardIDs {
		// Get CCRecord
		params := svc.GetCCRecordParams{
			WardID: wardID,
			Status: svc.CCrCheckInComplete,
		}
		ccRecord := svc.CCRecord{}
		err := svc.GetCCRecord(&params).Decode(&ccRecord)
		if err != nil {
			log.Printf("CCRecord with requested WardID and Status not Exist - %v\n", err)
			c.JSON(http.StatusMethodNotAllowed, gin.H{
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

func (s *CCServer) handleCCScanGuardianEvent(c *gin.Context,
	sPostingForm svc.ScanPostingForm, scanResultContent *parsedScanResult, statusParam int) bool {
	// "scanResultContent" contains "MemberID|WardID|checkin/out|single/all|timestamp"

	// Get Member
	memberToUpdate := svc.Member{}
	err := svc.GetMemberByID(scanResultContent.MemberTagID).Decode(&memberToUpdate)
	if err != nil {
		log.Printf("Error while Getting Member By ID - %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Something went wrong",
		})
		return false
	}

	// Make EventData
	gInfo := svc.MemberTagInfo{
		ID:       memberToUpdate.ID.Hex(),
		Name:     memberToUpdate.FirstName + " " + memberToUpdate.LastName,
		PhoneNum: memberToUpdate.PhoneNum,
		Relation: memberToUpdate.FamilyInfo.Relation,
		Group:    memberToUpdate.Group,
	}
	gEventToAdd := svc.GuardianEvent{
		IsSingleEvent: scanResultContent.isSingleEvent,
		GuardianInfo:  gInfo,
		ScanType:      sPostingForm.ScanType,
		Temperature:   sPostingForm.Temperature,
		Mask:          sPostingForm.Mask,
		Time:          scanResultContent.Time,
	}
	newEventData := svc.NewEventData{
		GuardianEvent: &gEventToAdd,
	}

	if scanResultContent.isSingleEvent {
		//Single Scan Event
		params := svc.GetCCRecordParams{
			WardID: scanResultContent.WardID,
			Status: statusParam,
		}
		return getAndUpdateCCRecordWithEvent(c, params, newEventData, svc.MemberTypeGuardian, scanResultContent.Stage)

	} else {
		//Family Scan Event
		// Get Family
		familyToUpdate := svc.Family{}
		err := svc.GetFamilyByID(memberToUpdate.FamilyInfo.ID).Decode(&familyToUpdate)
		if err != nil {
			log.Printf("Error while Getting Family By ID - %v\n", err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"message": "Something went wrong",
			})
			return false
		}
		// Update CCRecords
		for _, ward := range familyToUpdate.Wards {
			params := svc.GetCCRecordParams{
				WardID: ward.ID.Hex(),
				Status: statusParam,
			}
			if ok := getAndUpdateCCRecordWithEvent(c, params, newEventData, svc.MemberTypeGuardian, scanResultContent.Stage); !ok {
				return false
			}
		}
		return true
	}

}

func (s *CCServer) handleCCScanMemberEvent(c *gin.Context,
	sPostingForm svc.ScanPostingForm, scanResultContent *parsedScanResult, statusParam int) bool {
	// "scanResultContent" contains "MemberID|checkin/out|timestamp"

	// Make EventData
	mEventToAdd := svc.MemberTagEvent{
		ScanType:    sPostingForm.ScanType,
		Temperature: sPostingForm.Temperature,
		Mask:        sPostingForm.Mask,
		Time:        scanResultContent.Time,
	}
	newEventData := svc.NewEventData{
		MemberTagEvent: &mEventToAdd,
	}

	// Get CCRecord
	params := svc.GetCCRecordParams{
		MemberTagID: scanResultContent.MemberTagID,
		Status:      statusParam,
	}

	log.Printf("handleCCScanMemberEvent - getCCRecordParams: %v\n", params)
	return getAndUpdateCCRecordWithEvent(c, params, newEventData, svc.MemberTypeStandard, scanResultContent.Stage)
}

func (s *CCServer) handleCCScanTagEvent(c *gin.Context,
	sPostingForm svc.ScanPostingForm, scanResultContent *parsedScanResult) bool {
	// "scanResultContent" contains ONLY a "TagString" param
	//// Get Institution
	inst := svc.Institution{}
	err := svc.GetInstByIdentifier(scanResultContent.InstIdentifier).Decode(&inst)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			// If Institution does not exist, abort
			log.Printf("No Institution maching the Identifier! Tag Scan Failed.")
			c.JSON(http.StatusForbidden, gin.H{
				"message": "No Institution maching the Identifier! Tag Scan Failed.",
			})
			return false

		} else {
			log.Printf("Error while getting Institution by Identifier - %v\n", err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"message": "Something went wrong",
			})
			return false
		}
	}

	//// Get Tag
	tagToProcess := svc.Tag{}
	tParams := svc.GetTagParams{
		InstID:    inst.ID.Hex(),
		TagString: scanResultContent.MemberTagID,
	}
	err = svc.GetTag(&tParams).Decode(&tagToProcess)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			// If Tag not exist, create one under institution
			tRegForm := svc.TagRegForm{
				InstID:    inst.ID.Hex(),
				TagString: scanResultContent.MemberTagID,
			}
			_, err = svc.CreateTag(tRegForm)
			err = svc.GetTag(&tParams).Decode(&tagToProcess)

			log.Printf("Tag not found by TagString, New Tag Created!")
			// c.JSON(http.StatusForbidden, gin.H{
			// 	"message": "Tag not found by TagString, Tag Scan Failed!",
			// })
			// return false
		} else {
			log.Printf("Error while getting Tag by TagString - %v\n", err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"message": "Something went wrong",
			})
			return false
		}

	}

	//// Determine Status to Exclude
	var excludeStatus svc.CCRecordStatus
	if inst.WorkflowType == svc.WorkflowTypeCC {
		excludeStatus = svc.CCrCheckOutComplete
	} else if inst.WorkflowType == svc.WorkflowTypeCheckIn {
		excludeStatus = svc.CCrCheckInComplete
	}

	//// Get OR Create CCRecord & Determine Stage
	tagID := scanResultContent.MemberTagID
	ccParams := svc.GetCCRecordParams{
		InstID:        inst.ID.Hex(),
		MemberTagID:   tagID,
		Status:        -1,
		ExcludeStatus: int(excludeStatus),
	}
	log.Printf("getCCRecordParams - %v\n", ccParams)

	ccRecord := svc.CCRecord{}
	var stage string
	var statusParam int
	err = svc.GetCCRecord(&ccParams).Decode(&ccRecord)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			// When no CCRecord Found, Create a New One and return
			if ok := createCCRecordTByTag(c, tagToProcess); !ok {
				return false
			}
			stage = "checkin"
			statusParam = int(svc.CCrInit)
		} else {
			log.Printf("Error while getting CCRecord - %v\n", err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"message": "Something went wrong",
			})
			return false

		}
	} else {
		log.Printf("record found: - %v\n", ccRecord)
		if ccRecord.Status == svc.CCrInit {
			stage = "checkin"
			statusParam = int(svc.CCrInit)
		} else if ccRecord.Status == svc.CCrCheckInComplete {
			stage = "checkout"
			statusParam = int(svc.CCrCheckInComplete)
		}
	}

	// Make EventData
	mEventToAdd := svc.MemberTagEvent{
		ScanType:    sPostingForm.ScanType,
		Temperature: sPostingForm.Temperature,
		Mask:        sPostingForm.Mask,
		Time:        time.Now(),
	}
	newEventData := svc.NewEventData{
		MemberTagEvent: &mEventToAdd,
	}

	// Get CCRecord
	ccParams = svc.GetCCRecordParams{
		InstID:      inst.ID.Hex(),
		MemberTagID: scanResultContent.MemberTagID,
		Status:      statusParam,
	}
	log.Printf("getAndUpdateCCRecordParams - %v\n", ccParams)
	return getAndUpdateCCRecordWithEvent(c, ccParams, newEventData, svc.MemberTypeTag, stage)

}

func getAndUpdateCCRecordWithEvent(c *gin.Context,
	params svc.GetCCRecordParams, newEventData svc.NewEventData, mType svc.MemberType, stage string) bool {

	ccRecord := svc.CCRecord{}
	err := svc.GetCCRecord(&params).Decode(&ccRecord)
	if err != nil {
		log.Printf("CCRecord with requested WardID and Status not Exist - %v\n", err)
		c.JSON(http.StatusMethodNotAllowed, gin.H{
			"message": "CCRecord with requested WardID and Status not Exist",
		})
		return false
	}

	_, err = svc.UpdateCCRecordWithEvent(ccRecord, newEventData, mType, stage)
	if err != nil {
		log.Printf("Error when updating CCRecord with Event - %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Somthing went wrong",
		})
		return false
	}
	return true
}

func createCCRecordGWByWardID(c *gin.Context, wardID string, newCCRecord *svc.CCRecord) bool {
	// Get Family By WardID
	familyToProcess := svc.Family{}
	err := svc.GetFamilyByWardID(wardID).Decode(&familyToProcess)
	log.Printf("Decorded familyToProcess is - %v\n", familyToProcess)
	if err != nil {
		log.Printf("Error while Getting Family by WardID into DB - %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Something went wrong",
		})
		return false
	}
	// Get Ward And Create CCRecord
	wardToProcess := getWardInFamilyByID(familyToProcess, wardID)
	log.Printf("wardToProcess is - %v\n", wardToProcess)
	if wardToProcess == nil {
		log.Printf("Error when getting Ward In Family")
		return false
	}
	initData := svc.CreateCCRecordData{
		Ward: wardToProcess,
	}

	res, err := svc.CreateCCRecord(familyToProcess.InstID, initData)
	if err != nil {
		log.Printf("Error while inserting new CCRecord into DB - %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Something went wrong",
		})
		return false
	}

	err = svc.GetCCRecordByID(res.InsertedID.(primitive.ObjectID).Hex()).Decode(newCCRecord)

	if err != nil {
		log.Printf("Error while Getting new CCRecord By ID - %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Something went wrong",
		})
		return false
	}
	return true
}

func createCCRecordMByMemberID(c *gin.Context, memberID string, newCCRecord *svc.CCRecord) bool {
	// Get Member
	memberToProcess := svc.Member{}
	err := svc.GetMemberByID(memberID).Decode(&memberToProcess)
	if err != nil {
		log.Printf("Error while Getting Member by ID From DB - %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Something went wrong",
		})
		return false
	}

	initData := svc.CreateCCRecordData{
		Member: &memberToProcess,
	}
	res, err := svc.CreateCCRecord(memberToProcess.InstID, initData)
	if err != nil {
		log.Printf("Error while inserting new CCRecord into DB - %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Something went wrong",
		})
		return false
	}
	err = svc.GetCCRecordByID(res.InsertedID.(primitive.ObjectID).Hex()).Decode(newCCRecord)
	if err != nil {
		log.Printf("Error while Getting new CCRecord By ID - %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Something went wrong",
		})
		return false
	}
	return true
}

func createCCRecordTByTag(c *gin.Context, tag svc.Tag) bool {
	initData := svc.CreateCCRecordData{
		Tag: &tag,
	}
	log.Printf("createCCRecordTByTag")
	_, err := svc.CreateCCRecord(tag.InstID, initData)
	if err != nil {
		log.Printf("Error while inserting new CCRecord into DB - %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Something went wrong",
		})
		return false
	}
	return true
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

type ScanResultType int

const (
	ScanResultGWType     ScanResultType = 1
	ScanResultMemberType                = 2
	ScanResultTagType                   = 3
)

type parsedScanResult struct {
	InstIdentifier string
	MemberTagID    string
	WardID         string
	Stage          string
	isSingleEvent  bool
	Time           time.Time
	Type           ScanResultType
}

func parseScanResult(s string) *parsedScanResult {
	contents := strings.Split(s, "|")
	isSingleEvent := true
	var scanTime time.Time
	// GW Case
	if len(contents) == 5 {
		if contents[3] == "all" {
			isSingleEvent = false
		}
		timestamp, _ := strconv.ParseInt(contents[4], 10, 64)
		scanTime = time.Unix(timestamp/1000, 0)
		return &parsedScanResult{
			MemberTagID:   contents[0],
			WardID:        contents[1],
			Stage:         contents[2],
			isSingleEvent: isSingleEvent,
			Time:          scanTime,
			Type:          ScanResultGWType,
		}
	}
	// Member Case
	if len(contents) == 3 {
		timestamp, _ := strconv.ParseInt(contents[2], 10, 64)
		log.Println("scanned timestamp - ", timestamp)
		scanTime = time.Unix(timestamp/1000, 0)
		return &parsedScanResult{
			MemberTagID: contents[0],
			Stage:       contents[1],
			Time:        scanTime,
			Type:        ScanResultMemberType,
		}
	}

	// Tag Case
	if len(contents) == 4 {
		timestamp, _ := strconv.ParseInt(contents[3], 10, 64)
		log.Println("scanned timestamp - ", timestamp)
		scanTime = time.Unix(timestamp/1000, 0)
		return &parsedScanResult{
			InstIdentifier: contents[0],
			MemberTagID:    contents[1],
			Time:           scanTime,
			Type:           ScanResultTagType,
		}
	}

	return nil
}
