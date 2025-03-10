package routes

import (
	"context"
	"encoding/json"
	"goauthDemo/calendar"
	"goauthDemo/internal/auth"
	"goauthDemo/middleware"
	"io"
	"log"
	"net/http"
	"os"
	"text/template"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/mux"
	"github.com/markbates/goth/gothic"
)

type ContextKey string

const ProviderKey ContextKey = "provider"

// Home Page
func Home(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("index.html")
	if err != nil {
		log.Printf("Error parsing index.html template: %v", err)
		http.Error(w, "Failed to load home page", http.StatusInternalServerError)
		return
	}
	tmpl.Execute(w, nil)
}

// Google OAuth Login
func AuthProvider(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	provider := vars["provider"]

	if provider == "" {
		http.Error(w, "Provider is required", http.StatusBadRequest)
		return
	}

	log.Printf("Provider: %s", provider)
	ctxProvider := r.Context().Value(ProviderKey)
	log.Printf("Context Provider: %v", ctxProvider)

	r = r.WithContext(context.WithValue(r.Context(), "provider", provider))
	gothic.BeginAuthHandler(w, r)
}

// Google OAuth Callback
func AuthCallback(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	provider := vars["provider"]

	if provider == "" {
		http.Error(w, "Provider is required", http.StatusBadRequest)
		return
	}

	log.Printf("Provider: %s", provider)
	ctxProvider := r.Context().Value(ProviderKey)
	log.Printf("Context Provider: %v", ctxProvider)

	r = r.WithContext(context.WithValue(r.Context(), "provider", provider))
	user, err := gothic.CompleteUserAuth(w, r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
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
		http.Error(w, "Failed to generate token", http.StatusInternalServerError)
		return
	}

	// Redirect to Schedule Meeting Page with JWT Token
	http.Redirect(w, r, "/schedule-meeting?token="+token, http.StatusFound)
}

// Schedule Meeting Page
func ScheduleMeeting(w http.ResponseWriter, r *http.Request) {
	// First, check if the file exists
	templatePath := "templates/schedule-meeting.html"

	// Check if file exists
	_, err := os.Stat(templatePath)
	if os.IsNotExist(err) {
		log.Printf("Error: Schedule meeting template not found at path: %s", templatePath)

		// Try alternative paths
		alternativePaths := []string{
			"./templates/schedule-meeting.html",
			"../templates/schedule-meeting.html",
			"schedule-meeting.html",
		}

		found := false
		for _, path := range alternativePaths {
			if _, err := os.Stat(path); err == nil {
				templatePath = path
				found = true
				log.Printf("Found template at alternative path: %s", path)
				break
			}
		}

		if !found {
			log.Printf("Could not find schedule-meeting.html in any expected location")
			http.Error(w, "Schedule meeting page not found", http.StatusNotFound)
			return
		}
	} else if err != nil {
		log.Printf("Error checking template file: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Log that we're attempting to serve the file
	log.Printf("Serving schedule meeting template from: %s", templatePath)

	// Set content type explicitly
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	// Try to parse and execute the template
	tmpl, err := template.ParseFiles(templatePath)
	if err != nil {
		log.Printf("Error parsing template: %v", err)
		http.Error(w, "Failed to parse schedule meeting page", http.StatusInternalServerError)
		return
	}

	// Get token from query params to pass to template if needed
	token := r.URL.Query().Get("token")

	// Execute template with token data
	if err := tmpl.Execute(w, map[string]string{"Token": token}); err != nil {
		log.Printf("Error executing template: %v", err)
		http.Error(w, "Failed to render schedule meeting page", http.StatusInternalServerError)
		return
	}
}

// Create Meeting
func CreateMeeting(w http.ResponseWriter, r *http.Request) {
	var request struct {
		Title       string   `json:"title"`
		Description string   `json:"description"`
		Attendees   []string `json:"attendees"`
		StartTime   string   `json:"startTime"`
		EndTime     string   `json:"endTime"`
	}

	// Read the request body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request body: "+err.Error(), http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	// Parse the JSON
	if err := json.Unmarshal(body, &request); err != nil {
		http.Error(w, "Invalid request: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Get claims from context (set by middleware)
	claimsValue := r.Context().Value(middleware.UserCtxKey)
	if claimsValue == nil {
		http.Error(w, "User context missing", http.StatusUnauthorized)
		return
	}

	claims, ok := claimsValue.(jwt.MapClaims)
	if !ok {
		http.Error(w, "Invalid user context", http.StatusUnauthorized)
		return
	}

	accessToken, ok := claims["access_token"].(string)
	if !ok {
		http.Error(w, "Access token missing from claims", http.StatusUnauthorized)
		return
	}

	// Log details for debugging
	log.Printf("Creating calendar event: Title=%s, Attendees=%v", request.Title, request.Attendees)

	err = calendar.CreateEvent(accessToken, request.Title, request.StartTime, request.EndTime, request.Description, request.Attendees)
	if err != nil {
		log.Printf("Error creating event: %v", err)
		http.Error(w, "Failed to create meeting: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Send success response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Meeting created successfully!",
	})
}

// GetUpcomingMeetings fetches all upcoming meetings for the next week
func GetUpcomingMeetings(w http.ResponseWriter, r *http.Request) {
	// Get claims from context (set by middleware)
	claimsValue := r.Context().Value(middleware.UserCtxKey)
	if claimsValue == nil {
		http.Error(w, "User context missing", http.StatusUnauthorized)
		return
	}

	claims, ok := claimsValue.(jwt.MapClaims)
	if !ok {
		http.Error(w, "Invalid user context", http.StatusUnauthorized)
		return
	}

	accessToken, ok := claims["access_token"].(string)
	if !ok {
		http.Error(w, "Access token missing from claims", http.StatusUnauthorized)
		return
	}

	// Log for debugging
	log.Printf("Fetching calendar events for next week with token (length: %d)", len(accessToken))

	// Use the new function for upcoming week events
	events, err := calendar.GetUpcomingWeekEvents(accessToken)
	if err != nil {
		log.Printf("Error fetching upcoming meetings: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Format events for response
	var formattedEvents []map[string]interface{}
	for _, event := range events {
		// Parse start time for formatting
		var startTime, endTime time.Time
		var startTimeStr, endTimeStr string

		if event.Start.DateTime != "" {
			startTime, _ = time.Parse(time.RFC3339, event.Start.DateTime)
			startTimeStr = startTime.Format("Jan 02, 2006 03:04 PM")
		} else if event.Start.Date != "" {
			// All-day event
			startTimeStr = event.Start.Date + " (All day)"
		}

		if event.End.DateTime != "" {
			endTime, _ = time.Parse(time.RFC3339, event.End.DateTime)
			endTimeStr = endTime.Format("Jan 02, 2006 03:04 PM")
		} else if event.End.Date != "" {
			// All-day event
			endTimeStr = event.End.Date + " (All day)"
		}

		attendees := []string{}
		for _, attendee := range event.Attendees {
			attendees = append(attendees, attendee.Email)
		}

		formattedEvents = append(formattedEvents, map[string]interface{}{
			"id":          event.Id,
			"title":       event.Summary,
			"description": event.Description,
			"startTime":   startTimeStr,
			"endTime":     endTimeStr,
			"link":        event.HtmlLink,
			"attendees":   attendees,
		})
	}

	// Prepare the response
	response := map[string]interface{}{
		"events": formattedEvents,
		"period": map[string]string{
			"from": time.Now().Format("Jan 02, 2006"),
			"to":   time.Now().AddDate(0, 0, 7).Format("Jan 02, 2006"),
		},
	}

	// Send JSON response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
