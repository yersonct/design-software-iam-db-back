package jwtvalidator

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// GinMiddleware es un adaptador OPCIONAL para servicios que también usan
// Gin. Si un microservicio usa otro framework, usa Validate() directamente
// y arma su propio middleware -- Validator no depende de Gin en absoluto.
func GinMiddleware(v *Validator) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if header == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "missing_token", "message": "Se requiere autenticación.",
			})
			return
		}
		parts := strings.SplitN(header, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "invalid_authorization_header", "message": "El header Authorization debe tener el formato: Bearer <token>.",
			})
			return
		}

		claims, err := v.Validate(parts[1])
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "invalid_token", "message": "El token no es válido o expiró.",
			})
			return
		}

		c.Set("user_id", claims.UserID)
		c.Set("roles", claims.Roles)
		c.Set("role", claims.Role) // compat retro: primer rol, o ""
		c.Next()
	}
}