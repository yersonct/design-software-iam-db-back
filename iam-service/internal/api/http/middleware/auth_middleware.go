package middleware

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"

	"github.com/yersonct/iam-service/internal/application/authz" // ← agregar
)

// TokenValidator es el contrato mínimo que necesita este middleware.
// security.JWTGenerator ya lo implementa (método Validate).
type TokenValidator interface {
	Validate(tokenString string) (*jwt.Token, error)
}
type UserRoleChecker interface {
	HasActiveRole(ctx context.Context, userID string, roleNames []string) (bool, error)
}

var (
	ErrMissingToken = errors.New("missing authorization token")
	ErrInvalidToken = errors.New("invalid or expired token")
)

// RequireAuth valida el JWT del header "Authorization: Bearer <token>".
// Si es válido, deja disponibles en el contexto:
//   - "user_id" (claim "sub")
//   - "roles"   (claim "roles", []string con todos los roles vigentes)
//   - "role"    (compat: primer elemento de "roles", o "" si no tiene ninguno)
func RequireAuth(validator TokenValidator) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if header == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error":   "missing_token",
				"message": "Se requiere autenticación.",
			})
			return
		}

		parts := strings.SplitN(header, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error":   "invalid_authorization_header",
				"message": "El header Authorization debe tener el formato: Bearer <token>.",
			})
			return
		}

		token, err := validator.Validate(parts[1])
		if err != nil || token == nil || !token.Valid {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error":   "invalid_token",
				"message": "El token no es válido o expiró.",
			})
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error":   "invalid_token",
				"message": "El token no es válido.",
			})
			return
		}

		userID, _ := claims["sub"].(string)

		// El claim "roles" llega como []interface{} tras el parseo JSON
		// del JWT (jwt.MapClaims), no como []string directamente.
		var roles []string
		if rawRoles, ok := claims["roles"].([]interface{}); ok {
			roles = make([]string, 0, len(rawRoles))
			for _, rr := range rawRoles {
				if s, ok := rr.(string); ok {
					roles = append(roles, s)
				}
			}
		}

		var firstRole string
		if len(roles) > 0 {
			firstRole = roles[0]
		}

		c.Set("user_id", userID)
		c.Set("roles", roles)
		c.Set("role", firstRole) // compat: código legado que aún lea "role"

		c.Next()
	}
}

// RequireRole exige que RequireAuth se haya ejecutado antes (para tener
// "roles" en el contexto) y que AL MENOS UNO de los roles del usuario esté
// en la lista permitida (antes comparaba un único rol por igualdad exacta,
// lo que rompía a usuarios con más de un rol activo).
func RequireRole(allowedRoles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		rolesValue, exists := c.Get("roles")
		if !exists {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error":   "missing_token",
				"message": "Se requiere autenticación.",
			})
			return
		}

		roles, _ := rolesValue.([]string)

		for _, userRole := range roles {
			for _, allowed := range allowedRoles {
				if userRole == allowed {
					c.Next()
					return
				}
			}
		}

		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
			"error":   "forbidden",
			"message": "No tienes permisos para realizar esta acción.",
		})
	}
}

func RequireActiveRole(repo UserRoleChecker, allowedRoles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userIDValue, exists := c.Get("user_id")
		if !exists {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "missing_token", "message": "Se requiere autenticación.",
			})
			return
		}
		userID, _ := userIDValue.(string)

		active, err := repo.HasActiveRole(c.Request.Context(), userID, allowedRoles)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "internal_error"})
			return
		}
		if !active {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error": "forbidden", "message": "No tienes un rol vigente que permita esta acción.",
			})
			return
		}
		c.Next()
	}
}
// RequireFeature valida el permiso vía RBAC real (rol + overrides), no por
// nombre de rol hardcodeado. featureID se resuelve UNA VEZ al arrancar el
// servicio (ver main.go), no en cada request.
func RequireFeature(checkUC *authz.CheckPermissionUseCase, featureID string, scopeType string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userIDValue, exists := c.Get("user_id")
		if !exists {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "missing_token", "message": "Se requiere autenticación.",
			})
			return
		}
		userID, _ := userIDValue.(string)

		allowed, err := checkUC.Execute(c.Request.Context(), authz.CheckPermissionInput{
			UserID: userID, FeatureID: featureID, ScopeType: scopeType,
		})
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "internal_error"})
			return
		}
		if !allowed {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error": "forbidden", "message": "No tienes permiso para realizar esta acción.",
			})
			return
		}
		c.Next()
	}
}