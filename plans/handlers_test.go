package plans

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.New(
			log.New(io.Discard, "", log.LstdFlags), // io writer
			logger.Config{
				SlowThreshold:             time.Second,   // Slow SQL threshold
				LogLevel:                  logger.Silent, // Silent mode
				IgnoreRecordNotFoundError: true,          // Ignore record not found error
				Colorful:                  false,         // Disable color
			},
		),
	})
	if err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}

	// Auto migrate the schema
	err = db.AutoMigrate(&Plan{}, &Feature{}, &PlanFeature{})
	if err != nil {
		t.Fatalf("Failed to migrate test database: %v", err)
	}

	return db
}

func createTestPlan(t *testing.T, db *gorm.DB, suffix string) *Plan {
	plan := &Plan{
		Name:        "Test Plan " + suffix,
		Description: "Test Description",
		Price:       9.99,
		Interval:    "monthly",
		IsActive:    true,
	}

	if err := db.Create(plan).Error; err != nil {
		t.Fatalf("Failed to create test plan: %v", err)
	}

	return plan
}

func createTestFeatures(t *testing.T, db *gorm.DB, count int) []Feature {
	features := make([]Feature, count)
	for i := 0; i < count; i++ {
		feature := Feature{
			Name:        "Feature " + string(rune('A'+i)),
			Description: "Description " + string(rune('A'+i)),
			IsActive:    true,
		}
		if err := db.Create(&feature).Error; err != nil {
			t.Fatalf("Failed to create test feature: %v", err)
		}
		features[i] = feature
	}
	return features
}

func setupTestPlanManager(t *testing.T) (*PlanManager, *gorm.DB) {
	db := setupTestDB(t)
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	planManager := NewPlanManager(context.Background(), engine, db)
	return planManager, db
}

func TestGetPlansHandler(t *testing.T) {
	planManager, db := setupTestPlanManager(t)

	// Create test plans
	_ = createTestPlan(t, db, "Active") // active plan
	inactivePlan := createTestPlan(t, db, "Inactive")
	inactivePlan.IsActive = false
	db.Save(inactivePlan)

	tests := []struct {
		name         string
		queryParams  string
		expectedCode int
		checkResult  func(*testing.T, *httptest.ResponseRecorder)
	}{
		{
			name:         "get all plans",
			queryParams:  "",
			expectedCode: http.StatusOK,
			checkResult: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response struct {
					Plans []Plan `json:"plans"`
				}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Len(t, response.Plans, 2)
			},
		},
		{
			name:         "get active plans only",
			queryParams:  "active=true",
			expectedCode: http.StatusOK,
			checkResult: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response struct {
					Plans []Plan `json:"plans"`
				}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Len(t, response.Plans, 1)
				assert.True(t, response.Plans[0].IsActive)
			},
		},
		{
			name:         "invalid active parameter",
			queryParams:  "active=invalid",
			expectedCode: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			req := httptest.NewRequest(http.MethodGet, "/plans?"+tt.queryParams, nil)
			c.Request = req

			planManager.GetPlansHandler(c)

			assert.Equal(t, tt.expectedCode, w.Code)
			if tt.checkResult != nil {
				tt.checkResult(t, w)
			}
		})
	}
}

func TestGetPlanHandler(t *testing.T) {
	planManager, db := setupTestPlanManager(t)

	testPlan := createTestPlan(t, db, "Single")
	features := createTestFeatures(t, db, 2)
	db.Model(testPlan).Association("Features").Replace(features)

	tests := []struct {
		name         string
		planID       string
		expectedCode int
		checkResult  func(*testing.T, *httptest.ResponseRecorder)
	}{
		{
			name:         "valid plan",
			planID:       "1",
			expectedCode: http.StatusOK,
			checkResult: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response struct {
					Plan Plan `json:"plan"`
				}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, testPlan.ID, response.Plan.ID)
				assert.Len(t, response.Plan.Features, 2)
			},
		},
		{
			name:         "plan not found",
			planID:       "999",
			expectedCode: http.StatusNotFound,
		},
		{
			name:         "invalid plan ID",
			planID:       "invalid",
			expectedCode: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.AddParam("id", tt.planID)

			planManager.GetPlanHandler(c)

			assert.Equal(t, tt.expectedCode, w.Code)
			if tt.checkResult != nil {
				tt.checkResult(t, w)
			}
		})
	}
}

func TestUpdatePlanHandler(t *testing.T) {
	planManager, db := setupTestPlanManager(t)

	_ = createTestPlan(t, db, "Update") // Create initial plan
	features := createTestFeatures(t, db, 3)

	newName := "Updated Plan"
	newPrice := 19.99
	newInterval := "yearly"

	tests := []struct {
		name         string
		planID       string
		requestBody  UpdatePlanRequest
		expectedCode int
		checkResult  func(*testing.T, *httptest.ResponseRecorder)
	}{
		{
			name:   "update all fields",
			planID: "1",
			requestBody: UpdatePlanRequest{
				Name:       &newName,
				Price:      &newPrice,
				Interval:   &newInterval,
				FeatureIDs: []uint{features[0].ID, features[1].ID},
			},
			expectedCode: http.StatusOK,
			checkResult: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response struct {
					Plan Plan `json:"plan"`
				}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, newName, response.Plan.Name)
				assert.Equal(t, newPrice, response.Plan.Price)
				assert.Equal(t, newInterval, response.Plan.Interval)
				assert.Len(t, response.Plan.Features, 2)
			},
		},
		{
			name:   "update partial fields",
			planID: "1",
			requestBody: UpdatePlanRequest{
				Name: &newName,
			},
			expectedCode: http.StatusOK,
			checkResult: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response struct {
					Plan Plan `json:"plan"`
				}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, newName, response.Plan.Name)
			},
		},
		{
			name:   "invalid feature IDs",
			planID: "1",
			requestBody: UpdatePlanRequest{
				FeatureIDs: []uint{999},
			},
			expectedCode: http.StatusBadRequest,
		},
		{
			name:         "plan not found",
			planID:       "999",
			requestBody:  UpdatePlanRequest{},
			expectedCode: http.StatusNotFound,
		},
		{
			name:         "invalid plan ID",
			planID:       "invalid",
			requestBody:  UpdatePlanRequest{},
			expectedCode: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.AddParam("id", tt.planID)

			body, _ := json.Marshal(tt.requestBody)
			c.Request = httptest.NewRequest(http.MethodPatch, "/plans/"+tt.planID, bytes.NewBuffer(body))
			c.Request.Header.Set("Content-Type", "application/json")

			planManager.UpdatePlanHandler(c)

			assert.Equal(t, tt.expectedCode, w.Code)
			if tt.checkResult != nil {
				tt.checkResult(t, w)
			}
		})
	}
}
