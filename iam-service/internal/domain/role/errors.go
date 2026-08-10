package role

import "errors"

var (
	ErrRoleNotFound          = errors.New("role not found")
	ErrRoleNameAlreadyExists = errors.New("role name already exists")

	// ErrSystemRoleNotEditable / ErrSystemRoleNotDeletable son el corazón
	// del criterio de aceptación: "Intento de editar/eliminar un rol de
	// sistema → rechazado con mensaje explícito".
	ErrSystemRoleNotEditable  = errors.New("system role cannot be edited")
	ErrSystemRoleNotDeletable = errors.New("system role cannot be deleted")

	// ErrRoleInUse: el esquema define fk_user_role_role con ON DELETE
	// RESTRICT, así que Postgres rechaza el DELETE si hay usuarios con
	// ese rol asignado. Lo traducimos a un error de dominio legible.
	ErrRoleInUse = errors.New("role is assigned to one or more users and cannot be deleted")
)