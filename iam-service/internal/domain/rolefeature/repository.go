package rolefeature

import "context"

type Repository interface {
	Assign(ctx context.Context, rf *RoleFeature) error
	Remove(ctx context.Context, roleID, featureID string) error
	ListByRole(ctx context.Context, roleID string) ([]RoleFeatureItem, error)
	ReplaceAll(ctx context.Context, roleID string, items []RoleFeature) error // NUEVO
}