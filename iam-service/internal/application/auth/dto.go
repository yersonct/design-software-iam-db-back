package auth

type LoginInput struct {
	Email    string
	Password string
}

type LoginOutput struct {
	AccessToken   string
	RefreshToken  string
	Roles         []string
	UserID        string
	HasActiveRole bool
}