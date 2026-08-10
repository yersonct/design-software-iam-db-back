package dto

// AssignFeatureRequest: scope_type validado con binding:"oneof=..." --
// mismo patrón que action_level en catalog_dto.go. Esta es la primera
// línea de defensa del criterio de aceptación "scope_type solo acepta
// los valores del enum"; la segunda línea está en el dominio
// (ScopeType.IsValid()).
type AssignFeatureRequest struct {
	FeatureID string `json:"feature_id" binding:"required,uuid"`
	ScopeType string `json:"scope_type" binding:"required,oneof=GLOBAL TRAINING_CENTER AREA OWN_FICHAS OWN_SCHEDULE OWN_PROFILE OWN_FICHA_AS_LEARNER"`
}

// RoleFeatureResponse: usado por List() para la matriz módulos -> features.
type RoleFeatureResponse struct {
	ID          string `json:"id"`
	FeatureID   string `json:"feature_id"`
	FeatureCode string `json:"feature_code"`
	FeatureName string `json:"feature_name"`
	ModuleID    string `json:"module_id"`
	ModuleCode  string `json:"module_code"`
	ModuleName  string `json:"module_name"`
	ScopeType   string `json:"scope_type"`
}

// --- Batch (guardado por lote) ---

type FeatureAssignmentItem struct {
	FeatureID string `json:"feature_id" binding:"required,uuid"`
	ScopeType string `json:"scope_type" binding:"required,oneof=GLOBAL TRAINING_CENTER AREA OWN_FICHAS OWN_SCHEDULE OWN_PROFILE OWN_FICHA_AS_LEARNER"`
}

type BatchAssignFeaturesRequest struct {
	Features []FeatureAssignmentItem `json:"features" binding:"required,dive"`
}