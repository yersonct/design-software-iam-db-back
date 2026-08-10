package trainingcenter

import domaintc "github.com/yersonct/iam-service/internal/domain/trainingcenter"

// Catálogo quemado de Centros de Formación SENA reales (nombre/ciudad),
// mientras no exista un microservicio académico que los administre.
// IDs fijos y estables (no gen_random_uuid) para que no cambien entre
// despliegues -- rbac.user_role.training_center_id no tiene FK en BD,
// así que la única validación posible de existencia es contra este catálogo.
var centers = []domaintc.TrainingCenter{
	{ID: "a1000000-0000-0000-0000-000000000001", Code: "CFA-BOG-01", Name: "Centro de Gestión de Mercados, Logística y Tecnologías de la Información", City: "Bogotá D.C.", Department: "Cundinamarca"},
	{ID: "a1000000-0000-0000-0000-000000000002", Code: "CFA-BOG-02", Name: "Centro de Electricidad, Electrónica y Telecomunicaciones", City: "Bogotá D.C.", Department: "Cundinamarca"},
	{ID: "a1000000-0000-0000-0000-000000000003", Code: "CFA-CUN-01", Name: "Centro de Biotecnología Agropecuaria", City: "Mosquera", Department: "Cundinamarca"},
	{ID: "a1000000-0000-0000-0000-000000000004", Code: "CFA-MED-01", Name: "Centro de Comercio", City: "Medellín", Department: "Antioquia"},
	{ID: "a1000000-0000-0000-0000-000000000005", Code: "CFA-ANT-01", Name: "Centro Agropecuario La Salada", City: "Caldas", Department: "Antioquia"},
	{ID: "a1000000-0000-0000-0000-000000000006", Code: "CFA-CAL-01", Name: "Centro de Tecnología de la Manufactura Avanzada", City: "Cali", Department: "Valle del Cauca"},
	{ID: "a1000000-0000-0000-0000-000000000007", Code: "CFA-CAL-02", Name: "Centro de Electricidad y Automatización Industrial", City: "Cali", Department: "Valle del Cauca"},
	{ID: "a1000000-0000-0000-0000-000000000008", Code: "CFA-BAQ-01", Name: "Centro Industrial y de Aviación", City: "Barranquilla", Department: "Atlántico"},
	{ID: "a1000000-0000-0000-0000-000000000009", Code: "CFA-BUC-01", Name: "Centro de Comercio y Servicios", City: "Bucaramanga", Department: "Santander"},
	{ID: "a1000000-0000-0000-0000-00000000000a", Code: "CFA-HUI-01", Name: "Centro de la Industria, la Empresa y los Servicios", City: "Neiva", Department: "Huila"},
}

type StaticRepository struct{}

func NewStaticRepository() *StaticRepository { return &StaticRepository{} }

func (r *StaticRepository) List() []domaintc.TrainingCenter { return centers }

func (r *StaticRepository) Exists(id string) bool {
	for _, c := range centers {
		if c.ID == id {
			return true
		}
	}
	return false
}