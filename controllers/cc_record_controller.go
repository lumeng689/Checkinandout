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

//GetManyCCRecords - as is
func (s *CCServer) GetManyCCRecords(c *gin.Context) {
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
	var CCRecordsForm svc.CCSyncPostingForm
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

	var excludeStatusList []int
	if inst.WorkflowType == svc.WorkflowTypeCC {
		excludeStatusList = []int{int(svc.CCrCheckOutComplete), int(svc.CCrFailed)}
	} else if inst.WorkflowType == svc.WorkflowTypeCheckIn {
		excludeStatusList = []int{int(svc.CCrCheckInComplete), int(svc.CCrFailed)}
	}

	// Case 2 - Member
	if CCRecordsForm.MemberID != nil {
		memberID := *CCRecordsForm.MemberID
		getParams := svc.GetCCRecordParams{
			MemberTagID:       memberID,
			Status:            -1, // set Status to "-1" to disable status filter
			ExcludeStatusList: excludeStatusList,
		}
		ccRecord, ok := getOrCreateCCRecordMember(c, &getParams)
		if !ok {
			return
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
				WardID:            wardID,
				Status:            -1, // set Status to "-1" to disable status filter
				ExcludeStatusList: excludeStatusList,
			}
			ccRecord, ok := getOrCreateCCRecordGW(c, &getParams)
			if !ok {
				return
			}
			// // Append the obtained CC Record to List
			ccRecords = append(ccRecords, *ccRecord)
		}
		c.JSON(http.StatusOK, gin.H{
			"message": "All CCEvents",
			"data":    ccRecords,
		})
		return
	}
	c.JSON(http.StatusBadRequest, gin.H{
		"message": "Sync CCRecord Query given is not Supported",
	})
}

// HandleCheckoutScheduleEvent - process "Schedule Checkout" request from MobileApp
func (s *CCServer) HandleCheckoutScheduleEvent(c *gin.Context) {
	var sPostingForm svc.SchedulePostingForm
	c.BindJSON(&sPostingForm)

	log.Printf("Schedule Form is - %v\n", sPostingForm)
	for _, wardID := range sPostingForm.WardIDs {
		// Get CCRecord
		params := svc.GetCCRecordParams{
			WardID: wardID,
			Status: int(svc.CCrCheckInComplete),
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

func getAndUpdateCCRecordWithEvent(c *gin.Context,
	params svc.GetCCRecordParams, newEventData svc.NewEventData) bool {

	ccRecord := svc.CCRecord{}
	err := svc.GetCCRecord(&params).Decode(&ccRecord)
	if err != nil {
		log.Printf("CCRecord with requested WardID and Status not Exist - %v\n", err)
		c.JSON(http.StatusMethodNotAllowed, gin.H{
			"message": "CCRecord with requested WardID and Status not Exist",
		})
		return false
	}

	_, err = svc.UpdateCCRecordWithEvent(ccRecord, newEventData)
	if err != nil {
		log.Printf("Error when updating CCRecord with Event - %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Somthing went wrong",
		})
		return false
	}
	return true
}

func getOrCreateCCRecordGW(c *gin.Context, ccParams *svc.GetCCRecordParams) (*svc.CCRecord, bool) {
	ccRecord := svc.CCRecord{}
	err := svc.GetCCRecord(ccParams).Decode(&ccRecord)
	if err == nil {
		return &ccRecord, true
	}
	// When no CCRecord Found, Create a New One and return
	if err == mongo.ErrNoDocuments {
		if ok := createAndGetCCRecordGWByWardID(c, ccParams.WardID, &ccRecord); !ok {
			return nil, false
		}
		return &ccRecord, true
	}
	log.Printf("Error while finding CCRecords - %v\n", err)
	c.JSON(http.StatusInternalServerError, gin.H{
		"message": "Something went wrong",
	})
	return nil, false
}

func getOrCreateCCRecordMember(c *gin.Context, ccParams *svc.GetCCRecordParams) (*svc.CCRecord, bool) {
	ccRecord := svc.CCRecord{}
	err := svc.GetCCRecord(ccParams).Decode(&ccRecord)
	if err == nil {
		return &ccRecord, true
	}
	// When no CCRecord Found, Create a New One and return
	if err == mongo.ErrNoDocuments {
		if ok := createAndGetCCRecordMByMemberID(c, ccParams.MemberTagID, &ccRecord); !ok {
			return nil, false
		}
		return &ccRecord, true
	}
	log.Printf("Error while finding CCRecords - %v\n", err)
	c.JSON(http.StatusInternalServerError, gin.H{
		"message": "Something went wrong",
	})
	return nil, false

}

func createAndGetCCRecordGWByWardID(c *gin.Context, wardID string, newCCRecord *svc.CCRecord) bool {
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

func createAndGetCCRecordMByMemberID(c *gin.Context, memberID string, newCCRecord *svc.CCRecord) bool {
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
	statusList, ok := c.GetQuery("excludeStatus")
	if ok {
		statusList := strings.Split(statusList, ",")
		statusListCvt := []int{}
		for _, s := range statusList {
			sCvt, _ := strconv.ParseInt(s, 10, 0)
			statusListCvt = append(statusListCvt, int(sCvt))
		}

		// param, _ := strconv.ParseInt(param, 10, 0)
		params.ExcludeStatusList = statusListCvt
	} else {
		params.ExcludeStatusList = []int{}
	}
	return nil
}
