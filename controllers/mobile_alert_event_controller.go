package controllers

import (
	"log"
	"net/http"

	svc "cloudminds.com/harix/cc-server/services"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
)

// GetScanNameByDeviceID - Used by MobileAlert. Given DeviceIMEI (Gatekeeper), return the most recent Customer Name
func (s *CCServer) GetScanNameByDeviceID(c *gin.Context) {

	// TODO - Check By MemberType of the Affliated Institution

	ccParams := svc.GetCCRecordParams{
		GetLatest: true,
		DeviceID:  c.DefaultQuery("deviceID", "0000000000000000"),
		Status:    -1, // set to -1 to disable status filter
	}
	var ccRecord svc.CCRecord
	if err := svc.GetCCRecordByDeviceID(&ccParams, svc.MemberTypeGuardian).Decode(&ccRecord); err != nil {
		if err != mongo.ErrNoDocuments {
			log.Printf("Error while getting CCRecord by DeviceID - %v\n", err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"message": "Something went wrong",
			})
			return
		}
		if err := svc.GetCCRecordByDeviceID(&ccParams, svc.MemberTypeStandard).Decode(&ccRecord); err != nil {
			if err != mongo.ErrNoDocuments {
				log.Printf("Error while getting CCRecord by DeviceID - %v\n", err)
				c.JSON(http.StatusInternalServerError, gin.H{
					"message": "Something went wrong",
				})
				return
			}
			log.Printf("CCRecord with DeviceID not found, Getting CCRecord Failed!")
			c.JSON(http.StatusOK, gin.H{
				"data":    "",
				"time":    -1,
				"message": "CCRecord not found, Getting CCRecords Failed!",
			})
			return
		}
	}

	var name string
	var timestamp int
	if ccRecord.MT != nil {
		name = ccRecord.MT.Info.Name
		timestamp = int(ccRecord.MT.CheckInEvent.Time.Unix())
	} else if ccRecord.GW != nil {
		name = ccRecord.GW.CheckInEvent.GuardianInfo.Name
		timestamp = int(ccRecord.GW.CheckInEvent.Time.Unix())
	}

	c.JSON(http.StatusOK, gin.H{
		"data":    name,
		"time":    timestamp,
		"message": "Guest Name",
	})
}
