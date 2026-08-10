package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/lib/pq"

	"github.com/yersonct/iam-service/internal/domain/catalog"
)

type FeatureRepository struct {
	db *sql.DB
}

func NewFeatureRepository(db *sql.DB) *FeatureRepository {
	return &FeatureRepository{db: db}
}

func (r *FeatureRepository) Create(ctx context.Context, f *catalog.Feature) error {
	const query = `
		INSERT INTO rbac_catalog.feature (module_id, code, name, description, action_level)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, is_active, created_at, updated_at
	`

	err := r.db.QueryRowContext(ctx, query, f.ModuleID, f.Code, f.Name, f.Description, string(f.ActionLevel)).
		Scan(&f.ID, &f.IsActive, &f.CreatedAt, &f.UpdatedAt)

	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) {
			switch pqErr.Code {
			case "23505":
				return catalog.ErrFeatureCodeAlreadyExists
			case "23503": // FK violation: module_id no existe
				return catalog.ErrModuleNotFound
			case "23514": // CHECK violation: action_level fuera del enum (defensa final)
				return catalog.ErrInvalidActionLevel
			}
		}
		return fmt.Errorf("create feature: %w", err)
	}

	return nil
}

func (r *FeatureRepository) Update(
	ctx context.Context,
	id string,
	name string,
	description *string,
	actionLevel catalog.ActionLevel,
	isActive bool,
) error {
	const query = `
		UPDATE rbac_catalog.feature
		SET
			name = $1,
			description = $2,
			action_level = $3,
			is_active = $4,
			updated_at = now()
		WHERE id = $5
	`

	result, err := r.db.ExecContext(ctx, query, name, description, string(actionLevel), isActive, id)
	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) && pqErr.Code == "23514" {
			return catalog.ErrInvalidActionLevel
		}
		return fmt.Errorf("update feature: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("check updated feature: %w", err)
	}
	if rowsAffected == 0 {
		return catalog.ErrFeatureNotFound
	}

	return nil
}

func (r *FeatureRepository) FindByID(ctx context.Context, id string) (*catalog.Feature, error) {
	const query = `
		SELECT id, module_id, code, name, description, action_level, is_active, created_at, updated_at
		FROM rbac_catalog.feature
		WHERE id = $1
	`

	var f catalog.Feature
	var actionLevel string
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&f.ID, &f.ModuleID, &f.Code, &f.Name, &f.Description, &actionLevel, &f.IsActive, &f.CreatedAt, &f.UpdatedAt,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, catalog.ErrFeatureNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("find feature by id: %w", err)
	}

	f.ActionLevel = catalog.ActionLevel(actionLevel)
	return &f, nil
}

func (r *FeatureRepository) List(ctx context.Context) ([]catalog.Feature, error) {
	return r.queryFeatures(ctx, `
		SELECT id, module_id, code, name, description, action_level, is_active, created_at, updated_at
		FROM rbac_catalog.feature
		ORDER BY code ASC
	`)
}

func (r *FeatureRepository) ListByModuleID(ctx context.Context, moduleID string) ([]catalog.Feature, error) {
	return r.queryFeatures(ctx, `
		SELECT id, module_id, code, name, description, action_level, is_active, created_at, updated_at
		FROM rbac_catalog.feature
		WHERE module_id = $1
		ORDER BY code ASC
	`, moduleID)
}

func (r *FeatureRepository) queryFeatures(ctx context.Context, query string, args ...any) ([]catalog.Feature, error) {
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list features: %w", err)
	}
	defer rows.Close()

	var features []catalog.Feature
	for rows.Next() {
		var f catalog.Feature
		var actionLevel string
		if err := rows.Scan(
			&f.ID, &f.ModuleID, &f.Code, &f.Name, &f.Description, &actionLevel, &f.IsActive, &f.CreatedAt, &f.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan feature row: %w", err)
		}
		f.ActionLevel = catalog.ActionLevel(actionLevel)
		features = append(features, f)
	}

	return features, rows.Err()
}

func (r *FeatureRepository) ExistsByCode(ctx context.Context, code string) (bool, error) {
	const query = `SELECT EXISTS(SELECT 1 FROM rbac_catalog.feature WHERE code = $1)`

	var exists bool
	if err := r.db.QueryRowContext(ctx, query, code).Scan(&exists); err != nil {
		return false, fmt.Errorf("check feature code existence: %w", err)
	}

	return exists, nil
}

func (r *FeatureRepository) ModuleExists(ctx context.Context, moduleID string) (bool, error) {
	const query = `SELECT EXISTS(SELECT 1 FROM rbac_catalog.module WHERE id = $1)`

	var exists bool
	if err := r.db.QueryRowContext(ctx, query, moduleID).Scan(&exists); err != nil {
		return false, fmt.Errorf("check module existence: %w", err)
	}

	return exists, nil
}
func (r *FeatureRepository) FindByCode(ctx context.Context, code string) (*catalog.Feature, error) {
	const query = `
		SELECT id, module_id, code, name, description, action_level, is_active, created_at, updated_at
		FROM rbac_catalog.feature
		WHERE code = $1
	`

	var f catalog.Feature
	var actionLevel string
	err := r.db.QueryRowContext(ctx, query, code).Scan(
		&f.ID, &f.ModuleID, &f.Code, &f.Name, &f.Description, &actionLevel, &f.IsActive, &f.CreatedAt, &f.UpdatedAt,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, catalog.ErrFeatureNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("find feature by code: %w", err)
	}

	f.ActionLevel = catalog.ActionLevel(actionLevel)
	return &f, nil
}