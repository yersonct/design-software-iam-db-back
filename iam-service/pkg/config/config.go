package config

import (
	"fmt"
	"os"
)

type Config struct {
	Port       string
	DBHost     string
	DBPort     string
	DBName     string
	DBUser     string
	DBPassword string
	DBSchema   string
	JWTSecret  string
	Env        string

	SMTPHost         string
	SMTPPort         string
	SMTPUser         string
	SMTPPassword     string
	SMTPFrom         string
	FrontendResetURL string
	FrontendLoginURL string

	SeedAdminEmail    string
	SeedAdminPassword string
	
}

func Load() (*Config, error) {
	cfg := &Config{
		Port:       getEnv("PORT", "8001"),
		DBHost:     os.Getenv("DB_HOST"),
		DBPort:     getEnv("DB_PORT", "5432"),
		DBName:     os.Getenv("DB_NAME"),
		DBUser:     os.Getenv("DB_USER"),
		DBPassword: os.Getenv("DB_PASSWORD"),
		DBSchema:   getEnv("DB_SCHEMA", "identity"),
		JWTSecret:  os.Getenv("JWT_SECRET"),
		Env:        getEnv("ENV", "development"),

		SMTPHost:         os.Getenv("SMTP_HOST"),
		SMTPPort:         getEnv("SMTP_PORT", "2525"),
		SMTPUser:         os.Getenv("SMTP_USER"),
		SMTPPassword:     os.Getenv("SMTP_PASSWORD"),
		SMTPFrom:         getEnv("SMTP_FROM", "no-reply@iam-sena.local"),
		FrontendResetURL: os.Getenv("FRONTEND_RESET_URL"),
		FrontendLoginURL: getEnv("FRONTEND_LOGIN_URL", "http://localhost:4200/auth/login"),

		SeedAdminEmail:    os.Getenv("SEED_ADMIN_EMAIL"),
		SeedAdminPassword: os.Getenv("SEED_ADMIN_PASSWORD"),
	}

	if cfg.DBHost == "" {
		return nil, fmt.Errorf("DB_HOST no está definida")
	}

	if cfg.DBName == "" {
		return nil, fmt.Errorf("DB_NAME no está definida")
	}

	if cfg.DBUser == "" {
		return nil, fmt.Errorf("DB_USER no está definida")
	}

	if cfg.DBPassword == "" {
		return nil, fmt.Errorf("DB_PASSWORD no está definida")
	}

	if cfg.JWTSecret == "" {
		return nil, fmt.Errorf("JWT_SECRET no está definida")
	}

	if cfg.SMTPHost == "" {
		return nil, fmt.Errorf("SMTP_HOST no está definida")
	}

	if cfg.SMTPUser == "" {
		return nil, fmt.Errorf("SMTP_USER no está definida")
	}

	if cfg.SMTPPassword == "" {
		return nil, fmt.Errorf("SMTP_PASSWORD no está definida")
	}

	if cfg.FrontendResetURL == "" {
		return nil, fmt.Errorf("FRONTEND_RESET_URL no está definida")
	}

	return cfg, nil
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}

	return fallback
}