package controllers

import (
	"log"
	"net/http"

	svc "cloudminds.com/harix/cc-server/services"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
)

// AddVehicle - add a new Vehicle within a Family.
func (s *CCServer) AddVehicle(c *gin.Context) {
	// retrieve Family
	var queryParams svc.AddVehicleParams
	queryParams.FamilyID = c.DefaultQuery("familyID", "000000000000000000000000")
	familyToAppend := svc.Family{}
	err := svc.GetFamilyByID(queryParams.FamilyID).Decode(&familyToAppend)
	if err != nil {
		// When no family found, return failed
		if err == mongo.ErrNoDocuments {
			c.JSON(http.StatusForbidden, gin.H{
				"message": "Family does not exist, Adding Vehicle failed",
			})
			return
		}
		log.Printf("Error while getting Family from DB - %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Something went wrong",
		})
		return
	}

	// Assign new ID to the Vehicle
	var vAddForm svc.VehicleForm
	c.BindJSON(&vAddForm)
	newVehicle := svc.GetNewVehicle(vAddForm)

	// Append New Vehicle to Family and Update in DB
	vehicles := append(familyToAppend.Vehicles, newVehicle)
	_, err = svc.ReplaceFamily(familyToAppend, familyToAppend.ContactGuardianInfo, familyToAppend.Wards, vehicles)
	if err != nil {
		log.Printf("Error while adding Vehicle in DB - %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Something went wrong",
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Vehicle added Successfully",
	})
	return
}

// UpdateVehicleByID - as is
func (s *CCServer) UpdateVehicleByID(c *gin.Context) {
	// Get Family
	idToUpdate := c.Param("id")
	familyToUpdate := svc.Family{}
	err := svc.GetFamilyByVehicleID(idToUpdate).Decode(&familyToUpdate)
	if err != nil {
		// When no family found, return failed
		if err == mongo.ErrNoDocuments {
			c.JSON(http.StatusForbidden, gin.H{
				"message": "Family does not Exist, Vehicle Update failed",
			})
			return
		}
		log.Printf("Error while Getting Family By VehicleID - %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Something went wrong",
		})
		return
	}

	// Replace Vehicle in Family
	var vForm svc.VehicleForm
	c.BindJSON(&vForm)
	vehicles := familyToUpdate.Vehicles
	for index, prev := range vehicles {
		// replace vehicle if ID matches
		if idToUpdate == prev.ID.Hex() {
			vehicle := svc.Vehicle{
				ID:       prev.ID,
				Make:     vForm.Make,
				Model:    vForm.Model,
				Color:    vForm.Color,
				PlateNum: vForm.PlateNum,
			}
			vehicles = append(vehicles[:index], vehicles[index+1:]...)
			vehicles = append(vehicles, vehicle)
			break
		}
	}

	// Save updated family to DB
	_, err = svc.ReplaceFamily(familyToUpdate, familyToUpdate.ContactGuardianInfo, familyToUpdate.Wards, vehicles)
	if err != nil {
		log.Printf("Error while updating Vehicle in DB - %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Something went wrong",
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "Vehicle updated Successfully",
	})
	return
}

// DeleteVehicleByID - as is
func (s *CCServer) DeleteVehicleByID(c *gin.Context) {
	// Get Family
	idToDelete := c.Param("id")
	familyToUpdate := svc.Family{}
	err := svc.GetFamilyByVehicleID(idToDelete).Decode(&familyToUpdate)
	if err != nil {
		// When no family found, return failed
		if err == mongo.ErrNoDocuments {
			c.JSON(http.StatusForbidden, gin.H{
				"message": "Family does not Exist, Vehicle Deleting failed",
			})
			return
		}
		log.Printf("Error while Getting Family By VehicleID - %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Something went wrong",
		})
		return
	}

	// Remove Vehicle from Family
	vehicles := familyToUpdate.Vehicles
	for index, prev := range vehicles {
		// remove vehicle if ID matches
		if idToDelete == prev.ID.Hex() {
			vehicles = append(vehicles[:index], vehicles[index+1:]...)
			break
		}
	}

	// Save updated family to DB
	_, err = svc.ReplaceFamily(familyToUpdate, familyToUpdate.ContactGuardianInfo, familyToUpdate.Wards, vehicles)

	if err != nil {
		log.Printf("Error while removing Vehicle in DB - %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Something went wrong",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Vehicle deleted Successfully",
	})
	return
}
