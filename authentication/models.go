package authentication

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/gsarmaonline/goweb/core"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
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

		Email    string `json:"email" gorm:"uniqueIndex;not null"`
		Password string `json:"password,omitempty" gorm:"not null"`
	}

	Session struct {
		core.BaseModel

		User      *SessionUser `json:"-" gorm:"foreignKey:UserID"`
		UserID    uint         `json:"user_id" gorm:"not null"`
		Token     string       `json:"token" gorm:"-"`
		SecretKey []byte       `json:"-" gorm:"-"`

		ExpiresAt   time.Time `json:"expires_at"`
		LastUsedAt  time.Time `json:"last_used_at"`
		LastUsedIP  string    `json:"last_used_ip"`
		LastUsedLoc string    `json:"last_used_loc"`
	}
)

// BeforeSave hook for SessionUser to hash password before saving
func (u *SessionUser) BeforeSave(tx *gorm.DB) error {
	if u.Password == "" {
		return nil // Skip if password is empty (e.g., when updating other fields)
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	u.Password = string(hashedPassword)
	return nil
}

// ComparePassword compares the given password with the hashed password
func (u *SessionUser) ComparePassword(password string) error {
	return bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password))
}

// NewSession creates and initializes a new session for the user
func NewSession(secretKey []byte, user *SessionUser, clientIP, userAgent string) (*Session, error) {
	session := &Session{
		SecretKey: secretKey,
		User:      user,
		UserID:    user.ID,
	}

	// Create JWT token
	claims := claims{
		UserID: user.ID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(defaultTokenDuration)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(secretKey)
	if err != nil {
		return nil, err
	}

	session.Token = tokenString
	session.ExpiresAt = time.Now().Add(defaultTokenDuration)

	// Update session with client info
	session.UpdateLastUsed(clientIP, userAgent)
	return session, nil
}

// InitializeSession creates a new session for the user with token and tracking info
func (s *Session) InitializeSession(user *SessionUser, clientIP, userAgent string) error {
	s.User = user
	s.UserID = user.ID

	// Create JWT token
	if err := s.createToken(defaultTokenDuration); err != nil {
		return err
	}

	// Update session with client info
	s.UpdateLastUsed(clientIP, userAgent)
	return nil
}
