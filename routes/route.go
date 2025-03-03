package routes

import (
	"context"
	"goauthDemo/calendar"
	"goauthDemo/internal/auth"
	"goauthDemo/middleware"
	"log"
	"net/http"
	"strings"
	"text/template"

	"github.com/gin-gonic/gin"
	"github.com/markbates/goth/gothic"
)

// Home Page
func Home(c *gin.Context) {
	tmpl, _ := template.ParseFiles("index.html")
	tmpl.Execute(c.Writer, nil)
}

// Google OAuth Login
func AuthProvider(c *gin.Context) {
	provider := c.Param("provider")
	if provider == "" {
		c.JSON(400, gin.H{"error": "Provider is required"})
		return
	}

	c.Request = c.Request.WithContext(context.WithValue(c.Request.Context(), "provider", provider))
	gothic.BeginAuthHandler(c.Writer, c.Request)
}

// Google OAuth Callback
func AuthCallback(c *gin.Context) {
	provider := c.Param("provider")
	if provider == "" {
		c.JSON(400, gin.H{"error": "Provider is required"})
		return
	}

	c.Request = c.Request.WithContext(context.WithValue(c.Request.Context(), "provider", provider))
	user, err := gothic.CompleteUserAuth(c.Writer, c.Request)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	// Save user details to database
	err = auth.SaveUserToDB(user)
	if err != nil {
		log.Printf("Error saving user to database: %v", err)
		// Continue anyway - don't fail the authentication
	}

	token, err := auth.GenerateJWT(user.UserID, user.AccessToken)
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to generate token"})
		return
	}

	// For web clients, redirect to the frontend
	if c.GetHeader("Accept") == "text/html" {
		c.Redirect(http.StatusFound, "/schedule-meeting?token="+token)
		return
	}

	// For API clients (like Thunder Client), return JSON
	c.JSON(http.StatusOK, gin.H{
		"token": token,
		"user": gin.H{
			"id":    user.UserID,
			"name":  user.Name,
			"email": user.Email,
		},
	})
}

// Schedule Meeting Page
func ScheduleMeeting(c *gin.Context) {
	c.File("templates/schedule-meeting.html")
}

// Create Meeting
func CreateMeeting(c *gin.Context) {
	var request struct {
		Title       string   `json:"title"`
		Description string   `json:"description"`
		Attendees   []string `json:"attendees"`
		StartTime   string   `json:"startTime"`
		EndTime     string   `json:"endTime"`
	}

	if err := c.BindJSON(&request); err != nil {
		c.JSON(400, gin.H{"error": "Invalid request"})
		return
	}
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header missing"})
		c.Abort()
		return
	}

	tokenString := strings.TrimPrefix(authHeader, "Bearer ")

	claims, err := middleware.ValidateJWT(tokenString)
	if err != nil {
		c.JSON(401, gin.H{"error": "Invalid token"})
		c.Abort()
		return
	}

	accessToken := claims["access_token"].(string)
	// request.Attendees = append(request.Attendees, userID)
	// Use the token to create a meeting in Google Calendar

	err = calendar.CreateEvent(accessToken, request.Title, request.StartTime, request.EndTime, request.Description, request.Attendees)
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to create meeting"})
		return
	}

	c.JSON(200, gin.H{"message": "Meeting created successfully!"})
}

// Fetch Upcoming Meetings
func GetUpcomingMeetings(c *gin.Context) {
	// The middleware already validated the JWT and put the access_token in the context
	accessToken, exists := c.Get("access_token")
	if !exists {
		c.JSON(401, gin.H{"error": "Access token not found"})
		return
	}

	events, err := calendar.GetCalendarEvents(accessToken.(string))
	if err != nil {
		log.Printf("Error fetching upcoming meetings: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"events": events})
}
