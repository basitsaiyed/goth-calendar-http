package main

import (
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
	// Load environment variables first
	err := godotenv.Load()
	if err != nil {
		log.Println("Warning: Error loading .env file, will use environment variables")
	}

	// Check for required environment variables
	secretKey := os.Getenv("SECRET_KEY")
	if secretKey == "" {
		log.Fatal("SECRET_KEY environment variable is not set")
	}
	log.Printf("SECRET_KEY is set (length: %d)", len(secretKey))

	dbURL := os.Getenv("DB_URL")
	if dbURL == "" {
		log.Fatal("DB_URL environment variable is not set")
	}
	log.Println("DB_URL is set")

	// Initialize database
	db.InitDB()
	log.Println("Database initialized")

	// Initialize authentication
	auth.NewAuth()
	log.Println("Authentication initialized")

	// Setup Gin router
	r := gin.Default()

	// Web routes
	r.GET("/", routes.Home)
	r.GET("/auth/:provider", routes.AuthProvider)
	r.GET("/auth/:provider/callback", routes.AuthCallback)
	r.GET("/schedule-meeting", routes.ScheduleMeeting)

	// API routes (protected by JWT)
	r.POST("/create-meeting", middleware.JWTAuthMiddleware(), routes.CreateMeeting)
	r.GET("/upcoming-meetings", middleware.JWTAuthMiddleware(), routes.GetUpcomingMeetings)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Server starting on :%s", port)
	r.Run(":" + port)
}
