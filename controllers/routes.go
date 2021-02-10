package controllers

import (
	"github.com/gin-gonic/gin"
)

// Routes - All API definitions
func (s *CCServer) Routes(router *gin.Engine) {

	// Institution APIs
	router.GET("/api/institutions", s.GetManyInsts)
	router.GET("/api/institution/:id", s.GetInstByID)
	router.POST("api/institution", s.CreateInst)
	router.PUT("api/institution/:id", s.UpdateInstByID)
	router.DELETE("api/institution/:id", s.DeleteInstByID)

	// Admin APIs
	router.GET("/api/admins", s.GetManyAdminsByInstID)
	router.POST("api/admin/register", s.RegisterAdmin)
	router.PUT("api/admin/:id", s.UpdateAdminByID)
	router.DELETE("api/admin/:id", s.DeleteAdminByID)
	router.GET("api/admin", s.GetAdminByFrasUsername)
	router.POST("api/admin/login", s.AdminLogin)

	// CC-Records APIs
	router.GET("/api/cc-records", s.GetCCRecords)
	router.POST("/api/cc-record/sync", s.GetOrCreateManyCCRecordsByManyWardIDs)
	router.POST("/api/cc-record/scan", s.HandleCCScanEvent)
	router.POST("/api/cc-record/schedule", s.HandleCheckoutScheduleEvent)
	router.DELETE("/api/cc-record/:id", s.DeleteCCRecordByID)

	// Family APIs
	router.GET("/api/families", s.GetManyFamilies)
	router.GET("/api/family/:id", s.GetFamilyByID)
	router.POST("/api/family", s.CreateFamily)
	// router.GET("/api/families/:deviceId", controllers.GetFamilyByDeviceId)
	router.PUT("/api/family/:id", s.UpdateFamilyInfoByID)
	router.DELETE("/api/family/:id", s.DeleteFamilyByID)
	router.GET("/api/family", s.GetFamilyByWardID)

	// Guardian APIs
	router.POST("/api/guardian/add-new", s.AddGuardian)
	router.PUT("/api/guardian/:id", s.UpdateGuardianByID)
	router.DELETE("/api/guardian/:id", s.DeleteGuardianByID)
	router.POST("/api/guardian/login", s.LoginGuardian)
	router.POST("/api/guardian/activate", s.ActivateGuardian)

	// Ward APIs
	router.POST("/api/ward/add-new", s.AddWard)
	router.PUT("/api/ward/:id", s.UpdateWardByID)
	router.DELETE("/api/ward/:id", s.DeleteWardByID)

	// Vehicle APIs
	router.POST("/api/vehicle/add-new", s.AddVehicle)
	router.PUT("/api/vehicle/:id", s.UpdateVehicleByID)
	router.DELETE("/api/vehicle/:id", s.DeleteVehicleByID)

	// RegCode APIs
	router.GET("/api/reg-codes", s.GetManyRegCodes)
	router.GET("/api/reg-code", s.GetRegCodeByGuardianID)
	router.POST("/api/reg-code/email", s.SendRegCodeWithEmail)
	router.POST("/api/reg-code/sms", s.SendRegCodeWithSMS)

	// Survey APIs
	router.GET("/api/surveys", s.GetManySurveys)
	router.POST("/api/survey", s.CreateSurvey)

	// Export APIs
	router.GET("/api/export/cc-records", s.ExportManyCCRecords)
	router.GET("/api/export/wards", s.ExportManyWards)
	router.GET("/api/export/families", s.ExportManyFamilies)
	router.GET("/api/export/surveys", s.ExportManySurveys)
}
