package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/lib/pq"

	"github.com/yersonct/iam-service/internal/domain/catalog"
)

type ModuleRepository struct {
	db *sql.DB
}

func NewModuleRepository(db *sql.DB) *ModuleRepository {
	return &ModuleRepository{db: db}
}

func (r *ModuleRepository) Create(ctx context.Context, m *catalog.Module) error {
	const query = `
		INSERT INTO rbac_catalog.module (code, name, description, display_order, icon_key)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, is_active, created_at, updated_at
	`

	err := r.db.QueryRowContext(ctx, query, m.Code, m.Name, m.Description, m.DisplayOrder, m.IconKey).
		Scan(&m.ID, &m.IsActive, &m.CreatedAt, &m.UpdatedAt)

	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) && pqErr.Code == "23505" {
			return catalog.ErrModuleCodeAlreadyExists
		}
		return fmt.Errorf("create module: %w", err)
	}

	return nil
}

func (r *ModuleRepository) Update(
	ctx context.Context,
	id string,
	name string,
	description *string,
	displayOrder int16,
	iconKey *string,
	isActive bool,
) error {
	const query = `
		UPDATE rbac_catalog.module
		SET
			name = $1,
			description = $2,
			display_order = $3,
			icon_key = $4,
			is_active = $5,
			updated_at = now()
		WHERE id = $6
	`

	result, err := r.db.ExecContext(ctx, query, name, description, displayOrder, iconKey, isActive, id)
	if err != nil {
		return fmt.Errorf("update module: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("check updated module: %w", err)
	}
	if rowsAffected == 0 {
		return catalog.ErrModuleNotFound
	}

	return nil
}

func (r *ModuleRepository) FindByID(ctx context.Context, id string) (*catalog.Module, error) {
	const query = `
		SELECT id, code, name, description, display_order, icon_key, is_active, created_at, updated_at
		FROM rbac_catalog.module
		WHERE id = $1
	`

	var m catalog.Module
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&m.ID, &m.Code, &m.Name, &m.Description, &m.DisplayOrder, &m.IconKey, &m.IsActive, &m.CreatedAt, &m.UpdatedAt,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, catalog.ErrModuleNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("find module by id: %w", err)
	}

	return &m, nil
}

func (r *ModuleRepository) List(ctx context.Context) ([]catalog.Module, error) {
	const query = `
		SELECT id, code, name, description, display_order, icon_key, is_active, created_at, updated_at
		FROM rbac_catalog.module
		ORDER BY display_order ASC
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("list modules: %w", err)
	}
	defer rows.Close()

	var modules []catalog.Module
	for rows.Next() {
		var m catalog.Module
		if err := rows.Scan(
			&m.ID, &m.Code, &m.Name, &m.Description, &m.DisplayOrder, &m.IconKey, &m.IsActive, &m.CreatedAt, &m.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan module row: %w", err)
		}
		modules = append(modules, m)
	}

	return modules, rows.Err()
}

func (r *ModuleRepository) ExistsByCode(ctx context.Context, code string) (bool, error) {
	const query = `SELECT EXISTS(SELECT 1 FROM rbac_catalog.module WHERE code = $1)`

	var exists bool
	if err := r.db.QueryRowContext(ctx, query, code).Scan(&exists); err != nil {
		return false, fmt.Errorf("check module code existence: %w", err)
	}

	return exists, nil
}