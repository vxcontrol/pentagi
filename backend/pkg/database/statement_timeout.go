package database

import (
	"fmt"
	"net/url"
	"strings"
	"time"
)

// StatementTimeout is the ceiling Postgres puts on one statement.
const StatementTimeout = 25 * time.Second

// WithStatementTimeout returns dsn with a ceiling Postgres itself enforces.
//
// It is the only ceiling the REST services have: they reach the database
// through GORM v1, whose API takes no context, so cancelling a request cannot
// reach a query already in flight.
func WithStatementTimeout(dsn string, timeout time.Duration) (string, error) {
	parsed, err := url.Parse(dsn)
	if err != nil {
		return "", fmt.Errorf("failed to parse the database URL: %w", err)
	}

	query := parsed.Query()

	if strings.Contains(query.Get("options"), "statement_timeout") {
		return dsn, nil
	}

	setting := fmt.Sprintf("-c statement_timeout=%d", timeout.Milliseconds())
	if existing := query.Get("options"); existing != "" {
		setting = existing + " " + setting
	}

	query.Set("options", setting)
	parsed.RawQuery = query.Encode()

	return parsed.String(), nil
}
