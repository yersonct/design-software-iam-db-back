package user

import "context"

// Solo el CONTRATO.
// La implementación real vive en infrastructure/persistence/postgres.
type Repository interface {
	FindByEmail(ctx context.Context, email string) (*User, error)
	FindByID(ctx context.Context, id string) (*User, error)
	Update(ctx context.Context, u *User) error
	UpdatePassword(ctx context.Context, userID string, passwordHash string) error

	// Create inserta un nuevo usuario. Debe retornar ErrEmailAlreadyExists
	// si el email (case-insensitive) ya existe.
	Create(ctx context.Context, u *User) error

	// ExistsByEmail permite validar unicidad ANTES de intentar el insert,
	// para poder devolver un 409 claro sin depender de parsear el error
	// de constraint de Postgres.
	ExistsByEmail(ctx context.Context, email string) (bool, error)
	List(ctx context.Context, filter UserFilter, limit, offset int) ([]*User, int, error)
	// UpdateProfile edita nombre y correo. Debe retornar ErrEmailAlreadyExists
	// si el nuevo email ya pertenece a otro usuario.
	UpdateProfile(ctx context.Context, id string, email string, firstName string, lastName string) error

	// SetActive activa o desactiva la cuenta.
	SetActive(ctx context.Context, id string, isActive bool) error

	// Unlock limpia locked_until y reinicia failed_attempts a 0.
	Unlock(ctx context.Context, id string) error
}
// UserFilter agrupa los filtros opcionales de List. Un puntero nil
// significa "sin filtrar por este campo".
type UserFilter struct {
	ActorType *ActorType
	IsActive  *bool
}