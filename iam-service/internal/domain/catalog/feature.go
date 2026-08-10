package catalog

import (
	"errors"
	"time"
)

type ActionLevel string

const (
	ActionRead    ActionLevel = "READ"
	ActionWrite   ActionLevel = "WRITE"
	ActionDelete  ActionLevel = "DELETE"
	ActionPublish ActionLevel = "PUBLISH"
	ActionApprove ActionLevel = "APPROVE"
)

// Valid replica en Go el CHECK constraint de la base de datos, para poder
// rechazar con un 400 claro antes de siquiera tocar la BD.
func (a ActionLevel) Valid() bool {
	switch a {
	case ActionRead, ActionWrite, ActionDelete, ActionPublish, ActionApprove:
		return true
	default:
		return false
	}
}

type Feature struct {
	ID          string
	ModuleID    string
	Code        string
	Name        string
	Description *string
	ActionLevel ActionLevel
	IsActive    bool
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

var (
	ErrFeatureNotFound          = errors.New("feature_not_found")
	ErrFeatureCodeAlreadyExists = errors.New("feature_code_already_exists")
	ErrInvalidActionLevel       = errors.New("invalid_action_level")
)