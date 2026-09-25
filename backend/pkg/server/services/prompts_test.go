package services

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"pentagi/pkg/server/models"

	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func promptsTestDB(t *testing.T, promptColumn string) *gorm.DB {
	t.Helper()
	db := setupTestDB(t)
	require.NoError(t, db.Exec(`
		CREATE TABLE prompts (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			type TEXT NOT NULL,
			user_id INTEGER NOT NULL,
			prompt `+promptColumn+`,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)
	`).Error)
	return db
}

func promptsCall(handler gin.HandlerFunc, promptType string, privs []string, body string) *httptest.ResponseRecorder {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set("uid", uint64(1))
	c.Set("prm", privs)
	c.Params = gin.Params{{Key: "promptType", Value: promptType}}
	c.Request = httptest.NewRequest(http.MethodPut, "/prompts/"+promptType, strings.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")

	handler(c)

	return w
}

// promptsStored joins every stored prompt of the type, so a duplicate row shows as well as a missing one.
func promptsStored(t *testing.T, db *gorm.DB, promptType string) string {
	t.Helper()
	var stored []string
	require.NoError(t, db.Table("prompts").Where("type = ?", promptType).Pluck("prompt", &stored).Error)
	return strings.Join(stored, "|")
}

func TestPrompts_PatchPrompt_StoresOnlyAValidTemplateForAnEditor(t *testing.T) {
	edit := []string{"settings.prompts.edit"}

	for _, tc := range []struct {
		name       string
		existing   string
		promptType string
		prompt     string
		privs      []string
		wantCode   int
		wantStored string
	}{
		{"a valid template creates the prompt", "", "assistant", "You are a helpful assistant.", edit,
			http.StatusCreated, "You are a helpful assistant."},
		{"a valid template replaces the stored one", "old", "assistant", "You are a refreshed assistant.", edit,
			http.StatusOK, "You are a refreshed assistant."},
		{"an unbalanced end", "", "assistant", "broken {{end}}", edit, http.StatusBadRequest, ""},
		{"an undeclared variable", "", "assistant", "Hello {{.TotallyMadeUpVariable}}", edit, http.StatusBadRequest, ""},
		{"a whitespace-only template, which passes the required tag", "", "assistant", "   ", edit, http.StatusBadRequest, ""},
		{"an empty template", "", "assistant", "", edit, http.StatusBadRequest, ""},
		{"an unknown prompt type", "", "not_a_real_prompt_type", "anything", edit, http.StatusBadRequest, ""},
		{"a valid template without the edit privilege", "", "assistant", "You are a helpful assistant.",
			[]string{"settings.prompts.view"}, http.StatusForbidden, ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			db := promptsTestDB(t, "TEXT NOT NULL")
			defer db.Close()
			if tc.existing != "" {
				require.NoError(t, db.Exec("INSERT INTO prompts (type, user_id, prompt) VALUES (?, 1, ?)",
					tc.promptType, tc.existing).Error)
			}
			body, err := json.Marshal(models.PatchPrompt{Prompt: tc.prompt})
			require.NoError(t, err)

			w := promptsCall(NewPromptService(db).PatchPrompt, tc.promptType, tc.privs, string(body))

			assert.Equal(t, tc.wantCode, w.Code, w.Body.String())
			assert.Equal(t, tc.wantStored, promptsStored(t, db, tc.promptType), "a refused template must not be stored")
		})
	}
}

func TestPrompts_ResetPrompt_AnswersAFailedCreateOnce(t *testing.T) {
	db := promptsTestDB(t, "TEXT NOT NULL CHECK (length(prompt) < 10)")
	defer db.Close()

	w := promptsCall(NewPromptService(db).ResetPrompt, "primary_agent", []string{"settings.prompts.edit"}, "")

	assert.Equal(t, http.StatusInternalServerError, w.Code)

	decoder := json.NewDecoder(bytes.NewReader(w.Body.Bytes()))
	var first map[string]any
	require.NoError(t, decoder.Decode(&first), "the response must be one JSON document")
	assert.Equal(t, "error", first["status"])
	assert.False(t, decoder.More(), "a failed reset must not append a success body: %s", w.Body.String())

	assert.Empty(t, promptsStored(t, db, "primary_agent"))
}
