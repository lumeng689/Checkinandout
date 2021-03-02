package controllers

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"

	svc "cloudminds.com/harix/cc-server/services"
)

type FamilyWithMembersResponse struct {
	ID                string            `json:"_id"`
	InstID            string            `json:"institution_id"`
	AllRegCodeSent    bool              `json:"all_reg_code_sent"`
	ContactMemberInfo svc.MemberTagInfo `json:"contact_member_info"`
	Members           []svc.Member      `json:"members"`
	Wards             []svc.Ward        `json:"wards"`
	Vehicles          []svc.Vehicle     `json:"vehicles"`
	ModifiedAt        time.Time         `json:"modified_at"`
}

// GetManyFamilies - as is
func (s *CCServer) GetManyFamilies(c *gin.Context) {
	var queryParams svc.GetFamilyParams
	queryParams.InstID = c.DefaultQuery("instID", "000000000000000000000000")

	cursor, err := svc.GetManyFamilies(&queryParams)

	if err != nil {
		log.Printf("Error while getting all families - %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{
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

// GetFamilyWithMembers - as is
func (s *CCServer) GetFamilyWithMembersByID(c *gin.Context) {
	id := c.Param("id")

	// Get Family By ID
	family := svc.Family{}
	err := svc.GetFamilyByID(id).Decode(&family)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			c.JSON(http.StatusForbidden, gin.H{
				"message": "Family does not exist",
			})
			return
		}
		log.Printf("Error while Getting Family by ID - %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Something went wrong",
		})
		return
	}

	// Get Members By FamilyID
	params := svc.GetMemberParams{
		FamilyID: family.ID.Hex(),
	}
	cursor, err := svc.GetManyMembers(&params)

	members := []svc.Member{}
	if err = cursor.All(context.TODO(), &members); err != nil {
		panic(err)
	}

	familyWithMembers := FamilyWithMembersResponse{
		ID:                family.ID.Hex(),
		InstID:            family.InstID,
		AllRegCodeSent:    family.AllRegCodeSent,
		ContactMemberInfo: family.ContactMemberInfo,
		Members:           members,
		Wards:             family.Wards,
		Vehicles:          family.Vehicles,
		ModifiedAt:        family.ModifiedAt,
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Family with Members",
		"data":    familyWithMembers,
	})
	return
}

// GetFamilyWithMembers - as is
func (s *CCServer) GetFamilyWithMembers(c *gin.Context) {
	// id := c.Param("id")

	wardID, ok := c.GetQuery("wardID")
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "GetFamilyWithMembers Query given is not Supported",
		})
		return
	}
	// Get Family By WardID
	family := svc.Family{}
	err := svc.GetFamilyByWardID(wardID).Decode(&family)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			c.JSON(http.StatusForbidden, gin.H{
				"message": "Family does not exist",
			})
			return
		}
		log.Printf("Error while Getting Family by wardID - %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Something went wrong",
		})
		return
	}

	// Get Members By FamilyID
	params := svc.GetMemberParams{
		FamilyID: family.ID.Hex(),
	}
	cursor, err := svc.GetManyMembers(&params)

	members := []svc.Member{}
	if err = cursor.All(context.TODO(), &members); err != nil {
		panic(err)
	}

	familyWithMembers := FamilyWithMembersResponse{
		ID:                family.ID.Hex(),
		InstID:            family.InstID,
		AllRegCodeSent:    family.AllRegCodeSent,
		ContactMemberInfo: family.ContactMemberInfo,
		Members:           members,
		Wards:             family.Wards,
		Vehicles:          family.Vehicles,
		ModifiedAt:        family.ModifiedAt,
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Family with Members",
		"data":    familyWithMembers,
	})
	return
}

// GetFamily - as is
func (s *CCServer) GetFamily(c *gin.Context) {
	memberID, ok := c.GetQuery("memberID")
	if ok {
		s.getFamilyByMemberID(c, memberID)
		return
	}
	wardID, ok := c.GetQuery("wardID")
	if ok {
		s.getFamilyByWardID(c, wardID)
		return
	}

}

// CreateFamily - as is
func (s *CCServer) CreateFamily(c *gin.Context) {
	var fRegForm svc.FamilyRegForm
	c.BindJSON(&fRegForm)

	// Validation
	err := s.Validator.v.Struct(fRegForm)
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
	for _, mInFamilyRegForm := range fRegForm.Members {
		count, err := svc.CountMembersByPhoneNum(mInFamilyRegForm.PhoneNum)
		if err != nil {
			log.Printf("Error while finding Member by PhoneNum - %v\n", err)
		}
		if count > 0 {
			c.JSON(http.StatusForbidden, gin.H{
				"message": "Phone # has been used! Try another one",
			})
			return
		}
	}

	// Preparing for Creating Family
	wardsToCreate := []svc.Ward{}
	vehiclesToCreate := []svc.Vehicle{}
	for _, wForm := range fRegForm.Wards {
		newWard := svc.GetNewWard(wForm)
		wardsToCreate = append(wardsToCreate, newWard)
	}
	for _, vForm := range fRegForm.Vehicles {
		newVehicle := svc.GetNewVehicle(vForm)
		vehiclesToCreate = append(vehiclesToCreate, newVehicle)
	}

	// Create Family, but without family info
	cMemberInFamilyRegForm := fRegForm.Members[0]
	fRes, err := svc.CreateFamily(fRegForm, cMemberInFamilyRegForm, wardsToCreate, vehiclesToCreate)
	if err != nil {
		log.Printf("Error while inserting new Family into DB - %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Something went wrong",
		})
		return
	}

	// Create Members with FamilyID
	familyID := fRes.InsertedID.(primitive.ObjectID).Hex()
	var contactMemberID string
	for index, mInFamilyRegForm := range fRegForm.Members {
		fInfo := svc.FamilyInfo{
			ID:       familyID,
			Relation: mInFamilyRegForm.Relation,
		}
		mRegForm := svc.MemberRegForm{
			InstID:     fRegForm.InstID,
			FamilyInfo: &fInfo,
			PhoneNum:   mInFamilyRegForm.PhoneNum,
			Email:      mInFamilyRegForm.Email,
			FirstName:  mInFamilyRegForm.FirstName,
			LastName:   mInFamilyRegForm.LastName,
		}
		insertedMemberID, ok := s.handleCreateMember(c, &mRegForm)
		if !ok {
			return
		}
		if err != nil {
			log.Printf("Error while inserting new Member into DB - %v\n", err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"message": "Something went wrong",
			})
			return
		}
		if index == 0 {
			contactMemberID = insertedMemberID
		}
	}

	// Set Contact Member ID to Family
	_, err = svc.SetFamilyContactMemberID(familyID, contactMemberID)
	if err != nil {
		log.Printf("Error while setting Contact MemberID to Family - %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Something went wrong",
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Family Registered Successfully",
	})
	return
}

// DeleteFamilyByID - as is
func (s *CCServer) DeleteFamilyByID(c *gin.Context) {
	idToDelete := c.Param("id")

	_, err := svc.DeleteFamilyByID(idToDelete)

	if err != nil {
		log.Printf("Error while deleting Family in DB - %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Something went wrong",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Family deleted Successfully",
	})
	return
}

func (s *CCServer) getFamilyByMemberID(c *gin.Context, memberID string) {
	var family svc.Family
	err := svc.GetFamilyByMemberID(memberID).Decode(&family)
	if err != nil {
		log.Printf("Error while Getting Family by Member ID - %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Something went wrong",
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "Found Family BY MemberID",
		"data":    family,
	})
	return
}

// GetFamilyByWardID - as is
func (s *CCServer) getFamilyByWardID(c *gin.Context, wardID string) {
	family := svc.Family{}
	err := svc.GetFamilyByWardID(wardID).Decode(&family)
	if err != nil {
		log.Printf("Error while Getting Family by wardID - %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{
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
