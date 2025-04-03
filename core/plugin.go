package core

import "gorm.io/gorm"

type (
	Plugin interface {
		RegisterModels(*gorm.DB) error
	}
)
