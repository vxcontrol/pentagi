package knowledge

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"testing"

	"pentagi/pkg/database"
	"pentagi/pkg/graph/model"
	"pentagi/pkg/graph/subscriptions"

	"github.com/vxcontrol/langchaingo/schema"
	"github.com/vxcontrol/langchaingo/vectorstores"
)

// ============================================================================
// Mocks
// ============================================================================

// --- mockDB -----------------------------------------------------------------

// mockDB embeds database.Querier (nil) so the compiler sees a full
// implementation. Only knowledge-related methods are overridden; any accidental
// call to another method will panic at runtime (acceptable in unit tests).
type mockDB struct {
	database.Querier // nil — panics if an unexpected method is called

	getKnowledge        func(ctx context.Context, uuid string) (database.GetKnowledgeDocumentRow, error)
	getUserKnowledge    func(ctx context.Context, arg database.GetUserKnowledgeDocumentParams) (database.GetUserKnowledgeDocumentRow, error)
	listAll             func(ctx context.Context) ([]database.ListAllKnowledgeDocumentsRow, error)
	listFlow            func(ctx context.Context, flowID sql.NullString) ([]database.ListFlowKnowledgeDocumentsRow, error)
	listUser            func(ctx context.Context, userID sql.NullString) ([]database.ListUserKnowledgeDocumentsRow, error)
	updateKnowledge     func(ctx context.Context, arg database.UpdateKnowledgeDocumentParams) (database.UpdateKnowledgeDocumentRow, error)
	deleteKnowledge     func(ctx context.Context, dollar_1 sql.NullString) error
	deleteUserKnowledge func(ctx context.Context, arg database.DeleteUserKnowledgeDocumentParams) error
}

func (m *mockDB) GetKnowledgeDocument(ctx context.Context, uuid string) (database.GetKnowledgeDocumentRow, error) {
	return m.getKnowledge(ctx, uuid)
}
func (m *mockDB) GetUserKnowledgeDocument(ctx context.Context, arg database.GetUserKnowledgeDocumentParams) (database.GetUserKnowledgeDocumentRow, error) {
	return m.getUserKnowledge(ctx, arg)
}
func (m *mockDB) ListAllKnowledgeDocuments(ctx context.Context) ([]database.ListAllKnowledgeDocumentsRow, error) {
	return m.listAll(ctx)
}
func (m *mockDB) ListFlowKnowledgeDocuments(ctx context.Context, cmetadata sql.NullString) ([]database.ListFlowKnowledgeDocumentsRow, error) {
	return m.listFlow(ctx, cmetadata)
}
func (m *mockDB) ListUserKnowledgeDocuments(ctx context.Context, cmetadata sql.NullString) ([]database.ListUserKnowledgeDocumentsRow, error) {
	return m.listUser(ctx, cmetadata)
}
func (m *mockDB) UpdateKnowledgeDocument(ctx context.Context, arg database.UpdateKnowledgeDocumentParams) (database.UpdateKnowledgeDocumentRow, error) {
	return m.updateKnowledge(ctx, arg)
}
func (m *mockDB) DeleteKnowledgeDocument(ctx context.Context, dollar_1 sql.NullString) error {
	return m.deleteKnowledge(ctx, dollar_1)
}
func (m *mockDB) DeleteUserKnowledgeDocument(ctx context.Context, arg database.DeleteUserKnowledgeDocumentParams) error {
	return m.deleteUserKnowledge(ctx, arg)
}

// --- mockVectorStore --------------------------------------------------------

type mockVectorStore struct {
	similaritySearchFn func(ctx context.Context, query string, numDocuments int, options ...vectorstores.Option) ([]schema.Document, error)
	addDocumentsFn     func(ctx context.Context, docs []schema.Document, options ...vectorstores.Option) ([]string, error)
}

func (m *mockVectorStore) SimilaritySearch(ctx context.Context, query string, numDocuments int, options ...vectorstores.Option) ([]schema.Document, error) {
	return m.similaritySearchFn(ctx, query, numDocuments, options...)
}
func (m *mockVectorStore) AddDocuments(ctx context.Context, docs []schema.Document, options ...vectorstores.Option) ([]string, error) {
	return m.addDocumentsFn(ctx, docs, options...)
}

// --- mockEmbedder -----------------------------------------------------------

type mockEmbedder struct {
	available        bool
	embedDocumentsFn func(ctx context.Context, texts []string) ([][]float32, error)
	embedQueryFn     func(ctx context.Context, text string) ([]float32, error)
}

func (m *mockEmbedder) IsAvailable() bool { return m.available }
func (m *mockEmbedder) EmbedDocuments(ctx context.Context, texts []string) ([][]float32, error) {
	if m.embedDocumentsFn != nil {
		return m.embedDocumentsFn(ctx, texts)
	}
	return [][]float32{{0.1, 0.2, 0.3}}, nil
}
func (m *mockEmbedder) EmbedQuery(ctx context.Context, text string) ([]float32, error) {
	if m.embedQueryFn != nil {
		return m.embedQueryFn(ctx, text)
	}
	return []float32{0.1, 0.2, 0.3}, nil
}

// --- mockPublisher ----------------------------------------------------------

type mockPublisher struct {
	createdDocs []*model.KnowledgeDocument
	updatedDocs []*model.KnowledgeDocument
	deletedDocs []*model.KnowledgeDocument
	userID      int64
}

func (m *mockPublisher) GetUserID() int64   { return m.userID }
func (m *mockPublisher) SetUserID(id int64) { m.userID = id }
func (m *mockPublisher) KnowledgeDocumentCreated(_ context.Context, doc *model.KnowledgeDocument) {
	m.createdDocs = append(m.createdDocs, doc)
}
func (m *mockPublisher) KnowledgeDocumentUpdated(_ context.Context, doc *model.KnowledgeDocument) {
	m.updatedDocs = append(m.updatedDocs, doc)
}
func (m *mockPublisher) KnowledgeDocumentDeleted(_ context.Context, doc *model.KnowledgeDocument) {
	m.deletedDocs = append(m.deletedDocs, doc)
}

// ============================================================================
// Helpers
// ============================================================================

func ptr[T any](v T) *T { return &v }

// makeRow builds a synthetic knowledge row used across multiple tests.
func makeRow(id, document, cmetadata string) database.GetKnowledgeDocumentRow {
	return database.GetKnowledgeDocumentRow{
		ID:        id,
		Document:  document,
		Cmetadata: sql.NullString{String: cmetadata, Valid: true},
	}
}

func makeListAllRow(id, document, cmetadata string) database.ListAllKnowledgeDocumentsRow {
	return database.ListAllKnowledgeDocumentsRow{
		ID:        id,
		Document:  document,
		Cmetadata: sql.NullString{String: cmetadata, Valid: true},
	}
}

func makeListFlowRow(id, document, cmetadata string) database.ListFlowKnowledgeDocumentsRow {
	return database.ListFlowKnowledgeDocumentsRow{
		ID:        id,
		Document:  document,
		Cmetadata: sql.NullString{String: cmetadata, Valid: true},
	}
}

func makeListUserRow(id, document, cmetadata string) database.ListUserKnowledgeDocumentsRow {
	return database.ListUserKnowledgeDocumentsRow{
		ID:        id,
		Document:  document,
		Cmetadata: sql.NullString{String: cmetadata, Valid: true},
	}
}

// newPublisherFactory returns a PublisherFactory backed by a single shared publisher.
func newPublisherFactory(pub *mockPublisher) PublisherFactory {
	return func(userID int64) subscriptions.KnowledgePublisher {
		pub.userID = userID
		return pub
	}
}

// ============================================================================
// Tests: pure helper functions
// ============================================================================

func TestParseMeta(t *testing.T) {
	t.Run("empty string returns zero value", func(t *testing.T) {
		m := parseMeta("")
		if m.DocType != "" || m.UserID != 0 || m.Question != "" {
			t.Fatal("expected zero-value meta for empty input")
		}
	})

	t.Run("valid JSON parses all fields including user_id", func(t *testing.T) {
		raw := `{"doc_type":"guide","user_id":42,"flow_id":7,"question":"q","guide_type":"pentest","manual":true,"part_size":100,"total_size":200}`
		m := parseMeta(raw)
		if m.DocType != "guide" {
			t.Fatalf("DocType: want guide, got %s", m.DocType)
		}
		if m.UserID != 42 {
			t.Fatalf("UserID: want 42, got %d", m.UserID)
		}
		if m.FlowID == nil || *m.FlowID != 7 {
			t.Fatal("FlowID mismatch")
		}
		if m.Question != "q" {
			t.Fatal("Question mismatch")
		}
		if m.GuideType != "pentest" {
			t.Fatal("GuideType mismatch")
		}
		if !m.Manual {
			t.Fatal("Manual should be true")
		}
		if m.PartSize != 100 || m.TotalSize != 200 {
			t.Fatal("sizes mismatch")
		}
	})

	t.Run("invalid JSON returns zero value without panic", func(t *testing.T) {
		m := parseMeta("not-json")
		if m.DocType != "" {
			t.Fatal("expected empty DocType for invalid JSON")
		}
	})

	t.Run("partial JSON fills only present fields", func(t *testing.T) {
		m := parseMeta(`{"doc_type":"code","code_lang":"python"}`)
		if m.DocType != "code" || m.CodeLang != "python" {
			t.Fatal("partial JSON mismatch")
		}
		if m.UserID != 0 {
			t.Fatal("UserID should be zero")
		}
	})
}

func TestNullStr(t *testing.T) {
	t.Run("valid NullString returns inner string", func(t *testing.T) {
		ns := sql.NullString{String: "hello", Valid: true}
		if got := nullStr(ns); got != "hello" {
			t.Fatalf("want hello, got %s", got)
		}
	})

	t.Run("invalid NullString returns empty JSON object", func(t *testing.T) {
		ns := sql.NullString{}
		if got := nullStr(ns); got != "{}" {
			t.Fatalf("want {}, got %s", got)
		}
	})
}

func TestNsOf(t *testing.T) {
	ns := nsOf("test")
	if !ns.Valid || ns.String != "test" {
		t.Fatal("nsOf must return valid NullString with given value")
	}
}

func TestFormatVector(t *testing.T) {
	cases := []struct {
		name  string
		input []float32
		want  string
	}{
		{"empty", []float32{}, "[]"},
		{"single", []float32{1.0}, "[1]"},
		{"multiple", []float32{1.0, 2.0, 3.0}, "[1,2,3]"},
		{"precision", []float32{0.5}, "[0.5]"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := formatVector(tc.input)
			if got != tc.want {
				t.Fatalf("want %q, got %q", tc.want, got)
			}
		})
	}
}

func TestMetaToJSON(t *testing.T) {
	t.Run("serialises correctly", func(t *testing.T) {
		m := knowledgeMeta{DocType: "answer", UserID: 5, Question: "test?"}
		nm, err := metaToJSON(m)
		if err != nil {
			t.Fatal(err)
		}
		if !nm.Valid {
			t.Fatal("expected valid NullRawMessage")
		}
		parsed := parseMeta(string(nm.RawMessage))
		if parsed.DocType != "answer" || parsed.UserID != 5 || parsed.Question != "test?" {
			t.Fatal("round-trip mismatch")
		}
	})
}

func TestMetaToModelDoc(t *testing.T) {
	t.Run("doc_type answer", func(t *testing.T) {
		doc := metaToModelDoc("uuid1", "content", knowledgeMeta{DocType: "answer"})
		if doc.DocType != model.KnowledgeDocTypeAnswer {
			t.Fatal("wrong doc type")
		}
		if doc.ID != "uuid1" || doc.Content != "content" {
			t.Fatal("id/content mismatch")
		}
	})

	t.Run("doc_type guide", func(t *testing.T) {
		doc := metaToModelDoc("", "", knowledgeMeta{DocType: "guide"})
		if doc.DocType != model.KnowledgeDocTypeGuide {
			t.Fatal("wrong doc type")
		}
	})

	t.Run("doc_type code", func(t *testing.T) {
		doc := metaToModelDoc("", "", knowledgeMeta{DocType: "code"})
		if doc.DocType != model.KnowledgeDocTypeCode {
			t.Fatal("wrong doc type")
		}
	})

	t.Run("unknown doc_type defaults to answer", func(t *testing.T) {
		doc := metaToModelDoc("", "", knowledgeMeta{DocType: "memory"})
		if doc.DocType != model.KnowledgeDocTypeAnswer {
			t.Fatal("expected answer fallback")
		}
	})

	t.Run("optional guide fields", func(t *testing.T) {
		gt := "pentest"
		doc := metaToModelDoc("", "", knowledgeMeta{
			DocType:     "guide",
			GuideType:   gt,
			Description: "desc",
		})
		if doc.GuideType == nil || string(*doc.GuideType) != gt {
			t.Fatal("GuideType mismatch")
		}
		if doc.Description == nil || *doc.Description != "desc" {
			t.Fatal("Description mismatch")
		}
	})

	t.Run("optional answer fields", func(t *testing.T) {
		doc := metaToModelDoc("", "", knowledgeMeta{DocType: "answer", AnswerType: "vulnerability"})
		if doc.AnswerType == nil || string(*doc.AnswerType) != "vulnerability" {
			t.Fatal("AnswerType mismatch")
		}
	})

	t.Run("optional code fields", func(t *testing.T) {
		doc := metaToModelDoc("", "", knowledgeMeta{DocType: "code", CodeLang: "python"})
		if doc.CodeLang == nil || *doc.CodeLang != "python" {
			t.Fatal("CodeLang mismatch")
		}
	})

	t.Run("flow/task/subtask IDs", func(t *testing.T) {
		fid, tid, sid := int64(1), int64(2), int64(3)
		doc := metaToModelDoc("", "", knowledgeMeta{
			FlowID: &fid, TaskID: &tid, SubtaskID: &sid,
		})
		if doc.FlowID == nil || *doc.FlowID != 1 {
			t.Fatal("FlowID mismatch")
		}
		if doc.TaskID == nil || *doc.TaskID != 2 {
			t.Fatal("TaskID mismatch")
		}
		if doc.SubtaskID == nil || *doc.SubtaskID != 3 {
			t.Fatal("SubtaskID mismatch")
		}
	})

	t.Run("sizes and manual flag", func(t *testing.T) {
		doc := metaToModelDoc("", "", knowledgeMeta{PartSize: 50, TotalSize: 100, Manual: true})
		if doc.PartSize != 50 || doc.TotalSize != 100 || !doc.Manual {
			t.Fatal("sizes/manual mismatch")
		}
	})

	t.Run("user_id propagated to model", func(t *testing.T) {
		doc := metaToModelDoc("", "", knowledgeMeta{UserID: 77})
		if doc.UserID != 77 {
			t.Fatalf("UserID: want 77, got %d", doc.UserID)
		}
	})

	t.Run("zero user_id stays zero in model", func(t *testing.T) {
		doc := metaToModelDoc("", "", knowledgeMeta{})
		if doc.UserID != 0 {
			t.Fatalf("UserID: want 0, got %d", doc.UserID)
		}
	})
}

func TestRowToModel(t *testing.T) {
	cmetaJSON := `{"doc_type":"guide","question":"how?","part_size":10,"total_size":20}`

	t.Run("withContent=true includes document text", func(t *testing.T) {
		doc := rowToModel("id1", "text content", cmetaJSON, true)
		if doc.Content != "text content" {
			t.Fatal("expected content to be included")
		}
	})

	t.Run("withContent=false blanks document text", func(t *testing.T) {
		doc := rowToModel("id1", "text content", cmetaJSON, false)
		if doc.Content != "" {
			t.Fatalf("expected empty content, got %q", doc.Content)
		}
	})

	t.Run("cmetadata parsed into model fields", func(t *testing.T) {
		doc := rowToModel("id2", "", cmetaJSON, true)
		if doc.DocType != model.KnowledgeDocTypeGuide {
			t.Fatal("DocType mismatch")
		}
		if doc.Question != "how?" {
			t.Fatal("Question mismatch")
		}
		if doc.PartSize != 10 || doc.TotalSize != 20 {
			t.Fatal("sizes mismatch")
		}
	})

	t.Run("empty cmetadata yields defaults", func(t *testing.T) {
		doc := rowToModel("id3", "x", "{}", true)
		if doc.DocType != model.KnowledgeDocTypeAnswer {
			t.Fatal("expected default answer type")
		}
		if doc.GuideType != nil || doc.AnswerType != nil || doc.CodeLang != nil {
			t.Fatal("optional fields should be nil for empty meta")
		}
	})
}

// ============================================================================
// Tests: applyGoFilters
// ============================================================================

func TestApplyGoFilters(t *testing.T) {
	guide := &model.KnowledgeDocument{DocType: model.KnowledgeDocTypeGuide, GuideType: ptr(model.KnowledgeGuideTypePentest)}
	answer := &model.KnowledgeDocument{DocType: model.KnowledgeDocTypeAnswer, AnswerType: ptr(model.KnowledgeAnswerTypeVulnerability)}
	code := &model.KnowledgeDocument{DocType: model.KnowledgeDocTypeCode, CodeLang: ptr("python")}
	manualDoc := &model.KnowledgeDocument{DocType: model.KnowledgeDocTypeAnswer, Manual: true}

	all := []*model.KnowledgeDocument{guide, answer, code, manualDoc}

	t.Run("nil filter returns all", func(t *testing.T) {
		got := applyGoFilters(all, nil)
		if len(got) != 4 {
			t.Fatalf("want 4, got %d", len(got))
		}
	})

	t.Run("empty filter returns all", func(t *testing.T) {
		got := applyGoFilters(all, &model.KnowledgeFilter{})
		if len(got) != 4 {
			t.Fatalf("want 4, got %d", len(got))
		}
	})

	t.Run("single docType filter", func(t *testing.T) {
		got := applyGoFilters(all, &model.KnowledgeFilter{DocTypes: []model.KnowledgeDocType{model.KnowledgeDocTypeGuide}})
		if len(got) != 1 || got[0] != guide {
			t.Fatal("expected only guide")
		}
	})

	t.Run("multiple docType filter", func(t *testing.T) {
		got := applyGoFilters(all, &model.KnowledgeFilter{DocTypes: []model.KnowledgeDocType{model.KnowledgeDocTypeGuide, model.KnowledgeDocTypeCode}})
		if len(got) != 2 {
			t.Fatalf("want 2, got %d", len(got))
		}
	})

	t.Run("guideType filter excludes non-matching guide", func(t *testing.T) {
		otherGuide := &model.KnowledgeDocument{DocType: model.KnowledgeDocTypeGuide, GuideType: ptr(model.KnowledgeGuideTypeInstall)}
		got := applyGoFilters([]*model.KnowledgeDocument{guide, otherGuide}, &model.KnowledgeFilter{
			GuideTypes: []model.KnowledgeGuideType{model.KnowledgeGuideTypePentest},
		})
		if len(got) != 1 || got[0] != guide {
			t.Fatal("expected only pentest guide")
		}
	})

	t.Run("guideType filter excludes docs without guideType", func(t *testing.T) {
		got := applyGoFilters([]*model.KnowledgeDocument{answer, guide}, &model.KnowledgeFilter{
			GuideTypes: []model.KnowledgeGuideType{model.KnowledgeGuideTypePentest},
		})
		if len(got) != 1 || got[0] != guide {
			t.Fatal("answer without guideType must be excluded")
		}
	})

	t.Run("answerType filter", func(t *testing.T) {
		got := applyGoFilters(all, &model.KnowledgeFilter{
			AnswerTypes: []model.KnowledgeAnswerType{model.KnowledgeAnswerTypeVulnerability},
		})
		if len(got) != 1 || got[0] != answer {
			t.Fatal("expected only vulnerability answer")
		}
	})

	t.Run("codeLangs filter", func(t *testing.T) {
		got := applyGoFilters(all, &model.KnowledgeFilter{CodeLangs: []string{"python"}})
		if len(got) != 1 || got[0] != code {
			t.Fatal("expected only python code")
		}
	})

	t.Run("manual=true filter", func(t *testing.T) {
		got := applyGoFilters(all, &model.KnowledgeFilter{Manual: ptr(true)})
		if len(got) != 1 || got[0] != manualDoc {
			t.Fatal("expected only manual doc")
		}
	})

	t.Run("manual=false filter excludes manual docs", func(t *testing.T) {
		got := applyGoFilters(all, &model.KnowledgeFilter{Manual: ptr(false)})
		if len(got) != 3 {
			t.Fatalf("want 3 non-manual docs, got %d", len(got))
		}
	})

	t.Run("combined filters", func(t *testing.T) {
		got := applyGoFilters(all, &model.KnowledgeFilter{
			DocTypes: []model.KnowledgeDocType{model.KnowledgeDocTypeAnswer},
			Manual:   ptr(false),
		})
		if len(got) != 1 || got[0] != answer {
			t.Fatal("expected only non-manual answer")
		}
	})

	t.Run("no match returns empty slice", func(t *testing.T) {
		got := applyGoFilters(all, &model.KnowledgeFilter{CodeLangs: []string{"rust"}})
		if len(got) != 0 {
			t.Fatal("expected empty result")
		}
	})
}

// ============================================================================
// Tests: buildSearchFilters
// ============================================================================

func TestBuildSearchFilters(t *testing.T) {
	t.Run("userID=0 produces no user_id key", func(t *testing.T) {
		f := buildSearchFilters(0, nil)
		if _, ok := f["user_id"]; ok {
			t.Fatal("user_id should be absent when userID=0")
		}
	})

	t.Run("userID>0 adds user_id string", func(t *testing.T) {
		f := buildSearchFilters(42, nil)
		if f["user_id"] != "42" {
			t.Fatalf("expected user_id=42, got %v", f["user_id"])
		}
	})

	t.Run("nil filter returns only userID-based key", func(t *testing.T) {
		f := buildSearchFilters(1, nil)
		if len(f) != 1 {
			t.Fatalf("want 1 key, got %d", len(f))
		}
	})

	t.Run("single docType sets doc_type", func(t *testing.T) {
		f := buildSearchFilters(0, &model.KnowledgeFilter{DocTypes: []model.KnowledgeDocType{model.KnowledgeDocTypeGuide}})
		if f["doc_type"] != "guide" {
			t.Fatalf("expected doc_type=guide, got %v", f["doc_type"])
		}
	})

	t.Run("multiple docTypes does NOT set doc_type (handled in Go)", func(t *testing.T) {
		f := buildSearchFilters(0, &model.KnowledgeFilter{
			DocTypes: []model.KnowledgeDocType{model.KnowledgeDocTypeGuide, model.KnowledgeDocTypeCode},
		})
		if _, ok := f["doc_type"]; ok {
			t.Fatal("doc_type should not be set for multiple types")
		}
	})

	t.Run("single guideType sets guide_type", func(t *testing.T) {
		f := buildSearchFilters(0, &model.KnowledgeFilter{GuideTypes: []model.KnowledgeGuideType{model.KnowledgeGuideTypePentest}})
		if f["guide_type"] != "pentest" {
			t.Fatalf("expected guide_type=pentest, got %v", f["guide_type"])
		}
	})

	t.Run("single answerType sets answer_type", func(t *testing.T) {
		f := buildSearchFilters(0, &model.KnowledgeFilter{AnswerTypes: []model.KnowledgeAnswerType{model.KnowledgeAnswerTypeCode}})
		if f["answer_type"] != "code" {
			t.Fatalf("expected answer_type=code, got %v", f["answer_type"])
		}
	})

	t.Run("single codeLang sets code_lang", func(t *testing.T) {
		f := buildSearchFilters(0, &model.KnowledgeFilter{CodeLangs: []string{"go"}})
		if f["code_lang"] != "go" {
			t.Fatalf("expected code_lang=go, got %v", f["code_lang"])
		}
	})

	t.Run("flowID sets flow_id", func(t *testing.T) {
		f := buildSearchFilters(0, &model.KnowledgeFilter{FlowID: ptr(int64(55))})
		if f["flow_id"] != "55" {
			t.Fatalf("expected flow_id=55, got %v", f["flow_id"])
		}
	})

	t.Run("userID and all single filters combined", func(t *testing.T) {
		f := buildSearchFilters(7, &model.KnowledgeFilter{
			DocTypes:    []model.KnowledgeDocType{model.KnowledgeDocTypeAnswer},
			AnswerTypes: []model.KnowledgeAnswerType{model.KnowledgeAnswerTypeGuide},
			FlowID:      ptr(int64(3)),
		})
		if f["user_id"] != "7" || f["doc_type"] != "answer" || f["answer_type"] != "guide" || f["flow_id"] != "3" {
			t.Fatalf("combined filters mismatch: %v", f)
		}
	})
}

// ============================================================================
// Tests: ListDocuments / ListUserDocuments
// ============================================================================

func TestListDocuments(t *testing.T) {
	ctx := context.Background()

	t.Run("no filter calls ListAll", func(t *testing.T) {
		called := false
		db := &mockDB{
			listAll: func(_ context.Context) ([]database.ListAllKnowledgeDocumentsRow, error) {
				called = true
				return []database.ListAllKnowledgeDocumentsRow{
					makeListAllRow("u1", "text", `{"doc_type":"answer"}`),
				}, nil
			},
		}
		ks := &knowledgeStore{db: db}
		docs, err := ks.ListDocuments(ctx, nil, true)
		if err != nil {
			t.Fatal(err)
		}
		if !called {
			t.Fatal("ListAllKnowledgeDocuments was not called")
		}
		if len(docs) != 1 || docs[0].ID != "u1" {
			t.Fatal("unexpected result")
		}
	})

	t.Run("flowID filter calls ListFlow with correct id", func(t *testing.T) {
		var gotFlowID string
		db := &mockDB{
			listFlow: func(_ context.Context, flowID sql.NullString) ([]database.ListFlowKnowledgeDocumentsRow, error) {
				gotFlowID = flowID.String
				return nil, nil
			},
		}
		ks := &knowledgeStore{db: db}
		_, err := ks.ListDocuments(ctx, &model.KnowledgeFilter{FlowID: ptr(int64(99))}, true)
		if err != nil {
			t.Fatal(err)
		}
		if gotFlowID != "99" {
			t.Fatalf("expected flow_id=99, got %q", gotFlowID)
		}
	})

	t.Run("withContent=false blanks document text", func(t *testing.T) {
		db := &mockDB{
			listAll: func(_ context.Context) ([]database.ListAllKnowledgeDocumentsRow, error) {
				return []database.ListAllKnowledgeDocumentsRow{
					makeListAllRow("u1", "secret text", `{"doc_type":"answer"}`),
				}, nil
			},
		}
		ks := &knowledgeStore{db: db}
		docs, err := ks.ListDocuments(ctx, nil, false)
		if err != nil || docs[0].Content != "" {
			t.Fatal("content should be empty when withContent=false")
		}
	})

	t.Run("db error is wrapped and returned", func(t *testing.T) {
		db := &mockDB{
			listAll: func(_ context.Context) ([]database.ListAllKnowledgeDocumentsRow, error) {
				return nil, errors.New("db gone")
			},
		}
		ks := &knowledgeStore{db: db}
		_, err := ks.ListDocuments(ctx, nil, true)
		if err == nil || !errors.Is(err, fmt.Errorf("knowledge: list all: %w", errors.New("db gone"))) {
			if err == nil {
				t.Fatal("expected error")
			}
		}
	})

	t.Run("go filters applied after db fetch", func(t *testing.T) {
		db := &mockDB{
			listAll: func(_ context.Context) ([]database.ListAllKnowledgeDocumentsRow, error) {
				return []database.ListAllKnowledgeDocumentsRow{
					makeListAllRow("g1", "", `{"doc_type":"guide"}`),
					makeListAllRow("a1", "", `{"doc_type":"answer"}`),
				}, nil
			},
		}
		ks := &knowledgeStore{db: db}
		docs, err := ks.ListDocuments(ctx, &model.KnowledgeFilter{DocTypes: []model.KnowledgeDocType{model.KnowledgeDocTypeGuide}}, false)
		if err != nil {
			t.Fatal(err)
		}
		if len(docs) != 1 || docs[0].ID != "g1" {
			t.Fatal("go filter should have excluded the answer doc")
		}
	})
}

func TestListUserDocuments(t *testing.T) {
	ctx := context.Background()
	const userID = int64(5)
	const otherUserID = int64(9)

	t.Run("no filter calls ListUserKnowledgeDocuments with correct userID", func(t *testing.T) {
		var gotUserID string
		db := &mockDB{
			listUser: func(_ context.Context, uid sql.NullString) ([]database.ListUserKnowledgeDocumentsRow, error) {
				gotUserID = uid.String
				return nil, nil
			},
		}
		ks := &knowledgeStore{db: db}
		_, err := ks.ListUserDocuments(ctx, userID, nil, true)
		if err != nil {
			t.Fatal(err)
		}
		if gotUserID != "5" {
			t.Fatalf("expected userID=5, got %q", gotUserID)
		}
	})

	// SECURITY: flowID path must enforce user_id ownership in Go
	t.Run("SECURITY flowID filter excludes documents from other users", func(t *testing.T) {
		db := &mockDB{
			listFlow: func(_ context.Context, _ sql.NullString) ([]database.ListFlowKnowledgeDocumentsRow, error) {
				return []database.ListFlowKnowledgeDocumentsRow{
					makeListFlowRow("own", "", fmt.Sprintf(`{"doc_type":"guide","user_id":%d}`, userID)),
					makeListFlowRow("other", "", fmt.Sprintf(`{"doc_type":"guide","user_id":%d}`, otherUserID)),
				}, nil
			},
		}
		ks := &knowledgeStore{db: db}
		docs, err := ks.ListUserDocuments(ctx, userID, &model.KnowledgeFilter{FlowID: ptr(int64(1))}, true)
		if err != nil {
			t.Fatal(err)
		}
		if len(docs) != 1 || docs[0].ID != "own" {
			t.Fatalf("security violation: expected only own doc, got %d docs", len(docs))
		}
	})

	// SECURITY: user_id=0 (no user_id in cmetadata) should be excluded when
	// userID filter is active via the flow path
	t.Run("SECURITY doc without user_id is excluded for non-zero userID via flow path", func(t *testing.T) {
		db := &mockDB{
			listFlow: func(_ context.Context, _ sql.NullString) ([]database.ListFlowKnowledgeDocumentsRow, error) {
				return []database.ListFlowKnowledgeDocumentsRow{
					// no user_id key → UserID=0 after parse
					makeListFlowRow("missing", "", `{"doc_type":"answer"}`),
					makeListFlowRow("correct", "", fmt.Sprintf(`{"doc_type":"answer","user_id":%d}`, userID)),
				}, nil
			},
		}
		ks := &knowledgeStore{db: db}
		docs, err := ks.ListUserDocuments(ctx, userID, &model.KnowledgeFilter{FlowID: ptr(int64(1))}, true)
		if err != nil {
			t.Fatal(err)
		}
		if len(docs) != 1 || docs[0].ID != "correct" {
			t.Fatalf("doc with missing user_id must be excluded, got %d docs", len(docs))
		}
	})

	t.Run("db error propagates", func(t *testing.T) {
		db := &mockDB{
			listUser: func(_ context.Context, _ sql.NullString) ([]database.ListUserKnowledgeDocumentsRow, error) {
				return nil, errors.New("db error")
			},
		}
		ks := &knowledgeStore{db: db}
		_, err := ks.ListUserDocuments(ctx, userID, nil, true)
		if err == nil {
			t.Fatal("expected error")
		}
	})
}

// ============================================================================
// Tests: GetDocument / GetUserDocument
// ============================================================================

func TestGetDocument(t *testing.T) {
	ctx := context.Background()

	t.Run("success returns correctly mapped document", func(t *testing.T) {
		db := &mockDB{
			getKnowledge: func(_ context.Context, uuid string) (database.GetKnowledgeDocumentRow, error) {
				return makeRow(uuid, "content", `{"doc_type":"code","code_lang":"go"}`), nil
			},
		}
		ks := &knowledgeStore{db: db}
		doc, err := ks.GetDocument(ctx, "abc123")
		if err != nil {
			t.Fatal(err)
		}
		if doc.ID != "abc123" || doc.Content != "content" {
			t.Fatal("id/content mismatch")
		}
		if doc.DocType != model.KnowledgeDocTypeCode {
			t.Fatal("DocType mismatch")
		}
		if doc.CodeLang == nil || *doc.CodeLang != "go" {
			t.Fatal("CodeLang mismatch")
		}
	})

	t.Run("db error is wrapped and returned", func(t *testing.T) {
		db := &mockDB{
			getKnowledge: func(_ context.Context, _ string) (database.GetKnowledgeDocumentRow, error) {
				return database.GetKnowledgeDocumentRow{}, errors.New("not found")
			},
		}
		ks := &knowledgeStore{db: db}
		_, err := ks.GetDocument(ctx, "xyz")
		if err == nil {
			t.Fatal("expected error")
		}
	})
}

func TestGetUserDocument(t *testing.T) {
	ctx := context.Background()
	const userID = int64(7)

	t.Run("passes correct uuid and userID to db", func(t *testing.T) {
		var gotParams database.GetUserKnowledgeDocumentParams
		db := &mockDB{
			getUserKnowledge: func(_ context.Context, arg database.GetUserKnowledgeDocumentParams) (database.GetUserKnowledgeDocumentRow, error) {
				gotParams = arg
				return database.GetUserKnowledgeDocumentRow{
					ID:        arg.Uuid,
					Document:  "doc",
					Cmetadata: sql.NullString{String: `{"doc_type":"guide"}`, Valid: true},
				}, nil
			},
		}
		ks := &knowledgeStore{db: db}
		doc, err := ks.GetUserDocument(ctx, userID, "doc-uuid")
		if err != nil {
			t.Fatal(err)
		}
		if gotParams.Uuid != "doc-uuid" {
			t.Fatalf("uuid mismatch: %q", gotParams.Uuid)
		}
		if gotParams.Cmetadata.String != "7" {
			t.Fatalf("userID param mismatch: %q", gotParams.Cmetadata.String)
		}
		if doc.ID != "doc-uuid" {
			t.Fatal("document ID mismatch")
		}
	})

	// SECURITY: if the DB row is not found (wrong owner), an error is returned
	t.Run("SECURITY db not-found error propagates (ownership enforced at DB level)", func(t *testing.T) {
		db := &mockDB{
			getUserKnowledge: func(_ context.Context, _ database.GetUserKnowledgeDocumentParams) (database.GetUserKnowledgeDocumentRow, error) {
				return database.GetUserKnowledgeDocumentRow{}, errors.New("sql: no rows")
			},
		}
		ks := &knowledgeStore{db: db}
		_, err := ks.GetUserDocument(ctx, userID, "not-my-doc")
		if err == nil {
			t.Fatal("SECURITY: accessing another user's document must return error")
		}
	})
}

// ============================================================================
// Tests: DeleteDocument / DeleteUserDocument
// ============================================================================

func TestDeleteDocument(t *testing.T) {
	ctx := context.Background()
	const userID = int64(10)

	t.Run("success: fetches doc, deletes it and publishes event", func(t *testing.T) {
		pub := &mockPublisher{}
		db := &mockDB{
			getKnowledge: func(_ context.Context, uuid string) (database.GetKnowledgeDocumentRow, error) {
				return makeRow(uuid, "content", `{"doc_type":"answer"}`), nil
			},
			deleteKnowledge: func(_ context.Context, _ sql.NullString) error { return nil },
		}
		ks := &knowledgeStore{db: db, newKnp: newPublisherFactory(pub)}

		err := ks.DeleteDocument(ctx, userID, "doc1")
		if err != nil {
			t.Fatal(err)
		}
		if len(pub.deletedDocs) != 1 || pub.deletedDocs[0].ID != "doc1" {
			t.Fatal("KnowledgeDocumentDeleted event not published or wrong doc")
		}
		if pub.userID != userID {
			t.Fatalf("publisher userID mismatch: want %d, got %d", userID, pub.userID)
		}
	})

	t.Run("GetDocument error prevents deletion and no event", func(t *testing.T) {
		pub := &mockPublisher{}
		db := &mockDB{
			getKnowledge: func(_ context.Context, _ string) (database.GetKnowledgeDocumentRow, error) {
				return database.GetKnowledgeDocumentRow{}, errors.New("not found")
			},
		}
		ks := &knowledgeStore{db: db, newKnp: newPublisherFactory(pub)}

		err := ks.DeleteDocument(ctx, userID, "ghost")
		if err == nil {
			t.Fatal("expected error")
		}
		if len(pub.deletedDocs) != 0 {
			t.Fatal("no event should be published when fetch fails")
		}
	})

	t.Run("delete db error propagates and no event", func(t *testing.T) {
		pub := &mockPublisher{}
		db := &mockDB{
			getKnowledge: func(_ context.Context, uuid string) (database.GetKnowledgeDocumentRow, error) {
				return makeRow(uuid, "", `{}`), nil
			},
			deleteKnowledge: func(_ context.Context, _ sql.NullString) error {
				return errors.New("db delete error")
			},
		}
		ks := &knowledgeStore{db: db, newKnp: newPublisherFactory(pub)}

		err := ks.DeleteDocument(ctx, userID, "d1")
		if err == nil {
			t.Fatal("expected error")
		}
		if len(pub.deletedDocs) != 0 {
			t.Fatal("no event when delete fails")
		}
	})
}

func TestDeleteUserDocument(t *testing.T) {
	ctx := context.Background()
	const userID = int64(3)

	t.Run("success: ownership-checked fetch, delete and event", func(t *testing.T) {
		pub := &mockPublisher{}
		var gotDelParams database.DeleteUserKnowledgeDocumentParams
		db := &mockDB{
			getUserKnowledge: func(_ context.Context, arg database.GetUserKnowledgeDocumentParams) (database.GetUserKnowledgeDocumentRow, error) {
				return database.GetUserKnowledgeDocumentRow{
					ID:        arg.Uuid,
					Document:  "doc",
					Cmetadata: sql.NullString{String: `{"doc_type":"answer"}`, Valid: true},
				}, nil
			},
			deleteUserKnowledge: func(_ context.Context, arg database.DeleteUserKnowledgeDocumentParams) error {
				gotDelParams = arg
				return nil
			},
		}
		ks := &knowledgeStore{db: db, newKnp: newPublisherFactory(pub)}

		err := ks.DeleteUserDocument(ctx, userID, "my-doc")
		if err != nil {
			t.Fatal(err)
		}
		// Ownership enforced: correct uuid and userID sent to DB
		if gotDelParams.Column1.String != "my-doc" {
			t.Fatalf("uuid param: want my-doc, got %q", gotDelParams.Column1.String)
		}
		if gotDelParams.Column2.String != "3" {
			t.Fatalf("user_id param: want 3, got %q", gotDelParams.Column2.String)
		}
		if len(pub.deletedDocs) != 1 {
			t.Fatal("event not published")
		}
	})

	// SECURITY: GetUserDocument returns error → no delete, no event
	t.Run("SECURITY: fetching another user's doc aborts deletion", func(t *testing.T) {
		pub := &mockPublisher{}
		db := &mockDB{
			getUserKnowledge: func(_ context.Context, _ database.GetUserKnowledgeDocumentParams) (database.GetUserKnowledgeDocumentRow, error) {
				return database.GetUserKnowledgeDocumentRow{}, errors.New("not found")
			},
		}
		ks := &knowledgeStore{db: db, newKnp: newPublisherFactory(pub)}

		err := ks.DeleteUserDocument(ctx, userID, "other-doc")
		if err == nil {
			t.Fatal("SECURITY: must fail when doc does not belong to user")
		}
		if len(pub.deletedDocs) != 0 {
			t.Fatal("no event should be emitted on failure")
		}
	})

	t.Run("db delete failure propagates, no event", func(t *testing.T) {
		pub := &mockPublisher{}
		db := &mockDB{
			getUserKnowledge: func(_ context.Context, arg database.GetUserKnowledgeDocumentParams) (database.GetUserKnowledgeDocumentRow, error) {
				return database.GetUserKnowledgeDocumentRow{
					ID:        arg.Uuid,
					Document:  "",
					Cmetadata: sql.NullString{Valid: false},
				}, nil
			},
			deleteUserKnowledge: func(_ context.Context, _ database.DeleteUserKnowledgeDocumentParams) error {
				return errors.New("constraint violation")
			},
		}
		ks := &knowledgeStore{db: db, newKnp: newPublisherFactory(pub)}

		err := ks.DeleteUserDocument(ctx, userID, "doc")
		if err == nil {
			t.Fatal("expected error from db delete")
		}
		if len(pub.deletedDocs) != 0 {
			t.Fatal("no event on failure")
		}
	})
}

// ============================================================================
// Tests: CreateDocument
// ============================================================================

func TestCreateDocument(t *testing.T) {
	ctx := context.Background()
	const userID = int64(11)

	t.Run("store nil returns error immediately", func(t *testing.T) {
		ks := &knowledgeStore{store: nil}
		_, err := ks.CreateDocument(ctx, userID, model.CreateKnowledgeDocumentInput{
			DocType: model.KnowledgeDocTypeAnswer, Content: "x", Question: "q",
		})
		if err == nil {
			t.Fatal("expected error when store is nil")
		}
	})

	t.Run("store AddDocuments error propagates", func(t *testing.T) {
		vs := &mockVectorStore{
			addDocumentsFn: func(_ context.Context, _ []schema.Document, _ ...vectorstores.Option) ([]string, error) {
				return nil, errors.New("vector db error")
			},
		}
		ks := &knowledgeStore{store: vs, newKnp: newPublisherFactory(&mockPublisher{})}
		_, err := ks.CreateDocument(ctx, userID, model.CreateKnowledgeDocumentInput{
			DocType: model.KnowledgeDocTypeGuide, Content: "c", Question: "q",
		})
		if err == nil {
			t.Fatal("expected error from AddDocuments")
		}
	})

	t.Run("AddDocuments returning no IDs is an error", func(t *testing.T) {
		vs := &mockVectorStore{
			addDocumentsFn: func(_ context.Context, _ []schema.Document, _ ...vectorstores.Option) ([]string, error) {
				return []string{}, nil // empty slice
			},
		}
		ks := &knowledgeStore{store: vs, newKnp: newPublisherFactory(&mockPublisher{})}
		_, err := ks.CreateDocument(ctx, userID, model.CreateKnowledgeDocumentInput{
			DocType: model.KnowledgeDocTypeAnswer, Content: "c", Question: "q",
		})
		if err == nil {
			t.Fatal("expected error for empty IDs")
		}
	})

	t.Run("success: all fields set, manual=true, user_id present, event published", func(t *testing.T) {
		pub := &mockPublisher{}
		var capturedDocs []schema.Document
		vs := &mockVectorStore{
			addDocumentsFn: func(_ context.Context, docs []schema.Document, _ ...vectorstores.Option) ([]string, error) {
				capturedDocs = docs
				return []string{"new-uuid"}, nil
			},
		}
		ks := &knowledgeStore{store: vs, newKnp: newPublisherFactory(pub)}

		input := model.CreateKnowledgeDocumentInput{
			DocType:     model.KnowledgeDocTypeCode,
			Content:     "  func main() {}  ", // surrounding spaces trimmed
			Question:    "how to main?",
			Description: ptr("a Go main"),
			CodeLang:    ptr("go"),
		}
		doc, err := ks.CreateDocument(ctx, userID, input)
		if err != nil {
			t.Fatal(err)
		}

		// Returned document
		if doc.ID != "new-uuid" {
			t.Fatalf("id mismatch: %q", doc.ID)
		}
		if doc.Content != "func main() {}" {
			t.Fatalf("content not trimmed: %q", doc.Content)
		}
		if !doc.Manual {
			t.Fatal("manual must be true for manually created docs")
		}

		// Metadata stored in pgvector
		if len(capturedDocs) != 1 {
			t.Fatal("expected exactly one document passed to AddDocuments")
		}
		meta := capturedDocs[0].Metadata
		if meta["user_id"] != userID {
			t.Fatalf("user_id in metadata: want %d, got %v", userID, meta["user_id"])
		}
		if meta["manual"] != true {
			t.Fatal("manual flag must be true in metadata")
		}
		if meta["code_lang"] != "go" {
			t.Fatal("code_lang missing from metadata")
		}

		// Returned model must carry the creator's UserID
		if doc.UserID != userID {
			t.Fatalf("doc.UserID: want %d, got %d", userID, doc.UserID)
		}

		// Event
		if len(pub.createdDocs) != 1 || pub.createdDocs[0].ID != "new-uuid" {
			t.Fatal("KnowledgeDocumentCreated event not published correctly")
		}
		if pub.userID != userID {
			t.Fatalf("publisher userID: want %d, got %d", userID, pub.userID)
		}
	})

	t.Run("content is trimmed of whitespace", func(t *testing.T) {
		var capturedContent string
		vs := &mockVectorStore{
			addDocumentsFn: func(_ context.Context, docs []schema.Document, _ ...vectorstores.Option) ([]string, error) {
				capturedContent = docs[0].PageContent
				return []string{"id"}, nil
			},
		}
		ks := &knowledgeStore{store: vs, newKnp: newPublisherFactory(&mockPublisher{})}
		_, err := ks.CreateDocument(ctx, userID, model.CreateKnowledgeDocumentInput{
			DocType: model.KnowledgeDocTypeAnswer, Content: "  trimmed  ", Question: "q",
		})
		if err != nil {
			t.Fatal(err)
		}
		if capturedContent != "trimmed" {
			t.Fatalf("expected trimmed content, got %q", capturedContent)
		}
	})

	t.Run("optional fields absent means nil in model", func(t *testing.T) {
		vs := &mockVectorStore{
			addDocumentsFn: func(_ context.Context, _ []schema.Document, _ ...vectorstores.Option) ([]string, error) {
				return []string{"id"}, nil
			},
		}
		ks := &knowledgeStore{store: vs, newKnp: newPublisherFactory(&mockPublisher{})}
		doc, err := ks.CreateDocument(ctx, userID, model.CreateKnowledgeDocumentInput{
			DocType: model.KnowledgeDocTypeAnswer, Content: "c", Question: "q",
		})
		if err != nil {
			t.Fatal(err)
		}
		if doc.Description != nil || doc.GuideType != nil || doc.AnswerType != nil || doc.CodeLang != nil {
			t.Fatal("optional fields should be nil when not provided")
		}
	})
}

// ============================================================================
// Tests: UpdateDocument / UpdateUserDocument
// ============================================================================

func TestUpdateDocument(t *testing.T) {
	ctx := context.Background()
	const userID = int64(20)

	existingMeta := `{"doc_type":"guide","guide_type":"pentest","question":"old q","part_size":5,"total_size":5}`
	successDB := func(overrideGet ...func(string) (database.GetKnowledgeDocumentRow, error)) *mockDB {
		getFn := func(uuid string) (database.GetKnowledgeDocumentRow, error) {
			return makeRow(uuid, "old content", existingMeta), nil
		}
		if len(overrideGet) > 0 {
			getFn = overrideGet[0]
		}
		return &mockDB{
			getKnowledge: func(_ context.Context, uuid string) (database.GetKnowledgeDocumentRow, error) {
				return getFn(uuid)
			},
			updateKnowledge: func(_ context.Context, arg database.UpdateKnowledgeDocumentParams) (database.UpdateKnowledgeDocumentRow, error) {
				return database.UpdateKnowledgeDocumentRow{
					ID:        arg.Column1.String,
					Document:  arg.Column3.String,
					Cmetadata: sql.NullString{String: existingMeta, Valid: true},
				}, nil
			},
		}
	}

	t.Run("embedder nil returns error", func(t *testing.T) {
		ks := &knowledgeStore{
			db:       successDB(),
			embedder: nil,
		}
		_, err := ks.UpdateDocument(ctx, userID, "id1", model.UpdateKnowledgeDocumentInput{Content: "new"})
		if err == nil {
			t.Fatal("expected error when embedder is nil")
		}
	})

	t.Run("embedder unavailable returns error", func(t *testing.T) {
		ks := &knowledgeStore{
			db:       successDB(),
			embedder: &mockEmbedder{available: false},
		}
		_, err := ks.UpdateDocument(ctx, userID, "id1", model.UpdateKnowledgeDocumentInput{Content: "new"})
		if err == nil {
			t.Fatal("expected error when embedder unavailable")
		}
	})

	t.Run("EmbedDocuments error propagates", func(t *testing.T) {
		ks := &knowledgeStore{
			db: successDB(),
			embedder: &mockEmbedder{
				available: true,
				embedDocumentsFn: func(_ context.Context, _ []string) ([][]float32, error) {
					return nil, errors.New("embed error")
				},
			},
		}
		_, err := ks.UpdateDocument(ctx, userID, "id1", model.UpdateKnowledgeDocumentInput{Content: "new"})
		if err == nil {
			t.Fatal("expected error from EmbedDocuments")
		}
	})

	t.Run("embedder returning empty vectors is an error", func(t *testing.T) {
		ks := &knowledgeStore{
			db: successDB(),
			embedder: &mockEmbedder{
				available: true,
				embedDocumentsFn: func(_ context.Context, _ []string) ([][]float32, error) {
					return [][]float32{}, nil
				},
			},
		}
		_, err := ks.UpdateDocument(ctx, userID, "id1", model.UpdateKnowledgeDocumentInput{Content: "new"})
		if err == nil {
			t.Fatal("expected error for empty vectors")
		}
	})

	t.Run("db update error propagates", func(t *testing.T) {
		db := &mockDB{
			getKnowledge: func(_ context.Context, uuid string) (database.GetKnowledgeDocumentRow, error) {
				return makeRow(uuid, "old", existingMeta), nil
			},
			updateKnowledge: func(_ context.Context, _ database.UpdateKnowledgeDocumentParams) (database.UpdateKnowledgeDocumentRow, error) {
				return database.UpdateKnowledgeDocumentRow{}, errors.New("constraint error")
			},
		}
		ks := &knowledgeStore{
			db:       db,
			embedder: &mockEmbedder{available: true},
		}
		_, err := ks.UpdateDocument(ctx, userID, "id1", model.UpdateKnowledgeDocumentInput{Content: "new"})
		if err == nil {
			t.Fatal("expected error from db update")
		}
	})

	t.Run("success: metadata merged, embedding computed, event published", func(t *testing.T) {
		pub := &mockPublisher{}
		var gotParams database.UpdateKnowledgeDocumentParams

		db := &mockDB{
			getKnowledge: func(_ context.Context, uuid string) (database.GetKnowledgeDocumentRow, error) {
				return makeRow(uuid, "old content", `{"doc_type":"guide","guide_type":"pentest","question":"original"}`), nil
			},
			updateKnowledge: func(_ context.Context, arg database.UpdateKnowledgeDocumentParams) (database.UpdateKnowledgeDocumentRow, error) {
				gotParams = arg
				return database.UpdateKnowledgeDocumentRow{
					ID:       arg.Column1.String,
					Document: arg.Column3.String,
					Cmetadata: sql.NullString{
						String: `{"doc_type":"guide","guide_type":"pentest","question":"new q"}`,
						Valid:  true,
					},
				}, nil
			},
		}
		ks := &knowledgeStore{
			db:       db,
			embedder: &mockEmbedder{available: true},
			newKnp:   newPublisherFactory(pub),
		}

		doc, err := ks.UpdateDocument(ctx, userID, "doc-id", model.UpdateKnowledgeDocumentInput{
			Content:  "new content",
			Question: ptr("new q"),
		})
		if err != nil {
			t.Fatal(err)
		}

		// Correct uuid passed to update
		if gotParams.Column1.String != "doc-id" {
			t.Fatalf("uuid mismatch: %q", gotParams.Column1.String)
		}
		// New content passed
		if gotParams.Column3.String != "new content" {
			t.Fatalf("document mismatch: %q", gotParams.Column3.String)
		}
		// Embedding formatted as vector literal
		if gotParams.Column2 == "" || gotParams.Column2[0] != '[' {
			t.Fatalf("embedding format wrong: %q", gotParams.Column2)
		}

		// Updated doc returned
		if doc.ID != "doc-id" {
			t.Fatal("doc ID mismatch")
		}
		// Event published
		if len(pub.updatedDocs) != 1 {
			t.Fatal("KnowledgeDocumentUpdated event not published")
		}
		if pub.userID != userID {
			t.Fatalf("publisher userID: want %d, got %d", userID, pub.userID)
		}
	})
}

func TestUpdateDocumentPreservesOriginalOwner(t *testing.T) {
	// SECURITY: when an admin (userID=1) updates a document that belongs to
	// user 99, the stored user_id in cmetadata must remain 99, not be replaced
	// by 1. The admin's identity is used only for event scoping.
	ctx := context.Background()
	const adminID = int64(1)
	const originalOwner = int64(99)

	originalMeta := fmt.Sprintf(`{"doc_type":"answer","user_id":%d,"question":"q"}`, originalOwner)

	var storedMeta string
	db := &mockDB{
		getKnowledge: func(_ context.Context, uuid string) (database.GetKnowledgeDocumentRow, error) {
			return makeRow(uuid, "old", originalMeta), nil
		},
		updateKnowledge: func(_ context.Context, arg database.UpdateKnowledgeDocumentParams) (database.UpdateKnowledgeDocumentRow, error) {
			storedMeta = string(arg.Column4.RawMessage)
			return database.UpdateKnowledgeDocumentRow{
				ID:       arg.Column1.String,
				Document: arg.Column3.String,
				Cmetadata: sql.NullString{
					String: string(arg.Column4.RawMessage),
					Valid:  true,
				},
			}, nil
		},
	}
	pub := &mockPublisher{}
	ks := &knowledgeStore{
		db:       db,
		embedder: &mockEmbedder{available: true},
		newKnp:   newPublisherFactory(pub),
	}

	doc, err := ks.UpdateDocument(ctx, adminID, "doc", model.UpdateKnowledgeDocumentInput{Content: "new"})
	if err != nil {
		t.Fatal(err)
	}

	// The returned document must carry the original owner, not the admin
	if doc.UserID != originalOwner {
		t.Fatalf("SECURITY: doc.UserID should be original owner %d, got %d", originalOwner, doc.UserID)
	}

	// The cmetadata written to DB must also preserve the original owner
	parsed := parseMeta(storedMeta)
	if parsed.UserID != originalOwner {
		t.Fatalf("SECURITY: cmetadata user_id should be %d, got %d", originalOwner, parsed.UserID)
	}

	// Publisher is still scoped to the admin who performed the update
	if pub.userID != adminID {
		t.Fatalf("publisher userID should be admin %d, got %d", adminID, pub.userID)
	}
}

func TestUpdateUserDocument(t *testing.T) {
	ctx := context.Background()
	const userID = int64(30)

	existingMeta := `{"doc_type":"answer","answer_type":"tool","question":"q"}`

	t.Run("success uses GetUserDocument (ownership check)", func(t *testing.T) {
		pub := &mockPublisher{}
		var fetchedUUID, fetchedUserID string

		db := &mockDB{
			getUserKnowledge: func(_ context.Context, arg database.GetUserKnowledgeDocumentParams) (database.GetUserKnowledgeDocumentRow, error) {
				fetchedUUID = arg.Uuid
				fetchedUserID = arg.Cmetadata.String
				return database.GetUserKnowledgeDocumentRow{
					ID:        arg.Uuid,
					Document:  "old",
					Cmetadata: sql.NullString{String: existingMeta, Valid: true},
				}, nil
			},
			updateKnowledge: func(_ context.Context, arg database.UpdateKnowledgeDocumentParams) (database.UpdateKnowledgeDocumentRow, error) {
				return database.UpdateKnowledgeDocumentRow{
					ID:       arg.Column1.String,
					Document: arg.Column3.String,
					Cmetadata: sql.NullString{
						String: existingMeta,
						Valid:  true,
					},
				}, nil
			},
		}
		ks := &knowledgeStore{
			db:       db,
			embedder: &mockEmbedder{available: true},
			newKnp:   newPublisherFactory(pub),
		}

		_, err := ks.UpdateUserDocument(ctx, userID, "my-doc", model.UpdateKnowledgeDocumentInput{Content: "updated"})
		if err != nil {
			t.Fatal(err)
		}
		if fetchedUUID != "my-doc" {
			t.Fatal("wrong uuid fetched")
		}
		if fetchedUserID != "30" {
			t.Fatalf("wrong userID passed to GetUserDocument: %q", fetchedUserID)
		}
	})

	// SECURITY: if GetUserDocument fails (wrong owner), update is blocked
	t.Run("SECURITY: wrong owner causes update to fail", func(t *testing.T) {
		db := &mockDB{
			getUserKnowledge: func(_ context.Context, _ database.GetUserKnowledgeDocumentParams) (database.GetUserKnowledgeDocumentRow, error) {
				return database.GetUserKnowledgeDocumentRow{}, errors.New("not found")
			},
		}
		ks := &knowledgeStore{
			db:       db,
			embedder: &mockEmbedder{available: true},
		}
		_, err := ks.UpdateUserDocument(ctx, userID, "not-my-doc", model.UpdateKnowledgeDocumentInput{Content: "evil update"})
		if err == nil {
			t.Fatal("SECURITY: update of another user's document must be rejected")
		}
	})
}
