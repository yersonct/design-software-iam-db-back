package main

import (
	"context"
	"log"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	httpapi "github.com/yersonct/iam-service/internal/api/http"
	"github.com/yersonct/iam-service/internal/api/http/handlers"
	"github.com/yersonct/iam-service/internal/application/auth"
	applicationauthz "github.com/yersonct/iam-service/internal/application/authz"
	applicationcatalog "github.com/yersonct/iam-service/internal/application/catalog"
	applicationrole "github.com/yersonct/iam-service/internal/application/role"
	applicationrolefeature "github.com/yersonct/iam-service/internal/application/rolefeature"
	applicationscopeoverride "github.com/yersonct/iam-service/internal/application/scopeoverride"
	applicationuser "github.com/yersonct/iam-service/internal/application/user"
	applicationuserrole "github.com/yersonct/iam-service/internal/application/userrole"
	applicationaudit "github.com/yersonct/iam-service/internal/application/audit"
	"github.com/yersonct/iam-service/internal/infrastructure/notification"
	"github.com/yersonct/iam-service/internal/infrastructure/persistence/postgres"
	"github.com/yersonct/iam-service/internal/infrastructure/security"
	"github.com/yersonct/iam-service/internal/infrastructure/trainingcenter"
	"github.com/yersonct/iam-service/pkg/config"
)

func main() {
	// 1. Configuración
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("error cargando configuración: %v", err)
	}

	log.Printf("Entorno: %s", cfg.Env)

	// 2. Conexión a PostgreSQL
	db, err := postgres.NewConnection(
		cfg.DBHost,
		cfg.DBPort,
		cfg.DBUser,
		cfg.DBPassword,
		cfg.DBName,
	)

	if err != nil {
		log.Fatalf("error conectando a la base de datos: %v", err)
	}

	defer db.Close()

	log.Println("conexión a PostgreSQL establecida correctamente")

	// 3. Infraestructura
	userRepo := postgres.NewUserRepository(db)
	roleRepo := postgres.NewRoleRepository(db)
	moduleRepo := postgres.NewModuleRepository(db)
	featureRepo := postgres.NewFeatureRepository(db)
	roleFeatureRepo := postgres.NewRoleFeatureRepository(db)
	sessionRepo := postgres.NewSessionRepository(db)
	passwordResetRepo := postgres.NewPasswordResetRepository(db)
	hasher := security.BcryptHasher{}
	tokens := security.NewJWTGenerator(cfg.JWTSecret)
	audit := postgres.NewAuditRepository(db)
	updateUserUC := applicationuser.NewUpdateUserUseCase(userRepo)
	setUserStatusUC := applicationuser.NewSetUserStatusUseCase(userRepo)
	unlockUserUC := applicationuser.NewUnlockUserUseCase(userRepo)

	emailSender := notification.NewSMTPSender(
		cfg.SMTPHost,
		cfg.SMTPPort,
		cfg.SMTPUser,
		cfg.SMTPPassword,
		cfg.SMTPFrom,
	)

	// 4. Casos de uso
	loginUC := auth.NewLoginUseCase(
		userRepo,
		sessionRepo,
		hasher,
		tokens,
		audit,
	)
	refreshUC := auth.NewRefreshTokenUseCase(
		sessionRepo,
		userRepo,
		tokens,
	)
	logoutUC := auth.NewLogoutUseCase(sessionRepo)
	forgotPasswordUC := auth.NewForgotPasswordUseCase(
		userRepo,
		passwordResetRepo,
		tokens,
		emailSender,
		cfg.FrontendResetURL,
	)
	resetPasswordUC := auth.NewResetPasswordUseCase(
		passwordResetRepo,
		userRepo,
		sessionRepo,
		hasher,
	)
	createUserUC := applicationuser.NewCreateUserUseCase(
		userRepo,
		hasher,
		emailSender,
		cfg.FrontendLoginURL,
	)
	listUsersUC := applicationuser.NewListUsersUseCase(userRepo)
	getUserUC := applicationuser.NewGetUserUseCase(userRepo, roleRepo)
	createModuleUC := applicationcatalog.NewCreateModuleUseCase(moduleRepo)
	updateModuleUC := applicationcatalog.NewUpdateModuleUseCase(moduleRepo)
	listModulesUC := applicationcatalog.NewListModulesUseCase(moduleRepo)

	createFeatureUC := applicationcatalog.NewCreateFeatureUseCase(featureRepo)
	updateFeatureUC := applicationcatalog.NewUpdateFeatureUseCase(featureRepo)
	listFeaturesUC := applicationcatalog.NewListFeaturesUseCase(featureRepo)
	listFeaturesByModuleUC := applicationcatalog.NewListFeaturesByModuleUseCase(featureRepo)
	createRoleUC := applicationrole.NewCreateRoleUseCase(roleRepo)
	updateRoleUC := applicationrole.NewUpdateRoleUseCase(roleRepo)
	deleteRoleUC := applicationrole.NewDeleteRoleUseCase(roleRepo)
	listRolesUC := applicationrole.NewListRolesUseCase(roleRepo)

	// Historia: "definir qué features puede usar cada rol y con qué scope".
	assignFeatureUC := applicationrolefeature.NewAssignFeatureUseCase(roleFeatureRepo)
	removeFeatureUC := applicationrolefeature.NewRemoveFeatureUseCase(roleFeatureRepo)
	listRoleFeaturesUC := applicationrolefeature.NewListRoleFeaturesUseCase(roleFeatureRepo)
	batchAssignUC := applicationrolefeature.NewBatchAssignFeaturesUseCase(roleFeatureRepo)

	// Historia: "asignar/quitar roles a un usuario, con centro de formación".
	trainingCenterRepo := trainingcenter.NewStaticRepository()
	userRoleRepo := postgres.NewUserRoleRepository(db)
	assignUserRoleUC := applicationuserrole.NewAssignRoleUseCase(userRoleRepo, trainingCenterRepo)
	removeUserRoleUC := applicationuserrole.NewRemoveRoleUseCase(userRoleRepo)
	listUserRolesUC := applicationuserrole.NewListUserRolesUseCase(userRoleRepo)

	// Historia: "otorgar/denegar excepción puntual de acceso a un usuario".
	scopeOverrideRepo := postgres.NewScopeOverrideRepository(db)
	createOverrideUC := applicationscopeoverride.NewCreateOverrideUseCase(scopeOverrideRepo)
	removeOverrideUC := applicationscopeoverride.NewRemoveOverrideUseCase(scopeOverrideRepo)
	listOverridesUC := applicationscopeoverride.NewListOverridesUseCase(scopeOverrideRepo)

	// Motor de autorización: el override (si vigente) gana sobre el rol.
	// roleFeatureRepo ya implementa HasFeatureViaRole -> satisface
	// authz.RoleFeatureChecker por duck typing, sin tocar su interfaz de dominio.
	checkPermissionUC := applicationauthz.NewCheckPermissionUseCase(scopeOverrideRepo, roleFeatureRepo)

	// 5. Handlers
	authHandler := handlers.NewAuthHandler(
		loginUC,
		refreshUC,
		logoutUC,
		forgotPasswordUC,
		resetPasswordUC,
	)
	userHandler := handlers.NewUserHandler(
		createUserUC,
		listUsersUC,
		getUserUC,
		updateUserUC,
		setUserStatusUC,
		unlockUserUC,
	)
	moduleHandler := handlers.NewModuleHandler(createModuleUC, updateModuleUC, listModulesUC)
	featureHandler := handlers.NewFeatureHandler(createFeatureUC, updateFeatureUC, listFeaturesUC, listFeaturesByModuleUC)
	roleHandler := handlers.NewRoleHandler(createRoleUC, updateRoleUC, deleteRoleUC, listRolesUC)
	roleFeatureHandler := handlers.NewRoleFeatureHandler(assignFeatureUC, removeFeatureUC, listRoleFeaturesUC, batchAssignUC)
	userRoleHandler := handlers.NewUserRoleHandler(assignUserRoleUC, removeUserRoleUC, listUserRolesUC)
	scopeOverrideHandler := handlers.NewScopeOverrideHandler(createOverrideUC, removeOverrideUC, listOverridesUC)
	trainingCenterHandler := handlers.NewTrainingCenterHandler(trainingCenterRepo)
	authzHandler := handlers.NewAuthzHandler(checkPermissionUC)

	// Historia: "GET /auth/me -- permisos efectivos para los demás microservicios".
	getEffectivePermissionsUC := applicationauthz.NewGetEffectivePermissionsUseCase(
		userRepo, userRoleRepo, roleFeatureRepo, scopeOverrideRepo,
	)
	meHandler := handlers.NewMeHandler(getEffectivePermissionsUC)

	// Historia: "ver historial de intentos de inicio de sesión".
	listLoginAuditUC := applicationaudit.NewListLoginAuditUseCase(audit)
	auditHandler := handlers.NewAuditHandler(listLoginAuditUC)

	auditFeature, err := featureRepo.FindByCode(context.Background(), "AUDIT_LOGIN_VIEW")
	if err != nil {
		log.Fatalf("no se encontró la feature AUDIT_LOGIN_VIEW -- ¿corriste el seed?: %v", err)
	}
	auditLoginViewFeatureID := auditFeature.ID

	// 6. Router
	router := httpapi.NewRouter(httpapi.Handlers{
		Auth:           authHandler,
		User:           userHandler,
		Module:         moduleHandler,
		Feature:        featureHandler,
		Role:           roleHandler,
		RoleFeature:    roleFeatureHandler,
		UserRole:       userRoleHandler,
		TrainingCenter: trainingCenterHandler,
		ScopeOverride:  scopeOverrideHandler,
		Authz:          authzHandler,
		Me:             meHandler,
		Audit:          auditHandler,
	}, tokens, userRoleRepo, checkPermissionUC, auditLoginViewFeatureID)

	// 7. Servidor
	srv := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      router,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	go func() {
		log.Printf(
			"iam-service escuchando en :%s (env=%s)",
			cfg.Port,
			cfg.Env,
		)

		if err := srv.ListenAndServe(); err != nil &&
			err != http.ErrServerClosed {
			log.Fatalf("error iniciando servidor: %v", err)
		}
	}()

	// 8. Graceful shutdown
	ctx, stop := signal.NotifyContext(
		context.Background(),
		syscall.SIGINT,
		syscall.SIGTERM,
	)

	defer stop()

	<-ctx.Done()

	log.Println("apagando servidor...")

	shutdownCtx, cancel := context.WithTimeout(
		context.Background(),
		10*time.Second,
	)

	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("error en shutdown: %v", err)
	}

	log.Println("servidor detenido correctamente")
}