package rolefeature

// RoleFeature representa una fila de rbac.role_feature: qué feature puede
// usar un rol y con qué alcance (scope_type). La tabla no tiene created_at.
type RoleFeature struct {
	ID        string
	RoleID    string
	FeatureID string
	ScopeType ScopeType
}

// RoleFeatureItem es una fila "enriquecida" para la matriz completa del rol
// (GET /roles/{id}/features): incluye datos de la feature y su módulo
// (que viven en el schema rbac_catalog) para que el frontend pueda armar
// el checklist módulos -> features sin tener que cruzar con /features aparte.
type RoleFeatureItem struct {
	ID          string
	FeatureID   string
	FeatureCode string
	FeatureName string
	ModuleID    string
	ModuleCode  string
	ModuleName  string
	ScopeType   ScopeType
}