package postgres

import (
	"database/sql"
	"fmt"
	"net/url"
	"time"

	_ "github.com/lib/pq"
)

func NewConnection(
	host string,
	port string,
	user string,
	password string,
	dbname string,
) (*sql.DB, error) {

	dsn := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=disable",
		url.QueryEscape(user),
		url.QueryEscape(password),
		host,
		port,
		url.PathEscape(dbname),
	)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("error abriendo conexión: %w", err)
	}

	db.SetMaxOpenConns(20)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(30 * time.Minute)

	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("error haciendo ping a la BD: %w", err)
	}

	return db, nil
}
