package session

import "time"

type Session struct {
	ID               string
	UserID           string
	RefreshTokenHash string
	ExpiresAt        time.Time
	IsRevoked        bool
	CreatedAt        time.Time
}
