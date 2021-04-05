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

// GetManyTags - as is
func (s *CCServer) GetManyTags(c *gin.Context) {
	var queryParams svc.GetTagParams
	queryParams.InstID = c.DefaultQuery("instID", "000000000000000000000000")

	cursor, err := svc.GetManyTags(&queryParams)

	if err != nil {
		log.Printf("Error while getting all Tags - %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Something went wrong",
		})
		return
	}

	// Iterate through the returned cursor
	tags := []svc.Tag{}
	for cursor.Next(context.TODO()) {
		var tag svc.Tag
		cursor.Decode(&tag)
		tags = append(tags, tag)
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "All Tags",
		"data":    tags,
	})

	return
}

// GetTag - as is
// func (s *CCServer) GetTag(c *gin.Context) {
// 	var queryParams svc.GetTagParams
// 	queryParams.TagString = c.DefaultQuery("tagString", "")

// 	tag := svc.Tag{}
// 	params := svc.GetTagParams{
// 		TagString: queryParams.TagString,
// 	}

// 	err := svc.GetTag(&params).Decode(&tag)

// 	if err != nil {
// 		// When no institution found, return failed
// 		if err == mongo.ErrNoDocuments {
// 			c.JSON(http.StatusForbidden, gin.H{
// 				"message": "Tag not found by TagString, Get Tag Failed",
// 			})
// 			return
// 		}
// 		log.Printf("Error while finding Tag - %v\n", err)
// 		c.JSON(http.StatusInternalServerError, gin.H{
// 			"message": "Something went wrong",
// 		})
// 		return
// 	}

// 	c.JSON(http.StatusOK, gin.H{
// 		"message": "All Tags",
// 		"data":    tag,
// 	})

// 	return
// }

// CreateTag - as is
func (s *CCServer) CreateTag(c *gin.Context) {
	var tRegForm svc.TagRegForm
	c.BindJSON(&tRegForm)

	// Get Institution
	var inst svc.Institution
	err := svc.GetInstByID(tRegForm.InstID).Decode(&inst)

	if err != nil {
		// When no institution found, return failed
		if err == mongo.ErrNoDocuments {
			c.JSON(http.StatusForbidden, gin.H{
				"message": "Institution does not exist, Tag Registration failed",
			})
			return
		}
		log.Printf("Error while finding institution - %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Something went wrong",
		})
		return
	}

	// Validation
	s.RegisterTagStringValidator(inst.CustomTagStringRegex)
	err = s.Validator.v.Struct(tRegForm)
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

	// Check if the TagString exists
	countTagParams := svc.CountTagParams{
		InstID:    tRegForm.InstID,
		TagString: tRegForm.TagString,
	}
	count, err := svc.CountTag(countTagParams)
	if err != nil {
		log.Printf("Error while finding Tag by TagString - %v\n", err)
	}
	if count > 0 {
		c.JSON(http.StatusForbidden, gin.H{
			"message": "TagString has been registered!",
		})
		return
	}

	// Create Tag
	_, err = svc.CreateTag(tRegForm)
	if err != nil {
		log.Printf("Error while inserting new Tag into DB - %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Something went wrong",
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Tag registered Successfully",
	})
	return
}

// UpdateTagByID - as is
func (s *CCServer) UpdateTagByID(c *gin.Context) {

	var tForm svc.TagEditForm
	c.BindJSON(&tForm)

	// Validation
	err := s.Validator.v.Struct(tForm)
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

	// Perform Update
	idToUpdate := c.Param("id")
	_, err = svc.UpdateTagByID(tForm, idToUpdate)
	if err != nil {
		log.Printf("Error while updating Tag in DB - %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Something went wrong",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Tag updated Successfully",
	})

	// Get Tag & Institution
	tag := svc.Tag{}
	err = svc.GetTagByID(idToUpdate).Decode(&tag)

	inst := svc.Institution{}
	err = svc.GetInstByID(tag.InstID).Decode(&inst)

	// TODO - update Display Names in CCRecords
	tagInfo := svc.MemberTagInfo{
		ID:       tag.TagString,
		Name:     tag.FirstName + " " + tag.LastName,
		PhoneNum: tag.PhoneNum,
		Group:    tag.Group,
	}
	_, err = svc.UpdateManyCCRecordsMTInfoByMTID(tag.TagString, tagInfo, inst.MemberType)
	if err != nil {
		log.Printf("Error while updating CCRecord in DB - %v\n", err)
		// c.JSON(http.StatusInternalServerError, gin.H{
		// 	"message": "Something went wrong",
		// })
		return
	}
}

// DeleteTagByID - as is
func (s *CCServer) DeleteTagByID(c *gin.Context) {
	idToDelete := c.Param("id")
	_, err := svc.DeleteTagByID(idToDelete)
	if err != nil {
		log.Printf("Error while deleting Tag in DB - %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Something went wrong",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Tag deleted Successfully",
	})

	// TODO - set CC-Records to Expire
}

func getOrCreateTag(c *gin.Context, tParams *svc.GetTagParams) (*svc.Tag, bool) {

	tag := svc.Tag{}
	err := svc.GetTag(tParams).Decode(&tag)
	if err == nil {
		return &tag, true
	}

	if err == mongo.ErrNoDocuments {
		// If Tag not exist, create one under institution
		tRegForm := svc.TagRegForm{
			InstID:    tParams.InstID,
			TagString: tParams.TagString,
		}
		_, err = svc.CreateTag(tRegForm)
		if err != nil {
			log.Printf("Error while inserting new Tag into DB - %v\n", err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"message": "Something went wrong",
			})
			return nil, false
		}
		err = svc.GetTag(tParams).Decode(&tag)
		// log.Printf("Tag not found by TagString, New Tag Created!")
		return &tag, true
	}

	log.Printf("Error while getting Tag by TagString - %v\n", err)
	c.JSON(http.StatusInternalServerError, gin.H{
		"message": "Something went wrong",
	})
	return nil, false

}
