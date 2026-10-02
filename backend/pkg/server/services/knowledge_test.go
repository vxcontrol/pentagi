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

// recordingKnowledgeStore keeps what the door handed it, and hands nothing back once its context is done.
type recordingKnowledgeStore struct {
	knowledgepkg.KnowledgeStore
	contents []string
}

func (s *recordingKnowledgeStore) record(ctx context.Context, content string) (*gqlmodel.KnowledgeDocument, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	s.contents = append(s.contents, content)
	return &gqlmodel.KnowledgeDocument{ID: "doc", DocType: gqlmodel.KnowledgeDocTypeAnswer, Content: content}, nil
}

func (s *recordingKnowledgeStore) CreateDocument(
	ctx context.Context, _ int64, input gqlmodel.CreateKnowledgeDocumentInput,
) (*gqlmodel.KnowledgeDocument, error) {
	return s.record(ctx, input.Content)
}

func (s *recordingKnowledgeStore) UpdateUserDocument(
	ctx context.Context, _ int64, _ string, input gqlmodel.UpdateKnowledgeDocumentInput,
) (*gqlmodel.KnowledgeDocument, error) {
	return s.record(ctx, input.Content)
}

func TestKnowledge_CreateDocument_RefusesABlankContentAndStoresTheRestAsSent(t *testing.T) {
	knowledgeContentAtTheDoor(t, (*KnowledgeService).CreateDocument, http.StatusCreated,
		`{"doc_type":"answer","answer_type":"other","question":"q","content":%s}`)
}

func TestKnowledge_UpdateDocument_RefusesABlankContentAndStoresTheRestAsSent(t *testing.T) {
	knowledgeContentAtTheDoor(t, (*KnowledgeService).UpdateDocument, http.StatusOK, `{"content":%s}`)
}

func knowledgeContentAtTheDoor(t *testing.T, handler func(*KnowledgeService, *gin.Context), stored int, body string) {
	t.Helper()
	gin.SetMode(gin.TestMode)

	for _, tc := range []struct {
		name       string
		content    string
		wantStatus int
		wantStored []string
	}{
		{name: "a content of whitespace is refused before the store", content: " \n\t", wantStatus: http.StatusBadRequest},
		{name: "a padded content reaches the store as it was sent", content: "  22\n443\n", wantStatus: stored, wantStored: []string{"  22\n443\n"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			store := &recordingKnowledgeStore{}
			content, err := json.Marshal(tc.content)
			require.NoError(t, err)

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Set("uid", uint64(7))
			c.Params = gin.Params{{Key: "id", Value: "doc"}}
			c.Request = httptest.NewRequest(http.MethodPost, "/knowledge/", strings.NewReader(fmt.Sprintf(body, content)))
			c.Request.Header.Set("Content-Type", "application/json")

			handler(NewKnowledgeService(nil, store), c)

			assert.Equal(t, tc.wantStatus, w.Code, w.Body.String())
			assert.Equal(t, tc.wantStored, store.contents)
		})
	}
}
