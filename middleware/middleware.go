package middleware

import (
	"errors"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

const UserCtxKey = "user"

func JWTAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header missing"})
			c.Abort()
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")

		claims, err := ValidateJWT(tokenString)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token: " + err.Error()})
			c.Abort()
			return
		}

		// Store claims in request context
		c.Set(UserCtxKey, claims)
		c.Next()
	}
}

func ValidateJWT(tokenStr string) (jwt.MapClaims, error) {
	jwtSecret := []byte(os.Getenv("SECRET_KEY"))
	
	// Log token length for debugging without exposing the full token
	log.Printf("Validating token (length: %d)", len(tokenStr))
	
	// Log secret key length for debugging without exposing the full secret
	log.Printf("Using secret key (length: %d)", len(jwtSecret))

	token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return jwtSecret, nil
	})

	if err != nil {
		log.Printf("JWT Parsing Error: %v", err)
		return nil, err
	}

	if !token.Valid {
		return nil, errors.New("invalid token")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, errors.New("invalid claims in token")
	}

	// Log claim details for debugging without exposing sensitive information
	log.Printf("JWT Claims successfully validated for user_id: %v", claims["user_id"])
	return claims, nil
}
