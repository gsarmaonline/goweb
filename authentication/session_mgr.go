package authentication

import (
	"context"
	"errors"
	"os"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type (
	SessionManager struct {
		ctx       context.Context
		db        *gorm.DB
		apiEngine *gin.Engine

		secretKey []byte
	}
)

func NewSessionManager(ctx context.Context, db *gorm.DB, apiEngine *gin.Engine) (sessionMgr *SessionManager, err error) {
	secretKey := []byte(os.Getenv("JWT_SECRET_KEY"))
	if len(secretKey) == 0 {
		return nil, errors.New("JWT_SECRET_KEY environment variable is not set")
	}

	sessionMgr = &SessionManager{
		ctx:       ctx,
		db:        db,
		apiEngine: apiEngine,
		secretKey: secretKey,
	}
	return
}

func (sessionMgr *SessionManager) RegisterModels(db *gorm.DB) (err error) {
	err = db.AutoMigrate(&SessionUser{}, &Session{})
	return
}
