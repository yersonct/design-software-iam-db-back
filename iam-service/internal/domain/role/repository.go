package role

import "context"

type Repository interface {
	// FindByUserID retorna todos los roles activos (no expirados)
	// asignados a un usuario. Se usa en el detalle de usuario y en /users/me.
	FindByUserID(ctx context.Context, userID string) ([]*Role, error)

	// --- CRUD del catálogo de roles (historia "crear y editar roles") ---

	Create(ctx context.Context, r *Role) error
	Update(ctx context.Context, id string, displayName string, description string) error
	Delete(ctx context.Context, id string) error
	FindByID(ctx context.Context, id string) (*Role, error)
	List(ctx context.Context) ([]Role, error)
	ExistsByName(ctx context.Context, name string) (bool, error)
}