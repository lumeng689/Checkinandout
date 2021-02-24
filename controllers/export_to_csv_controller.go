package controllers

import (
	"bytes"
	"context"
	"encoding/csv"
	"log"
	"math"
	"net/http"
	"strconv"
	"time"

	svc "cloudminds.com/harix/cc-server/services"
	"github.com/gin-gonic/gin"
)

// ExportManyCCRecords - as is
func (s *CCServer) ExportManyCCRecords(c *gin.Context) {

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

	cursor, err := svc.GetManyCCRecords(&queryParams, inst.MemberType)

	if err != nil {
		log.Printf("Error while getting all CCEvents - %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{
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

	// export
	b := &bytes.Buffer{}
	w := csv.NewWriter(b)

	firstLine := []string{"Member Name", "Group", "Temperature", "Phone #", "Drop Off At"}
	// firstLine := []string{"Ward Name", "Group", "Guardian Name", "Temperature", "Phone #", "Drop Off At", "Scheduled Pickup At", "Actual Pickup At"}
	if err := w.Write(firstLine); err != nil {
		log.Printf("eoor writing record to csv - %v\n", err)
	}

	for _, ccRecord := range ccRecords {
		var record []string
		record = append(record, ccRecord.MT.Info.Name)
		record = append(record, ccRecord.MT.Info.Group)
		// record = append(record, ccRecord.CheckInEvent.GuardianInfo.Name)
		record = append(record, strconv.FormatFloat(float64(ccRecord.Temperature), 'f', 1, 64))
		record = append(record, ccRecord.MT.Info.PhoneNum)
		record = append(record, ccRecord.MT.CheckInEvent.Time.In(time.Now().Location()).Format("01/02/2006 03:04:05PM"))
		// record = append(record, ccRecord.CheckOutScheduledAt.In(time.Now().Location()).Format("01/02/2006 03:04:05PM"))
		// record = append(record, ccRecord.CheckOutEvent.Time.In(time.Now().Location()).Format("01/02/2006 03:04:05PM"))
		if err := w.Write(record); err != nil {
			log.Printf("error writing record to csv - %v\n", err)
		}
	}
	w.Flush()

	if err := w.Error(); err != nil {
		log.Printf("error while flushing writer- %v\n", err)
	}

	c.Header("Content-Description", "File Transfer")
	c.Header("Content-Disposition", "attachment; filename=contacts.csv")
	c.Data(http.StatusOK, "text/csv", b.Bytes())
}

// ExportManyWards - as is
// func (s *CCServer) ExportManyWards(c *gin.Context) {
// 	var families *[]svc.Family
// 	if families = getFamilies(c); families == nil {
// 		log.Println("Error when getting Failies")
// 		return
// 	}

// 	// collecting Wards
// 	var wards []svc.Ward
// 	for _, family := range *families {
// 		wards = append(wards, family.Wards...)
// 	}

// 	// export
// 	b := &bytes.Buffer{}
// 	w := csv.NewWriter(b)

// 	firstLine := []string{"Name", "Group"}
// 	if err := w.Write(firstLine); err != nil {
// 		log.Printf("error writing record to csv - %v\n", err)
// 	}

// 	for _, ward := range wards {

// 		var record []string
// 		record = append(record, ward.FirstName+" "+ward.LastName)
// 		record = append(record, ward.Group)
// 		if err := w.Write(record); err != nil {
// 			log.Printf("error writing wards to csv - %v\n", err)
// 		}
// 	}
// 	w.Flush()

// 	if err := w.Error(); err != nil {
// 		log.Printf("error while flushing writer- %v\n", err)
// 	}

// 	c.Header("Content-Description", "File Transfer")
// 	c.Header("Content-Disposition", "attachment; filename=contacts.csv")
// 	c.Data(http.StatusOK, "text/csv", b.Bytes())
// }

// ExportManyFamilies - as is
func (s *CCServer) ExportManyFamilies(c *gin.Context) {

	// var families *[]svc.Family
	// if families = getFamilies(c); families == nil {
	// 	log.Println("Error when getting Failies")
	// 	return
	// }
	// // export
	// b := &bytes.Buffer{}
	// w := csv.NewWriter(b)

	// firstLine := []string{"Family Contact", "Phone #", "Email", "Relation"}
	// if err := w.Write(firstLine); err != nil {
	// 	log.Printf("error writing record to csv - %v\n", err)
	// }

	// for _, family := range *families {

	// 	contactGuardian := getContactGuardianInFamily(family)
	// 	var record []string
	// 	record = append(record, contactGuardian.FirstName+" "+contactGuardian.LastName)
	// 	record = append(record, contactGuardian.PhoneNum)
	// 	record = append(record, contactGuardian.Email)
	// 	record = append(record, contactGuardian.Relation)
	// 	if err := w.Write(record); err != nil {
	// 		log.Printf("error writing family to csv - %v\n", err)
	// 	}
	// }
	// w.Flush()

	// if err := w.Error(); err != nil {
	// 	log.Printf("error while flushing writer- %v\n", err)
	// }

	// c.Header("Content-Description", "File Transfer")
	// c.Header("Content-Disposition", "attachment; filename=contacts.csv")
	// c.Data(http.StatusOK, "text/csv", b.Bytes())
}

// ExportManySurveys = as is
func (s *CCServer) ExportManySurveys(c *gin.Context) {

}

// func getFamilies(c *gin.Context) *[]svc.Family {
// 	var queryParams svc.GetFamilyParams
// 	err := extractFamilyParams(c, &queryParams)

// 	if err != nil {
// 		log.Printf("Error while parsing Family Query Parameters - %v\n", err)
// 		c.JSON(http.StatusOK, gin.H{
// 			"status":  http.StatusOK,
// 			"message": "Bad CCEvents Query Parameters",
// 		})
// 		return nil
// 	}
// 	log.Printf("families query params - %v\n", queryParams)

// 	cursor, err := svc.GetManyFamilies(&queryParams)
// 	if err != nil {
// 		log.Printf("Error while getting all families - %v\n", err)
// 		c.JSON(http.StatusInternalServerError, gin.H{
// 			"status":  http.StatusInternalServerError,
// 			"message": "Something went wrong",
// 		})
// 		return nil
// 	}
// 	families := []svc.Family{}
// 	if err = cursor.All(context.TODO(), &families); err != nil {
// 		panic(err)
// 	}
// 	return &families
// }

func getSurveyAnswers(qa svc.QuestionAnswer) string {
	// If Text Answer is non-empty, return TextAnswer
	if len(qa.AnswerText) > 0 {
		return qa.AnswerText
	}
	// if Numerical Answer is Empty, return Boolean
	if math.Abs(qa.AnswerNumerical-0) < 1e-10 {
		return mapAnswerBool(qa.AnswerBool)
	}
	return strconv.FormatFloat(qa.AnswerNumerical, 'f', 1, 64)
}

func mapAnswerBool(ans bool) string {
	if ans {
		return "YES"
	}
	return "NO"
}
