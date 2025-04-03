package authentication

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func TestCreateToken(t *testing.T) {
	secretKey := []byte("test-secret-key")
	user := &SessionUser{Email: "test@example.com"}
	user.ID = 123

	session, err := NewSession(secretKey, user, "127.0.0.1", "test-agent")
	if err != nil {
		t.Fatalf("Failed to create session: %v", err)
	}

	if session.Token == "" {
		t.Error("Expected token to be set")
	}
}

func TestParseToken(t *testing.T) {
	secretKey := []byte("test-secret-key")
	user := &SessionUser{Email: "test@example.com"}
	user.ID = 123

	tests := []struct {
		name        string
		setupToken  func() (*Session, error)
		expectError error
	}{
		{
			name: "valid token",
			setupToken: func() (*Session, error) {
				return NewSession(secretKey, user, "127.0.0.1", "test-agent")
			},
			expectError: nil,
		},
		{
			name: "expired token",
			setupToken: func() (*Session, error) {
				session, err := NewSession(secretKey, user, "127.0.0.1", "test-agent")
				if err != nil {
					return nil, err
				}
				// Manually create an expired token
				claims := claims{
					UserID: user.ID,
					RegisteredClaims: jwt.RegisteredClaims{
						ExpiresAt: jwt.NewNumericDate(time.Now().Add(-time.Hour)),
						IssuedAt:  jwt.NewNumericDate(time.Now().Add(-time.Hour)),
						NotBefore: jwt.NewNumericDate(time.Now().Add(-time.Hour)),
					},
				}
				token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
				tokenString, err := token.SignedString(secretKey)
				if err != nil {
					return nil, err
				}
				session.Token = tokenString
				session.ExpiresAt = time.Now().Add(-time.Hour)
				return session, nil
			},
			expectError: errExpiredToken,
		},
		{
			name: "invalid secret",
			setupToken: func() (*Session, error) {
				session, err := NewSession([]byte("wrong-secret"), user, "127.0.0.1", "test-agent")
				if err != nil {
					return nil, err
				}
				session.SecretKey = secretKey // Switch back to correct key for validation
				return session, nil
			},
			expectError: errInvalidToken,
		},
		{
			name: "malformed token",
			setupToken: func() (*Session, error) {
				session, err := NewSession(secretKey, user, "127.0.0.1", "test-agent")
				if err != nil {
					return nil, err
				}
				session.Token = "malformed.token.string"
				return session, nil
			},
			expectError: errInvalidToken,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			session, err := tt.setupToken()
			if err != nil {
				t.Fatalf("Failed to setup token: %v", err)
			}

			_, err = session.parseToken()
			if err != tt.expectError {
				t.Errorf("Expected error %v, got %v", tt.expectError, err)
			}
		})
	}
}

func TestValidateToken(t *testing.T) {
	secretKey := []byte("test-secret-key")
	user := &SessionUser{Email: "test@example.com"}
	user.ID = 123

	tests := []struct {
		name        string
		setupToken  func() (*Session, error)
		expectError error
	}{
		{
			name: "valid token",
			setupToken: func() (*Session, error) {
				return NewSession(secretKey, user, "127.0.0.1", "test-agent")
			},
			expectError: nil,
		},
		{
			name: "expired token",
			setupToken: func() (*Session, error) {
				session, err := NewSession(secretKey, user, "127.0.0.1", "test-agent")
				if err != nil {
					return nil, err
				}
				// Manually create an expired token
				claims := claims{
					UserID: user.ID,
					RegisteredClaims: jwt.RegisteredClaims{
						ExpiresAt: jwt.NewNumericDate(time.Now().Add(-time.Hour)),
						IssuedAt:  jwt.NewNumericDate(time.Now().Add(-time.Hour)),
						NotBefore: jwt.NewNumericDate(time.Now().Add(-time.Hour)),
					},
				}
				token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
				tokenString, err := token.SignedString(secretKey)
				if err != nil {
					return nil, err
				}
				session.Token = tokenString
				session.ExpiresAt = time.Now().Add(-time.Hour)
				return session, nil
			},
			expectError: errExpiredToken,
		},
		{
			name: "invalid token",
			setupToken: func() (*Session, error) {
				session, err := NewSession(secretKey, user, "127.0.0.1", "test-agent")
				if err != nil {
					return nil, err
				}
				session.Token = "invalid.token.string"
				return session, nil
			},
			expectError: errInvalidToken,
		},
		{
			name: "wrong secret",
			setupToken: func() (*Session, error) {
				session, err := NewSession([]byte("wrong-secret"), user, "127.0.0.1", "test-agent")
				if err != nil {
					return nil, err
				}
				session.SecretKey = secretKey // Switch back to correct key for validation
				return session, nil
			},
			expectError: errInvalidToken,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			session, err := tt.setupToken()
			if err != nil {
				t.Fatalf("Failed to setup token: %v", err)
			}

			err = session.validateToken()
			if err != tt.expectError {
				t.Errorf("Expected error %v, got %v", tt.expectError, err)
			}
		})
	}
}

func TestUpdateLastUsed(t *testing.T) {
	secretKey := []byte("test-secret-key")
	user := &SessionUser{Email: "test@example.com"}
	user.ID = 123

	session, err := NewSession(secretKey, user, "127.0.0.1", "test-agent")
	if err != nil {
		t.Fatalf("Failed to create session: %v", err)
	}

	ip := "192.168.1.1"
	loc := "Mozilla/5.0"
	session.UpdateLastUsed(ip, loc)

	if session.LastUsedIP != ip {
		t.Errorf("Expected LastUsedIP to be %s, got %s", ip, session.LastUsedIP)
	}

	if session.LastUsedLoc != loc {
		t.Errorf("Expected LastUsedLoc to be %s, got %s", loc, session.LastUsedLoc)
	}

	if session.LastUsedAt.IsZero() {
		t.Error("Expected LastUsedAt to be set")
	}
}
