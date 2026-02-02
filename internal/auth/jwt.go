package auth

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/yuanweize/RouteLens/pkg/logging"
	"golang.org/x/crypto/bcrypt"
)

var secretKey []byte

func init() {
	// Load secret from env or generate cryptographically secure random secret
	sk := os.Getenv("RS_JWT_SECRET")
	if sk == "" {
		// SECURITY: Generate random 32-byte secret instead of hardcoded fallback
		// Note: This means all tokens will be invalidated on server restart
		randomBytes := make([]byte, 32)
		if _, err := rand.Read(randomBytes); err != nil {
			// This should never happen, but panic if crypto/rand fails
			panic(fmt.Sprintf("CRITICAL: Failed to generate random JWT secret: %v", err))
		}
		sk = hex.EncodeToString(randomBytes)
		logging.Warn("auth", "RS_JWT_SECRET not set - using random secret (tokens will expire on restart)")
	}
	secretKey = []byte(sk)
}

// GenerateToken creates a new JWT token for the given username
// In single-user mode, username can be any identifier
func GenerateToken(username string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": username,
		"exp": time.Now().Add(24 * time.Hour).Unix(),
	})
	return token.SignedString(secretKey)
}

// AuthMiddleware validates the JWT token and stores user info in context
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required"})
			c.Abort()
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid authorization format"})
			c.Abort()
			return
		}

		tokenString := parts[1]
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return secretKey, nil
		})

		if err != nil || !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
			c.Abort()
			return
		}

		// Extract username from token claims and store in context
		if claims, ok := token.Claims.(jwt.MapClaims); ok {
			if sub, exists := claims["sub"].(string); exists {
				c.Set("username", sub)
			}
		}

		c.Next()
	}
}

// HashPassword hashes a raw password
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}

// ComparePassword compares hashed password with raw input
func ComparePassword(hashed, raw string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashed), []byte(raw))
	return err == nil
}
