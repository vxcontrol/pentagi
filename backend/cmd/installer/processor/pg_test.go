package processor

import (
	"database/sql"
	"strings"
	"testing"

	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPg_ResetAdminPassword_ReplacesThePasswordAndEndsSessionsWhereTracked(t *testing.T) {
	for _, tc := range []struct {
		name        string
		generations bool
	}{
		{"a stack with session generations", true},
		{"a stack older than session generations", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			db, err := sql.Open("sqlite3", ":memory:")
			require.NoError(t, err)
			defer db.Close()
			db.SetMaxOpenConns(1)

			schema := `CREATE TABLE users (id INTEGER PRIMARY KEY, mail TEXT NOT NULL, password TEXT, status TEXT NOT NULL`
			if tc.generations {
				schema += `, session_generation INTEGER NOT NULL DEFAULT 1`
			}
			_, err = db.Exec(schema + `)`)
			require.NoError(t, err)
			_, err = db.Exec(`INSERT INTO users (mail, password, status) VALUES (?, 'old', 'blocked'), ('other@pentagi.com', 'x', 'active')`,
				AdminEmail)
			require.NoError(t, err)

			rows, err := resetAdminPassword(t.Context(), db, "new-hash")
			require.NoError(t, err)
			assert.EqualValues(t, 1, rows)

			var password, status string
			require.NoError(t, db.QueryRow(`SELECT password, status FROM users WHERE mail = ?`, AdminEmail).
				Scan(&password, &status))
			assert.Equal(t, "new-hash", password)
			assert.Equal(t, "active", status)
			if !tc.generations {
				return
			}

			var admin, other int
			require.NoError(t, db.QueryRow(`SELECT session_generation FROM users WHERE mail = ?`, AdminEmail).Scan(&admin))
			require.NoError(t, db.QueryRow(`SELECT session_generation FROM users WHERE mail = 'other@pentagi.com'`).Scan(&other))
			assert.Equal(t, 2, admin, "a cookie issued before the reset, which is why the reset is run, must stop working")
			assert.Equal(t, 1, other)
		})
	}
}

func TestPg_PerformPasswordReset_RefusesWhatThePolicyRefuses(t *testing.T) {
	for _, tc := range []struct {
		label, password string
		want            error
	}{
		{"short and plain", "admin", errPasswordTooWeak},
		{"eight characters without every class", "abcdefgh", errPasswordTooWeak},
		{"fifteen characters without every class", strings.Repeat("a", 15), errPasswordTooWeak},
		{"longer than bcrypt hashes", strings.Repeat("a", 73), errPasswordTooLong},
	} {
		t.Run(tc.label, func(t *testing.T) {
			h := processorHarnessWith(t, func(c *mockCheckConfig) { c.PentagiRunning = true })
			// A role no database has, so a missing refusal fails at login instead of writing.
			require.NoError(t, h.p.state.SetVar(EnvPostgreSQLUser, "pagi-policy-test-no-such-role"))
			require.NoError(t, h.p.state.SetVar(EnvPostgreSQLPassword, "not-a-password-of-any-database"))

			err := h.p.ResetPassword(t.Context(), ProductStackPentagi, WithPasswordValue(tc.password))
			require.ErrorIs(t, err, tc.want)
		})
	}
}

func TestPg_CheckPasswordPolicy_LetsThroughWhatThePolicyAccepts(t *testing.T) {
	for _, pw := range []string{
		strings.Repeat("a", 16),
		"Abcdef1!",
		strings.Repeat("a", 72),
	} {
		require.NoError(t, checkPasswordPolicy(pw), pw)
	}
}
