package scopeoverride

import "context"

type Repository interface {
	Create(ctx context.Context, o *ScopeOverride) error
	Remove(ctx context.Context, id string) error
	ListByUser(ctx context.Context, userID string) ([]ScopeOverrideItem, error)

	// FindActiveOverride: usado por el motor de autorización (application/authz).
	// Busca un override vigente (no vencido) para (user_id, feature_id, scope_type).
	FindActiveOverride(ctx context.Context, userID, featureID, scopeType string) (*ScopeOverride, error)
}