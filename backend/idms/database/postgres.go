package database

import (
	"context"
	"database/sql"
	"fmt"
	"os"

	_ "github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
)

var DB *sql.DB

func InitPostgres() error {
	dsn := os.Getenv("POSTGRES_URL")
	if dsn == "" {
		return fmt.Errorf("POSTGRES_URL not set")
	}

	var err error
	DB, err = sql.Open("postgres", dsn)
	if err != nil {
		return fmt.Errorf("failed to open postgres: %w", err)
	}

	if err = DB.Ping(); err != nil {
		return fmt.Errorf("failed to ping postgres: %w\n(Check POSTGRES_URL in .env - on macOS, the default role is usually your macOS username e.g. 'postgres://priyanshu@127.0.0.1:5432/postgres?sslmode=disable')", err)
	}

	// Create users table if it doesn't exist
	_, err = DB.ExecContext(context.Background(), `
		CREATE TABLE IF NOT EXISTS users (
			id SERIAL PRIMARY KEY,
			username VARCHAR(50) UNIQUE NOT NULL,
			password_hash VARCHAR(255) NOT NULL,
			created_at TIMESTAMP DEFAULT NOW()
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to create users table: %w", err)
	}

	fmt.Println("PostgreSQL connected and users table ready")
	return nil
}

func CreateUser(username, password string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	_, err = DB.ExecContext(context.Background(),
		"INSERT INTO users (username, password_hash) VALUES ($1, $2)",
		username, string(hash))
	return err
}

func AuthenticateUser(username, password string) (bool, error) {
	var hash string
	err := DB.QueryRowContext(context.Background(),
		"SELECT password_hash FROM users WHERE username = $1",
		username).Scan(&hash)
	if err == sql.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) == nil, nil
}

