package auth

import (
	"goauthDemo/models"
	"log"
	"os"
	"time"

	db "goauthDemo/database"

	"github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/sessions"
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
	googleClientId := os.Getenv("GOOGLE_CLIENT_ID")
	googleClientSecret := os.Getenv("GOOGLE_CLIENT_SECRET")
	
	if googleClientId == "" || googleClientSecret == "" {
		log.Fatal("Google OAuth credentials are missing. Ensure GOOGLE_CLIENT_ID and GOOGLE_CLIENT_SECRET are set.")
	}

	store := sessions.NewCookieStore([]byte(key))
	store.MaxAge(maxAge)
	store.Options.HttpOnly = true
	store.Options.Secure = isProd

	gothic.Store = store
	goth.UseProviders(
		google.New(googleClientId, googleClientSecret, "http://localhost:8080/auth/google/callback", "email", "profile", "https://www.googleapis.com/auth/calendar.events"),
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
		AccessToken:  user.AccessToken,
		RefreshToken: user.RefreshToken,
		TokenExpiry:  tokenExpiry,
	}

	return userRepo.CreateOrUpdateUser(dbUser)
}

// GenerateJWT generates a JWT token for authenticated users
func GenerateJWT(userID, accessToken string) (string, error) {
	jwtSecret := []byte(os.Getenv("SECRET_KEY"))
	
	if len(jwtSecret) == 0 {
		log.Println("WARNING: SECRET_KEY is empty or not set")
	} else {
		log.Printf("Using JWT secret key (length: %d)", len(jwtSecret))
	}
	
	claims := jwt.MapClaims{
		"user_id":      userID,
		"access_token": accessToken,
		"exp":          time.Now().Add(time.Hour * 24).Unix(), // Token expires in 24 hours
	}
	
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret)
}