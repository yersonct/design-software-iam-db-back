package userrole

import "context"

type Repository interface {
	Assign(ctx context.Context, ur *UserRole) error
	// Remove borra la fila EXACTA (user_id, role_id, training_center_id).
	// trainingCenterID == nil significa "el rol global, sin centro".
	Remove(ctx context.Context, userID, roleID string, trainingCenterID *string) error
	ListByUser(ctx context.Context, userID string) ([]UserRoleItem, error)

	// HasActiveRole: usado por el middleware de autorización dinámico.
	// "Activa" = existe fila con ese rol y (expires_at IS NULL OR expires_at > now()).
	// Criterio de aceptación: expiración validada en middleware, no solo visual.
	HasActiveRole(ctx context.Context, userID string, roleNames []string) (bool, error)
}