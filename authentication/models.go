package authentication

import (
	"time"

	"github.com/gsarmaonline/goweb/core"
)

type (
	SessionUser struct {
		core.BaseModel

		Email    string `json:"email"`
		Password string `json:"password"`
	}

	Session struct {
		core.BaseModel

		User  *SessionUser `json:"-"`
		Token string       `json:"token"`

		ExpiresAt   time.Time `json:"expires_at"`
		LastUsedAt  time.Time `json:"last_used_at"`
		LastUsedIP  string    `json:"last_used_ip"`
		LastUsedLoc string    `json:"last_used_loc"`
	}
)
