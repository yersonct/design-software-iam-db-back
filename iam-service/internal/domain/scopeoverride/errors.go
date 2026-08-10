package scopeoverride

import "errors"

var (
	ErrOverrideAlreadyExists = errors.New("scope_override_already_exists")
	ErrOverrideNotFound      = errors.New("scope_override_not_found")
	ErrUserOrFeatureNotFound = errors.New("user_or_feature_not_found")
	ErrInvalidExpiresAt      = errors.New("invalid_expires_at")
	ErrInvalidScopeType      = errors.New("invalid_scope_type")
	ErrReasonRequired        = errors.New("reason_required")
)