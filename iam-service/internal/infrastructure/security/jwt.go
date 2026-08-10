package security

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type JWTGenerator struct {
	secret []byte
}

func NewJWTGenerator(secret string) *JWTGenerator {
	return &JWTGenerator{
		secret: []byte(secret),
	}
}

// GenerateAccessToken genera el JWT que Angular utilizará
// para autenticarse contra los endpoints protegidos.
// El claim "roles" lleva TODOS los roles activos y vigentes del usuario
// (antes era "role", string único con solo el más reciente). Si el
// usuario no tiene ningún rol, roles llega como slice vacío, nunca nil,
// para que el claim serialice como [] y no como null.
func (g *JWTGenerator) GenerateAccessToken(userID string, roles []string) (string, error) {
	now := time.Now()

	if roles == nil {
		roles = []string{}
	}

	claims := jwt.MapClaims{
		"sub":   userID,
		"roles": roles,
		"iat":   now.Unix(),
		"exp":   now.Add(15 * time.Minute).Unix(),
	}

	token := jwt.NewWithClaims(
		jwt.SigningMethodHS256,
		claims,
	)

	return token.SignedString(g.secret)
}

// GenerateRefreshToken genera:
// 1. token plano -> se envía al cliente mediante cookie HttpOnly
// 2. hash -> se guarda en PostgreSQL
func (g *JWTGenerator) GenerateRefreshToken() (plain string, hash string, err error) {
	randomBytes := make([]byte, 32)

	if _, err := rand.Read(randomBytes); err != nil {
		return "", "", err
	}

	plain = hex.EncodeToString(randomBytes)

	sum := sha256.Sum256([]byte(plain))
	hash = hex.EncodeToString(sum[:])

	return plain, hash, nil
}

// Generate se conserva si otras partes del proyecto lo utilizan.
func (g *JWTGenerator) Generate(userID string, expiresIn time.Duration) (string, error) {
	now := time.Now()

	claims := jwt.MapClaims{
		"sub": userID,
		"iat": now.Unix(),
		"exp": now.Add(expiresIn).Unix(),
	}

	token := jwt.NewWithClaims(
		jwt.SigningMethodHS256,
		claims,
	)

	return token.SignedString(g.secret)
}

func (g *JWTGenerator) Validate(tokenString string) (*jwt.Token, error) {
	return jwt.Parse(
		tokenString,
		func(token *jwt.Token) (interface{}, error) {
			if token.Method != jwt.SigningMethodHS256 {
				return nil, jwt.ErrSignatureInvalid
			}

			return g.secret, nil
		},
	)
}
func (g *JWTGenerator) GenerateResetToken() (plain string, hash string, err error) {
	return g.GenerateRefreshToken()
}