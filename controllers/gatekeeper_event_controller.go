package controllers

import (
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	svc "cloudminds.com/harix/cc-server/services"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
)

// HandleCCScanEvent - Accepts Gatekeeper Scan Postings, and Process it in one of three modes ["guardian", "standard" or "tag"]
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
	sResultContent := parseScanResult(sPostingForm.ScanResult)

	// Get Stage Param
	stage := sResultContent.Stage
	var statusParam int = -1
	if stage == "checkin" {
		statusParam = int(svc.CCrInit)
	} else if stage == "checkout" {
		if sResultContent.Type == ScanResultGWType {
			statusParam = int(svc.CCrScheduleComplete)
		} else {
			statusParam = int(svc.CCrCheckInComplete)
		}
	}

	// Check if Failed
	var isScanFailed = false
	if sPostingForm.Temperature > s.Config.TempThrd {
		isScanFailed = true
	}

	var ok bool
	var tagStage string
	if sResultContent.Type == ScanResultGWType {
		ok = s.handleCCScanGuardianEvent(c, sPostingForm, sResultContent, statusParam, isScanFailed)
	} else if sResultContent.Type == ScanResultMemberType {
		ok = s.handleCCScanMemberEvent(c, sPostingForm, sResultContent, statusParam, isScanFailed)
	} else if sResultContent.Type == ScanResultTagType {
		ok, tagStage = s.handleCCScanTagEvent(c, sPostingForm, sResultContent, isScanFailed)
	}
	if !ok {
		return
	}

	// Response to Temp Scanner
	// surveyURL, err := http.NewRequest("GET", s.Config.ServerAddr+surveyBaseAddr+"check-in-survey.html", nil)
	// if err != nil {
	// 	log.Printf("HandleCCEvent - error occurs when making survey url %v\n", err)
	// }
	// q := surveyURL.URL.Query()
	// q.Add("memberID", scanResultContent.MemberTagID)
	// surveyURL.URL.RawQuery = q.Encode()
	// log.Printf("SurveyURL: %v\n", surveyURL.URL.String())
	var responseStage string
	if sResultContent.Type == ScanResultTagType {
		responseStage = tagStage
	} else {
		responseStage = stage
	}

	if responseStage == "checkin" {
		if sPostingForm.Temperature < tempThrd {
			// TODO - generate a url with guardianID
			log.Println("Checkin Scan Received, returning Success & Survey URL")
			c.JSON(http.StatusOK, gin.H{
				"success": true,
				"data":    s.Config.ServerAddr + surveyBaseAddr + "succeed-page.html",
				// "data":  s.Config.ServerAddr + surveyBaseAddr + "check-in-survey.html",
				"stage": responseStage,
			})
			return
		}
		if sPostingForm.Temperature >= tempThrd {
			log.Println("Checkin Scan Received, returning Failed")
			c.JSON(http.StatusOK, gin.H{
				"success": false,
				"data":    s.Config.ServerAddr + surveyBaseAddr + "failed-page.html",
				"stage":   responseStage,
			})
			return
		}
	}
	if responseStage == "checkout" {
		if !s.Config.RequireCheckOutTemp || sPostingForm.Temperature < tempThrd {
			log.Println("CheckOut Scan Received, returning Success")
			c.JSON(http.StatusOK, gin.H{
				"success": true,
				"stage":   responseStage,
			})
			return
		}
		if sPostingForm.Temperature > tempThrd {
			log.Println("CheckOut Scan Received, returning Failed")
			c.JSON(http.StatusOK, gin.H{
				"success": false,
				"stage":   responseStage,
			})
			return
		}
	}
}

func (s *CCServer) handleCCScanGuardianEvent(c *gin.Context,
	sPostingForm svc.ScanPostingForm, sResultContent *parsedScanResult, statusParam int, scanFailed bool) bool {
	// "scanResultContent" contains "MemberID|WardID|checkin/out|single/all|timestamp"

	// Get Member
	memberToUpdate := svc.Member{}
	err := svc.GetMemberByID(sResultContent.MemberTagID).Decode(&memberToUpdate)
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
		IsSingleEvent: sResultContent.isSingleEvent,
		GuardianInfo:  gInfo,
		ScanType:      sPostingForm.ScanType,
		DeviceID:      sPostingForm.DeviceID,
		Temperature:   sPostingForm.Temperature,
		Mask:          sPostingForm.Mask,
		Time:          sResultContent.Time,
	}
	newEventData := svc.NewEventData{
		GuardianEvent: &gEventToAdd,
		Stage:         sResultContent.Stage,
		IsScanFailed:  scanFailed,
	}

	if sResultContent.isSingleEvent {
		//Single Scan Event
		ccParams := svc.GetCCRecordParams{
			WardID: sResultContent.WardID,
			Status: statusParam,
		}
		return getAndUpdateCCRecordWithEvent(c, ccParams, newEventData)
	}
	//Family Scan Event
	// Get Family
	familyToUpdate := svc.Family{}
	err = svc.GetFamilyByID(memberToUpdate.FamilyInfo.ID).Decode(&familyToUpdate)
	if err != nil {
		log.Printf("Error while Getting Family By ID - %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Something went wrong",
		})
		return false
	}
	// Update CCRecords
	for _, ward := range familyToUpdate.Wards {
		ccParams := svc.GetCCRecordParams{
			WardID: ward.ID.Hex(),
			Status: statusParam,
		}
		if ok := getAndUpdateCCRecordWithEvent(c, ccParams, newEventData); !ok {
			return false
		}
	}
	return true
}

func (s *CCServer) handleCCScanMemberEvent(c *gin.Context,
	sPostingForm svc.ScanPostingForm, sResultContent *parsedScanResult, statusParam int, isScanFailed bool) bool {
	// "scanResultContent" contains "MemberID|checkin/out|timestamp"

	// Make EventData
	mEventToAdd := svc.MemberTagEvent{
		ScanType:    sPostingForm.ScanType,
		DeviceID:    sPostingForm.DeviceID,
		Temperature: sPostingForm.Temperature,
		Mask:        sPostingForm.Mask,
		Time:        sResultContent.Time,
	}
	newEventData := svc.NewEventData{
		MemberTagEvent: &mEventToAdd,
		Stage:          sResultContent.Stage,
		IsScanFailed:   isScanFailed,
	}

	// Get CCRecord
	params := svc.GetCCRecordParams{
		MemberTagID: sResultContent.MemberTagID,
		Status:      statusParam,
	}

	log.Printf("handleCCScanMemberEvent - getCCRecordParams: %v\n", params)
	return getAndUpdateCCRecordWithEvent(c, params, newEventData)
}

func (s *CCServer) handleCCScanTagEvent(c *gin.Context,
	sPostingForm svc.ScanPostingForm, sResultContent *parsedScanResult, isScanFailed bool) (bool, string) {
	// "scanResultContent" contains ONLY a "TagString" param
	//// Get Institution
	inst := svc.Institution{}
	err := svc.GetInstByIdentifier(sResultContent.InstIdentifier).Decode(&inst)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			// If Institution does not exist, abort
			log.Printf("No Institution maching the Identifier! Tag Scan Failed.")
			c.JSON(http.StatusForbidden, gin.H{
				"message": "No Institution maching the Identifier! Tag Scan Failed.",
			})
			return false, ""

		}
		log.Printf("Error while getting Institution by Identifier - %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Something went wrong",
		})
		return false, ""

	}

	//// Get Tag
	tParams := svc.GetTagParams{
		InstID:    inst.ID.Hex(),
		TagString: sResultContent.MemberTagID,
	}

	tagToProcess, ok := getOrCreateTag(c, &tParams)
	if !ok {
		return false, ""
	}

	//// Determine Status to Exclude
	var excludeStatusList []int
	if inst.WorkflowType == svc.WorkflowTypeCC {
		excludeStatusList = []int{int(svc.CCrCheckOutComplete), int(svc.CCrFailed)}
	} else if inst.WorkflowType == svc.WorkflowTypeCheckIn {
		excludeStatusList = []int{int(svc.CCrCheckInComplete), int(svc.CCrFailed)}
	}

	//// Get OR Create CCRecord & Determine Stage
	tagID := sResultContent.MemberTagID
	ccParams := svc.GetCCRecordParams{
		InstID:            inst.ID.Hex(),
		MemberTagID:       tagID,
		Status:            -1,
		ExcludeStatusList: excludeStatusList,
	}
	log.Printf("getCCRecordParams - %v\n", ccParams)

	ccRecord := svc.CCRecord{}
	var stage string
	var statusParam int
	err = svc.GetCCRecord(&ccParams).Decode(&ccRecord)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			// When no CCRecord Found, Create a New One and return
			if ok := createCCRecordTByTag(c, *tagToProcess); !ok {
				return false, ""
			}
			stage = "checkin"
			statusParam = int(svc.CCrInit)
		} else {
			log.Printf("Error while getting CCRecord - %v\n", err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"message": "Something went wrong",
			})
			return false, ""

		}
	} else {
		// If CCRecord already exists, figure out the next stage for the record
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
		DeviceID:    sPostingForm.DeviceID,
		Temperature: sPostingForm.Temperature,
		Mask:        sPostingForm.Mask,
		Time:        time.Now(),
	}
	newEventData := svc.NewEventData{
		MemberTagEvent: &mEventToAdd,
		Stage:          stage,
		IsScanFailed:   isScanFailed,
	}

	// Get CCRecord
	ccParams = svc.GetCCRecordParams{
		InstID:      inst.ID.Hex(),
		MemberTagID: sResultContent.MemberTagID,
		Status:      statusParam,
	}
	log.Printf("getAndUpdateCCRecordParams - %v\n", ccParams)
	return getAndUpdateCCRecordWithEvent(c, ccParams, newEventData), stage

}

// ScanResultType - as is
type ScanResultType int

// ScanResultType Enum Defs
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
	timestamp, _ := strconv.ParseInt(contents[len(contents)-1], 10, 64)
	scanTime := time.Unix(timestamp/1000, 0)
	// log.Println("scanned timestamp - ", timestamp)

	// Member Case
	if len(contents) == 3 {
		return &parsedScanResult{
			MemberTagID: contents[0],
			Stage:       contents[1],
			Time:        scanTime,
			Type:        ScanResultMemberType,
		}
	}

	// Tag Case
	// If [length of Identifier] < 24, regard it as Tag case
	if len(contents) == 4 && len(contents[0]) < 24 {
		return &parsedScanResult{
			InstIdentifier: contents[0],
			MemberTagID:    contents[1],
			Time:           scanTime,
			Type:           ScanResultTagType,
		}
	}

	// GW Case - All
	if len(contents) == 4 && contents[2] == "all" {
		return &parsedScanResult{
			MemberTagID:   contents[0],
			Stage:         contents[1],
			isSingleEvent: false,
			Time:          scanTime,
			Type:          ScanResultGWType,
		}
	}
	// GW Case - Single
	if len(contents) == 5 && contents[3] == "single" {
		return &parsedScanResult{
			MemberTagID:   contents[0],
			WardID:        contents[1],
			Stage:         contents[2],
			isSingleEvent: true,
			Time:          scanTime,
			Type:          ScanResultGWType,
		}
	}

	return nil
}
