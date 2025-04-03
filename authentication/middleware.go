package authentication

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

const (
	bearerSchema         = "Bearer "
	userKey              = "user_id"
	defaultTokenDuration = time.Hour * 24 // 24 hours
)

// AuthMiddleware creates a gin middleware for JWT authentication
func (sessMgr *SessionManager) AuthMiddleware(c *gin.Context) {
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
			"error": "Authorization header is required",
		})
		return
	}

	if !strings.HasPrefix(authHeader, bearerSchema) {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
			"error": "Authorization header must start with 'Bearer'",
		})
		return
	}

	tokenString := strings.TrimPrefix(authHeader, bearerSchema)
	if tokenString == "" {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
			"error": "Token is required",
		})
		return
	}

	session := NewSession(sessMgr.secretKey)
	session.Token = tokenString

	claims, err := session.parseToken()
	if err != nil {
		status := http.StatusUnauthorized
		message := "Invalid token"

		if err == errExpiredToken {
			message = "Token has expired"
		}

		c.AbortWithStatusJSON(status, gin.H{
			"error": message,
		})
		return
	}

	// Store user ID in context
	c.Set(userKey, claims.UserID)
	c.Next()
}

// GetUserID retrieves the authenticated user ID from the context
// Returns 0 if no user ID is found in context
func (sessMgr *SessionManager) GetUserID(c *gin.Context) uint {
	if id, exists := c.Get(userKey); exists {
		if userID, ok := id.(uint); ok {
			return userID
		}
	}
	return 0
}
