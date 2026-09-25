package database

import (
	"database/sql"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"

	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStatementTimeout_WithStatementTimeout_AddsTheCeilingUnlessOneIsSet(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name    string
		dsn     string
		want    url.Values
		wantErr string
	}{
		{
			name: "the ceiling is asked of the postgres side and the rest of the url survives",
			dsn:  "postgres://u:p@host:5432/db?sslmode=disable",
			want: url.Values{"options": {"-c statement_timeout=25000"}, "sslmode": {"disable"}},
		},
		{
			name: "a ceiling the operator already set is kept",
			dsn:  "postgres://u:p@host:5432/db?options=-c+statement_timeout%3D90000",
			want: url.Values{"options": {"-c statement_timeout=90000"}},
		},
		{
			name: "other startup options are kept beside the ceiling",
			dsn:  "postgres://u:p@host/db?options=-c+search_path%3Dtenant",
			want: url.Values{"options": {"-c search_path=tenant -c statement_timeout=25000"}},
		},
		{
			name:    "an unreadable url is refused",
			dsn:     "postgres://u:p@host:not-a-port/db",
			wantErr: "failed to parse the database URL",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			dsn, err := WithStatementTimeout(tc.dsn, 25*time.Second)
			if tc.wantErr != "" {
				assert.ErrorContains(t, err, tc.wantErr)
				return
			}
			require.NoError(t, err)

			parsed, err := url.Parse(dsn)
			require.NoError(t, err)
			assert.Equal(t, tc.want, parsed.Query())
		})
	}
}

func TestStatementTimeout_WithStatementTimeout_LetsPostgresEndAStatementPastTheCeiling(t *testing.T) {
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("TEST_DATABASE_URL is not set")
	}

	bounded, err := WithStatementTimeout(dsn, 2*time.Second)
	require.NoError(t, err)

	db, err := sql.Open("postgres", bounded)
	require.NoError(t, err)
	defer db.Close()

	started := time.Now()
	_, err = db.Query("SELECT pg_sleep(30)")

	require.Error(t, err, "a statement past the ceiling has to be ended, not waited out")
	assert.Contains(t, strings.ToLower(err.Error()), "statement timeout")
	assert.Less(t, time.Since(started), 20*time.Second)
}
