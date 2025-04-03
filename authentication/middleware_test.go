package authentication

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"gorm.io/gorm"
)

func setupTestSessionManager(t *testing.T) *SessionManager {
	// Set test secret key
	testSecretKey := []byte("test-secret-key")
	os.Setenv("JWT_SECRET_KEY", string(testSecretKey))
	defer os.Unsetenv("JWT_SECRET_KEY")

	// Create test Gin engine
	gin.SetMode(gin.TestMode)
	engine := gin.New()

	// Create session manager
	sessMgr, err := NewSessionManager(context.Background(), &gorm.DB{}, engine)
	if err != nil {
		t.Fatalf("Failed to create session manager: %v", err)
	}

	return sessMgr
}

func TestAuthMiddleware(t *testing.T) {
	sessMgr := setupTestSessionManager(t)
	secretKey := sessMgr.secretKey

	tests := []struct {
		name           string
		setupAuth      func(*http.Request)
		expectedCode   int
		expectedError  string
		expectedUserID uint
	}{
		{
			name: "valid token",
			setupAuth: func(req *http.Request) {
				user := &SessionUser{Email: "test@example.com"}
				user.ID = 123
				session, err := NewSession(secretKey, user, "127.0.0.1", "test-agent")
				if err != nil {
					t.Fatalf("Failed to create session: %v", err)
				}
				req.Header.Set("Authorization", bearerSchema+session.Token)
			},
			expectedCode:   http.StatusOK,
			expectedUserID: 123,
		},
		{
			name:          "missing auth header",
			setupAuth:     func(req *http.Request) {},
			expectedCode:  http.StatusUnauthorized,
			expectedError: "Authorization header is required",
		},
		{
			name: "invalid auth schema",
			setupAuth: func(req *http.Request) {
				req.Header.Set("Authorization", "Basic token123")
			},
			expectedCode:  http.StatusUnauthorized,
			expectedError: "Authorization header must start with 'Bearer'",
		},
		{
			name: "empty token",
			setupAuth: func(req *http.Request) {
				req.Header.Set("Authorization", bearerSchema)
			},
			expectedCode:  http.StatusUnauthorized,
			expectedError: "Token is required",
		},
		{
			name: "invalid token",
			setupAuth: func(req *http.Request) {
				req.Header.Set("Authorization", bearerSchema+"invalid.token.string")
			},
			expectedCode:  http.StatusUnauthorized,
			expectedError: "Invalid token",
		},
		{
			name: "expired token",
			setupAuth: func(req *http.Request) {
				user := &SessionUser{Email: "test@example.com"}
				user.ID = 123
				// Create claims directly without creating a session
				claims := claims{
					UserID: user.ID,
					RegisteredClaims: jwt.RegisteredClaims{
						ExpiresAt: jwt.NewNumericDate(time.Now().Add(-time.Hour)),
						IssuedAt:  jwt.NewNumericDate(time.Now().Add(-time.Hour)),
						NotBefore: jwt.NewNumericDate(time.Now().Add(-time.Hour)),
					},
				}
				token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
				tokenString, err := token.SignedString(secretKey)
				if err != nil {
					t.Fatalf("Failed to create expired token: %v", err)
				}
				req.Header.Set("Authorization", bearerSchema+tokenString)
			},
			expectedCode:  http.StatusUnauthorized,
			expectedError: "Token has expired",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := gin.New()
			var capturedUserID uint

			router.GET("/test", sessMgr.AuthMiddleware, func(c *gin.Context) {
				capturedUserID = sessMgr.GetUserID(c)
				c.Status(http.StatusOK)
			})

			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			tt.setupAuth(req)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			if w.Code != tt.expectedCode {
				t.Errorf("Expected status code %d, got %d", tt.expectedCode, w.Code)
			}

			if tt.expectedError != "" {
				if w.Body.String() == "" {
					t.Error("Expected error response, got empty body")
				} else if !strings.Contains(w.Body.String(), tt.expectedError) {
					t.Errorf("Expected error message '%s', got '%s'", tt.expectedError, w.Body.String())
				}
			}

			if tt.expectedUserID > 0 && capturedUserID != tt.expectedUserID {
				t.Errorf("Expected user ID %d, got %d", tt.expectedUserID, capturedUserID)
			}
		})
	}
}

func TestGetUserID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	sessMgr := setupTestSessionManager(t)
	c, _ := gin.CreateTestContext(httptest.NewRecorder())

	// Test with no user ID in context
	if id := sessMgr.GetUserID(c); id != 0 {
		t.Errorf("Expected user ID 0 when not set, got %d", id)
	}

	// Test with valid user ID
	expectedID := uint(123)
	c.Set(userKey, expectedID)
	if id := sessMgr.GetUserID(c); id != expectedID {
		t.Errorf("Expected user ID %d, got %d", expectedID, id)
	}

	// Test with invalid type in context
	c.Set(userKey, "invalid")
	if id := sessMgr.GetUserID(c); id != 0 {
		t.Errorf("Expected user ID 0 for invalid type, got %d", id)
	}
}
