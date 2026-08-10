package passwordreset

import "time"

type PasswordResetRequest struct {
	ID          string
	UserID      string
	TokenHash   string
	ExpiresAt   time.Time
	IsUsed      bool
	RequestedAt time.Time
	IPAddress   string
}