package auth

import (
	"fmt"
	"goauthDemo/models"
	"log"
	"os"
	"time"

	db "goauthDemo/database"

	"github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/sessions"
	"github.com/joho/godotenv"
	"github.com/markbates/goth"
	"github.com/markbates/goth/gothic"
	"github.com/markbates/goth/providers/google"
)

var userRepo *db.UserRepository

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
	// SecretKEY := os.Getenv("SECRET_KEY")
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

	if db.DB == nil {
		log.Fatal("Database not initialized")
	}
	userRepo = db.NewUserRepository(db.DB)
	log.Println("Auth initialized successfully")
}

// SaveUserToDB saves the user details to the database
func SaveUserToDB(user goth.User) error {
	// Calculate token expiry (if the provider doesn't provide it)
	tokenExpiry := time.Now().Add(time.Hour)
	if exp := user.ExpiresAt; !exp.IsZero() {
		tokenExpiry = exp
	}

	dbUser := &models.User{
		GoogleID:     user.UserID,
		Email:        user.Email,
		Name:         user.Name,
		Picture:      user.AvatarURL,
		AccessToken:  user.AccessToken,
		RefreshToken: user.RefreshToken,
		TokenExpiry:  tokenExpiry,
	}

	return userRepo.CreateOrUpdateUser(dbUser)
}

// GenerateJWT generates a JWT token for authenticated users
func GenerateJWT(userID, accessToken string) (string, error) {
	claims := jwt.MapClaims{
		"user_id":      userID,
		"access_token": accessToken,
		"exp":          time.Now().Add(time.Hour * 24).Unix(), // Token expires in 24 hours
	}
	jwtSecret := []byte(os.Getenv("SECRET_KEY"))
	fmt.Println("JWT Secret:", string(jwtSecret)) // Debug
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret)
}
