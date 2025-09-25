package processor

import (
	"context"
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
)

const (
	// PostgreSQL connection constants (fixed for installer on host)
	PostgreSQLHost = "127.0.0.1"
	PostgreSQLPort = "5432"

	// Default values for PostgreSQL configuration
	DefaultPostgreSQLUser     = "postgres"
	DefaultPostgreSQLPassword = "postgres"
	DefaultPostgreSQLDatabase = "pentagidb"

	// Admin user email
	AdminEmail = "admin@pentagi.com"

	// Environment variable names
	EnvPostgreSQLUser     = "PENTAGI_POSTGRES_USER"
	EnvPostgreSQLPassword = "PENTAGI_POSTGRES_PASSWORD"
	EnvPostgreSQLDatabase = "PENTAGI_POSTGRES_DB"
)

// performPasswordReset updates the admin password in PostgreSQL
func (p *processor) performPasswordReset(ctx context.Context, newPassword string, state *operationState) error {
	// get database configuration from state
	dbUser := DefaultPostgreSQLUser
	if envVar, ok := p.state.GetVar(EnvPostgreSQLUser); ok && envVar.Value != "" {
		dbUser = envVar.Value
	}

	dbPassword := DefaultPostgreSQLPassword
	if envVar, ok := p.state.GetVar(EnvPostgreSQLPassword); ok && envVar.Value != "" {
		dbPassword = envVar.Value
	}

	dbName := DefaultPostgreSQLDatabase
	if envVar, ok := p.state.GetVar(EnvPostgreSQLDatabase); ok && envVar.Value != "" {
		dbName = envVar.Value
	}

	// create connection string
	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		PostgreSQLHost, PostgreSQLPort, dbUser, dbPassword, dbName)

	// open database connection
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	defer db.Close()

	// test connection
	if err := db.PingContext(ctx); err != nil {
		return fmt.Errorf("failed to ping database: %w", err)
	}

	p.appendLog(fmt.Sprintf("Connected to PostgreSQL at %s:%s (database: %s)", PostgreSQLHost, PostgreSQLPort, dbName), ProductStackPentagi, state)

	// hash the new password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	// update the admin user password and status
	query := `UPDATE users SET password = $1, status = 'active' WHERE mail = $2`
	result, err := db.ExecContext(ctx, query, string(hashedPassword), AdminEmail)
	if err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}

	// check if any rows were affected
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("no admin user found with email %s", AdminEmail)
	}

	p.appendLog(fmt.Sprintf("Password updated for %s", AdminEmail), ProductStackPentagi, state)

	return nil
}
