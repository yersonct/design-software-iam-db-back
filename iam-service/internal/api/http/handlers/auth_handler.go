package handlers

import (
	"errors"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/yersonct/iam-service/internal/api/dto"
	"github.com/yersonct/iam-service/internal/application/auth"
	"github.com/yersonct/iam-service/internal/domain/user"
)

type AuthHandler struct {
	loginUC          *auth.LoginUseCase
	refreshUC        *auth.RefreshTokenUseCase
	logoutUC         *auth.LogoutUseCase
	forgotPasswordUC *auth.ForgotPasswordUseCase
	resetPasswordUC  *auth.ResetPasswordUseCase
}

func NewAuthHandler(
	loginUC *auth.LoginUseCase,
	refreshUC *auth.RefreshTokenUseCase,
	logoutUC *auth.LogoutUseCase,
	forgotPasswordUC *auth.ForgotPasswordUseCase,
	resetPasswordUC *auth.ResetPasswordUseCase,
) *AuthHandler {
	return &AuthHandler{
		loginUC:          loginUC,
		refreshUC:        refreshUC,
		logoutUC:         logoutUC,
		forgotPasswordUC: forgotPasswordUC,
		resetPasswordUC:  resetPasswordUC,
	}
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req dto.LoginRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_payload",
			"details": err.Error(),
		})
		return
	}

	out, err := h.loginUC.Execute(
		c.Request.Context(),
		auth.LoginInput{
			Email:    req.Email,
			Password: req.Password,
		},
	)

	if err != nil {
		switch {
		case errors.Is(err, user.ErrAccountLocked):
			c.JSON(http.StatusForbidden, gin.H{
				"error":   "account_locked",
				"message": "Tu cuenta está bloqueada temporalmente por múltiples intentos fallidos.",
			})

		case errors.Is(err, user.ErrAccountInactive):
			c.JSON(http.StatusForbidden, gin.H{
				"error":   "account_inactive",
				"message": "Tu cuenta se encuentra inactiva. Contacta al administrador.",
			})

		case errors.Is(err, user.ErrInvalidCredentials):
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "invalid_credentials",
				"message": "Correo o contraseña incorrectos.",
			})

		default:
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "internal_error",
			})
		}

		return
	}

	// El refresh token solamente se entrega mediante cookie HttpOnly.
	c.SetCookie(
		"refresh_token",
		out.RefreshToken,
		7*24*60*60,
		"/",
		"",
		false, // Secure=false en desarrollo HTTP
		true,  // HttpOnly=true
	)

		c.JSON(http.StatusOK, dto.LoginResponse{
		AccessToken:   out.AccessToken,
		Roles:         out.Roles,
		HasActiveRole: out.HasActiveRole,
	})
}

func (h *AuthHandler) Refresh(c *gin.Context) {
	refreshToken, err := c.Cookie("refresh_token")

	if err != nil || refreshToken == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "invalid_refresh_token",
		})
		return
	}

	accessToken, roles, err := h.refreshUC.Execute(
		c.Request.Context(),
		refreshToken,
	)

	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "invalid_refresh_token",
		})
		return
	}

	c.JSON(http.StatusOK, dto.LoginResponse{
		AccessToken:   accessToken,
		Roles:         roles,
		HasActiveRole: len(roles) > 0,
	})
}
func (h *AuthHandler) Logout(c *gin.Context) {
	refreshToken, _ := c.Cookie("refresh_token")

	if err := h.logoutUC.Execute(c.Request.Context(), refreshToken); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "internal_error",
		})
		return
	}

	// Borra la cookie del lado del cliente (MaxAge negativo).
	c.SetCookie(
		"refresh_token",
		"",
		-1,
		"/",
		"",
		false,
		true,
	)

	c.Status(http.StatusNoContent)
}
func (h *AuthHandler) ForgotPassword(c *gin.Context) {
	var req dto.ForgotPasswordRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_payload",
			"details": err.Error(),
		})
		return
	}

if err := h.forgotPasswordUC.Execute(c.Request.Context(), req.Email); err != nil {
		log.Printf("forgot-password error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "internal_error",
		})
		return
	}

	// Respuesta SIEMPRE genérica, exista o no el correo — evita
	// que alguien deduzca qué correos están registrados en el sistema.
	c.JSON(http.StatusOK, gin.H{
		"message": "Si el correo existe, se ha enviado un enlace de recuperación.",
	})
}

func (h *AuthHandler) ResetPassword(c *gin.Context) {
	var req dto.ResetPasswordRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_payload",
			"details": err.Error(),
		})
		return
	}

	err := h.resetPasswordUC.Execute(c.Request.Context(), req.Token, req.NewPassword)

	if err != nil {
		switch {
		case errors.Is(err, auth.ErrInvalidResetToken):
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "invalid_reset_token",
				"message": "El enlace de recuperación no es válido.",
			})

		case errors.Is(err, auth.ErrResetTokenExpired):
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "reset_token_expired",
				"message": "El enlace de recuperación ha expirado. Solicita uno nuevo.",
			})

		case errors.Is(err, auth.ErrResetTokenUsed):
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "reset_token_used",
				"message": "Este enlace ya fue utilizado. Solicita uno nuevo.",
			})

		case errors.Is(err, auth.ErrWeakPassword):
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "weak_password",
				"message": "La contraseña debe tener al menos 8 caracteres.",
			})

		default:
			log.Printf("reset-password error: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "internal_error",
			})
		}

		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Contraseña actualizada correctamente. Todas las sesiones activas fueron cerradas.",
	})
}