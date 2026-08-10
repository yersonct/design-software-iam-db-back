package main

import (
	"log"

	"github.com/yersonct/iam-service/internal/infrastructure/persistence/postgres"
	"github.com/yersonct/iam-service/internal/infrastructure/security"
	"github.com/yersonct/iam-service/pkg/config"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("error cargando configuración: %v", err)
	}

	db, err := postgres.NewConnection(
		cfg.DBHost,
		cfg.DBPort,
		cfg.DBUser,
		cfg.DBPassword,
		cfg.DBName,
	)
	if err != nil {
		log.Fatalf("error conectando a PostgreSQL: %v", err)
	}
	defer db.Close()

	hasher := security.BcryptHasher{}

	email := cfg.SeedAdminEmail
	password := cfg.SeedAdminPassword

	if email == "" || password == "" {
		log.Fatalf("SEED_ADMIN_EMAIL y SEED_ADMIN_PASSWORD son obligatorias para correr el seed")
	}

	hash, err := hasher.Hash(password)
	if err != nil {
		log.Fatalf("error generando hash: %v", err)
	}

	const query = `
		UPDATE identity."user"
		SET
			password_hash = $1,
			failed_attempts = 0,
			locked_until = NULL,
			is_active = true,
			updated_at = now()
		WHERE lower(email) = lower($2)
	`

	result, err := db.Exec(query, hash, email)
	if err != nil {
		log.Fatalf("error actualizando usuario: %v", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		log.Fatalf("error verificando actualización: %v", err)
	}

	if rows == 0 {
		log.Fatalf("no existe el usuario %s", email)
	}

	log.Printf("usuario %s actualizado correctamente", email)
	log.Printf("password configurada desde SEED_ADMIN_PASSWORD")

	// Seed: feature AUDIT_LOGIN_VIEW (HU: historial de intentos de login).
	// Se siembra desde aquí -- no desde el repo de BD (Liquibase) -- porque
	// ese repo pertenece a otro equipo y no debe modificarse desde iam-service.
	// Idempotente: ON CONFLICT DO NOTHING permite correr el seed muchas veces.
	const seedFeatureQuery = `
		INSERT INTO rbac_catalog.feature (module_id, code, name, action_level)
		SELECT m.id, 'AUDIT_LOGIN_VIEW', 'Ver historial de logins', 'READ'
		FROM rbac_catalog.module m
		WHERE m.code = 'MOD_AUDIT'
		ON CONFLICT (code) DO NOTHING
	`
	if _, err := db.Exec(seedFeatureQuery); err != nil {
		log.Fatalf("error sembrando feature AUDIT_LOGIN_VIEW: %v", err)
	}
	log.Println("feature AUDIT_LOGIN_VIEW verificada/creada")

	const seedRoleFeatureQuery = `
		INSERT INTO rbac.role_feature (role_id, feature_id, scope_type)
		SELECT r.id, f.id, 'GLOBAL'
		FROM rbac.role r
		JOIN rbac_catalog.feature f ON f.code = 'AUDIT_LOGIN_VIEW'
		WHERE r.name = 'SYSTEM_ADMIN'
		ON CONFLICT (role_id, feature_id) DO NOTHING
	`
	if _, err := db.Exec(seedRoleFeatureQuery); err != nil {
		log.Fatalf("error asignando AUDIT_LOGIN_VIEW a SYSTEM_ADMIN: %v", err)
	}
	log.Println("AUDIT_LOGIN_VIEW asignada a SYSTEM_ADMIN")
}