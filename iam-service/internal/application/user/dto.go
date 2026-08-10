package user

// CreateUserInput es lo que recibe el caso de uso, ya desacoplado del
// transporte HTTP (el handler traduce el DTO de la API a esto).
type CreateUserInput struct {
	Email     string
	FirstName string
	LastName  string
	ActorType string
	ActorID   *string // nil si no aplica (p. ej. actor_type = USER)
}

type CreateUserOutput struct {
	ID        string
	Email     string
	FirstName string
	LastName  string
	ActorType string
	ActorID   *string
	IsActive  bool

	// EmailSent indica si el correo de bienvenida (con la contraseña
	// temporal) salió correctamente. false no significa que la creación
	// falló -- el usuario sí quedó creado -- solo que el admin debe
	// avisarle al usuario que use "olvidé mi contraseña" como plan B.
	EmailSent bool
}