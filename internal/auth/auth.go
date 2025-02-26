package auth

import (
	"log"
	"os"

	"github.com/gorilla/sessions"
	"github.com/joho/godotenv"
	"github.com/markbates/goth"
	"github.com/markbates/goth/gothic"
	"github.com/markbates/goth/providers/github"
	"github.com/markbates/goth/providers/google"
)

const (
	key    = "randomstring"
	maxAge = 86400 * 30
	isProd = true
)

func NewAuth() {
	// Initialize auth
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	googleClientId := os.Getenv("GOOGLE_CLIENT_ID")
	googleClientSecret := os.Getenv("GOOGLE_CLIENT_SECRET")

	store := sessions.NewCookieStore([]byte(key))
	store.MaxAge(maxAge)
	store.Options.HttpOnly = true
	store.Options.Secure = isProd

	gothic.Store = store
	goth.UseProviders(
		google.New(googleClientId, googleClientSecret, "http://localhost:8080/auth/google/callback"),
		github.New(os.Getenv("GITHUB_CLIENT_ID"), os.Getenv("GITHUB_CLIENT_SECRET"), "http://localhost:8080/auth/github/callback"),
	)
	log.Println("Auth initialized successfully")
}
