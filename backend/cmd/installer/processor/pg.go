package processor

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"slices"
	"strconv"
	"strings"

	"pentagi/pkg/config"
	"pentagi/pkg/database"
	"pentagi/pkg/password"

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

var (
	errPasswordTooLong = fmt.Errorf("password is longer than the %d bytes bcrypt hashes", password.MaxBytes)
	errPasswordTooWeak = errors.New("password does not meet the password policy")
)

func checkPasswordPolicy(newPassword string) error {
	switch {
	case !password.FitsHashLimit(newPassword):
		return errPasswordTooLong
	case !password.IsStrong(newPassword):
		return errPasswordTooWeak
	default:
		return nil
	}
}

// performPasswordReset updates the admin password in PostgreSQL
func (p *processor) performPasswordReset(ctx context.Context, newPassword string, state *operationState) error {
	if err := checkPasswordPolicy(newPassword); err != nil {
		return err
	}

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

	cfg := &config.Config{
		DatabaseURL: fmt.Sprintf(
			"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
			PostgreSQLHost, PostgreSQLPort, dbUser, dbPassword, dbName,
		),
	}

	if envVar, ok := p.state.GetVar("TENANT_ID"); ok {
		cfg.TenantID = strings.TrimSpace(envVar.Value)
	}
	if err := cfg.ValidateTenantID(); err != nil {
		return err
	}
	if envVar, ok := p.state.GetVar("DATABASE_EXTENSIONS_SCHEMA"); ok {
		cfg.DatabaseExtensionsSchema = strings.TrimSpace(envVar.Value)
	}
	if envVar, ok := p.state.GetVar("DATABASE_SEARCH_PATH_VIA_OPTIONS"); ok && envVar.Value != "" {
		viaOptions, err := strconv.ParseBool(envVar.Value)
		if err != nil {
			return fmt.Errorf("invalid DATABASE_SEARCH_PATH_VIA_OPTIONS %q: %w", envVar.Value, err)
		}
		cfg.DatabaseSearchPathViaOptions = viaOptions
	}

	// Point the DSN at the tenant schema before opening a connection. Without
	// this, an UPDATE on unqualified "users" would hit public.users while the
	// running instance owns <tenant>.users — a silent no-op or wrong-tenant write.
	if err := database.RewriteDatabaseURLForTenant(cfg); err != nil {
		return fmt.Errorf("failed to apply tenant search_path: %w", err)
	}

	db, err := sql.Open("postgres", cfg.DatabaseURL)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	defer db.Close()

	db.SetMaxOpenConns(1)

	if err := db.PingContext(ctx); err != nil {
		return fmt.Errorf("failed to ping database: %w", err)
	}

	if err := database.VerifySearchPath(ctx, db, cfg); err != nil {
		return fmt.Errorf("tenant schema verification failed: %w", err)
	}

	p.appendLog(fmt.Sprintf(
		"Connected to PostgreSQL at %s:%s (database: %s, schema: %s)",
		PostgreSQLHost, PostgreSQLPort, dbName, cfg.SchemaName(),
	), ProductStackPentagi, state)

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	rowsAffected, err := resetAdminPassword(ctx, db, string(hashedPassword))
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return fmt.Errorf("no admin user found with email %s", AdminEmail)
	}

	p.appendLog(fmt.Sprintf("Password updated for %s", AdminEmail), ProductStackPentagi, state)

	return nil
}

func resetAdminPassword(ctx context.Context, db *sql.DB, hashedPassword string) (int64, error) {
	hasGenerations, err := usersHaveSessionGeneration(ctx, db)
	if err != nil {
		return 0, err
	}

	query := `UPDATE users SET password = $1, status = 'active', session_generation = session_generation + 1 WHERE mail = $2`
	if !hasGenerations {
		query = `UPDATE users SET password = $1, status = 'active' WHERE mail = $2`
	}
	result, err := db.ExecContext(ctx, query, hashedPassword, AdminEmail)
	if err != nil {
		return 0, fmt.Errorf("failed to update password: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("failed to get rows affected: %w", err)
	}

	return rowsAffected, nil
}

func usersHaveSessionGeneration(ctx context.Context, db *sql.DB) (bool, error) {
	rows, err := db.QueryContext(ctx, `SELECT * FROM users LIMIT 0`)
	if err != nil {
		return false, fmt.Errorf("failed to read the users columns: %w", err)
	}
	defer rows.Close()

	columns, err := rows.Columns()
	if err != nil {
		return false, fmt.Errorf("failed to read the users columns: %w", err)
	}

	return slices.Contains(columns, "session_generation"), nil
}
