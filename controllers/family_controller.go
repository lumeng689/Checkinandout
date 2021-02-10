package controllers

import (
	"context"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"

	svc "cloudminds.com/harix/cc-server/services"
)

// GetManyFamilies - as is
func (s *CCServer) GetManyFamilies(c *gin.Context) {
	var queryParams svc.GetFamilyParams
	err := extractFamilyParams(c, &queryParams)

	if err != nil {
		log.Printf("Error while parsing Family Query Parameters - %v\n", err)
		c.JSON(http.StatusOK, gin.H{
			"status":  http.StatusOK,
			"message": "Bad CCEvents Query Parameters",
		})
		return
	}
	log.Printf("families query params - %v\n", queryParams)

	cursor, err := svc.GetManyFamilies(&queryParams)

	if err != nil {
		log.Printf("Error while getting all families - %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  http.StatusInternalServerError,
			"message": "Something went wrong",
		})
		return
	}
	// Iterate through the returned cursor
	families := []svc.Family{}
	for cursor.Next(context.TODO()) {
		var family svc.Family
		cursor.Decode(&family)
		families = append(families, family)
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "All Families",
		"data":    families,
	})

	return
}

// GetFamilyByID - as is
func (s *CCServer) GetFamilyByID(c *gin.Context) {
	id := c.Param("id")
	family := svc.Family{}
	err := svc.GetFamilyByID(id).Decode(&family)
	if err != nil {
		log.Printf("Error while Getting Family by ID - %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  http.StatusInternalServerError,
			"message": "Something went wrong",
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "Found Family BY ID",
		"data":    family,
	})
	return
}

// GetFamilyByWardID - as is
func (s *CCServer) GetFamilyByWardID(c *gin.Context) {
	wardID := c.DefaultQuery("wardID", "000000000000000000000000")
	family := svc.Family{}
	err := svc.GetFamilyByWardID(wardID).Decode(&family)
	if err != nil {
		log.Printf("Error while Getting Family by ID - %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  http.StatusInternalServerError,
			"message": "Something went wrong",
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "Found Family BY wardID",
		"data":    family,
	})
	return
}

// CreateFamily - as is
func (s *CCServer) CreateFamily(c *gin.Context) {
	var fRegForm svc.FamilyRegForm
	c.BindJSON(&fRegForm)

	guardiansToCreate := []svc.Guardian{}
	wardsToCreate := []svc.Ward{}
	vehiclesToCreate := []svc.Vehicle{}
	for _, gForm := range fRegForm.Guardians {
		newGuardian := svc.GetNewGuardian(gForm)
		guardiansToCreate = append(guardiansToCreate, newGuardian)
		// register RegCode for the Guardian in DB
		_, err := svc.CreateRegCodeByGuardianID(newGuardian.ID.Hex())
		if err != nil {
			log.Printf("Error while Creating new RegCode into DB - %v\n", err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"status":  http.StatusInternalServerError,
				"message": "Something went wrong",
			})
			return
		}
	}
	for _, wForm := range fRegForm.Wards {
		newWard := svc.GetNewWard(wForm)
		wardsToCreate = append(wardsToCreate, newWard)

	}
	for _, vForm := range fRegForm.Vehicles {
		newVehicle := svc.GetNewVehicle(vForm)
		vehiclesToCreate = append(vehiclesToCreate, newVehicle)

	}
	// gForm := fRegForm.Guardians[0]

	// register Family in DB
	_, err := svc.CreateFamily(fRegForm, guardiansToCreate, wardsToCreate, vehiclesToCreate)

	if err != nil {
		log.Printf("Error while inserting new Family into DB - %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  http.StatusInternalServerError,
			"message": "Something went wrong",
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Family created Successfully",
	})
	return
}

// func GetFamilyByDeviceId(c *gin.Context) {

// }

// UpdateFamilyInfoByID modifies Family fields, EXCEPT embedded Guardians / Wards
func (s *CCServer) UpdateFamilyInfoByID(c *gin.Context) {
	var fEditForm svc.FamilyEditForm
	c.BindJSON(&fEditForm)
	idToUpdate := c.Param("id")
	log.Printf("family update form - %v\n", fEditForm)
	log.Printf("family ID to Update - %v\n", idToUpdate)
	_, err := svc.UpdateFamilyInfoByID(fEditForm, idToUpdate)

	if err != nil {
		log.Printf("Error while updating Family in DB - %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  http.StatusInternalServerError,
			"message": "Something went wrong",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Family updated Successfully",
	})
	return

}

// DeleteFamilyByID deletes a Family Object
func (s *CCServer) DeleteFamilyByID(c *gin.Context) {
	idToDelete := c.Param("id")

	_, err := svc.DeleteFamilyByID(idToDelete)

	if err != nil {
		log.Printf("Error while deleting Family in DB - %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  http.StatusInternalServerError,
			"message": "Something went wrong",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Family deleted Successfully",
	})
	return
}

func extractFamilyParams(c *gin.Context, params *svc.GetFamilyParams) error {
	params.InstID = c.DefaultQuery("instID", "000000000000000000000000")

	return nil
}
