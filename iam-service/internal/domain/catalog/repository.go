package catalog

import "context"

type ModuleRepository interface {
	Create(ctx context.Context, m *Module) error

	// Update edita name, description, display_order, icon_key, is_active.
	// code NO es editable por diseño (ver nota de la historia).
	Update(ctx context.Context, id string, name string, description *string, displayOrder int16, iconKey *string, isActive bool) error

	FindByID(ctx context.Context, id string) (*Module, error)
	List(ctx context.Context) ([]Module, error)
	ExistsByCode(ctx context.Context, code string) (bool, error)
}

type FeatureRepository interface {
	Create(ctx context.Context, f *Feature) error

	// Update edita name, description, action_level, is_active.
	// code y module_id NO son editables por diseño.
	Update(ctx context.Context, id string, name string, description *string, actionLevel ActionLevel, isActive bool) error

	FindByID(ctx context.Context, id string) (*Feature, error)
		FindByCode(ctx context.Context, code string) (*Feature, error)
	List(ctx context.Context) ([]Feature, error)
	ListByModuleID(ctx context.Context, moduleID string) ([]Feature, error)
	ExistsByCode(ctx context.Context, code string) (bool, error)
	ModuleExists(ctx context.Context, moduleID string) (bool, error)
}