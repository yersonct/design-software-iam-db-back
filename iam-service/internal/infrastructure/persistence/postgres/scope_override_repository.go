package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/lib/pq"

	"github.com/yersonct/iam-service/internal/domain/scopeoverride"
)

type ScopeOverrideRepository struct {
	db *sql.DB
}

func NewScopeOverrideRepository(db *sql.DB) *ScopeOverrideRepository {
	return &ScopeOverrideRepository{db: db}
}

func (r *ScopeOverrideRepository) Create(ctx context.Context, o *scopeoverride.ScopeOverride) error {
	const query = `
		INSERT INTO rbac.user_scope_override
			(user_id, feature_id, scope_type, is_allowed, reason, granted_by, expires_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, created_at
	`
	err := r.db.QueryRowContext(ctx, query,
		o.UserID, o.FeatureID, o.ScopeType, o.IsAllowed, o.Reason, o.GrantedBy, o.ExpiresAt,
	).Scan(&o.ID, &o.CreatedAt)

	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) {
			switch pqErr.Code {
			case "23505":
				return scopeoverride.ErrOverrideAlreadyExists
			case "23503":
				return scopeoverride.ErrUserOrFeatureNotFound
			case "23502":
				return scopeoverride.ErrReasonRequired
			}
		}
		return fmt.Errorf("create scope override: %w", err)
	}
	return nil
}

func (r *ScopeOverrideRepository) Remove(ctx context.Context, id string) error {
	const query = `DELETE FROM rbac.user_scope_override WHERE id = $1`
	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("remove scope override: %w", err)
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("check removed scope override: %w", err)
	}
	if rows == 0 {
		return scopeoverride.ErrOverrideNotFound
	}
	return nil
}

func (r *ScopeOverrideRepository) ListByUser(ctx context.Context, userID string) ([]scopeoverride.ScopeOverrideItem, error) {
	const query = `
		SELECT so.id, so.feature_id, f.code, f.name,
		       so.scope_type, so.is_allowed, so.reason,
		       so.granted_by, so.created_at, so.expires_at
		FROM rbac.user_scope_override so
		INNER JOIN rbac_catalog.feature f ON f.id = so.feature_id
		WHERE so.user_id = $1
		ORDER BY so.created_at DESC
	`
	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("list scope overrides: %w", err)
	}
	defer rows.Close()

	var items []scopeoverride.ScopeOverrideItem
	for rows.Next() {
		var it scopeoverride.ScopeOverrideItem
		if err := rows.Scan(&it.ID, &it.FeatureID, &it.FeatureName, &it.FeatureDisplayName,
			&it.ScopeType, &it.IsAllowed, &it.Reason,
			&it.GrantedBy, &it.CreatedAt, &it.ExpiresAt); err != nil {
			return nil, fmt.Errorf("scan scope override row: %w", err)
		}
		items = append(items, it)
	}
	return items, rows.Err()
}

// FindActiveOverride: si hay más de un override vigente para el mismo
// user+feature+scope (ej. uno permanente de una prueba vieja y uno nuevo
// con expiración), ORDER BY created_at DESC asegura que gane el más
// reciente, no uno arbitrario elegido por Postgres. El filtro de
// expiración sigue resuelto en la propia query (WHERE expires_at IS NULL
// OR expires_at > now()), que es el corazón del criterio "override
// expirado deja de aplicar automáticamente".
func (r *ScopeOverrideRepository) FindActiveOverride(ctx context.Context, userID, featureID, scopeType string) (*scopeoverride.ScopeOverride, error) {
	const query = `
		SELECT id, user_id, feature_id, scope_type, is_allowed, reason, granted_by, created_at, expires_at
		FROM rbac.user_scope_override
		WHERE user_id = $1
		  AND feature_id = $2
		  AND scope_type = $3
		  AND (expires_at IS NULL OR expires_at > now())
		ORDER BY created_at DESC
		LIMIT 1
	`
	var o scopeoverride.ScopeOverride
	err := r.db.QueryRowContext(ctx, query, userID, featureID, scopeType).Scan(
		&o.ID, &o.UserID, &o.FeatureID, &o.ScopeType, &o.IsAllowed, &o.Reason, &o.GrantedBy, &o.CreatedAt, &o.ExpiresAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("find active scope override: %w", err)
	}
	return &o, nil
}