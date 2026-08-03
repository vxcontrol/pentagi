package services

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"pentagi/pkg/server/models"

	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupPromptsTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db := setupTestDB(t)
	require.NoError(t, db.Exec(`
		CREATE TABLE prompts (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			type TEXT NOT NULL,
			user_id INTEGER NOT NULL,
			prompt TEXT NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)
	`).Error)
	return db
}

func callPatchPrompt(t *testing.T, db *gorm.DB, promptType, prompt string, privs []string) *httptest.ResponseRecorder {
	t.Helper()
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set("uid", uint64(1))
	c.Set("prm", privs)
	c.Params = gin.Params{{Key: "promptType", Value: promptType}}

	body, err := json.Marshal(models.PatchPrompt{Prompt: prompt})
	require.NoError(t, err)
	c.Request, _ = http.NewRequest("PUT", "/prompts/"+promptType, bytes.NewBuffer(body))
	c.Request.Header.Set("Content-Type", "application/json")

	NewPromptService(db).PatchPrompt(c)
	return w
}

func countPrompts(t *testing.T, db *gorm.DB, promptType string) int {
	t.Helper()
	var n int
	require.NoError(t, db.Raw("SELECT COUNT(*) FROM prompts WHERE type = ?", promptType).Row().Scan(&n))
	return n
}

func TestPatchPrompt_TemplateValidation(t *testing.T) {
	const editPriv = "settings.prompts.edit"
	edit := []string{editPriv}

	t.Run("valid template creates a new prompt", func(t *testing.T) {
		db := setupPromptsTestDB(t)
		defer db.Close()

		w := callPatchPrompt(t, db, "assistant", "You are a helpful assistant.", edit)

		assert.Equal(t, http.StatusCreated, w.Code, "a valid template should be accepted")
		assert.Equal(t, 1, countPrompts(t, db, "assistant"))
	})

	t.Run("valid template updates an existing prompt", func(t *testing.T) {
		db := setupPromptsTestDB(t)
		defer db.Close()
		require.NoError(t, db.Exec(
			"INSERT INTO prompts (type, user_id, prompt) VALUES (?, 1, ?)", "assistant", "old",
		).Error)

		w := callPatchPrompt(t, db, "assistant", "You are a refreshed assistant.", edit)

		assert.Equal(t, http.StatusOK, w.Code)
		var stored string
		require.NoError(t, db.Raw(
			"SELECT prompt FROM prompts WHERE type = ? AND user_id = 1", "assistant",
		).Row().Scan(&stored))
		assert.Equal(t, "You are a refreshed assistant.", stored)
	})

	t.Run("syntax error is rejected and not persisted", func(t *testing.T) {
		db := setupPromptsTestDB(t)
		defer db.Close()

		w := callPatchPrompt(t, db, "assistant", "broken {{end}}", edit)

		assert.Equal(t, http.StatusBadRequest, w.Code, "an unbalanced {{end}} must be rejected")
		assert.Equal(t, 0, countPrompts(t, db, "assistant"), "a rejected template must not be stored")
	})

	t.Run("undeclared variable is rejected", func(t *testing.T) {
		db := setupPromptsTestDB(t)
		defer db.Close()

		w := callPatchPrompt(t, db, "assistant", "Hello {{.TotallyMadeUpVariable}}", edit)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Equal(t, 0, countPrompts(t, db, "assistant"))
	})

	// "   " passes the struct-tag `required` check (non-empty string) but must
	// still be rejected by ValidatePrompt's empty-template guard.
	t.Run("whitespace-only template is rejected", func(t *testing.T) {
		db := setupPromptsTestDB(t)
		defer db.Close()

		w := callPatchPrompt(t, db, "assistant", "   ", edit)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Equal(t, 0, countPrompts(t, db, "assistant"))
	})

	t.Run("empty template is rejected", func(t *testing.T) {
		db := setupPromptsTestDB(t)
		defer db.Close()

		w := callPatchPrompt(t, db, "assistant", "", edit)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("invalid prompt type is rejected", func(t *testing.T) {
		db := setupPromptsTestDB(t)
		defer db.Close()

		w := callPatchPrompt(t, db, "not_a_real_prompt_type", "anything", edit)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("valid template without the edit privilege is forbidden", func(t *testing.T) {
		db := setupPromptsTestDB(t)
		defer db.Close()

		w := callPatchPrompt(t, db, "assistant", "You are a helpful assistant.", []string{"settings.prompts.view"})

		assert.Equal(t, http.StatusForbidden, w.Code)
		assert.Equal(t, 0, countPrompts(t, db, "assistant"))
	})
}
