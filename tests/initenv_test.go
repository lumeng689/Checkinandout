package tests

import (
	"log"
	"os"
	"testing"

	"cloudminds.com/harix/cc-server/controllers"
	"github.com/gin-gonic/gin"
)

// TestMain is a golang-testing reserved function. TestMain gets execute before other Test- Functions
func TestMain(m *testing.M) {

	testCCServer = controllers.InitServer()
	testCCServer.InitConfig(appName, true)
	// Database
	testCCServer.Connect()
	testCCServer.ReloadConfigFromDB()

	testCCServer.InitValidator()
	// Init router
	testRouter = gin.Default()
	testCCServer.Routes(testRouter)

	log.Println("Test Setup Complete")
	os.Exit(m.Run())
}
