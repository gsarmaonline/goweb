package authentication

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/gsarmaonline/goweb/core"
)

var (
	errInvalidToken = errors.New("invalid token")
	errExpiredToken = errors.New("token has expired")
)

type claims struct {
	UserID uint `json:"user_id"`
	jwt.RegisteredClaims
}

type (
	SessionUser struct {
		core.BaseModel

		Email    string `json:"email"`
		Password string `json:"password"`
	}

	Session struct {
		core.BaseModel

		User      *SessionUser `json:"-"`
		Token     string       `json:"token"`
		SecretKey []byte       `json:"-"`

		ExpiresAt   time.Time `json:"expires_at"`
		LastUsedAt  time.Time `json:"last_used_at"`
		LastUsedIP  string    `json:"last_used_ip"`
		LastUsedLoc string    `json:"last_used_loc"`
	}
)

// NewSession creates a new session with the given secret key
func NewSession(secretKey []byte) *Session {
	return &Session{
		SecretKey: secretKey,
	}
}
