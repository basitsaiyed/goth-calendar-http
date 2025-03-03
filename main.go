package main

import (
	"fmt"
	db "goauthDemo/database"
	"goauthDemo/internal/auth"
	"goauthDemo/middleware"
	"goauthDemo/routes"
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	db.InitDB()
	var SecretKEY = os.Getenv("SECRET_KEY")
	fmt.Println("Secret Key:", SecretKEY)
	err := godotenv.Load()
	if err != nil {
		log.Println("Warning: Error loading .env file")
	}
	fmt.Println("SECRET_KEY:", os.Getenv("SECRET_KEY")) // Debug
	auth.NewAuth()
	r := gin.Default()

	// Web routes
	r.GET("/", routes.Home)
	r.GET("/auth/:provider", routes.AuthProvider)
	r.GET("/auth/:provider/callback", routes.AuthCallback)
	r.GET("/schedule-meeting", routes.ScheduleMeeting)

	// API routes (protected by JWT)
	r.POST("/create-meeting", middleware.JWTAuthMiddleware(), routes.CreateMeeting)
	r.GET("/upcoming-meetings", middleware.JWTAuthMiddleware(), routes.GetUpcomingMeetings)

	// Test endpoint - would be disabled in production
	// r.POST("/api/test-token", routes.GetTestToken)

	log.Println("Server starting on :8080")
	r.Run()
}
