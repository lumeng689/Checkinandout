package controllers

import (
	"context"
	"log"
	"net/http"

	svc "cloudminds.com/harix/cc-server/services"
	"github.com/gin-gonic/gin"
)

// GetManySurveys - as is
func (s *CCServer) GetManySurveys(c *gin.Context) {
	var queryParams svc.GetSurveyParams
	queryParams.InstID = c.DefaultQuery("instID", "000000000000000000000000")

	cursor, err := svc.GetManySurveys(&queryParams)

	if err != nil {
		log.Printf("Error while getting all surveys - %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  http.StatusInternalServerError,
			"message": "Something went wrong",
		})
		return
	}

	surveys := []svc.Survey{}
	if err = cursor.All(context.TODO(), &surveys); err != nil {
		panic(err)
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "All Surveys",
		"data":    surveys,
	})
	return
}

// CreateSurvey - as is
func (s *CCServer) CreateSurvey(c *gin.Context) {
	sRegForm := svc.SurveyRegForm{}
	c.BindJSON(&sRegForm)

	log.Printf("survey Reg Form - %v\n", sRegForm)

	_, err := svc.CreateSurvey(sRegForm, sRegForm.QAList)

	if err != nil {
		log.Printf("Error while inserting new Survey into DB - %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Something went wrong",
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Survey created Successfully",
	})
	return

}
