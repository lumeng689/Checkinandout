package controllers

import (
	"context"
	"log"
	"net/http"

	svc "cloudminds.com/harix/cc-server/services"
	"github.com/gin-gonic/gin"
)

// RunReloadConfig - as is
func (s *CCServer) RunReloadConfig(c *gin.Context) {
	s.ReloadConfigFromDB()
	c.JSON(http.StatusOK, gin.H{
		"message": "Config Reload is Successful",
	})
}

// GetManyConfigs - as is
func (s *CCServer) GetManyConfigs(c *gin.Context) {
	cursor, err := svc.GetManyConfigs()
	if err != nil {
		log.Printf("Error while getting all admins - %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Something went wrong",
		})
		return
	}
	configs := []svc.Config{}
	if err = cursor.All(context.TODO(), &configs); err != nil {
		panic(err)
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "All Configs",
		"data":    configs,
	})
	return
}

// CreateConfig - as is
func (s *CCServer) CreateConfig(c *gin.Context) {
	var configForm svc.ConfigForm
	c.BindJSON(&configForm)

	// Check if Config exists
	count, err := svc.CountConfigByName(configForm.Name)
	if err != nil {
		log.Printf("Error while counting Config by Name - %v\n", err)
	}
	if count > 0 {
		c.JSON(http.StatusForbidden, gin.H{
			"message": "Config Name has been used! Try another one",
		})
		return
	}

	// Create Config in DB
	_, err = svc.CreateConfig(configForm)
	if err != nil {
		log.Printf("Error while creating new Config in DB - %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Something went wrong",
		})
		return
	}
	c.JSON(http.StatusCreated, gin.H{
		"message": "Config created Successfully",
	})
	return
}

// UpdateConfigByID - as is
func (s *CCServer) UpdateConfigByID(c *gin.Context) {
	var cEditForm svc.ConfigEditForm
	c.BindJSON(&cEditForm)

	idToUpdate := c.Param("id")
	res, err := svc.UpdateConfigByID(cEditForm, idToUpdate)
	if err != nil {
		log.Printf("Error while updating Config in DB - %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Something went wrong",
		})
		return
	}
	if res.ModifiedCount == 0 {
		c.JSON(http.StatusForbidden, gin.H{
			"message": "ID To Update Config Not Found",
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "Config updated Successfully",
	})
	return
}
