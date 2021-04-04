package controllers

import (
	"log"
	"net/http"
	"time"

	svc "cloudminds.com/harix/cc-server/services"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
)

var surveyBaseAddr = "/surveys/"

// GetOrCreateManyCCRecords - For "syncing" on Mobile CC App
func (s *CCServer) GetOrCreateManyCCRecords(c *gin.Context) {
	var CCRecordsForm svc.CCSyncPostingForm
	c.BindJSON(&CCRecordsForm)
	log.Printf("GetOrCreateManyCCRecords - CCRecordForm is - %v\n", CCRecordsForm)

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
	if CCRecordsForm.WardIDList != nil {
		var ccRecords []svc.CCRecord
		for _, wardID := range *CCRecordsForm.WardIDList {

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
