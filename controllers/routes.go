package controllers

import (
	"github.com/gin-gonic/gin"
)

// Routes - All API definitions
func (s *CCServer) Routes(router *gin.Engine) {

	superAdminTokenNeeded := router.Group("/")
	adminTokenNeeded := router.Group("/")
	mobileTokenNeeded := router.Group("/")
	authNotNeeded := router.Group("/")

	superAdminTokenNeeded.Use(s.tokenAuth(s.Config.DebugTokenL.SuperAdmin))
	adminTokenNeeded.Use(s.tokenAuth(s.Config.DebugTokenL.Admin))
	mobileTokenNeeded.Use(s.tokenAuth(s.Config.DebugTokenL.Mobile))

	// Check-Me MobileApp APIs
	mobileTokenNeeded.POST("api/cc-record/sync", s.GetOrCreateManyCCRecords)
	mobileTokenNeeded.POST("api/cc-record/schedule", s.HandleCheckoutScheduleEvent)
	authNotNeeded.POST("api/member/login", s.LoginMember)
	authNotNeeded.POST("api/member/activate", s.ActivateMember)
	authNotNeeded.POST("api/member/register-and-sms", s.CreateMemberAndSendSMS)

	// Gatekeeper APIs
	authNotNeeded.POST("api/cc-record/scan", s.HandleCCScanEvent)

	// MobileAlert APIS
	authNotNeeded.GET("api/cc-record/get-name", s.GetScanNameByDeviceID)

	// Institution APIs
	superAdminTokenNeeded.GET("api/institutions", s.GetManyInsts)
	authNotNeeded.GET("api/institution/:id", s.GetInstByID)
	superAdminTokenNeeded.POST("api/institution", s.CreateInst)
	superAdminTokenNeeded.PUT("api/institution/:id", s.UpdateInstByID)
	superAdminTokenNeeded.DELETE("api/institution/:id", s.DeleteInstByID)

	// Admin APIs
	superAdminTokenNeeded.GET("api/admins", s.GetManyAdminsByInstID)
	superAdminTokenNeeded.POST("api/admin/register", s.RegisterAdmin)
	superAdminTokenNeeded.PUT("api/admin/:id", s.UpdateAdminByID)
	superAdminTokenNeeded.DELETE("api/admin/:id", s.DeleteAdminByID)
	adminTokenNeeded.GET("api/admin", s.GetAdminByFrasUsername)
	authNotNeeded.POST("api/admin/login", s.AdminLogin)

	// CC-Records APIs
	adminTokenNeeded.GET("api/cc-records", s.GetManyCCRecords)
	adminTokenNeeded.DELETE("api/cc-record/:id", s.DeleteCCRecordByID)

	// Tag APIs
	adminTokenNeeded.GET("api/tags", s.GetManyTags)
	adminTokenNeeded.GET("api/tag", s.GetTag)
	adminTokenNeeded.POST("api/tag", s.CreateTag)
	adminTokenNeeded.PUT("api/tag/:id", s.UpdateTagByID)
	adminTokenNeeded.DELETE("api/tag/:id", s.DeleteTagByID)

	// Member APIs
	adminTokenNeeded.GET("api/members", s.GetManyMembers)
	adminTokenNeeded.POST("api/member", s.CreateMember)
	adminTokenNeeded.PUT("api/member/:id", s.UpdateMemberByID)
	adminTokenNeeded.DELETE("api/member/:id", s.DeleteMemberByID)

	// Family APIs
	adminTokenNeeded.GET("api/families", s.GetManyFamilies)
	adminTokenNeeded.POST("api/family", s.CreateFamily)
	adminTokenNeeded.DELETE("api/family/:id", s.DeleteFamilyByID)
	adminTokenNeeded.GET("api/family", s.GetFamily)
	adminTokenNeeded.GET("api/family-with-members", s.GetFamilyWithMembers)
	adminTokenNeeded.GET("api/family-with-members/:id", s.GetFamilyWithMembersByID)

	// Ward APIs
	adminTokenNeeded.POST("api/ward/add-new", s.AddWard)
	adminTokenNeeded.PUT("api/ward/:id", s.UpdateWardByID)
	adminTokenNeeded.DELETE("api/ward/:id", s.DeleteWardByID)

	// Vehicle APIs
	adminTokenNeeded.POST("api/vehicle/add-new", s.AddVehicle)
	adminTokenNeeded.PUT("api/vehicle/:id", s.UpdateVehicleByID)
	adminTokenNeeded.DELETE("api/vehicle/:id", s.DeleteVehicleByID)

	// RegCode APIs
	superAdminTokenNeeded.GET("api/reg-codes", s.GetManyRegCodes)
	adminTokenNeeded.GET("api/reg-code", s.GetRegCodeByMemberID)
	adminTokenNeeded.POST("api/reg-code/email", s.SendRegCodeWithEmail)
	adminTokenNeeded.POST("api/reg-code/sms", s.SendRegCodeWithSMS)

	// Survey APIs
	adminTokenNeeded.GET("api/surveys", s.GetManySurveys)
	authNotNeeded.POST("api/survey", s.CreateSurvey)

	// Export APIs
	adminTokenNeeded.GET("api/export/cc-records", s.ExportManyCCRecords)
	adminTokenNeeded.GET("api/export/members", s.ExportManyMembers)
	adminTokenNeeded.GET("api/export/families", s.ExportManyFamilies)
	adminTokenNeeded.GET("api/export/wards", s.ExportManyWards)
	adminTokenNeeded.GET("api/export/surveys", s.ExportManySurveys)

	// Import APIs
	adminTokenNeeded.POST("api/import/tags", s.ImportManyTags)
	adminTokenNeeded.POST("api/import/members", s.ImportManyMembers)

	// Config APIs
	adminTokenNeeded.GET("api/configs", s.GetManyConfigs)
	adminTokenNeeded.POST("api/config", s.CreateConfig)
	adminTokenNeeded.PUT("api/config/:id", s.UpdateConfigByID)
	adminTokenNeeded.POST("api/config/reload", s.RunReloadConfig)

}
