package core

import "gorm.io/gorm"

type (
	BaseModel struct {
		gorm.Model
		OwnedBy string `json:"owned_by"`
	}
)
