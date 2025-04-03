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

// createToken generates a new JWT token for the session
func (s *Session) createToken(expirationTime time.Duration) error {
	if s.User == nil {
		return errors.New("session user not set")
	}

	claims := claims{
		UserID: s.User.ID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(expirationTime)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(s.SecretKey)
	if err != nil {
		return err
	}

	s.Token = tokenString
	s.ExpiresAt = time.Now().Add(expirationTime)
	return nil
}

// parseToken validates and parses the JWT token
func (s *Session) parseToken() (*claims, error) {
	token, err := jwt.ParseWithClaims(s.Token, &claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errInvalidToken
		}
		return s.SecretKey, nil
	})

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, errExpiredToken
		}
		return nil, errInvalidToken
	}

	claims, ok := token.Claims.(*claims)
	if !ok || !token.Valid {
		return nil, errInvalidToken
	}

	return claims, nil
}

// validateToken checks if the session token is valid
func (s *Session) validateToken() error {
	_, err := s.parseToken()
	return err
}

// UpdateLastUsed updates the last used timestamp and location information
func (s *Session) UpdateLastUsed(ip, location string) {
	s.LastUsedAt = time.Now()
	s.LastUsedIP = ip
	s.LastUsedLoc = location
}
