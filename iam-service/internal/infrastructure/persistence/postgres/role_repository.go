package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/lib/pq"

	"github.com/yersonct/iam-service/internal/domain/role"
)

type RoleRepository struct {
	db *sql.DB
}

func NewRoleRepository(db *sql.DB) *RoleRepository {
	return &RoleRepository{db: db}
}

func (r *RoleRepository) FindByUserID(ctx context.Context, userID string) ([]*role.Role, error) {
	const query = `
		SELECT
			r.id,
			r.name,
			r.display_name,
			r.description,
			r.is_system_role,
			r.created_at
		FROM rbac.role r
		INNER JOIN rbac.user_role ur ON ur.role_id = r.id
		WHERE ur.user_id = $1
			AND (ur.expires_at IS NULL OR ur.expires_at > now())
		ORDER BY r.name
	`

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("find roles by user id: %w", err)
	}
	defer rows.Close()

	var roles []*role.Role

	for rows.Next() {
		var rl role.Role
		var description sql.NullString

		if err := rows.Scan(
			&rl.ID, &rl.Name, &rl.DisplayName, &description, &rl.IsSystemRole, &rl.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan role row: %w", err)
		}

		rl.Description = description.String
		roles = append(roles, &rl)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate role rows: %w", err)
	}

	return roles, nil
}

// Create inserta un nuevo rol. is_system_role siempre se crea en FALSE:
// los roles de sistema solo existen por seed (001_seed_rbac.sql), nunca
// se crean desde la API — así que ni siquiera lo recibimos en el input.
func (r *RoleRepository) Create(ctx context.Context, rl *role.Role) error {
	const query = `
		INSERT INTO rbac.role (name, display_name, description, is_system_role)
		VALUES ($1, $2, $3, FALSE)
		RETURNING id, is_system_role, created_at
	`

	err := r.db.QueryRowContext(ctx, query, rl.Name, rl.DisplayName, rl.Description).
		Scan(&rl.ID, &rl.IsSystemRole, &rl.CreatedAt)

	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) && pqErr.Code == "23505" {
			return role.ErrRoleNameAlreadyExists
		}
		return fmt.Errorf("create role: %w", err)
	}

	return nil
}

// Update solo toca display_name y description. name (identificador) e
// is_system_role nunca se actualizan desde acá — el caso de uso ya
// verificó antes de llamar a este método que el rol no es de sistema.
func (r *RoleRepository) Update(ctx context.Context, id string, displayName string, description string) error {
	const query = `
		UPDATE rbac.role
		SET display_name = $1,
		    description  = $2
		WHERE id = $3
	`

	result, err := r.db.ExecContext(ctx, query, displayName, description, id)
	if err != nil {
		return fmt.Errorf("update role: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("check updated role: %w", err)
	}
	if rowsAffected == 0 {
		return role.ErrRoleNotFound
	}

	return nil
}

// Delete es un borrado físico real (rbac.role no tiene is_active).
// Si el rol tiene usuarios asignados (rbac.user_role), la FK
// fk_user_role_role (ON DELETE RESTRICT) hace que Postgres rechace el
// DELETE con un error 23503 (foreign_key_violation).
func (r *RoleRepository) Delete(ctx context.Context, id string) error {
	const query = `DELETE FROM rbac.role WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) && pqErr.Code == "23503" {
			return role.ErrRoleInUse
		}
		return fmt.Errorf("delete role: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("check deleted role: %w", err)
	}
	if rowsAffected == 0 {
		return role.ErrRoleNotFound
	}

	return nil
}

func (r *RoleRepository) FindByID(ctx context.Context, id string) (*role.Role, error) {
	const query = `
		SELECT id, name, display_name, description, is_system_role, created_at
		FROM rbac.role
		WHERE id = $1
	`

	var rl role.Role
	var description sql.NullString

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&rl.ID, &rl.Name, &rl.DisplayName, &description, &rl.IsSystemRole, &rl.CreatedAt,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, role.ErrRoleNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("find role by id: %w", err)
	}

	rl.Description = description.String
	return &rl, nil
}

func (r *RoleRepository) List(ctx context.Context) ([]role.Role, error) {
	const query = `
		SELECT id, name, display_name, description, is_system_role, created_at
		FROM rbac.role
		ORDER BY is_system_role DESC, name ASC
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("list roles: %w", err)
	}
	defer rows.Close()

	var roles []role.Role
	for rows.Next() {
		var rl role.Role
		var description sql.NullString

		if err := rows.Scan(
			&rl.ID, &rl.Name, &rl.DisplayName, &description, &rl.IsSystemRole, &rl.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan role row: %w", err)
		}

		rl.Description = description.String
		roles = append(roles, rl)
	}

	return roles, rows.Err()
}

func (r *RoleRepository) ExistsByName(ctx context.Context, name string) (bool, error) {
	const query = `SELECT EXISTS(SELECT 1 FROM rbac.role WHERE name = $1)`

	var exists bool
	if err := r.db.QueryRowContext(ctx, query, name).Scan(&exists); err != nil {
		return false, fmt.Errorf("check role name existence: %w", err)
	}

	return exists, nil
}