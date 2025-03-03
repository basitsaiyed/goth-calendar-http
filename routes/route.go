package routes

import (
	"context"
	"goauthDemo/calendar"
	"goauthDemo/internal/auth"
	"log"
	"net/http"
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

	token, err := auth.GenerateJWT(user.UserID, user.AccessToken)
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to generate token"})
		return
	}

	c.Redirect(http.StatusFound, "/schedule-meeting?token="+token)
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

	accessToken, exists := c.Get("access_token")
	if !exists {
		c.JSON(401, gin.H{"error": "Access token not found"})
		return
	}

	err := calendar.CreateEvent(accessToken.(string), request.Title, request.StartTime, request.EndTime, request.Description, request.Attendees)
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
