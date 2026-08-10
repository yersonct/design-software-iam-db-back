// Package jwtvalidator provee la validación LOCAL (sin llamar a
// iam-service) del JWT que iam-service emite. Pensado para ser copiado o
// publicado como módulo Go independiente y usado por scheduling-service,
// academic-management-service, y cualquier otro microservicio del sistema.
//
// Por qué no hay /.well-known/jwks.json: el JWT actual se firma con HS256
// (secreto simétrico, ver internal/infrastructure/security/jwt.go). JWKS
// solo aplica a algoritmos asimétricos (RS256/ES256), donde existe una
// clave PÚBLICA segura de publicar. Con HS256 el mismo secreto sirve para
// firmar y validar -- publicarlo anularía la seguridad del sistema. Por
// eso la validación local se resuelve compartiendo JWT_SECRET por variable
// de entorno entre servicios, no vía un endpoint de claves públicas. Si el
// equipo decide migrar a RS256 en el futuro, ese es el momento de agregar
// JWKS real -- decisión que amerita su propio ADR, no se resuelve aquí.
//
// Uso típico en otro microservicio (ejemplo con Gin):
//
//	validator := jwtvalidator.New(os.Getenv("JWT_SECRET"))
//	router.Use(jwtvalidator.GinMiddleware(validator))
//
// El middleware deja en el contexto:
//   - "user_id" (claim "sub")
//   - "roles"   (claim "roles" -- TODOS los roles vigentes AL MOMENTO DE
//     EMITIR el access token (login o último refresh); no refleja roles
//     asignados/revocados a mitad de la vida del access token -- ese es
//     el motivo de que el access token dure poco (15 min) y el refresh
//     vuelva a consultar roles en vivo. Para el detalle más reciente
//     posible, llamar a GET {IAM_SERVICE_URL}/auth/me).
//   - "role"    (compat retro: primer elemento de "roles", o "" si el
//     usuario no tiene ninguno. Los consumidores nuevos deberían usar
//     "roles" -- este campo se mantiene para no romper integraciones
//     existentes durante la migración del contrato del token).
//
// BREAKING CHANGE (coordinar con todo microservicio que use este paquete):
// el JWT de iam-service dejó de llevar el claim "role" (string único) y
// ahora lleva "roles" (array). Si tu servicio solo lee el campo Role de
// Claims, sigue funcionando (se sigue llenando por compat), pero solo verá
// UN rol aunque el usuario tenga varios activos. Actualiza a Claims.Roles
// en cuanto puedas.
package jwtvalidator

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var (
	ErrMissingToken = errors.New("missing authorization token")
	ErrInvalidToken = errors.New("invalid or expired token")
)

type Claims struct {
	UserID string
	// Roles: todos los roles vigentes del usuario al emitir el token.
	Roles []string
	// Role: compat retro -- primer elemento de Roles, o "" si está vacío.
	// Deprecated: usar Roles.
	Role      string
	IssuedAt  time.Time
	ExpiresAt time.Time
}

type Validator struct {
	secret []byte
}

// New crea un validador con el mismo secreto HS256 que usa iam-service
// para firmar. DEBE ser el mismo valor de JWT_SECRET en ambos servicios.
func New(secret string) *Validator {
	return &Validator{secret: []byte(secret)}
}

// Validate verifica firma y expiración; no hace ninguna llamada de red.
// Por eso es apto para usarse en cada request de los otros microservicios
// sin penalidad de latencia (a diferencia de llamar a /auth/me siempre).
func (v *Validator) Validate(tokenString string) (*Claims, error) {
	token, err := jwt.Parse(tokenString, func(t *jwt.Token) (interface{}, error) {
		if t.Method != jwt.SigningMethodHS256 {
			return nil, jwt.ErrSignatureInvalid
		}
		return v.secret, nil
	})
	if err != nil || token == nil || !token.Valid {
		return nil, ErrInvalidToken
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, ErrInvalidToken
	}

	sub, _ := claims["sub"].(string)

	var roles []string
	if rawRoles, ok := claims["roles"].([]interface{}); ok {
		roles = make([]string, 0, len(rawRoles))
		for _, rr := range rawRoles {
			if s, ok := rr.(string); ok {
				roles = append(roles, s)
			}
		}
	}

	var role string
	if len(roles) > 0 {
		role = roles[0]
	}

	c := &Claims{UserID: sub, Roles: roles, Role: role}
	if iat, ok := claims["iat"].(float64); ok {
		c.IssuedAt = time.Unix(int64(iat), 0)
	}
	if exp, ok := claims["exp"].(float64); ok {
		c.ExpiresAt = time.Unix(int64(exp), 0)
	}
	return c, nil
}