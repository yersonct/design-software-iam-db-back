package dto

// CreateUserRequest es el payload que espera POST /users.
//
// actor_id es opcional a propósito: solo tiene sentido cuando ya existe
// el actor correspondiente en actors-service (instructor/aprendiz ya
// registrado ahí). Si no se manda, queda NULL en la base de datos y se
// puede vincular después.
type CreateUserRequest struct {
	Email     string  `json:"email" binding:"required,email"`
	FirstName string  `json:"first_name" binding:"required,max=100"`
	LastName  string  `json:"last_name" binding:"required,max=100"`
	ActorType string  `json:"actor_type" binding:"required,oneof=USER INSTRUCTOR LEARNER"`
	ActorID   *string `json:"actor_id"`
}

// CreateUserResponse devuelve los datos del usuario creado. La
// contraseña temporal NUNCA viaja en esta respuesta -- se envía
// directamente al correo del usuario. email_sent le dice al frontend
// si ese correo salió bien, para poder avisarle al admin si no.
type CreateUserResponse struct {
	ID        string  `json:"id"`
	Email     string  `json:"email"`
	FirstName string  `json:"first_name"`
	LastName  string  `json:"last_name"`
	ActorType string  `json:"actor_type"`
	ActorID   *string `json:"actor_id"`
	IsActive  bool    `json:"is_active"`
	EmailSent bool    `json:"email_sent"`
}
type UserSummaryResponse struct {
	ID        string  `json:"id"`
	Email     string  `json:"email"`
	FirstName string  `json:"first_name"`
	LastName  string  `json:"last_name"`
	ActorType string  `json:"actor_type"`
	ActorID   *string `json:"actor_id"`
	IsActive  bool    `json:"is_active"`
	IsLocked  bool    `json:"is_locked"`
	CreatedAt string  `json:"created_at"`
}

type RoleResponse struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	DisplayName string `json:"display_name"`
}

type UserDetailResponse struct {
	UserSummaryResponse
	Roles []RoleResponse `json:"roles"`
}

type ListUsersResponse struct {
	Data       []UserSummaryResponse `json:"data"`
	Page       int                    `json:"page"`
	PageSize   int                    `json:"page_size"`
	Total      int                    `json:"total"`
	TotalPages int                    `json:"total_pages"`


}


type UpdateUserRequest struct {
	Email     string `json:"email" binding:"required,email"`
	FirstName string `json:"first_name" binding:"required,max=100"`
	LastName  string `json:"last_name" binding:"required,max=100"`
}

type SetUserStatusRequest struct {
	IsActive bool `json:"is_active"`
}