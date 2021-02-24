package controllers

import (
	"context"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"

	svc "cloudminds.com/harix/cc-server/services"
)

// GetManyAdmins as name suggests
func (s *CCServer) GetManyAdmins(c *gin.Context) {
	cursor, err := svc.GetManyAdmins()
	if err != nil {
		log.Printf("Error while getting all admins - %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  http.StatusInternalServerError,
			"message": "Something went wrong",
		})
		return
	}
	admins := []svc.Admin{}
	if err = cursor.All(context.TODO(), &admins); err != nil {
		panic(err)
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "All Admins",
		"data":    admins,
	})
	return
}

// GetManyAdminsByInstID as name suggests
func (s *CCServer) GetManyAdminsByInstID(c *gin.Context) {
	instID := c.DefaultQuery("instID", "000000000000000000000000")
	cursor, err := svc.GetManyAdminsByInstID(instID)
	if err != nil {
		log.Printf("Error while getting all admins under the institution - %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  http.StatusInternalServerError,
			"message": "Something went wrong",
		})
		return
	}
	admins := []svc.Admin{}
	if err = cursor.All(context.TODO(), &admins); err != nil {
		panic(err)
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "All Admins under the Institution",
		"data":    admins,
	})
	return
}

// GetAdminByFrasUsername as name suggests
func (s *CCServer) GetAdminByFrasUsername(c *gin.Context) {
	var queryParams svc.GetAdminParams
	queryParams.FrasUsername = c.DefaultQuery("frasUsername", "000000000000000000000000")
	admin := svc.Admin{}
	err := svc.GetAdminByFrasUsername(queryParams.FrasUsername).Decode(&admin)
	if err != nil {
		// When no Doc found, create a new Admin in DB
		if err == mongo.ErrNoDocuments {
			log.Printf("Admin does not exist=")
			c.JSON(http.StatusForbidden, gin.H{
				"message": "Admin does not exist",
			})
			return
		}
		log.Printf("Error while Getting Admin by Fras Username - %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  http.StatusInternalServerError,
			"message": "Something went wrong",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "All Admins under the Institution",
		"data":    admin,
	})
	return
}

// AdminLogin as name suggests
func (s *CCServer) AdminLogin(c *gin.Context) {
	var aLoginForm svc.AdminLoginForm
	c.BindJSON(&aLoginForm)
	log.Printf("Login Info - %v\n", aLoginForm)

	// Get Admin
	adminToLogin := svc.Admin{}
	err := svc.GetAdminByFrasUsername(aLoginForm.FrasUsername).Decode(&adminToLogin)
	if err != nil {
		// When no Doc found, create a new Admin in DB
		if err == mongo.ErrNoDocuments {
			log.Printf("Admin does not exist, need to create a new one")
			c.JSON(http.StatusForbidden, gin.H{
				"message": "Admin does not exist, need to create a new one",
			})
			return
		}
		log.Printf("Error while getting Admin By Fras Username from DB - %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Something went wrong",
		})
		return
	}

	// Compare Password
	if s.Config.RequireAdminPswd && (aLoginForm.Password != adminToLogin.Password) {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Wrong Password, Login Failed",
			"success": false,
		})
		return
	}

	// Update Admin
	admin := svc.Admin{}
	err = svc.UpdateAdminLoginTime(adminToLogin.ID.Hex(), &admin)
	if err != nil {
		log.Printf("Error while loggin in admin - %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  http.StatusInternalServerError,
			"message": "Something went wrong",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Login Succeed",
		"success": true,
		"data":    admin,
	})
	return
}

// RegisterAdmin as name suggests
func (s *CCServer) RegisterAdmin(c *gin.Context) {
	var adminForm svc.AdminRegForm
	c.BindJSON(&adminForm)

	// Check if institution exists
	inst := svc.Institution{}
	err := svc.GetInstByID(adminForm.InstID).Decode(&inst)

	if err != nil {
		// When no institution found, return failed
		if err == mongo.ErrNoDocuments {
			c.JSON(http.StatusForbidden, gin.H{
				"message": "Institution does not exist, admin registration failed",
			})
			return
		}
		log.Printf("Error while finding institution - %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Something went wrong",
		})
		return
	}

	// Check if admin with same FRAS username exists (by validator)
	count, err := svc.CountAdminByFrasUsername(adminForm.FrasUsername)
	if err != nil {
		log.Printf("Error while counting Admin by FrasUsername - %v\n", err)
	}
	if count > 0 {
		c.JSON(http.StatusForbidden, gin.H{
			"message": "Username has been used! Try another one",
		})
		return
	}

	// Create Admin in DB
	_, err = svc.CreateAdmin(adminForm)
	if err != nil {
		log.Printf("Error while inserting new Admin into DB - %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Something went wrong",
		})
		return
	}
	c.JSON(http.StatusCreated, gin.H{
		"message": "Admin registered Successfully",
	})
	return
}

// UpdateAdminByID as name suggests
func (s *CCServer) UpdateAdminByID(c *gin.Context) {
	var adminForm svc.AdminEditForm
	c.BindJSON(&adminForm)

	idToUpdate := c.Param("id")
	res, err := svc.UpdateAdminByID(adminForm, idToUpdate)
	if err != nil {
		log.Printf("Error while updating Admin in DB - %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Something went wrong",
		})
		return
	}
	if res.ModifiedCount == 0 {
		c.JSON(http.StatusForbidden, gin.H{
			"message": "ID To Update Admin Not Found",
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "Admin updated Successfully",
	})
	return
}

// DeleteAdminByID as name suggests
func (s *CCServer) DeleteAdminByID(c *gin.Context) {
	// TODO: Set CC Records to "expired"
	idToDelete := c.Param("id")

	res, err := svc.DeleteAdminByID(idToDelete)
	if err != nil {
		log.Printf("Error while deleting Admin from DB - %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Something went wrong",
		})
		return
	}

	if res.DeletedCount == 0 {
		c.JSON(http.StatusForbidden, gin.H{
			"message": "ID To Delete Admin Not Found",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Admin Deleted Successfully",
	})
	return
}
