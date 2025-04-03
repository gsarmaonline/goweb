package plans

import (
	"github.com/gsarmaonline/goweb/core"
	"gorm.io/gorm"
)

type (
	// Plan represents a subscription plan with its features and pricing
	Plan struct {
		core.BaseModel

		Name        string    `json:"name" gorm:"uniqueIndex;not null"`
		Description string    `json:"description"`
		Price       float64   `json:"price" gorm:"not null"`
		Interval    string    `json:"interval" gorm:"not null"` // monthly, yearly, etc.
		IsActive    bool      `json:"is_active" gorm:"default:true"`
		Features    []Feature `json:"features" gorm:"many2many:plan_features"`
	}

	// Feature represents a single feature that can be included in multiple plans
	Feature struct {
		core.BaseModel

		Name        string `json:"name" gorm:"uniqueIndex;not null"`
		Description string `json:"description"`
		IsActive    bool   `json:"is_active" gorm:"default:true"`
		Plans       []Plan `json:"plans" gorm:"many2many:plan_features"`
	}

	// PlanFeature represents the many-to-many relationship between plans and features
	PlanFeature struct {
		PlanID    uint `gorm:"primaryKey"`
		FeatureID uint `gorm:"primaryKey"`
	}
)

// BeforeCreate hook for Plan to validate the interval
func (p *Plan) BeforeCreate(tx *gorm.DB) error {
	// Validate interval
	switch p.Interval {
	case "monthly", "yearly":
		return nil
	default:
		return core.ErrInvalidField{Field: "interval", Message: "must be either 'monthly' or 'yearly'"}
	}
}

// BeforeDelete hook for Plan to prevent deletion if it has active subscriptions
func (p *Plan) BeforeDelete(tx *gorm.DB) error {
	// TODO: Add check for active subscriptions when subscription model is added
	return nil
}

// BeforeDelete hook for Feature to prevent deletion if it's used in any active plans
func (f *Feature) BeforeDelete(tx *gorm.DB) error {
	var count int64
	if err := tx.Model(&Plan{}).
		Joins("JOIN plan_features ON plans.id = plan_features.plan_id").
		Where("plan_features.feature_id = ? AND plans.is_active = ?", f.ID, true).
		Count(&count).Error; err != nil {
		return err
	}

	if count > 0 {
		return core.ErrDeleteForbidden{Message: "feature is used in active plans"}
	}

	return nil
}
