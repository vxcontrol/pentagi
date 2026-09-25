package services

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"testing"

	"pentagi/pkg/server/models"

	"github.com/jinzhu/gorm"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupContainerServiceTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	db := setupFlowFileServiceTestDB(t)
	require.NoError(t, db.Exec(`
		CREATE TABLE containers (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			type TEXT NOT NULL DEFAULT 'primary',
			name TEXT NOT NULL DEFAULT '',
			image TEXT NOT NULL DEFAULT '',
			status TEXT NOT NULL DEFAULT 'starting',
			local_id TEXT,
			local_dir TEXT,
			flow_id INTEGER NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)
	`).Error)
	require.NoError(t, db.Exec(`
		INSERT INTO flows (id, user_id, model, model_provider_name, model_provider_type, tool_call_id_template)
		VALUES (1, 42, 'gpt', 'openai', 'openai', 'tcid')
	`).Error)

	return db
}

func TestContainers_GetContainers_ServesTheRowsTheSchemaAllows(t *testing.T) {
	db := setupContainerServiceTestDB(t)
	require.NoError(t, db.Exec(`
		INSERT INTO containers (id, type, name, image, status, local_id, local_dir, flow_id) VALUES
			(1, 'primary', 'pentagi-terminal-1', 'debian:latest', 'running', 'abc123', NULL,        1),
			(2, 'primary', 'pentagi-terminal-2', 'debian:latest', 'running', NULL,     '/tmp/work', 1)
	`).Error)

	c, w := newFlowFileTestContext(
		http.MethodGet,
		"/containers/?page=1&pageSize=5&type=init",
		nil,
		[]string{"containers.admin"},
		42,
		0,
	)

	NewContainerService(db).GetContainers(c)

	require.Equal(t, http.StatusOK, w.Code)
	var resp struct {
		Status string `json:"status"`
		Data   struct {
			Containers []models.Container `json:"containers"`
			Total      uint64             `json:"total"`
		} `json:"data"`
	}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	require.Equal(t, "success", resp.Status)
	require.Equal(t, uint64(2), resp.Data.Total)
	require.Len(t, resp.Data.Containers, 2)
	byID := map[uint64]models.Container{}
	for _, container := range resp.Data.Containers {
		byID[container.ID] = container
	}
	require.Empty(t, byID[1].LocalDir)
	require.Equal(t, "abc123", byID[1].LocalID)
	require.Empty(t, byID[2].LocalID)
	require.Equal(t, "/tmp/work", byID[2].LocalDir)
}

// The data filter concatenates a table's columns, so one NULL drops the row from every data search.
var searchBlobs = map[string]map[string]any{
	"agentlogs":     agentlogsSQLMappers,
	"assistantlogs": assistantlogsSQLMappers,
	"assistants":    assistantsSQLMappers,
	"containers":    containersSQLMappers,
	"flows":         flowsSQLMappers,
	"msglogs":       msglogsSQLMappers,
	"prompts":       promptsSQLMappers,
	"roles":         rolesSQLMappers,
	"screenshots":   screenshotsSQLMappers,
	"searchlogs":    searchlogsSQLMappers,
	"subtasks":      subtasksSQLMappers,
	"tasks":         tasksSQLMappers,
	"termlogs":      termlogsSQLMappers,
	"toolcalls":     toolcallsSQLMappers,
	"users":         usersSQLMappers,
	"vecstorelogs":  vecstorelogsSQLMappers,
}

var jsonDataMappers = map[string]string{"knowledge": "knowledge.go"}

var (
	createTable   = regexp.MustCompile(`(?is)CREATE TABLE (\w+)\s*\((.*?)\n\);`)
	addColumn     = regexp.MustCompile(`(?i)ALTER TABLE (\w+) ADD COLUMN (?:IF NOT EXISTS )?(\w+)\s+([^;]+);`)
	renameColumn  = regexp.MustCompile(`(?i)ALTER TABLE (\w+) RENAME COLUMN (\w+) TO (\w+)`)
	dropColumn    = regexp.MustCompile(`(?i)ALTER TABLE (\w+) DROP COLUMN (?:IF EXISTS )?(\w+)`)
	columnInBlob  = regexp.MustCompile(`\{\{table\}\}\.(\w+)`)
	coalescedCols = regexp.MustCompile(`COALESCE\(\{\{table\}\}\.(\w+),`)
)

func appliedSchema(migration string) string {
	if at := strings.Index(migration, "-- +goose Down"); at != -1 {
		return migration[:at]
	}

	return migration
}

type schemaEdit struct {
	at     int
	groups []string
	kind   string
}

func schemaEdits(migration string) []schemaEdit {
	edits := []schemaEdit{}

	collect := func(kind string, pattern *regexp.Regexp) {
		for _, match := range pattern.FindAllStringSubmatchIndex(migration, -1) {
			groups := []string{}

			for group := 0; group*2 < len(match); group++ {
				if match[group*2] == -1 {
					groups = append(groups, "")
					continue
				}

				groups = append(groups, migration[match[group*2]:match[group*2+1]])
			}

			edits = append(edits, schemaEdit{at: match[0], groups: groups, kind: kind})
		}
	}

	collect("create", createTable)
	collect("add", addColumn)
	collect("rename", renameColumn)
	collect("drop", dropColumn)

	sort.Slice(edits, func(i, j int) bool { return edits[i].at < edits[j].at })

	return edits
}

func nullableColumns(t *testing.T) map[string]map[string]bool {
	t.Helper()

	dir := filepath.Join("..", "..", "..", "migrations", "sql")
	entries, err := os.ReadDir(dir)
	if err != nil || len(entries) == 0 {
		t.Fatalf("migrations are not readable at %s (%v) — the schema this check reads has moved", dir, err)
	}

	migrations := make([]string, 0, len(entries))
	for _, entry := range entries {
		body, err := os.ReadFile(filepath.Join(dir, entry.Name()))
		if err != nil {
			t.Fatalf("reading %s: %v", entry.Name(), err)
		}

		migrations = append(migrations, string(body))
	}

	return schemaAfter(migrations)
}

func schemaAfter(migrations []string) map[string]map[string]bool {
	nullable := map[string]map[string]bool{}

	set := func(table, column string, isNullable bool) {
		if nullable[table] == nil {
			nullable[table] = map[string]bool{}
		}

		nullable[table][column] = isNullable
	}

	for _, migration := range migrations {
		for _, edit := range schemaEdits(appliedSchema(migration)) {
			switch edit.kind {
			case "create":
				for _, line := range strings.Split(edit.groups[2], "\n") {
					line = strings.TrimSpace(strings.TrimSuffix(strings.TrimSpace(line), ","))
					fields := strings.Fields(line)

					if len(fields) < 2 || strings.ContainsAny(fields[0], "(),") {
						continue
					}

					switch strings.ToUpper(fields[0]) {
					case "CONSTRAINT", "PRIMARY", "UNIQUE", "FOREIGN", "CHECK":
						continue
					}

					set(edit.groups[1], fields[0], !strings.Contains(strings.ToUpper(line), "NOT NULL"))
				}

			case "add":
				set(edit.groups[1], edit.groups[2], !strings.Contains(strings.ToUpper(edit.groups[3]), "NOT NULL"))

			case "rename":
				columns := nullable[edit.groups[1]]
				if columns == nil {
					continue
				}

				isNullable, known := columns[edit.groups[2]]
				if !known {
					continue
				}

				delete(columns, edit.groups[2])
				columns[edit.groups[3]] = isNullable

			case "drop":
				if columns := nullable[edit.groups[1]]; columns != nil {
					delete(columns, edit.groups[2])
				}
			}
		}
	}

	return nullable
}

func TestContainers_SchemaAfter_KeepsOnlyWhatTheForwardMigrationsLeave(t *testing.T) {
	tests := []struct {
		name       string
		migrations []string
		wantKnown  []string
		wantGone   []string
	}{
		{
			name: "a rename moves the column to its current name",
			migrations: []string{
				"-- +goose Up\nCREATE TABLE flows (\n  id BIGSERIAL PRIMARY KEY,\n  model_provider VARCHAR(50) NOT NULL\n);\n",
				"-- +goose Up\nALTER TABLE flows RENAME COLUMN model_provider TO model_provider_name;\n" +
					"-- +goose Down\nALTER TABLE flows RENAME COLUMN model_provider_name TO model_provider;\n",
			},
			wantKnown: []string{"model_provider_name"},
			wantGone:  []string{"model_provider"},
		},
		{
			name: "a column only the rollback restores is not credited",
			migrations: []string{
				"-- +goose Up\nCREATE TABLE flows (\n  id BIGSERIAL PRIMARY KEY\n);\n",
				"-- +goose Up\nALTER TABLE flows ADD COLUMN prompts JSON NULL;\n" +
					"ALTER TABLE flows DROP COLUMN prompts;\n" +
					"-- +goose Down\nALTER TABLE flows ADD COLUMN prompts JSON NULL;\n",
			},
			wantGone: []string{"prompts"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			flows := schemaAfter(tt.migrations)["flows"]
			for _, column := range tt.wantKnown {
				_, known := flows[column]
				assert.True(t, known, "%s must count as declared", column)
			}
			for _, column := range tt.wantGone {
				_, known := flows[column]
				assert.False(t, known, "%s must not count as declared", column)
			}
		})
	}
}

func TestContainers_SQLMappers_EveryDataBlobCoalescesANullableColumn(t *testing.T) {
	nullable := nullableColumns(t)

	for table, mappers := range searchBlobs {
		blob, ok := mappers["data"].(string)
		if !ok {
			t.Errorf("%s has no data mapper — this check no longer covers it", table)
			continue
		}

		columns, ok := nullable[table]
		if !ok {
			t.Errorf("no migration creates %s, so the nullability of its columns is unknown", table)
			continue
		}

		guarded := map[string]bool{}
		for _, match := range coalescedCols.FindAllStringSubmatch(blob, -1) {
			guarded[match[1]] = true
		}

		for _, match := range columnInBlob.FindAllStringSubmatch(blob, -1) {
			column := match[1]

			isNullable, known := columns[column]
			if !known {
				t.Errorf("%s.data reads %s, which no migration declares", table, column)
				continue
			}

			if isNullable && !guarded[column] {
				t.Errorf(
					"%s.data concatenates %s bare, and it may be NULL: every row without it drops out of the data filter entirely",
					table, column,
				)
			}
		}
	}
}

func TestContainers_SQLMappers_EveryDataMapperInThisPackageIsChecked(t *testing.T) {
	entries, err := os.ReadDir(".")
	if err != nil {
		t.Fatalf("reading the package directory: %v", err)
	}

	for _, entry := range entries {
		name := entry.Name()
		if !strings.HasSuffix(name, ".go") || strings.HasSuffix(name, "_test.go") {
			continue
		}

		body, err := os.ReadFile(name)
		if err != nil {
			t.Fatalf("reading %s: %v", name, err)
		}

		if !strings.Contains(string(body), `"data":`) {
			continue
		}

		table := strings.TrimSuffix(name, ".go")
		_, columnBased := searchBlobs[table]
		_, jsonBased := jsonDataMappers[table]

		if !columnBased && !jsonBased {
			t.Errorf("%s declares a data mapper that no nullability check covers", name)
		}
	}
}

func TestContainers_SQLMappers_JSONDataMappersCoalesceEveryMember(t *testing.T) {
	for table, file := range jsonDataMappers {
		body, err := os.ReadFile(file)
		if err != nil {
			t.Fatalf("reading %s: %v", file, err)
		}

		blob, ok := dataMapperValue(string(body))
		if !ok {
			t.Fatalf("%s declares no data mapper any more — this check has lost its subject", file)
		}

		members := strings.Count(blob, "->>")
		guarded := strings.Count(blob, "COALESCE(")

		if members != guarded {
			t.Errorf(
				"%s.data concatenates %d JSON members but guards only %d: a row missing one of them "+
					"drops out of the data filter entirely",
				table, members, guarded,
			)
		}
	}
}

func dataMapperValue(body string) (string, bool) {
	lines := strings.Split(body, "\n")

	for at, line := range lines {
		if !strings.Contains(line, `"data":`) {
			continue
		}

		value := []string{}
		for _, next := range lines[at:] {
			value = append(value, next)

			if strings.HasSuffix(strings.TrimSpace(next), `",`) {
				return strings.Join(value, "\n"), true
			}
		}
	}

	return "", false
}
