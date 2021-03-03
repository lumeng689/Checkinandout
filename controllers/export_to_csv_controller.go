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
	offsetHourRaw := c.DefaultQuery("hourOffset", "-8")
	offsetHours, _ := strconv.ParseInt(offsetHourRaw, 10, 0)

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
		var recordName, recordGroup, recordPhoneNum string
		var recordTime time.Time
		if inst.MemberType == svc.MemberTypeStandard || inst.MemberType == svc.MemberTypeTag {
			if ccRecord.MT == nil {
				continue
			}
			recordName = ccRecord.MT.Info.Name
			recordGroup = ccRecord.MT.Info.Group
			recordPhoneNum = ccRecord.MT.Info.PhoneNum
			recordTime = ccRecord.MT.CheckInEvent.Time
		} else if inst.MemberType == svc.MemberTypeGuardian {
			if ccRecord.GW == nil {
				continue
			}
			recordName = ccRecord.GW.WardInfo.Name
			recordGroup = ccRecord.GW.WardInfo.Group
			recordPhoneNum = ccRecord.GW.CheckInEvent.GuardianInfo.PhoneNum
			recordTime = ccRecord.GW.CheckInEvent.Time
		}
		var record []string
		record = append(record, recordName)
		record = append(record, recordGroup)
		// record = append(record, ccRecord.CheckInEvent.GuardianInfo.Name)
		record = append(record, strconv.FormatFloat(float64(ccRecord.Temperature), 'f', 1, 64))
		record = append(record, recordPhoneNum)
		record = append(record, recordTime.In(time.FixedZone("BROWSER", int(offsetHours)*60*60)).Format("01/02/2006 03:04:05PM"))
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

// ExportManyMembers - as is
func (s *CCServer) ExportManyMembers(c *gin.Context) {

	instID := c.DefaultQuery("instID", "000000000000000000000000")
	// Get Institution
	inst := svc.Institution{}
	err := svc.GetInstByID(instID).Decode(&inst)

	params := svc.GetMemberParams{
		InstID: instID,
	}

	cursor, err := svc.GetManyMembers(&params)

	if err != nil {
		log.Printf("Error while getting all CCEvents - %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Something went wrong",
		})
		return
	}

	members := []svc.Member{}
	for cursor.Next(context.TODO()) {
		var member svc.Member
		cursor.Decode(&member)
		members = append(members, member)
	}

	// export
	b := &bytes.Buffer{}
	w := csv.NewWriter(b)

	firstLine := []string{"Member Name", "Group", "Phone #"}
	// firstLine := []string{"Ward Name", "Group", "Guardian Name", "Temperature", "Phone #", "Drop Off At", "Scheduled Pickup At", "Actual Pickup At"}
	if err := w.Write(firstLine); err != nil {
		log.Printf("eoor writing record to csv - %v\n", err)
	}

	for _, member := range members {
		var record []string
		record = append(record, member.FirstName+" "+member.LastName)
		record = append(record, member.Group)
		record = append(record, member.PhoneNum)

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

// ExportManyFamilies - as is
func (s *CCServer) ExportManyFamilies(c *gin.Context) {

	var families *[]svc.Family
	if families = getFamilies(c); families == nil {
		log.Println("Error when getting Failies")
		return
	}
	// export
	b := &bytes.Buffer{}
	w := csv.NewWriter(b)

	firstLine := []string{"Family Contact", "Phone #", "Relation"}
	if err := w.Write(firstLine); err != nil {
		log.Printf("error writing record to csv - %v\n", err)
	}

	for _, family := range *families {

		contactGuardian := family.ContactMemberInfo
		var record []string
		record = append(record, contactGuardian.Name)
		record = append(record, contactGuardian.PhoneNum)
		record = append(record, contactGuardian.Relation)
		if err := w.Write(record); err != nil {
			log.Printf("error writing family to csv - %v\n", err)
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

//ExportManyWards - as is
func (s *CCServer) ExportManyWards(c *gin.Context) {
	var families *[]svc.Family
	if families = getFamilies(c); families == nil {
		log.Println("Error when getting Failies")
		return
	}

	// collecting Wards
	var wards []svc.Ward
	for _, family := range *families {
		wards = append(wards, family.Wards...)
	}

	// export
	b := &bytes.Buffer{}
	w := csv.NewWriter(b)

	firstLine := []string{"Name", "Group"}
	if err := w.Write(firstLine); err != nil {
		log.Printf("error writing record to csv - %v\n", err)
	}

	for _, ward := range wards {

		var record []string
		record = append(record, ward.FirstName+" "+ward.LastName)
		record = append(record, ward.Group)
		if err := w.Write(record); err != nil {
			log.Printf("error writing wards to csv - %v\n", err)
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

// ExportManySurveys = as is
func (s *CCServer) ExportManySurveys(c *gin.Context) {
	var queryParams svc.GetSurveyParams
	queryParams.InstID = c.DefaultQuery("instID", "000000000000000000000000")
	offsetHourRaw := c.DefaultQuery("hourOffset", "-8")
	offsetHours, _ := strconv.ParseInt(offsetHourRaw, 10, 0)

	cursor, err := svc.GetManySurveys(&queryParams)

	if err != nil {
		log.Printf("Error while getting all Surveys - %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Something went wrong",
		})
		return
	}
	surveys := []svc.Survey{}
	if err = cursor.All(context.TODO(), &surveys); err != nil {
		panic(err)
	}
	if len(surveys) == 0 {
		c.JSON(http.StatusOK, gin.H{
			"message": "No surveys found",
		})
	}

	//export
	b := &bytes.Buffer{}
	w := csv.NewWriter(b)
	var firstLine []string
	firstLine = append(firstLine, "Guardian Name")
	firstLine = append(firstLine, "Phone #")
	firstLine = append(firstLine, "Submitted At")
	for _, qa := range surveys[0].QAList {
		firstLine = append(firstLine, qa.QuestionIndex+" "+qa.Question)
	}
	if err = w.Write(firstLine); err != nil {
		log.Printf("error writing surveys to csv - %v\n", err)
	}

	for _, survey := range surveys {
		var record []string
		member := svc.Member{}
		err := svc.GetMemberByID(survey.MemberID).Decode(&member)
		if err != nil {
			record = append(record, []string{"", "", ""}...)
		} else {
			memberName := member.FirstName + " " + member.LastName
			record = append(record, memberName)
			record = append(record, member.PhoneNum)
			record = append(record, survey.CreatedAt.In(time.FixedZone("BROWSER", int(offsetHours)*60*60)).Format("Jan 2 2006 03:04:05PM"))
			// record = append(record, []string{memberName, member.PhoneNum, survey.CreatedAt.In(time.Now().Location()).Format("Jan 2 2006 03:04:05PM")}...)
		}
		// Export Survey Answers
		for _, qa := range survey.QAList {
			record = append(record, getSurveyAnswers(qa))
		}
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

func getFamilies(c *gin.Context) *[]svc.Family {
	var queryParams svc.GetFamilyParams
	err := extractFamilyParams(c, &queryParams)

	if err != nil {
		log.Printf("Error while parsing Family Query Parameters - %v\n", err)
		c.JSON(http.StatusOK, gin.H{
			"status":  http.StatusOK,
			"message": "Bad CCEvents Query Parameters",
		})
		return nil
	}
	log.Printf("families query params - %v\n", queryParams)

	cursor, err := svc.GetManyFamilies(&queryParams)
	if err != nil {
		log.Printf("Error while getting all families - %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  http.StatusInternalServerError,
			"message": "Something went wrong",
		})
		return nil
	}
	families := []svc.Family{}
	if err = cursor.All(context.TODO(), &families); err != nil {
		panic(err)
	}
	return &families
}

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

func extractFamilyParams(c *gin.Context, params *svc.GetFamilyParams) error {
	params.InstID = c.DefaultQuery("instID", "000000000000000000000000")

	return nil
}
