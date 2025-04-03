package authentication

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}

	// Auto migrate the schema
	err = db.AutoMigrate(&SessionUser{}, &Session{})
	if err != nil {
		t.Fatalf("Failed to migrate test database: %v", err)
	}

	return db
}

func TestRegister(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupTestDB(t)
	sessMgr := setupTestSessionManager(t)
	sessMgr.db = db

	tests := []struct {
		name         string
		requestBody  map[string]interface{}
		expectedCode int
	}{
		{
			name: "valid registration",
			requestBody: map[string]interface{}{
				"email":    "test@example.com",
				"password": "password123",
			},
			expectedCode: http.StatusCreated,
		},
		{
			name: "duplicate email",
			requestBody: map[string]interface{}{
				"email":    "test@example.com",
				"password": "password123",
			},
			expectedCode: http.StatusConflict,
		},
		{
			name: "invalid email",
			requestBody: map[string]interface{}{
				"email":    "invalid-email",
				"password": "password123",
			},
			expectedCode: http.StatusBadRequest,
		},
		{
			name: "short password",
			requestBody: map[string]interface{}{
				"email":    "test2@example.com",
				"password": "12345",
			},
			expectedCode: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			body, _ := json.Marshal(tt.requestBody)
			c.Request = httptest.NewRequest(http.MethodPost, "/register", bytes.NewBuffer(body))
			c.Request.Header.Set("Content-Type", "application/json")

			sessMgr.RegisterHandler(c)

			assert.Equal(t, tt.expectedCode, w.Code)
		})
	}
}

func TestLogin(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupTestDB(t)
	sessMgr := setupTestSessionManager(t)
	sessMgr.db = db

	// Create a test user
	testUser := &SessionUser{
		Email:    "test@example.com",
		Password: "password123",
	}
	db.Create(testUser)

	tests := []struct {
		name         string
		requestBody  map[string]interface{}
		expectedCode int
	}{
		{
			name: "valid login",
			requestBody: map[string]interface{}{
				"email":    "test@example.com",
				"password": "password123",
			},
			expectedCode: http.StatusOK,
		},
		{
			name: "invalid password",
			requestBody: map[string]interface{}{
				"email":    "test@example.com",
				"password": "wrongpassword",
			},
			expectedCode: http.StatusUnauthorized,
		},
		{
			name: "non-existent user",
			requestBody: map[string]interface{}{
				"email":    "nonexistent@example.com",
				"password": "password123",
			},
			expectedCode: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			body, _ := json.Marshal(tt.requestBody)
			c.Request = httptest.NewRequest(http.MethodPost, "/login", bytes.NewBuffer(body))
			c.Request.Header.Set("Content-Type", "application/json")

			sessMgr.LoginHandler(c)

			assert.Equal(t, tt.expectedCode, w.Code)

			if tt.expectedCode == http.StatusOK {
				var response LoginResponse
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.NotEmpty(t, response.Session.Token)
				assert.Equal(t, testUser.Email, response.User.Email)
				assert.Empty(t, response.User.Password)
			}
		})
	}
}

func TestLogout(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupTestDB(t)
	sessMgr := setupTestSessionManager(t)
	sessMgr.db = db

	// Create a test user and session
	testUser := &SessionUser{
		Email:    "test@example.com",
		Password: "password123",
	}
	db.Create(testUser)

	session, err := NewSession(sessMgr.secretKey, testUser, "127.0.0.1", "test-agent")
	if err != nil {
		t.Fatalf("Failed to create session: %v", err)
	}
	db.Create(session)

	tests := []struct {
		name         string
		setupAuth    func(*gin.Context)
		expectedCode int
	}{
		{
			name: "successful logout",
			setupAuth: func(c *gin.Context) {
				c.Set(userKey, testUser.ID)
			},
			expectedCode: http.StatusOK,
		},
		{
			name:         "not authenticated",
			setupAuth:    func(c *gin.Context) {},
			expectedCode: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodPost, "/logout", nil)
			tt.setupAuth(c)

			sessMgr.LogoutHandler(c)

			assert.Equal(t, tt.expectedCode, w.Code)

			if tt.expectedCode == http.StatusOK {
				var count int64
				db.Model(&Session{}).Where("user_id = ?", testUser.ID).Count(&count)
				assert.Equal(t, int64(0), count)
			}
		})
	}
}
