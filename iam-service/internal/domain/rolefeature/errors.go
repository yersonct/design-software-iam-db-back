package rolefeature

import "errors"

var (
	// ErrInvalidScopeType: doble validación en dominio, no solo en el
	// binding:"oneof" del DTO. Criterio de aceptación de la historia.
	ErrInvalidScopeType = errors.New("scope_type is not a valid value")

	// ErrRoleOrFeatureNotFound: el rol o la feature referenciados no
	// existen. Se traduce desde la violación de FK en Postgres (23503),
	// igual que ErrRoleInUse se traduce desde una violación de FK en
	// role_repository.go.
	ErrRoleOrFeatureNotFound = errors.New("role or feature not found")

	// ErrRoleFeatureAlreadyExists: viola uq_role_feature_role_id_feature_id.
	// Es el respaldo en dominio/BD del criterio de aceptación "no se puede
	// duplicar la combinación rol+feature" (el frontend además debe
	// deshabilitar la opción si ya está asignada).
	ErrRoleFeatureAlreadyExists = errors.New("this feature is already assigned to this role")

	ErrRoleFeatureNotFound = errors.New("role feature assignment not found")
)