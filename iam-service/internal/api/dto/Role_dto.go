package dto

import "time"

// name sigue el patrón de "code" en module/feature: identificador técnico,
// se define al crear y no se puede cambiar después (por eso no aparece
// en UpdateRoleRequest). display_name es lo que ve el usuario final.
type CreateRoleRequest struct {
	Name        string  `json:"name" binding:"required,max=50"`
	DisplayName string  `json:"display_name" binding:"required,max=100"`
	Description *string `json:"description" binding:"omitempty,max=1000"`
}

// No incluye is_system_role: nunca se setea desde la API (ver
// RoleRepository.Create). No incluye name: no es editable.
type UpdateRoleRequest struct {
	DisplayName string  `json:"display_name" binding:"required,max=100"`
	Description *string `json:"description" binding:"omitempty,max=1000"`
}

// RoleCatalogResponse: se llama distinto a RoleResponse (que ya existe
// en user_dto.go, usado dentro de UserDetailResponse.Roles) porque acá
// necesitamos más campos -- is_system_role y created_at -- para el
// catálogo de administración de roles.
type RoleCatalogResponse struct {
	ID           string    `json:"id"`
	Name         string    `json:"name"`
	DisplayName  string    `json:"display_name"`
	Description  *string   `json:"description"`
	IsSystemRole bool      `json:"is_system_role"`
	CreatedAt    time.Time `json:"created_at"`
}