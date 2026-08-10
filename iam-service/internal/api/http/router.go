package http

import (
	"github.com/gin-gonic/gin"

	"github.com/yersonct/iam-service/internal/api/http/handlers"
	"github.com/yersonct/iam-service/internal/api/http/middleware"
	"github.com/yersonct/iam-service/internal/application/authz"
)

type Handlers struct {
	Auth           *handlers.AuthHandler
	User           *handlers.UserHandler
	Module         *handlers.ModuleHandler
	Feature        *handlers.FeatureHandler
	Role           *handlers.RoleHandler
	RoleFeature    *handlers.RoleFeatureHandler
	UserRole       *handlers.UserRoleHandler
	TrainingCenter *handlers.TrainingCenterHandler
	ScopeOverride  *handlers.ScopeOverrideHandler
	Authz          *handlers.AuthzHandler
	Me             *handlers.MeHandler
	Audit          *handlers.AuditHandler 
}

// userRoleChecker: implementado por postgres.UserRoleRepository. Se pasa
// aparte de Handlers porque RequireActiveRole necesita el repo crudo,
// no un handler.
func NewRouter(h Handlers, tokenValidator middleware.TokenValidator, userRoleChecker middleware.UserRoleChecker,checkPermissionUC *authz.CheckPermissionUseCase,auditLoginViewFeatureID string) *gin.Engine {
	r := gin.New()
	r.Use(gin.Recovery(), middleware.CORS())
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "iam-service is running"})
	})

	authGroup := r.Group("/auth")
	{
		authGroup.POST("/login", h.Auth.Login)
		authGroup.POST("/refresh", h.Auth.Refresh)
		authGroup.POST("/logout", h.Auth.Logout)
		authGroup.POST("/forgot-password", h.Auth.ForgotPassword)
		authGroup.POST("/reset-password", h.Auth.ResetPassword)
		authGroup.GET("/me", middleware.RequireAuth(tokenValidator), h.Me.Me)
	}

	usersMeGroup := r.Group("/users")
	usersMeGroup.Use(middleware.RequireAuth(tokenValidator))
	{
		usersMeGroup.GET("/me", h.User.Me)
	}

	usersAdminGroup := r.Group("/users")
	usersAdminGroup.Use(
		middleware.RequireAuth(tokenValidator),
		middleware.RequireRole("SYSTEM_ADMIN"),
	)
	{
		usersAdminGroup.POST("", h.User.Create)
		usersAdminGroup.GET("", h.User.List)
		usersAdminGroup.GET("/:id", h.User.GetByID)
		usersAdminGroup.PUT("/:id", h.User.Update)
		usersAdminGroup.PATCH("/:id/status", h.User.SetStatus)
		usersAdminGroup.PATCH("/:id/unlock", h.User.Unlock)
		usersAdminGroup.GET("/:id/roles", h.UserRole.List)
		usersAdminGroup.POST("/:id/roles", h.UserRole.Assign)
		usersAdminGroup.DELETE("/:id/roles/:roleId", h.UserRole.Remove)

		usersAdminGroup.GET("/:id/scope-overrides", h.ScopeOverride.List)
		usersAdminGroup.POST("/:id/scope-overrides", h.ScopeOverride.Create)
		usersAdminGroup.DELETE("/:id/scope-overrides/:overrideId", h.ScopeOverride.Remove)
	}

	catalogAdminGroup := r.Group("")
	catalogAdminGroup.Use(
		middleware.RequireAuth(tokenValidator),
		middleware.RequireRole("SYSTEM_ADMIN"),
	)
	{
		catalogAdminGroup.GET("/modules", h.Module.List)
		catalogAdminGroup.POST("/modules", h.Module.Create)
		catalogAdminGroup.PUT("/modules/:id", h.Module.Update)
		catalogAdminGroup.GET("/modules/:id/features", h.Feature.ListByModule)

		catalogAdminGroup.GET("/features", h.Feature.List)
		catalogAdminGroup.POST("/features", h.Feature.Create)
		catalogAdminGroup.PUT("/features/:id", h.Feature.Update)

		catalogAdminGroup.GET("/roles", h.Role.List)
		catalogAdminGroup.POST("/roles", h.Role.Create)
		catalogAdminGroup.PUT("/roles/:id", h.Role.Update)
		catalogAdminGroup.DELETE("/roles/:id", h.Role.Delete)

		catalogAdminGroup.GET("/roles/:id/features", h.RoleFeature.List)
		catalogAdminGroup.POST("/roles/:id/features", h.RoleFeature.Assign)
		catalogAdminGroup.DELETE("/roles/:id/features/:featureId", h.RoleFeature.Remove)
		catalogAdminGroup.PUT("/roles/:id/features/batch", h.RoleFeature.BatchAssign)
		catalogAdminGroup.GET("/training-centers", h.TrainingCenter.List)

		// Historia: "otorgar/denegar excepción puntual de acceso a un usuario".
		// Endpoint de verificación: prueba en vivo que el override le gana
		// al rol, y que un override vencido deja de aplicar.
		catalogAdminGroup.GET("/authz/check", h.Authz.Check)
	}
	auditGroup := r.Group("/audit")
	auditGroup.Use(
		middleware.RequireAuth(tokenValidator),
		middleware.RequireFeature(checkPermissionUC, auditLoginViewFeatureID, "GLOBAL"),
	)
	{
		auditGroup.GET("/logins", h.Audit.ListLogins)
}
	return r
}