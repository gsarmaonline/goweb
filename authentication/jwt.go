package authentication

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

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
