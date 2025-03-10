package db

import (
	"goauthDemo/models"
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

// InitDB initializes the database connection
func InitDB() {
		err := godotenv.Load()
		if err != nil {
			log.Println("Warning: Error loading .env file")
		}

	dsn := os.Getenv("DB_URL")
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}
	DB = db
	log.Println("Connected to database successfully")

	err = DB.AutoMigrate(&models.User{})
	if err != nil {
		panic("failed to migrate database")
	}
}

// UserRepository provides methods to interact with the users table
type UserRepository struct {
	DB *gorm.DB
}

// NewUserRepository creates a new UserRepository instance
func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{DB: db}
}

// CreateOrUpdateUser creates a new user or updates an existing one
func (r *UserRepository) CreateOrUpdateUser(user *models.User) error {
	var existingUser models.User
	if err := r.DB.Where("google_id = ?", user.GoogleID).First(&existingUser).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			// User not found, create a new record
			return r.DB.Create(user).Error
		}
		return err
	}

	// User exists, update fields
	existingUser.Email = user.Email
	existingUser.Name = user.Name
	existingUser.AccessToken = user.AccessToken
	existingUser.RefreshToken = user.RefreshToken
	existingUser.TokenExpiry = user.TokenExpiry
	existingUser.UpdatedAt = time.Now()

	return r.DB.Save(&existingUser).Error
}

// UpdateUserToken updates a user's tokens
func (r *UserRepository) UpdateUserToken(googleID, accessToken, refreshToken string, tokenExpiry time.Time) error {
	return r.DB.Model(&models.User{}).
		Where("google_id = ?", googleID).
		Updates(models.User{
			AccessToken:  accessToken,
			RefreshToken: refreshToken,
			TokenExpiry:  tokenExpiry,
			UpdatedAt:    time.Now(),
		}).Error
}
