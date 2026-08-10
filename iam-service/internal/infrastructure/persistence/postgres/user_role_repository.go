package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/lib/pq"

	"github.com/yersonct/iam-service/internal/domain/userrole"
)

type UserRoleRepository struct {
	db *sql.DB
}

func NewUserRoleRepository(db *sql.DB) *UserRoleRepository {
	return &UserRoleRepository{db: db}
}

func (r *UserRoleRepository) Assign(ctx context.Context, ur *userrole.UserRole) error {
	const query = `
		INSERT INTO rbac.user_role (user_id, role_id, training_center_id, assigned_by, expires_at)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, assigned_at
	`
	err := r.db.QueryRowContext(ctx, query,
		ur.UserID, ur.RoleID, ur.TrainingCenterID, ur.AssignedBy, ur.ExpiresAt,
	).Scan(&ur.ID, &ur.AssignedAt)

	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) {
			switch pqErr.Code {
			case "23505":
				return userrole.ErrUserRoleAlreadyExists
			case "23503":
				return userrole.ErrUserOrRoleNotFound
			}
		}
		return fmt.Errorf("assign user role: %w", err)
	}
	return nil
}

// Remove borra la fila exacta. trainingCenterID nil -> rol global (columna
// IS NULL); no nil -> ese centro específico. Así un usuario puede tener el
// mismo rol en 2 centros y borrar uno sin afectar el otro.
func (r *UserRoleRepository) Remove(ctx context.Context, userID, roleID string, trainingCenterID *string) error {
	var query string
	var args []interface{}

	if trainingCenterID == nil {
		query = `DELETE FROM rbac.user_role WHERE user_id = $1 AND role_id = $2 AND training_center_id IS NULL`
		args = []interface{}{userID, roleID}
	} else {
		query = `DELETE FROM rbac.user_role WHERE user_id = $1 AND role_id = $2 AND training_center_id = $3`
		args = []interface{}{userID, roleID, *trainingCenterID}
	}

	result, err := r.db.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("remove user role: %w", err)
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("check removed user role: %w", err)
	}
	if rows == 0 {
		return userrole.ErrUserRoleNotFound
	}
	return nil
}

func (r *UserRoleRepository) ListByUser(ctx context.Context, userID string) ([]userrole.UserRoleItem, error) {
	const query = `
		SELECT ur.id, ur.role_id, ro.name, ro.display_name,
		       ur.training_center_id, ur.assigned_by, ur.assigned_at, ur.expires_at
		FROM rbac.user_role ur
		INNER JOIN rbac.role ro ON ro.id = ur.role_id
		WHERE ur.user_id = $1
		ORDER BY ur.assigned_at DESC
	`
	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("list user roles: %w", err)
	}
	defer rows.Close()

	var items []userrole.UserRoleItem
	for rows.Next() {
		var it userrole.UserRoleItem
		if err := rows.Scan(&it.ID, &it.RoleID, &it.RoleName, &it.RoleDisplayName,
			&it.TrainingCenterID, &it.AssignedBy, &it.AssignedAt, &it.ExpiresAt); err != nil {
			return nil, fmt.Errorf("scan user role row: %w", err)
		}
		items = append(items, it)
	}
	return items, rows.Err()
}

// HasActiveRole: soporta el middleware RequireActiveRole (ver abajo).
func (r *UserRoleRepository) HasActiveRole(ctx context.Context, userID string, roleNames []string) (bool, error) {
	if len(roleNames) == 0 {
		return false, nil
	}
	const query = `
		SELECT EXISTS (
			SELECT 1
			FROM rbac.user_role ur
			INNER JOIN rbac.role ro ON ro.id = ur.role_id
			WHERE ur.user_id = $1
			  AND ro.name = ANY($2)
			  AND (ur.expires_at IS NULL OR ur.expires_at > now())
		)
	`
	var exists bool
	err := r.db.QueryRowContext(ctx, query, userID, pq.Array(roleNames)).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("check active role: %w", err)
	}
	return exists, nil
}