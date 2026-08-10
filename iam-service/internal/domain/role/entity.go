package role

import "time"

type Role struct {
	ID           string
	Name         string
	DisplayName  string
	Description  string
	IsSystemRole bool
	CreatedAt    time.Time
}

// EnsureEditable aplica la regla de negocio de la historia de usuario:
// un rol de sistema (is_system_role = true) no se puede editar/renombrar.
// Se valida en el dominio, no solo en la base de datos, porque la tabla
// rbac.role no tiene ningún CHECK ni trigger que lo impida.
func (r *Role) EnsureEditable() error {
	if r.IsSystemRole {
		return ErrSystemRoleNotEditable
	}
	return nil
}

// EnsureDeletable aplica la misma regla para el borrado.
func (r *Role) EnsureDeletable() error {
	if r.IsSystemRole {
		return ErrSystemRoleNotDeletable
	}
	return nil
}