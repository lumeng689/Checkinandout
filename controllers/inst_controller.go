package controllers

import (
	"context"
	"fmt"
	"log"
	"net/http"

	svc "cloudminds.com/harix/cc-server/services"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"go.mongodb.org/mongo-driver/mongo"
)

// GetManyInsts - as is
func (s *CCServer) GetManyInsts(c *gin.Context) {
	cursor, err := svc.GetManyInsts()

	if err != nil {
		log.Printf("Error while getting all institutions - %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  http.StatusInternalServerError,
			"message": "Something went wrong",
		})
		return
	}
	// Iterate through the returned cursor
	insts := []svc.Institution{}
	if err = cursor.All(context.TODO(), &insts); err != nil {
		panic(err)
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "All Institutions",
		"data":    insts,
	})
	return
}

// GetInstByID - as is
func (s *CCServer) GetInstByID(c *gin.Context) {
	id := c.Param("id")
	inst := svc.Institution{}
	err := svc.GetInstByID(id).Decode(&inst)

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

	c.JSON(http.StatusOK, gin.H{
		"message": "Institution Found",
		"data":    inst,
	})
	return
}

// CreateInst - as is
func (s *CCServer) CreateInst(c *gin.Context) {
	var instForm svc.InstitutionForm
	c.BindJSON(&instForm)
	log.Printf("instForm received - %v\n", instForm)

	// Validation
	err := s.Validator.v.Struct(instForm)
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

	_, err = svc.CreateInst(instForm)

	if err != nil {
		log.Printf("Error while inserting new Institution into DB - %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Something went wrong",
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Institution registered Successfully",
	})
	return
}

// UpdateInstByID - as is
func (s *CCServer) UpdateInstByID(c *gin.Context) {
	var instForm svc.InstitutionForm
	c.BindJSON(&instForm)
	idToUpdate := c.Param("id")

	res, err := svc.UpdateInstByID(instForm, idToUpdate)

	if err != nil {

		log.Printf("Error while updating Institution to DB - %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  http.StatusInternalServerError,
			"message": "Something went wrong",
		})
		return
	}
	if res.ModifiedCount == 0 {
		c.JSON(http.StatusForbidden, gin.H{
			"message": "ID To Update Institution Not Found",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Institution updated Successfully",
	})
	return
}

// DeleteInstByID - as is
func (s *CCServer) DeleteInstByID(c *gin.Context) {
	// TODO: Set CC Records to "expired"
	idToDelete := c.Param("id")

	// // Check Admins
	// count, err := svc.CountAdminsByInstID(idToDelete)
	// if err != nil {
	// 	log.Printf("Error while finding Admins in DB - %v\n", err)
	// 	return
	// }
	// if count > 0 {
	// 	c.JSON(http.StatusForbidden, gin.H{
	// 		"message": "Need to remove all affliated admins before deleting the institution",
	// 	})
	// 	return
	// }
	// // Check Families
	// count, err = svc.CountFamiliesByInstID(idToDelete)
	// if err != nil {
	// 	log.Printf("Error while finding Families in DB - %v\n", err)
	// 	return
	// }
	// if count > 0 {
	// 	c.JSON(http.StatusForbidden, gin.H{
	// 		"message": "Need to remove all affliated families before deleting the institution",
	// 	})
	// 	return
	// }

	// Delete the Institution
	res, err := svc.DeleteInstByID(idToDelete)
	if err != nil {
		log.Printf("Error while deleting Institution from DB - %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Something went wrong",
		})
		return
	}
	if res.DeletedCount == 0 {
		c.JSON(http.StatusForbidden, gin.H{
			"message": "ID To Delete Institution Not Found",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Institution Deleted Successfully",
	})
	return
}
