package plans

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type UpdatePlanRequest struct {
	Name        *string  `json:"name,omitempty"`
	Description *string  `json:"description,omitempty"`
	Price       *float64 `json:"price,omitempty"`
	Interval    *string  `json:"interval,omitempty"`
	IsActive    *bool    `json:"is_active,omitempty"`
	FeatureIDs  []uint   `json:"feature_ids,omitempty"`
}

// GetPlansHandler returns all active plans with their features
func (pm *PlanManager) GetPlansHandler(c *gin.Context) {
	var plans []Plan
	query := pm.db.Preload("Features", "is_active = ?", true)

	// Filter by active status if specified
	if active := c.Query("active"); active != "" {
		isActive, err := strconv.ParseBool(active)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid active parameter"})
			return
		}
		query = query.Where("is_active = ?", isActive)
	}

	if err := query.Find(&plans).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch plans"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"plans": plans})
}

// GetPlanHandler returns a single plan by ID
func (pm *PlanManager) GetPlanHandler(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid plan ID"})
		return
	}

	var plan Plan
	err = pm.db.Preload("Features", "is_active = ?", true).First(&plan, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "plan not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch plan"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"plan": plan})
}

// UpdatePlanHandler updates an existing plan
func (pm *PlanManager) UpdatePlanHandler(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid plan ID"})
		return
	}

	var req UpdatePlanRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Start a transaction
	tx := pm.db.Begin()
	if tx.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to start transaction"})
		return
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Fetch existing plan
	var plan Plan
	if err := tx.First(&plan, id).Error; err != nil {
		tx.Rollback()
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "plan not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch plan"})
		return
	}

	// Update fields if provided
	if req.Name != nil {
		plan.Name = *req.Name
	}
	if req.Description != nil {
		plan.Description = *req.Description
	}
	if req.Price != nil {
		plan.Price = *req.Price
	}
	if req.Interval != nil {
		plan.Interval = *req.Interval
	}
	if req.IsActive != nil {
		plan.IsActive = *req.IsActive
	}

	// Update plan
	if err := tx.Save(&plan).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update plan"})
		return
	}

	// Update features if provided
	if req.FeatureIDs != nil {
		// Clear existing features
		if err := tx.Model(&plan).Association("Features").Clear(); err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to clear features"})
			return
		}

		// Add new features
		var features []Feature
		if err := tx.Where("id IN ?", req.FeatureIDs).Find(&features).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch features"})
			return
		}

		if len(features) != len(req.FeatureIDs) {
			tx.Rollback()
			c.JSON(http.StatusBadRequest, gin.H{"error": "one or more features not found"})
			return
		}

		if err := tx.Model(&plan).Association("Features").Replace(features); err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update features"})
			return
		}
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to commit transaction"})
		return
	}

	// Fetch updated plan with features
	var updatedPlan Plan
	if err := pm.db.Preload("Features").First(&updatedPlan, id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch updated plan"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"plan": updatedPlan})
}
