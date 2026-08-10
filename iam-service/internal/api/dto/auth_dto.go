package dto

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type LoginResponse struct {
	AccessToken string `json:"access_token"`
	// Roles: TODOS los roles activos y vigentes del usuario, no solo el
	// más reciente. Puede venir como [] si no tiene ninguno.
	Roles         []string `json:"roles"`
	HasActiveRole bool     `json:"has_active_role"`
}
type ForgotPasswordRequest struct {
	Email string `json:"email" binding:"required,email"`
}

type ResetPasswordRequest struct {
	Token       string `json:"token" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=8"`
}