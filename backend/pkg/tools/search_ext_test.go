// External: it drives the GraphQL resolver, and pentagi/pkg/graph depends on this package.
package tools_test

import (
	"context"
	"database/sql"
	"encoding/json"
	"io"
	"strings"
	"testing"

	"pentagi/pkg/database"
	"pentagi/pkg/database/knowledge"
	"pentagi/pkg/database/knowledge/limits"
	"pentagi/pkg/graph"
	"pentagi/pkg/graph/model"
	obs "pentagi/pkg/observability"
	"pentagi/pkg/tools"

	"github.com/sirupsen/logrus"
	"github.com/vxcontrol/langchaingo/vectorstores/pgvector"
)

// The resolver sees only the content string, so one door proves the tool and resolver limits agree.
func TestSearch_Handle_StoresAnAnswerGraphQLCanUpdate(t *testing.T) {
	obs.InitObserver(context.Background(), nil, nil, []logrus.Level{})

	db := &searchContentDB{}
	door := tools.NewSearchTool(1, 1, nil, nil, searchKeepReplacer{}, &pgvector.Store{}, searchUnitEmbedder{}, db, 0, nil, nil)
	args, err := json.Marshal(map[string]any{
		"answer": strings.Repeat("я", limits.MaxContentLen+1), "message": "m", "question": "q", "type": "tool",
	})
	if err != nil {
		t.Fatal(err)
	}

	if _, err := door.Handle(context.Background(), tools.StoreAnswerToolName, args); err != nil {
		t.Fatalf("the tool returned an error instead of storing: %v", err)
	}
	if !db.document.Valid {
		t.Fatal("no document reached the store")
	}

	editor := &searchAcceptingKnowledge{}
	resolver := &graph.Resolver{Knowledge: editor, Logger: logrus.NewEntry(logrus.New())}
	ctx := graph.SetUserType(
		graph.SetUserPermissions(graph.SetUserID(context.Background(), 1), []string{"knowledge.edit"}),
		"local",
	)

	if _, err := resolver.Mutation().UpdateKnowledgeDocument(ctx, "doc-1", model.UpdateKnowledgeDocumentInput{
		Content: db.document.String,
	}); err != nil {
		t.Fatalf("a document this tool wrote cannot be updated again: %v", err)
	}
	if editor.updated != db.document.String {
		t.Fatal("the resolver did not pass the body through unchanged")
	}
}

type searchKeepReplacer struct{}

func (searchKeepReplacer) ReplaceString(s string) string    { return s }
func (searchKeepReplacer) ReplaceBytes(b []byte) []byte     { return b }
func (searchKeepReplacer) WrapReader(r io.Reader) io.Reader { return r }

// The doubles below refuse a done context, like the embedding API, table and store they replace.

type searchUnitEmbedder struct{}

func (searchUnitEmbedder) EmbedDocuments(ctx context.Context, texts []string) ([][]float32, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	vectors := make([][]float32, len(texts))
	for i := range vectors {
		vectors[i] = []float32{0.1}
	}
	return vectors, nil
}
func (searchUnitEmbedder) EmbedQuery(ctx context.Context, _ string) ([]float32, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	return []float32{0.1}, nil
}
func (searchUnitEmbedder) IsAvailable() bool { return true }

type searchContentDB struct {
	database.Querier
	document sql.NullString
}

func (c *searchContentDB) InsertKnowledgeDocument(
	ctx context.Context, arg database.InsertKnowledgeDocumentParams,
) (string, error) {
	if err := ctx.Err(); err != nil {
		return "", err
	}
	c.document = arg.Document
	return "doc-1", nil
}

type searchAcceptingKnowledge struct {
	knowledge.KnowledgeStore
	updated string
}

func (s *searchAcceptingKnowledge) UpdateUserDocument(
	ctx context.Context, _ int64, _ string, input model.UpdateKnowledgeDocumentInput,
) (*model.KnowledgeDocument, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	s.updated = input.Content
	return &model.KnowledgeDocument{ID: "doc-1", Content: input.Content}, nil
}
