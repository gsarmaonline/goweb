package authentication

import (
	"context"

	"gorm.io/gorm"
)

type (
	Authentication struct {
		ctx context.Context
		db  *gorm.DB
	}
)

func NewAuthentication(ctx context.Context, db *gorm.DB) (auth *Authentication, err error) {
	auth = &Authentication{
		ctx: ctx,
		db:  db,
	}
	return
}

func (auth *Authentication) RegisterModels(db *gorm.DB) (err error) {
	err = db.AutoMigrate(&SessionUser{}, &Session{})
	return
}
