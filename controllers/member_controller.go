package controllers

import (
	"context"
	"fmt"
	"log"
	"net/http"

	svc "cloudminds.com/harix/cc-server/services"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// MemberLoginRespose - as is
type MemberLoginRespose struct {
	Member svc.Member  `json:"member"`
	Family *svc.Family `json:"family"`
}

// MemberRegistrationResponse - as is
type MemberRegistrationResponse struct {
	MemberID string `json:"member_id"`
	RegCode  string `json:"reg_code"`
}

// GetManyMembers - as is
func (s *CCServer) GetManyMembers(c *gin.Context) {

	var queryParams svc.GetMemberParams
	queryParams.InstID = c.DefaultQuery("instID", "000000000000000000000000")

	cursor, err := svc.GetManyMembers(&queryParams)

	if err != nil {
		log.Printf("Error while getting all members - %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Something went wrong",
		})
		return
	}

	// Iterate through the returned cursor
	members := []svc.Member{}
	for cursor.Next(context.TODO()) {
		var member svc.Member
		cursor.Decode(&member)
		members = append(members, member)
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "All Members",
		"data":    members,
	})

	return
}

// CreateMember - as is
func (s *CCServer) CreateMember(c *gin.Context) {
	var mRegForm svc.MemberRegForm
	c.BindJSON(&mRegForm)
	// log.Printf("CreateMember - received Form: %v\n", mRegForm)

	// Validation
	err := s.Validator.v.Struct(mRegForm)
	if err != nil {
		var badInput bool = false
		for _, e := range err.(validator.ValidationErrors) {
			badInput = true
			log.Println(e)
			c.JSON(http.StatusBadRequest, gin.H{
				"message": fmt.Sprint(e.Translate(*s.Validator.trans)),
			})
		}
		if badInput {
			return
		}
	}

	// Check if the phone number exists
	count, err := svc.CountMembersByPhoneNum(mRegForm.PhoneNum)
	if err != nil {
		log.Printf("Error while finding Member by PhoneNum - %v\n", err)
	}
	if count > 0 {
		c.JSON(http.StatusForbidden, gin.H{
			"message": "Phone # has been used! Try another one",
		})
		return
	}

	_, ok := handleCreateMember(c, &mRegForm)

	if ok {
		c.JSON(http.StatusCreated, gin.H{
			"message": "Member registered Successfully",
		})
		return
	}

}

// ActivateMember - as is
func (s *CCServer) ActivateMember(c *gin.Context) {
	mActivateForm := svc.MemberActivateForm{}
	c.BindJSON(&mActivateForm)

	// Validation
	err := s.Validator.v.Struct(mActivateForm)
	if err != nil {
		var badInput bool = false
		for _, e := range err.(validator.ValidationErrors) {
			badInput = true
			log.Println(e)
			c.JSON(http.StatusBadRequest, gin.H{
				"message": fmt.Sprint(e.Translate(*s.Validator.trans)),
			})
		}
		if badInput {
			return
		}
	}

	// Get Member
	mToActivate := svc.Member{}
	err = svc.GetMemberByPhoneNum(mActivateForm.PhoneNum).Decode(&mToActivate)
	if err != nil {
		// When no RegCode found, return failed
		if err == mongo.ErrNoDocuments {
			c.JSON(http.StatusForbidden, gin.H{
				"message": "Member does not exist, Member Activation failed",
			})
			return
		}
		log.Printf("Error while getting Member giving PhoneNum - %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Something went wrong",
		})
		return
	}

	// Compare with RegCode in DB
	regCode := svc.RegCode{}
	// TODO - update func name
	err = svc.GetRegCodeByMemberID(mToActivate.ID.Hex()).Decode(&regCode)
	if err != nil {
		// When no RegCode found, return failed
		if err == mongo.ErrNoDocuments {
			c.JSON(http.StatusForbidden, gin.H{
				"message": "RegCode does not exist, Member Activation failed",
			})
			return
		}
		log.Printf("Error while getting regCode given MemberID - %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Something went wrong",
		})
		return
	}
	if mActivateForm.RegCode != regCode.RegCode {
		log.Printf("Cannot Activate Member - RegCode not Match, Try Another One")
		c.JSON(http.StatusForbidden, gin.H{
			"message": "Cannot Activate Member - RegCode not Match, Try Another One",
		})
		return
	}

	// Activate Member
	err = svc.ActivateMemberByID(mToActivate.ID.Hex())
	if err != nil {
		log.Printf("ActivateMember - Error while updating Member in DB - %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Something went wrong",
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"status":  http.StatusOK,
		"message": "Member Activated Successfully",
		"token":   s.Config.DebugTokenL.Mobile,
	})
	return
}

// LoginMember - as is
func (s *CCServer) LoginMember(c *gin.Context) {
	mLoginForm := svc.MemberLoginForm{}
	c.BindJSON(&mLoginForm)

	// Validation
	err := s.Validator.v.Struct(mLoginForm)
	if err != nil {
		var badInput bool = false
		for _, e := range err.(validator.ValidationErrors) {
			badInput = true
			log.Println(e)
			c.JSON(http.StatusBadRequest, gin.H{
				"message": fmt.Sprint(e.Translate(*s.Validator.trans)),
			})
		}
		if badInput {
			return
		}
	}

	// Get Member
	memberToLogin := svc.Member{}
	if len(mLoginForm.PhoneNum) > 0 {
		err = svc.GetMemberByPhoneNum(mLoginForm.PhoneNum).Decode(&memberToLogin)
	} else if len(mLoginForm.DeviceID) > 0 {
		//Check if DeviceID is available
	} else {
		log.Printf("Need More Info to Login Member")
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "Need More Info to Login Member",
		})
		return
	}
	//// When no Member found, return Failed
	if err != nil {
		if err == mongo.ErrNoDocuments {
			c.JSON(http.StatusForbidden, gin.H{
				"message": "Member does not Exist, Login failed",
			})
			return
		}
		log.Printf("Error while finding member from DB - %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Something went wrong",
		})
		return
	}

	// Check if Status is "Activated"
	if memberToLogin.Status != svc.MActivated {
		log.Printf("Member Not Activated, Login failed")
		c.JSON(http.StatusForbidden, gin.H{
			"message": "Member Not Activated, Login failed",
		})
		return
	}

	err = svc.UpdateMemberLoginTimeByID(memberToLogin.ID.Hex())
	if err != nil {
		log.Printf("Error while updating Member Login Time into DB - %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Something went wrong",
		})
		return
	}
	member := svc.Member{}
	svc.GetMemberByID(memberToLogin.ID.Hex()).Decode(&member)

	// return Member
	// TODO - Return Family if Member is Associated With One
	mLoginResponse := MemberLoginRespose{
		Member: member,
	}
	if member.FamilyInfo != nil {
		family := svc.Family{}
		svc.GetFamilyByID(member.FamilyInfo.ID).Decode(&family)
		mLoginResponse.Family = &family
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Member Login Succeed",
		"data":    mLoginResponse,
		"token":   s.Config.DebugTokenL.Mobile,
	})
	return
}

func (s *CCServer) CreateMemberAndSendSMS(c *gin.Context) {
	var mRegForm svc.MemberRegForm
	c.BindJSON(&mRegForm)

	// Validation
	err := s.Validator.v.Struct(mRegForm)
	if err != nil {
		var badInput bool = false
		for _, e := range err.(validator.ValidationErrors) {
			badInput = true
			log.Println(e)
			c.JSON(http.StatusBadRequest, gin.H{
				"message": fmt.Sprint(e.Translate(*s.Validator.trans)),
				"success": false,
			})
		}
		if badInput {
			return
		}
	}

	// Check if the phone number exists
	count, err := svc.CountMembersByPhoneNum(mRegForm.PhoneNum)
	if err != nil {
		log.Printf("Error while finding Member by PhoneNum - %v\n", err)
	}
	if count > 0 {
		c.JSON(http.StatusForbidden, gin.H{
			"message": "Phone # has been used! You can activate using the Registration Code in SMS",
			"success": false,
		})
		return
	}

	// Create Member in DB
	memberID, ok := handleCreateMember(c, &mRegForm)
	if !ok {
		return
	}

	// Get RegCode
	regCode := svc.RegCode{}
	err = svc.GetRegCodeByMemberID(memberID).Decode(&regCode)
	if err != nil {
		log.Printf("Error while getting regcode by ID - %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  http.StatusInternalServerError,
			"message": "Something went wrong",
			"success": false,
		})
		return
	}

	// Send SMS
	smsContent := RegCodeSMSContent{
		Name:    mRegForm.FirstName,
		RegCode: regCode.RegCode,
	}

	s.handleSendRegCodeWithSMS(smsContent, mRegForm.PhoneNum)

	// return MemberID and RegCode
	mRegResponse := MemberRegistrationResponse{
		MemberID: memberID,
		RegCode:  regCode.RegCode,
	}

	log.Println("SMS Sent!")
	c.JSON(http.StatusOK, gin.H{
		"data":    mRegResponse,
		"message": "SMS Sent!",
		"success": true,
	})
	s.sendRegCodePostProcessing(c, mRegForm.PhoneNum)
}

// UpdateMemberByID - as is
func (s *CCServer) UpdateMemberByID(c *gin.Context) {
	// perform Update
	idToUpdate := c.Param("id")
	var mForm svc.MemberEditForm
	c.BindJSON(&mForm)

	// Validation
	err := s.Validator.v.Struct(mForm)
	if err != nil {
		var badInput bool = false
		for _, e := range err.(validator.ValidationErrors) {
			badInput = true
			log.Println(e)
			c.JSON(http.StatusBadRequest, gin.H{
				"message": fmt.Sprint(e.Translate(*s.Validator.trans)),
			})
		}
		if badInput {
			return
		}
	}

	_, err = svc.UpdateMemberByID(mForm, idToUpdate)

	if err != nil {
		log.Printf("Error while updating Member in DB - %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Something went wrong",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Member updated Successfully",
	})

	// Update Display Names in CCRecords
	// TODO - update struct & func name
	// mInfo := svc.MemberInfo{
	// 	Name:     mForm.FirstName + " " + mForm.LastName,
	// 	PhoneNum: mForm.PhoneNum,
	// }
	// _, err = svc.UpdateManyCCRecordsGuardianInfoByGuardianID(idToUpdate, mInfo)
	// if err != nil {
	// 	log.Printf("Error when Updating Many CCRecords by Member ID - %v\n", err)
	// }
	return
}

// DeleteMemberByID - as is
func (s *CCServer) DeleteMemberByID(c *gin.Context) {
	// Get Member
	idToDelete := c.Param("id")
	_, err := svc.DeleteMemberByID(idToDelete)
	if err != nil {
		log.Printf("Error while deleting Member in DB - %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Something went wrong",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Member deleted Successfully",
	})

	// Set CC-Records to Expire
	// TODO - Update field name
	// ccrParams := svc.MarkCCRecordAsExpiredParams{
	// 	GuardianID: idToDelete,
	// }
	// _, err = svc.MarkCCRecordAsExpired(ccrParams)
	// // If any error, log it in server, since it is not fatal
	// if err != nil {
	// 	log.Printf("Error while marking CCRecord as Expired - %v\n", err)
	// }

	// Delete RegCode
	_, err = svc.DeleteRegCodeByMemberID(idToDelete)
	// If any error, log it in server, since it is not fatal
	if err != nil {
		log.Printf("Error while Deleting RegCode by GuardianID - %v\n", err)
	}

	return
}

// handleCreateMember - return (memberID, ok)
func handleCreateMember(c *gin.Context, mRegForm *svc.MemberRegForm) (string, bool) {

	// Create Member
	res, err := svc.CreateMember(*mRegForm)
	if err != nil {
		log.Printf("Error while inserting new Member into DB - %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Something went wrong",
		})
		return "", false
	}
	memberID := res.InsertedID.(primitive.ObjectID).Hex()

	// Register RegCode for the Guardian in DB
	// TODO - update func name
	res, err = svc.CreateRegCodeByMemberID(memberID)
	if err != nil {
		log.Printf("Error while Creating new RegCode into DB - %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Something went wrong",
		})
		return "", false
	}

	return memberID, true
}
