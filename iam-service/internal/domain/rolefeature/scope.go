package rolefeature

// ScopeType es el enum de rbac.role_feature.scope_type. Se valida tanto en
// el DTO (binding:"oneof=...") como acá en el dominio -- igual que
// domain/catalog.ActionLevel -- para cumplir el criterio de aceptación
// "scope_type solo acepta los valores del enum" sin depender únicamente
// del binding HTTP.
type ScopeType string

const (
	ScopeGlobal          ScopeType = "GLOBAL"
	ScopeTrainingCenter  ScopeType = "TRAINING_CENTER"
	ScopeArea            ScopeType = "AREA"
	ScopeOwnFichas       ScopeType = "OWN_FICHAS"
	ScopeOwnSchedule     ScopeType = "OWN_SCHEDULE"
	ScopeOwnProfile      ScopeType = "OWN_PROFILE"
	ScopeOwnFichaLearner ScopeType = "OWN_FICHA_AS_LEARNER"
)

// AllScopeTypes se usa para validar y también se puede exponer en un futuro
// endpoint de catálogo de scopes si el frontend lo necesita.
var AllScopeTypes = []ScopeType{
	ScopeGlobal,
	ScopeTrainingCenter,
	ScopeArea,
	ScopeOwnFichas,
	ScopeOwnSchedule,
	ScopeOwnProfile,
	ScopeOwnFichaLearner,
}

func (s ScopeType) IsValid() bool {
	for _, v := range AllScopeTypes {
		if s == v {
			return true
		}
	}
	return false
}