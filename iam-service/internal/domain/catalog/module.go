package catalog

import (
	"errors"
	"time"
)

type Module struct {
	ID           string
	Code         string
	Name         string
	Description  *string
	DisplayOrder int16
	IconKey      *string
	IsActive     bool
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

var (
	ErrModuleNotFound         = errors.New("module_not_found")
	ErrModuleCodeAlreadyExists = errors.New("module_code_already_exists")
)