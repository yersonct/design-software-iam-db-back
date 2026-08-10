package dto

import "time"

// ---------- Module ----------

type CreateModuleRequest struct {
	Code         string  `json:"code" binding:"required,max=30"`
	Name         string  `json:"name" binding:"required,max=100"`
	Description  *string `json:"description" binding:"omitempty,max=1000"`
	DisplayOrder int16   `json:"display_order" binding:"required,min=1"`
	IconKey      *string `json:"icon_key" binding:"omitempty,max=50"`
}

type UpdateModuleRequest struct {
	Name         string  `json:"name" binding:"required,max=100"`
	Description  *string `json:"description" binding:"omitempty,max=1000"`
	DisplayOrder int16   `json:"display_order" binding:"required,min=1"`
	IconKey      *string `json:"icon_key" binding:"omitempty,max=50"`
	IsActive     *bool   `json:"is_active" binding:"required"`
}

type ModuleResponse struct {
	ID           string    `json:"id"`
	Code         string    `json:"code"`
	Name         string    `json:"name"`
	Description  *string   `json:"description"`
	DisplayOrder int16     `json:"display_order"`
	IconKey      *string   `json:"icon_key"`
	IsActive     bool      `json:"is_active"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// ---------- Feature ----------

// El `oneof` es la primera línea de defensa: rechaza un action_level fuera
// del enum ANTES de llegar al caso de uso o a la base de datos, con un 400
// claro. Esto es lo que pide el criterio de aceptación de la historia.
type CreateFeatureRequest struct {
	ModuleID    string  `json:"module_id" binding:"required,uuid"`
	Code        string  `json:"code" binding:"required,max=60"`
	Name        string  `json:"name" binding:"required,max=120"`
	Description *string `json:"description" binding:"omitempty,max=1000"`
	ActionLevel string  `json:"action_level" binding:"required,oneof=READ WRITE DELETE PUBLISH APPROVE"`
}

type UpdateFeatureRequest struct {
	Name        string  `json:"name" binding:"required,max=120"`
	Description *string `json:"description" binding:"omitempty,max=1000"`
	ActionLevel string  `json:"action_level" binding:"required,oneof=READ WRITE DELETE PUBLISH APPROVE"`
	IsActive    *bool   `json:"is_active" binding:"required"`
}

type FeatureResponse struct {
	ID          string    `json:"id"`
	ModuleID    string    `json:"module_id"`
	Code        string    `json:"code"`
	Name        string    `json:"name"`
	Description *string   `json:"description"`
	ActionLevel string    `json:"action_level"`
	IsActive    bool      `json:"is_active"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}