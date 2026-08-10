package audit

import (
	"context"
	"time"
)

type LoginAttempt struct {
	ID             string
	EmailAttempted string
	Outcome        string // SUCCESS | INVALID_PASSWORD | USER_NOT_FOUND | ACCOUNT_LOCKED
	AttemptedAt    time.Time
}

type ListFilter struct {
	Email    *string
	Success  *bool // true = solo SUCCESS, false = solo fallidos, nil = todos
	From     *time.Time
	To       *time.Time
	Page     int
	PageSize int
}

type ListResult struct {
	Items    []LoginAttempt
	Total    int
	Page     int
	PageSize int
}

// Repository: solo lectura -- respeta el carácter append-only de la tabla.
// No hay Update ni Delete a propósito.
type Repository interface {
	ListLoginAttempts(ctx context.Context, filter ListFilter) (*ListResult, error)
}