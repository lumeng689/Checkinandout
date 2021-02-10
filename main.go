package main

import (
	"fmt"
	"log"

	"cloudminds.com/harix/cc-server/controllers"
	"github.com/gin-contrib/static"
	"github.com/gin-gonic/gin"
)

var appName = "cc-server"

func main() {

	var ccServer controllers.CCServer = controllers.InitServer()
	ccServer.InitConfig(appName)
	// Database
	ccServer.Connect()

	// Init router
	r := gin.Default()

	// Route Handlers
	r.Use(CORS())
	// Hosting survey pages
	r.Static("/surveys", "./surveys")

	// Hosting CC-Portal Pages
	// r.Static("/dist", "/")
	r.Use(static.Serve("/", static.LocalFile("./web/cc-portal/dist", true)))
	r.Use(static.Serve("/m", static.LocalFile("./web/mobile/dist", true)))
	// API Routes
	ccServer.Routes(r)

	// Fall Through Page
	r.NoRoute(func(c *gin.Context) {
		c.File("./dist/index.html")
	})
	fmt.Println("Initializing Router")

	log.Fatal(r.Run(":8000"))

}

// CORS - config CORS
func CORS() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}
