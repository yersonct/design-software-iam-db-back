package userrole

import "errors"

var (
	ErrUserRoleAlreadyExists = errors.New("user_role_already_exists")
	ErrUserRoleNotFound      = errors.New("user_role_not_found")
	ErrUserOrRoleNotFound    = errors.New("user_or_role_not_found")
	ErrInvalidTrainingCenter = errors.New("invalid_training_center")
	ErrInvalidExpiresAt      = errors.New("invalid_expires_at")
)