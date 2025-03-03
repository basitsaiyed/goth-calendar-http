package middleware

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/joho/godotenv"
)

// var jwtSecret = []byte("your_secret_key")

const UserCtxKey = "user"

func JWTAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		err := godotenv.Load()
		if err != nil {
			log.Fatal("Error loading .env file")
		}

		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header missing"})
			c.Abort()
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")

		claims, err := ValidateJWT(tokenString)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token key?"})
			c.Abort() // Important to stop further execution
			return
		}

		// Store claims in request context
		c.Set(UserCtxKey, claims)
		c.Next()
	}
}

func ValidateJWT(tokenStr string) (jwt.MapClaims, error) {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	var jwtSecret = []byte(os.Getenv("SECRET_KEY"))
	fmt.Println("JWT Secret:", string(jwtSecret)) // Debug

	fmt.Println("Validating Token:", tokenStr)

	token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
		fmt.Println("JWT Secret:", string(jwtSecret)) // Debug
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return jwtSecret, nil
	})

	if err != nil {
		fmt.Println("JWT Parsing Error:", err)
		return nil, errors.New("invalid token")
	}

	if !token.Valid {
		fmt.Println("Token is Invalid")
		return nil, errors.New("invalid token")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		fmt.Println("Invalid Claims")
		return nil, errors.New("invalid claims in token")
	}

	fmt.Println("JWT Claims:", claims)
	return claims, nil
}
