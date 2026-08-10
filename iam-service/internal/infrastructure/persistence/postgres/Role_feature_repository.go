package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/lib/pq"

	"github.com/yersonct/iam-service/internal/domain/rolefeature"
)

type RoleFeatureRepository struct {
	db *sql.DB
}

func NewRoleFeatureRepository(db *sql.DB) *RoleFeatureRepository {
	return &RoleFeatureRepository{db: db}
}

// Assign inserta la fila en rbac.role_feature. La tabla no tiene created_at,
// solo id (default gen_random_uuid()).
// - 23505 (unique_violation) -> viola uq_role_feature_role_id_feature_id
//   -> ErrRoleFeatureAlreadyExists (criterio: no duplicar rol+feature).
// - 23503 (foreign_key_violation) -> role_id o feature_id no existen
//   -> ErrRoleOrFeatureNotFound.
func (r *RoleFeatureRepository) Assign(ctx context.Context, rf *rolefeature.RoleFeature) error {
	const query = `
		INSERT INTO rbac.role_feature (role_id, feature_id, scope_type)
		VALUES ($1, $2, $3)
		RETURNING id
	`

	err := r.db.QueryRowContext(ctx, query, rf.RoleID, rf.FeatureID, string(rf.ScopeType)).
		Scan(&rf.ID)

	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) {
			switch pqErr.Code {
			case "23505":
				return rolefeature.ErrRoleFeatureAlreadyExists
			case "23503":
				return rolefeature.ErrRoleOrFeatureNotFound
			}
		}
		return fmt.Errorf("assign role feature: %w", err)
	}

	return nil
}

// Remove borra la asignación rol+feature.
func (r *RoleFeatureRepository) Remove(ctx context.Context, roleID string, featureID string) error {
	const query = `
		DELETE FROM rbac.role_feature
		WHERE role_id = $1 AND feature_id = $2
	`

	result, err := r.db.ExecContext(ctx, query, roleID, featureID)
	if err != nil {
		return fmt.Errorf("remove role feature: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("check removed role feature: %w", err)
	}
	if rowsAffected == 0 {
		return rolefeature.ErrRoleFeatureNotFound
	}

	return nil
}

// ListByRole trae la matriz completa (feature + módulo) asignada al rol,
// para armar el checklist módulos -> features en el frontend. feature y
// module viven en el schema rbac_catalog, no en rbac.
func (r *RoleFeatureRepository) ListByRole(ctx context.Context, roleID string) ([]rolefeature.RoleFeatureItem, error) {
	const query = `
		SELECT
			rf.id,
			f.id   AS feature_id,
			f.code AS feature_code,
			f.name AS feature_name,
			m.id   AS module_id,
			m.code AS module_code,
			m.name AS module_name,
			rf.scope_type
		FROM rbac.role_feature rf
		INNER JOIN rbac_catalog.feature f ON f.id = rf.feature_id
		INNER JOIN rbac_catalog.module m ON m.id = f.module_id
		WHERE rf.role_id = $1
		ORDER BY m.display_order, f.code
	`

	rows, err := r.db.QueryContext(ctx, query, roleID)
	if err != nil {
		return nil, fmt.Errorf("list role features: %w", err)
	}
	defer rows.Close()

	var items []rolefeature.RoleFeatureItem
	for rows.Next() {
		var it rolefeature.RoleFeatureItem
		var scope string

		if err := rows.Scan(
			&it.ID, &it.FeatureID, &it.FeatureCode, &it.FeatureName,
			&it.ModuleID, &it.ModuleCode, &it.ModuleName,
			&scope,
		); err != nil {
			return nil, fmt.Errorf("scan role feature row: %w", err)
		}

		it.ScopeType = rolefeature.ScopeType(scope)
		items = append(items, it)
	}

	return items, rows.Err()
}

// ReplaceAll reemplaza TODO el set de features del rol en una transacción.
// Usado por el guardado por lote (batch) del criterio de aceptación.
func (r *RoleFeatureRepository) ReplaceAll(ctx context.Context, roleID string, items []rolefeature.RoleFeature) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx replace all: %w", err)
	}
	defer tx.Rollback()

	if _, err := tx.ExecContext(ctx, `DELETE FROM rbac.role_feature WHERE role_id = $1`, roleID); err != nil {
		return fmt.Errorf("delete existing role features: %w", err)
	}

	for _, it := range items {
		_, err := tx.ExecContext(ctx, `
			INSERT INTO rbac.role_feature (role_id, feature_id, scope_type)
			VALUES ($1, $2, $3)
		`, roleID, it.FeatureID, string(it.ScopeType))
		if err != nil {
			var pqErr *pq.Error
			if errors.As(err, &pqErr) && pqErr.Code == "23503" {
				return rolefeature.ErrRoleOrFeatureNotFound
			}
			return fmt.Errorf("insert role feature in batch: %w", err)
		}
	}

	return tx.Commit()
}

// HasFeatureViaRole responde si el usuario tiene acceso a featureID con
// exactamente scopeType a través de alguno de sus roles activos (no
// expirados en rbac.user_role). Implementa authz.RoleFeatureChecker por
// duck typing -- no requiere tocar domain/rolefeature.Repository.
//
// Es la mitad "rol" del motor de autorización: CheckPermissionUseCase la
// consulta solo cuando NO hay un override vigente para ese usuario+feature+scope.
func (r *RoleFeatureRepository) HasFeatureViaRole(ctx context.Context, userID, featureID, scopeType string) (bool, error) {
	const query = `
		SELECT EXISTS (
			SELECT 1
			FROM rbac.role_feature rf
			INNER JOIN rbac.user_role ur ON ur.role_id = rf.role_id
			WHERE ur.user_id = $1
			  AND rf.feature_id = $2
			  AND rf.scope_type = $3
			  AND (ur.expires_at IS NULL OR ur.expires_at > now())
		)
	`

	var exists bool
	if err := r.db.QueryRowContext(ctx, query, userID, featureID, scopeType).Scan(&exists); err != nil {
		return false, fmt.Errorf("check feature via role: %w", err)
	}

	return exists, nil
}