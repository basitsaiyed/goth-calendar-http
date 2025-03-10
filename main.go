package main

import (
	db "goauthDemo/database"
	"goauthDemo/internal/auth"
	"goauthDemo/middleware"
	"goauthDemo/routes"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/mux"
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

	// Setup router using gorilla/mux
	r := mux.NewRouter()

	// Web routes
	r.HandleFunc("/", routes.Home).Methods("GET")
	r.HandleFunc("/auth/{provider}", routes.AuthProvider).Methods("GET")
	r.HandleFunc("/auth/{provider}/callback", routes.AuthCallback).Methods("GET")
	r.HandleFunc("/schedule-meeting", routes.ScheduleMeeting).Methods("GET")

	// API routes (protected by JWT)
	apiRouter := r.PathPrefix("").Subrouter()
	apiRouter.Use(middleware.JWTAuthMiddleware)
	apiRouter.HandleFunc("/create-meeting", routes.CreateMeeting).Methods("POST")
	apiRouter.HandleFunc("/upcoming-meetings", routes.GetUpcomingMeetings).Methods("GET")

	// Add some basic middleware for all routes
	// Similar to what Gin provides by default
	handler := logRequest(r)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	srv := &http.Server{
		Handler:      handler,
		Addr:         ":" + port,
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	log.Printf("Server starting on :%s", port)
	log.Fatal(srv.ListenAndServe())
}

// Simple logging middleware to replace Gin's default logger
func logRequest(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		log.Printf("Started %s %s", r.Method, r.URL.Path)
		next.ServeHTTP(w, r)
		log.Printf("Completed %s %s in %v", r.Method, r.URL.Path, time.Since(start))
	})
}
