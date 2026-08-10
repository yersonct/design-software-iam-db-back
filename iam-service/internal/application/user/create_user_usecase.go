package user

import (
	"context"
	"crypto/rand"
	"fmt"
	"math/big"
	"strings"

	domainuser "github.com/yersonct/iam-service/internal/domain/user"
)

// PasswordHasher es el mismo contrato que ya usa auth (security.BcryptHasher
// lo implementa), lo redeclaramos aquí para no acoplar application/user a
// application/auth.
type PasswordHasher interface {
	Hash(plain string) (string, error)
}

// EmailSender para el correo de bienvenida con la contraseña temporal.
// SMTPSender ya implementa esto (método SendWelcomeEmail).
type EmailSender interface {
	SendWelcomeEmail(ctx context.Context, to string, firstName string, tempPassword string, loginURL string) error
}

type CreateUserUseCase struct {
	userRepo domainuser.Repository
	hasher   PasswordHasher
	emailer  EmailSender
	loginURL string
}

func NewCreateUserUseCase(
	userRepo domainuser.Repository,
	hasher PasswordHasher,
	emailer EmailSender,
	loginURL string,
) *CreateUserUseCase {
	return &CreateUserUseCase{
		userRepo: userRepo,
		hasher:   hasher,
		emailer:  emailer,
		loginURL: loginURL,
	}
}

// Execute crea el usuario con una contraseña temporal generada por el
// sistema y la envía por correo. El admin nunca ve la contraseña en
// texto plano en ningún momento — ni en la respuesta HTTP, ni en logs.
//
// Si el correo falla DESPUÉS de crear el usuario, no revertimos la
// creación (el usuario ya existe y es válido); en cambio devolvemos
// EmailSent = false en el output para que el handler pueda avisarle
// al admin que debe usar "olvidé mi contraseña" como plan B.
func (uc *CreateUserUseCase) Execute(ctx context.Context, in CreateUserInput) (*CreateUserOutput, error) {
	actorType := domainuser.ActorType(in.ActorType)
	if !actorType.IsValid() {
		return nil, domainuser.ErrInvalidActorType
	}

	exists, err := uc.userRepo.ExistsByEmail(ctx, in.Email)
	if err != nil {
		return nil, fmt.Errorf("check email existence: %w", err)
	}
	if exists {
		return nil, domainuser.ErrEmailAlreadyExists
	}

	tempPassword, err := generateTempPassword()
	if err != nil {
		return nil, fmt.Errorf("generate temp password: %w", err)
	}

	hash, err := uc.hasher.Hash(tempPassword)
	if err != nil {
		return nil, fmt.Errorf("hash temp password: %w", err)
	}

	newUser := &domainuser.User{
		Email:        in.Email,
		PasswordHash: hash,
		FirstName:    in.FirstName,
		LastName:     in.LastName,
		ActorType:    actorType,
		ActorID:      in.ActorID,
		IsActive:     true, // criterio de aceptación: queda activo por defecto
	}

	if err := uc.userRepo.Create(ctx, newUser); err != nil {
		return nil, fmt.Errorf("create user: %w", err)
	}

	out := &CreateUserOutput{
		ID:        newUser.ID,
		Email:     newUser.Email,
		FirstName: newUser.FirstName,
		LastName:  newUser.LastName,
		ActorType: string(newUser.ActorType),
		ActorID:   newUser.ActorID,
		IsActive:  newUser.IsActive,
		EmailSent: true,
	}

	if err := uc.emailer.SendWelcomeEmail(ctx, newUser.Email, newUser.FirstName, tempPassword, uc.loginURL); err != nil {
		// No revertimos la creación del usuario: el registro ya es válido.
		// Solo marcamos que el correo no salió, para que el handler pueda
		// avisar al admin.
		out.EmailSent = false
	}

	return out, nil
}

// generateTempPassword produce una contraseña temporal legible pero segura:
// 12 caracteres tomados de un alfabeto sin ambigüedades (sin 0/O, 1/l, etc.)
func generateTempPassword() (string, error) {
	const alphabet = "ABCDEFGHJKMNPQRSTUVWXYZabcdefghijkmnpqrstuvwxyz23456789"
	const length = 12

	var sb strings.Builder
	sb.Grow(length)

	for i := 0; i < length; i++ {
		idx, err := rand.Int(rand.Reader, big.NewInt(int64(len(alphabet))))
		if err != nil {
			return "", err
		}
		sb.WriteByte(alphabet[idx.Int64()])
	}

	return sb.String(), nil
}