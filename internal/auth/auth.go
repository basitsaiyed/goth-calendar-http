package auth

import (
	"log"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/sessions"
	"github.com/joho/godotenv"
	"github.com/markbates/goth"
	"github.com/markbates/goth/gothic"
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
	// fmt.Println("This is Faiz")

	googleClientId := os.Getenv("GOOGLE_CLIENT_ID")
	googleClientSecret := os.Getenv("GOOGLE_CLIENT_SECRET")

	store := sessions.NewCookieStore([]byte(key))
	store.MaxAge(maxAge)
	store.Options.HttpOnly = true
	store.Options.Secure = isProd

	gothic.Store = store
	goth.UseProviders(
		google.New(googleClientId, googleClientSecret, "http://localhost:8080/auth/google/callback", "email", "profile", "https://www.googleapis.com/auth/calendar.events"),
		// github.New(os.Getenv("GITHUB_CLIENT_ID"), os.Getenv("GITHUB_CLIENT_SECRET"), "http://localhost:8080/auth/github/callback"),
	)
	log.Println("Auth initialized successfully")
}

var jwtSecret = []byte("your_secret_key")

// GenerateJWT generates a JWT token for authenticated users
func GenerateJWT(userID, accessToken string) (string, error) {
	claims := jwt.MapClaims{
		"user_id":      userID,
		"access_token": accessToken,
		"exp":          time.Now().Add(time.Hour * 24).Unix(), // Token expires in 24 hours
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret)
}
