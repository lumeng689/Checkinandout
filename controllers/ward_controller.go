package controllers

import (
	"log"
	"net/http"

	svc "cloudminds.com/harix/cc-server/services"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
)

// AddWard add a new Ward within a Family.
func (s *CCServer) AddWard(c *gin.Context) {
	// retrieve Family
	var queryParams svc.AddWardParams
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
	var wAddForm svc.WardForm
	c.BindJSON(&wAddForm)
	newWard := svc.GetNewWard(wAddForm)

	// Append New Ward to Family and Update in DB
	wards := append(familyToAppend.Wards, newWard)
	_, err = svc.ReplaceFamily(familyToAppend, familyToAppend.Guardians, wards, familyToAppend.Vehicles)
	if err != nil {
		log.Printf("Error while adding Ward in DB - %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Something went wrong",
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Ward added Successfully",
	})
	return
}

// UpdateWardByID - as is
func (s *CCServer) UpdateWardByID(c *gin.Context) {
	// Get Family
	idToUpdate := c.Param("id")
	familyToUpdate := svc.Family{}
	err := svc.GetFamilyByWardID(idToUpdate).Decode(&familyToUpdate)
	if err != nil {
		// When no family found, return failed
		if err == mongo.ErrNoDocuments {
			c.JSON(http.StatusForbidden, gin.H{
				"message": "Family does not Exist, Ward Update failed",
			})
			return
		}
		log.Printf("Error while Getting Family By WardID - %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Something went wrong",
		})
		return
	}

	// Replace Ward In Family
	var wForm svc.WardForm
	c.BindJSON(&wForm)
	wards := familyToUpdate.Wards
	for index, prevW := range wards {
		// replace ward if ID matches
		if idToUpdate == prevW.ID.Hex() {
			ward := svc.Ward{
				ID:        prevW.ID,
				FirstName: wForm.FirstName,
				LastName:  wForm.LastName,
				Group:     wForm.Group,
			}
			wards = append(wards[:index], wards[index+1:]...)
			wards = append(wards, ward)
			break
		}
	}

	// Save updated family to DB
	_, err = svc.ReplaceFamily(familyToUpdate, familyToUpdate.Guardians, wards, familyToUpdate.Vehicles)
	if err != nil {
		log.Printf("Error while updating Ward in DB - %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Something went wrong",
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "Ward updated Successfully",
	})

	// Update Display Names in CCRecords
	wInfo := svc.WardInfo{
		Name:  wForm.FirstName + " " + wForm.LastName,
		Group: wForm.Group,
	}
	_, err = svc.UpdateManyCCRecordsWardInfoByWardID(idToUpdate, wInfo)
	if err != nil {
		log.Printf("Error when Updating Many CCRecords by Ward ID - %v\n", err)
	}
	return
}

// DeleteWardByID - as is
func (s *CCServer) DeleteWardByID(c *gin.Context) {
	// Get Family
	idToDelete := c.Param("id")
	familyToUpdate := svc.Family{}
	err := svc.GetFamilyByWardID(idToDelete).Decode(&familyToUpdate)
	if err != nil {
		// When no family found, return failed
		if err == mongo.ErrNoDocuments {
			c.JSON(http.StatusForbidden, gin.H{
				"message": "Family does not Exist, Ward Deleting failed",
			})
			return
		}
		log.Printf("Error while Getting Family By WardID - %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Something went wrong",
		})
		return
	}

	// Remove Ward from Family
	wards := familyToUpdate.Wards
	for index, prevW := range wards {
		// replace ward if ID matches
		if idToDelete == prevW.ID.Hex() {
			wards = append(wards[:index], wards[index+1:]...)
			break
		}
	}
	// Save updated family to DB
	_, err = svc.ReplaceFamily(familyToUpdate, familyToUpdate.Guardians, wards, familyToUpdate.Vehicles)

	if err != nil {
		log.Printf("Error while deleting Ward in DB - %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Something went wrong",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Ward deleted Successfully",
	})

	// Set CC-Records to Expire
	ccrParams := svc.MarkCCRecordAsExpiredParams{
		WardID: idToDelete,
	}
	_, err = svc.MarkCCRecordAsExpired(ccrParams)
	// If any error, log it in server, since it is not fatal
	if err != nil {
		log.Printf("Error while marking CCRecord as Expired - %v\n", err)
	}

	return
}

func getWardInFamilyByID(f svc.Family, id string) *svc.Ward {
	for _, w := range f.Wards {
		if w.ID.Hex() == id {
			return &w
		}
	}
	return nil
}
