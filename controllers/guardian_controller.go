package controllers

import (
	"log"
	"net/http"

	svc "cloudminds.com/harix/cc-server/services"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
)

// AddGuardian register a new Guardian within a Family. Guardian's PhoneNum is REQUIRED
func (s *CCServer) AddGuardian(c *gin.Context) {
	// retrieve Family
	var queryParams svc.AddGuardianParams
	queryParams.FamilyID = c.DefaultQuery("familyID", "000000000000000000000000")
	familyToAppend := svc.Family{}
	err := svc.GetFamilyByID(queryParams.FamilyID).Decode(&familyToAppend)
	if err != nil {
		// When no family found, return failed
		if err == mongo.ErrNoDocuments {
			c.JSON(http.StatusForbidden, gin.H{
				"message": "Family does not exist, Adding Guardian failed",
			})
			return
		}
		log.Printf("Error while getting Family from DB - %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Something went wrong",
		})
		return
	}

	// Assign new ID to the guardian
	var gAddForm svc.GuardianAddForm
	c.BindJSON(&gAddForm)
	newGuardian := svc.GetNewGuardian(gAddForm)
	// register RegCode for the Guardian in DB
	_, err = svc.CreateRegCodeByGuardianID(newGuardian.ID.Hex())
	if err != nil {
		log.Printf("Error while Creating new RegCode into DB - %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Something went wrong",
		})
		return
	}

	// Append New Guardian to Family and Update in DB
	guardians := append(familyToAppend.Guardians, newGuardian)
	_, err = svc.ReplaceFamily(familyToAppend, guardians, familyToAppend.Wards, familyToAppend.Vehicles)
	if err != nil {
		log.Printf("Error while adding Guardian in DB - %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Something went wrong",
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Guardian added Successfully",
	})
	return
}

// ActivateGuardian - compare RegCode in DB and Request; If matches, Activate Guardian and Update Family in DB
func (s *CCServer) ActivateGuardian(c *gin.Context) {
	gActivateForm := svc.GuardianActivateForm{}
	c.BindJSON(&gActivateForm)

	// Get Family
	familyToUpdate := svc.Family{}
	err := svc.GetFamilyByPhoneNum(gActivateForm.PhoneNum).Decode(&familyToUpdate)
	if err != nil {
		// When no family found, return failed
		if err == mongo.ErrNoDocuments {
			c.JSON(http.StatusForbidden, gin.H{
				"message": "Family does not exist, Guardian Activation failed",
			})
			return
		}
		log.Printf("Error while getting Family given PhoneNum - %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Something went wrong",
		})
		return
	}

	// Get Guardian
	gToActivate := s.GetGuardianInFamilyByPhoneNum(familyToUpdate, gActivateForm.PhoneNum)

	// Compare with RegCode in DB
	regCode := svc.RegCode{}
	err = svc.GetRegCodeByGuardianID(gToActivate.ID.Hex()).Decode(&regCode)
	log.Printf(" getting regCode - GuardianID: %v\n", gToActivate.ID.Hex())
	if err != nil {
		// When no RegCode found, return failed
		if err == mongo.ErrNoDocuments {
			c.JSON(http.StatusForbidden, gin.H{
				"message": "RegCode does not exist, Guardian Activation failed",
			})
			return
		}
		log.Printf("Error while getting regCode given GuardianID - %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Something went wrong",
		})
		return
	}
	if gActivateForm.RegCode != regCode.RegCode {
		log.Printf("Cannot Activate Guardian - RegCode not Match, Try Another One")
		c.JSON(http.StatusForbidden, gin.H{
			"message": "Cannot Activate Guardian - RegCode not Match, Try Another One",
		})
		return
	}

	// Activate Gaurdian
	gToActivate.Status = svc.GActivated
	ReplaceGuardianInFamily(&familyToUpdate, *gToActivate)
	log.Printf("gToActivate: %v\n", gToActivate)
	log.Printf("familyToUpdate: %v\n", familyToUpdate)
	_, err = svc.ReplaceFamily(familyToUpdate, familyToUpdate.Guardians, familyToUpdate.Wards, familyToUpdate.Vehicles)

	if err != nil {
		log.Printf("ActivateGuardian - Error while updating Family in DB - %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Something went wrong",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  http.StatusOK,
		"message": "Guardian Activated Successfully",
	})
	return

}

// LoginGuardian - Login using PhoneNum or DeviceID(TODO),  return Family and GuardianID if success
func (s *CCServer) LoginGuardian(c *gin.Context) {
	var err error
	gLoginForm := svc.GuardianLoginForm{}
	c.BindJSON(&gLoginForm)

	familyToLogin := svc.Family{}
	var guardianToLogin *svc.Guardian

	// Check if PhoneNum is available - CAUTION: PhoneNumber MUST be Unique, or RegCode will NOT Found
	if len(gLoginForm.PhoneNum) > 0 {
		err = svc.GetFamilyByPhoneNum(gLoginForm.PhoneNum).Decode(&familyToLogin)
		guardianToLogin = s.GetGuardianInFamilyByPhoneNum(familyToLogin, gLoginForm.PhoneNum)
	} else if len(gLoginForm.DeviceID) > 0 {
		// Check if DeviceID is available

	} else {
		log.Printf("Need More Info to Login Guardian")
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  http.StatusBadRequest,
			"message": "Need More Info to Login Guardian",
		})
		return
	}

	if err != nil {
		// When no family found, return failed
		if err == mongo.ErrNoDocuments {
			c.JSON(http.StatusForbidden, gin.H{
				"message": "Family does not Exist, Guardian Login failed",
			})
			return
		}
		log.Printf("Error while finding family from DB - %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Something went wrong",
		})
		return
	}

	// Check if Status is "Activated"
	if guardianToLogin.Status != svc.GActivated {
		log.Printf("Guardian Not Activated, Guardian Login failed")
		c.JSON(http.StatusForbidden, gin.H{
			"message": "Guardian Not Activated, Guardian Login failed",
		})
		return
	}

	family := svc.Family{}
	err = svc.UpdateGuardianLoginTime(familyToLogin.ID.Hex(),
		guardianToLogin.ID.Hex(), &family)

	if err != nil {
		log.Printf("Error while updating Guardian Login Time into DB - %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Something went wrong",
		})
		return
	}

	// return Family & "GuardianID" as Login Response
	c.JSON(http.StatusOK, gin.H{
		"message": "Single Family",
		"data":    family,
		"id":      guardianToLogin.ID.Hex(),
	})
	return
}

// UpdateGuardianByID - as is
func (s *CCServer) UpdateGuardianByID(c *gin.Context) {
	// Get Family
	idToUpdate := c.Param("id")
	familyToUpdate := svc.Family{}
	err := svc.GetFamilyByGuardianID(idToUpdate).Decode(&familyToUpdate)
	if err != nil {
		// When no family found, return failed
		if err == mongo.ErrNoDocuments {
			c.JSON(http.StatusForbidden, gin.H{
				"message": "Family does not Exist, Guardian Update failed",
			})
			return
		}
		log.Printf("Error while Getting Family By GuardianID - %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Something went wrong",
		})
		return
	}

	// Replace Guardian In Family
	var gForm svc.GuardianEditForm
	c.BindJSON(&gForm)
	guardians := familyToUpdate.Guardians
	for index, prevG := range guardians {
		// replace guardian if ID matches
		if idToUpdate == prevG.ID.Hex() {
			guardian := svc.Guardian{
				ID:          prevG.ID,
				PhoneNum:    gForm.PhoneNum,
				Email:       gForm.Email,
				DeviceID:    gForm.DeviceID,
				FirstName:   gForm.FirstName,
				LastName:    gForm.LastName,
				Relation:    gForm.Relation,
				LastLoginAt: prevG.LastLoginAt,
				Status:      prevG.Status,
			}
			guardians = append(guardians[:index], guardians[index+1:]...)
			guardians = append(guardians, guardian)
			break
		}
	}

	// Save updated family to DB
	_, err = svc.ReplaceFamily(familyToUpdate, guardians, familyToUpdate.Wards, familyToUpdate.Vehicles)
	if err != nil {
		log.Printf("Error while updating Guardian in DB - %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Something went wrong",
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "Guardian updated Successfully",
	})

	// Update Display Names in CCRecords
	gInfo := svc.GuardianInfo{
		Name:     gForm.FirstName + " " + gForm.LastName,
		PhoneNum: gForm.PhoneNum,
	}
	_, err = svc.UpdateManyCCRecordsGuardianInfoByGuardianID(idToUpdate, gInfo)
	if err != nil {
		log.Printf("Error when Updating Many CCRecords by Guardian ID - %v\n", err)
	}
	return
}

// DeleteGuardianByID - as is
func (s *CCServer) DeleteGuardianByID(c *gin.Context) {
	// Get Family
	idToDelete := c.Param("id")
	familyToUpdate := svc.Family{}
	err := svc.GetFamilyByGuardianID(idToDelete).Decode(&familyToUpdate)
	if err != nil {
		// When no family found, return failed
		if err == mongo.ErrNoDocuments {
			c.JSON(http.StatusForbidden, gin.H{
				"message": "Family does not Exist, Guardian Deleting failed",
			})
			return
		}
		log.Printf("Error while Deleting Family By GuardianID - %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Something went wrong",
		})
		return
	}

	// Remove Guardian from Family, and Re-assign Family Contact
	guardians := familyToUpdate.Guardians
	// abort deletion if the family has only one guardian
	if len(guardians) < 2 {
		c.JSON(http.StatusForbidden, gin.H{
			"message": "Guardian Deleting Failed! - Family must have at least one guardian",
		})
		return
	}
	var hasContactIDReset = false
	for index, prevG := range guardians {
		// remove guardian if ID matches
		if idToDelete == prevG.ID.Hex() {
			guardians = append(guardians[:index], guardians[index+1:]...)
			if hasContactIDReset {
				break
			}
		} else if !hasContactIDReset {
			// set Family ContactID to another guardian
			familyToUpdate.ContactID = prevG.ID.Hex()
			hasContactIDReset = true
		}
	}
	// Save updated family to DB
	_, err = svc.ReplaceFamily(familyToUpdate, guardians, familyToUpdate.Wards, familyToUpdate.Vehicles)

	if err != nil {
		log.Printf("Error while deleting Guardian in DB - %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Something went wrong",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Guardian deleted Successfully",
	})

	// Set CC-Records to Expire
	ccrParams := svc.MarkCCRecordAsExpiredParams{
		GuardianID: idToDelete,
	}
	_, err = svc.MarkCCRecordAsExpired(ccrParams)
	// If any error, log it in server, since it is not fatal
	if err != nil {
		log.Printf("Error while marking CCRecord as Expired - %v\n", err)
	}

	// Delete RegCode
	_, err = svc.DeleteRegCodeByGuardianID(idToDelete)
	// If any error, log it in server, since it is not fatal
	if err != nil {
		log.Printf("Error while Deleting RegCode by GuardianID - %v\n", err)
	}

	return
}

// GetGuardianInFamilyByPhoneNum - as is
func (s *CCServer) GetGuardianInFamilyByPhoneNum(f svc.Family, phoneNum string) *svc.Guardian {
	for _, g := range f.Guardians {
		if g.PhoneNum == phoneNum {
			return &g
		}
	}
	return nil
}

func getGuardianInFamilyByID(f svc.Family, guardianID string) *svc.Guardian {
	for _, g := range f.Guardians {
		if g.ID.Hex() == guardianID {
			return &g
		}
	}
	return nil
}

// ReplaceGuardianInFamily - as is
func ReplaceGuardianInFamily(f *svc.Family, gToReplace svc.Guardian) {
	var guardians []svc.Guardian
	for index, g := range f.Guardians {
		if g.ID.Hex() == gToReplace.ID.Hex() {
			guardians = append(f.Guardians[:index], f.Guardians[index+1:]...)
			guardians = append(guardians, gToReplace)
			break
		}
	}
	f.Guardians = guardians
}
