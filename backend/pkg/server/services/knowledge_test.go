package services

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	knowledgepkg "pentagi/pkg/database/knowledge"
	gqlmodel "pentagi/pkg/graph/model"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type refusingKnowledgeStore struct {
	knowledgepkg.KnowledgeStore
	err error
}

func (s refusingKnowledgeStore) CreateDocument(
	_ context.Context, _ int64, _ gqlmodel.CreateKnowledgeDocumentInput,
) (*gqlmodel.KnowledgeDocument, error) {
	return nil, s.err
}

func (s refusingKnowledgeStore) UpdateUserDocument(
	_ context.Context, _ int64, _ string, _ gqlmodel.UpdateKnowledgeDocumentInput,
) (*gqlmodel.KnowledgeDocument, error) {
	return nil, s.err
}

func TestKnowledge_CreateDocument_AnswersAnUnstorableDocumentWithAClientError(t *testing.T) {
	knowledgeAnswerToAnUnstorableDocument(t, (*KnowledgeService).CreateDocument, `{"doc_type":"answer","content":"c","question":"q"}`)
}

func TestKnowledge_UpdateDocument_AnswersAnUnstorableDocumentWithAClientError(t *testing.T) {
	knowledgeAnswerToAnUnstorableDocument(t, (*KnowledgeService).UpdateDocument, `{"content":"c"}`)
}

func knowledgeAnswerToAnUnstorableDocument(t *testing.T, handler func(*KnowledgeService, *gin.Context), body string) {
	t.Helper()
	gin.SetMode(gin.TestMode)

	service := NewKnowledgeService(nil, refusingKnowledgeStore{
		err: fmt.Errorf("%w: answer document requires answer type", knowledgepkg.ErrInvalidDocument),
	})
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set("uid", uint64(7))
	c.Params = gin.Params{{Key: "id", Value: "doc"}}
	c.Request = httptest.NewRequest(http.MethodPost, "/knowledge/", strings.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")

	handler(service, c)

	var resp struct {
		Status string `json:"status"`
		Code   string `json:"code"`
	}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))

	assert.Equal(t, http.StatusBadRequest, w.Code, "the caller sent a document the door refuses, not a server fault")
	assert.Equal(t, "Knowledge.InvalidRequest", resp.Code)
}
