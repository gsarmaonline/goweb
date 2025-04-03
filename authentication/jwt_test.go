package authentication

import (
	"testing"
	"time"
)

var testSecretKey = []byte("test-secret-key")

func TestCreateToken(t *testing.T) {
	session := NewSession(testSecretKey)
	session.User = &SessionUser{
		Email: "test@example.com",
	}
	session.User.ID = 123
	expirationTime := 1 * time.Hour

	err := session.createToken(expirationTime)
	if err != nil {
		t.Errorf("createToken failed: %v", err)
	}
	if session.Token == "" {
		t.Error("createToken returned empty token")
	}

	// Verify the token can be parsed
	claims, err := session.parseToken()
	if err != nil {
		t.Errorf("Failed to parse created token: %v", err)
	}
	if claims.UserID != session.User.ID {
		t.Errorf("Expected UserID %d, got %d", session.User.ID, claims.UserID)
	}
}

func TestParseToken(t *testing.T) {
	defaultSession := NewSession(testSecretKey)
	defaultSession.User = &SessionUser{Email: "test@example.com"}
	defaultSession.User.ID = 123

	wrongSession := NewSession([]byte("wrong-secret"))
	wrongSession.User = &SessionUser{Email: "test@example.com"}
	wrongSession.User.ID = 123

	tests := []struct {
		name    string
		setup   func() *Session
		wantErr error
		wantID  uint
	}{
		{
			name: "valid token",
			setup: func() *Session {
				s := defaultSession
				s.createToken(time.Hour)
				return s
			},
			wantErr: nil,
			wantID:  123,
		},
		{
			name: "expired token",
			setup: func() *Session {
				s := defaultSession
				s.createToken(-time.Hour) // negative duration for expired token
				return s
			},
			wantErr: errExpiredToken,
			wantID:  0,
		},
		{
			name: "invalid secret",
			setup: func() *Session {
				s := defaultSession
				s.createToken(time.Hour)
				wrongSession.Token = s.Token // Copy token to session with wrong secret
				return wrongSession
			},
			wantErr: errInvalidToken,
			wantID:  0,
		},
		{
			name: "malformed token",
			setup: func() *Session {
				s := defaultSession
				s.Token = "malformed.token.string"
				return s
			},
			wantErr: errInvalidToken,
			wantID:  0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			session := tt.setup()
			claims, err := session.parseToken()

			if tt.wantErr != nil {
				if err != tt.wantErr {
					t.Errorf("parseToken() error = %v, wantErr %v", err, tt.wantErr)
				}
				return
			}

			if err != nil {
				t.Errorf("parseToken() unexpected error: %v", err)
				return
			}

			if claims.UserID != tt.wantID {
				t.Errorf("parseToken() got UserID = %v, want %v", claims.UserID, tt.wantID)
			}
		})
	}
}

func TestValidateToken(t *testing.T) {
	defaultSession := NewSession(testSecretKey)
	defaultSession.User = &SessionUser{Email: "test@example.com"}
	defaultSession.User.ID = 123

	wrongSession := NewSession([]byte("wrong-secret"))

	tests := []struct {
		name    string
		setup   func() *Session
		wantErr error
	}{
		{
			name: "valid token",
			setup: func() *Session {
				s := defaultSession
				s.createToken(time.Hour)
				return s
			},
			wantErr: nil,
		},
		{
			name: "expired token",
			setup: func() *Session {
				s := defaultSession
				s.createToken(-time.Hour)
				return s
			},
			wantErr: errExpiredToken,
		},
		{
			name: "invalid token",
			setup: func() *Session {
				s := defaultSession
				s.Token = "invalid.token.string"
				return s
			},
			wantErr: errInvalidToken,
		},
		{
			name: "wrong secret",
			setup: func() *Session {
				s := defaultSession
				s.createToken(time.Hour)
				wrongSession.Token = s.Token // Copy token to session with wrong secret
				return wrongSession
			},
			wantErr: errInvalidToken,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			session := tt.setup()
			err := session.validateToken()

			if tt.wantErr != nil {
				if err != tt.wantErr {
					t.Errorf("validateToken() error = %v, wantErr %v", err, tt.wantErr)
				}
				return
			}

			if err != nil {
				t.Errorf("validateToken() unexpected error: %v", err)
			}
		})
	}
}

func TestUpdateLastUsed(t *testing.T) {
	session := NewSession(testSecretKey)
	ip := "192.168.1.1"
	location := "Test Location"

	beforeUpdate := time.Now()
	session.UpdateLastUsed(ip, location)
	afterUpdate := time.Now()

	if session.LastUsedAt.Before(beforeUpdate) || session.LastUsedAt.After(afterUpdate) {
		t.Error("LastUsedAt not set correctly")
	}

	if session.LastUsedIP != ip {
		t.Errorf("Expected LastUsedIP %s, got %s", ip, session.LastUsedIP)
	}

	if session.LastUsedLoc != location {
		t.Errorf("Expected LastUsedLoc %s, got %s", location, session.LastUsedLoc)
	}
}
