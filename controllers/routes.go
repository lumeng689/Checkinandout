package controllers

import (
	"github.com/gin-gonic/gin"
)

// Routes - All API definitions
func (s *CCServer) Routes(router *gin.Engine) {

	// Institution APIs
	router.GET("api/institutions", s.GetManyInsts)
	router.GET("api/institution/:id", s.GetInstByID)
	router.POST("api/institution", s.CreateInst)
	router.PUT("api/institution/:id", s.UpdateInstByID)
	router.DELETE("api/institution/:id", s.DeleteInstByID)

	// Admin APIs
	router.GET("api/admins", s.GetManyAdminsByInstID)
	router.POST("api/admin/register", s.RegisterAdmin)
	router.PUT("api/admin/:id", s.UpdateAdminByID)
	router.DELETE("api/admin/:id", s.DeleteAdminByID)
	router.GET("api/admin", s.GetAdminByFrasUsername)
	router.POST("api/admin/login", s.AdminLogin)

	// CC-Records APIs
	router.GET("/api/cc-records", s.GetManyCCRecords)
	router.DELETE("/api/cc-record/:id", s.DeleteCCRecordByID)
	router.POST("/api/cc-record/sync", s.GetOrCreateManyCCRecords)
	router.POST("/api/cc-record/scan", s.HandleCCScanEvent)
	router.POST("/api/cc-record/schedule", s.HandleCheckoutScheduleEvent)

	// Tag APIs
	router.GET("api/tags", s.GetManyTags)
	router.GET("api/tag", s.GetTag)
	router.POST("api/tag", s.CreateTag)
	router.PUT("api/tag/:id", s.UpdateTagByID)
	router.DELETE("api/tag/:id", s.DeleteTagByID)

	// Member APIs
	router.GET("/api/members", s.GetManyMembers)
	router.POST("/api/member", s.CreateMember)
	router.PUT("/api/member/:id", s.UpdateMemberByID)
	router.DELETE("/api/member/:id", s.DeleteMemberByID)
	router.POST("/api/member/login", s.LoginMember)
	router.POST("/api/member/activate", s.ActivateMember)

	// Family APIs
	router.GET("/api/families", s.GetManyFamilies)
	router.POST("/api/family", s.CreateFamily)
	router.DELETE("/api/family/:id", s.DeleteFamilyByID)
	router.GET("/api/family", s.GetFamily)
	router.GET("/api/family-with-members", s.GetFamilyWithMembers)
	router.GET("/api/family-with-members/:id", s.GetFamilyWithMembersByID)

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
	router.GET("/api/reg-code", s.GetRegCodeByMemberID)
	router.POST("/api/reg-code/email", s.SendRegCodeWithEmail)
	router.POST("/api/reg-code/sms", s.SendRegCodeWithSMS)

	// Survey APIs
	router.GET("/api/surveys", s.GetManySurveys)
	router.POST("/api/survey", s.CreateSurvey)

	// Export APIs
	router.GET("/api/export/cc-records", s.ExportManyCCRecords)
	router.GET("/api/export/members", s.ExportManyMembers)
	router.GET("/api/export/families", s.ExportManyFamilies)
	router.GET("/api/export/wards", s.ExportManyWards)
	router.GET("/api/export/surveys", s.ExportManySurveys)

	// Import APIs
	router.POST("/api/import/tags", s.ImportManyTags)

	// Config APIs
	router.GET("/api/configs", s.GetManyConfigs)
	router.POST("/api/config", s.CreateConfig)
	router.PUT("/api/config/:id", s.UpdateConfigByID)
	router.POST("/api/config/reload", s.RunReloadConfig)

}
