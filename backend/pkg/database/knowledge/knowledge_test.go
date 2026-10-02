package knowledge

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"pentagi/pkg/database"
	"pentagi/pkg/graph/model"
	"pentagi/pkg/graph/subscriptions"
	"pentagi/pkg/providers/embeddings"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockDB embeds a nil Querier so an unstubbed query panics; its writes refuse a done context as the driver does.
type mockDB struct {
	database.Querier

	insertKnowledge     func(arg database.InsertKnowledgeDocumentParams) (string, error)
	getKnowledge        func(uuid string) (database.GetKnowledgeDocumentRow, error)
	getUserKnowledge    func(arg database.GetUserKnowledgeDocumentParams) (database.GetUserKnowledgeDocumentRow, error)
	listAll             func() ([]database.ListAllKnowledgeDocumentsRow, error)
	listFlow            func(flowID sql.NullString) ([]database.ListFlowKnowledgeDocumentsRow, error)
	listUser            func(userID sql.NullString) ([]database.ListUserKnowledgeDocumentsRow, error)
	updateKnowledge     func(arg database.UpdateKnowledgeDocumentParams) (database.UpdateKnowledgeDocumentRow, error)
	updateKnowledgeMeta func(arg database.UpdateKnowledgeDocumentMetadataParams) (database.UpdateKnowledgeDocumentMetadataRow, error)
	deleteKnowledge     func(uuid sql.NullString) error
	deleteUserKnowledge func(arg database.DeleteUserKnowledgeDocumentParams) error
	searchKnowledge     func(arg database.SearchKnowledgeDocumentsParams) ([]database.SearchKnowledgeDocumentsRow, error)
	searchUserKnowledge func(arg database.SearchUserKnowledgeDocumentsParams) ([]database.SearchUserKnowledgeDocumentsRow, error)
}

func (m *mockDB) InsertKnowledgeDocument(ctx context.Context, arg database.InsertKnowledgeDocumentParams) (string, error) {
	if err := ctx.Err(); err != nil {
		return "", err
	}
	return m.insertKnowledge(arg)
}
func (m *mockDB) GetKnowledgeDocument(_ context.Context, uuid string) (database.GetKnowledgeDocumentRow, error) {
	return m.getKnowledge(uuid)
}
func (m *mockDB) GetUserKnowledgeDocument(_ context.Context, arg database.GetUserKnowledgeDocumentParams) (database.GetUserKnowledgeDocumentRow, error) {
	return m.getUserKnowledge(arg)
}
func (m *mockDB) ListAllKnowledgeDocuments(context.Context) ([]database.ListAllKnowledgeDocumentsRow, error) {
	return m.listAll()
}
func (m *mockDB) ListFlowKnowledgeDocuments(_ context.Context, flowID sql.NullString) ([]database.ListFlowKnowledgeDocumentsRow, error) {
	return m.listFlow(flowID)
}
func (m *mockDB) ListUserKnowledgeDocuments(_ context.Context, userID sql.NullString) ([]database.ListUserKnowledgeDocumentsRow, error) {
	return m.listUser(userID)
}
func (m *mockDB) UpdateKnowledgeDocument(ctx context.Context, arg database.UpdateKnowledgeDocumentParams) (database.UpdateKnowledgeDocumentRow, error) {
	if err := ctx.Err(); err != nil {
		return database.UpdateKnowledgeDocumentRow{}, err
	}
	return m.updateKnowledge(arg)
}
func (m *mockDB) UpdateKnowledgeDocumentMetadata(ctx context.Context, arg database.UpdateKnowledgeDocumentMetadataParams) (database.UpdateKnowledgeDocumentMetadataRow, error) {
	if err := ctx.Err(); err != nil {
		return database.UpdateKnowledgeDocumentMetadataRow{}, err
	}
	return m.updateKnowledgeMeta(arg)
}
func (m *mockDB) DeleteKnowledgeDocument(ctx context.Context, uuid sql.NullString) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	return m.deleteKnowledge(uuid)
}
func (m *mockDB) DeleteUserKnowledgeDocument(ctx context.Context, arg database.DeleteUserKnowledgeDocumentParams) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	return m.deleteUserKnowledge(arg)
}
func (m *mockDB) SearchKnowledgeDocuments(_ context.Context, arg database.SearchKnowledgeDocumentsParams) ([]database.SearchKnowledgeDocumentsRow, error) {
	return m.searchKnowledge(arg)
}
func (m *mockDB) SearchUserKnowledgeDocuments(_ context.Context, arg database.SearchUserKnowledgeDocumentsParams) ([]database.SearchUserKnowledgeDocumentsRow, error) {
	return m.searchUserKnowledge(arg)
}

// mockEmbedder records what it embeds and answers [0.1,0.2,0.3] unless vectors or err say otherwise.
type mockEmbedder struct {
	unavailable bool
	vectors     [][]float32
	err         error
	embedded    []string
}

func (m *mockEmbedder) IsAvailable() bool { return !m.unavailable }
func (m *mockEmbedder) EmbedDocuments(_ context.Context, texts []string) ([][]float32, error) {
	m.embedded = append(m.embedded, texts...)
	if m.err != nil || m.vectors != nil {
		return m.vectors, m.err
	}
	return [][]float32{{0.1, 0.2, 0.3}}, nil
}
func (m *mockEmbedder) EmbedQuery(context.Context, string) ([]float32, error) {
	return []float32{0.1, 0.2, 0.3}, nil
}

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

// newPublisherFactory hands every caller one publisher, which keeps the id of the last caller.
func newPublisherFactory(pub *mockPublisher) PublisherFactory {
	return func(userID int64) subscriptions.KnowledgePublisher {
		pub.userID = userID
		return pub
	}
}

func ptr[T any](v T) *T { return &v }

func makeRow(id, document, cmetadata string) database.GetKnowledgeDocumentRow {
	return database.GetKnowledgeDocumentRow{
		ID:        id,
		Document:  document,
		Cmetadata: sql.NullString{String: cmetadata, Valid: true},
	}
}

func makeSearchRow(id, cmetadata string, score float64) database.SearchKnowledgeDocumentsRow {
	return database.SearchKnowledgeDocumentsRow{
		ID:        id,
		Document:  "content of " + id,
		Cmetadata: sql.NullString{String: cmetadata, Valid: true},
		Score:     score,
	}
}

// Driven through ListDocuments, the entry that hands rowToModel both NULL-able metadata and withContent.
func TestKnowledge_RowToModel_ReadsStoredMetadataIntoTheDocument(t *testing.T) {
	for _, tc := range []struct {
		name        string
		row         database.GetKnowledgeDocumentRow
		hideContent bool
		want        model.KnowledgeDocument
	}{
		{
			name: "every stored field reaches the document",
			row: makeRow("doc-1", "text", `{"doc_type":"guide","user_id":42,"flow_id":7,"task_id":8,"subtask_id":9,`+
				`"question":"q","description":"d","guide_type":"pentest","manual":true,"part_size":100,"total_size":200}`),
			want: model.KnowledgeDocument{
				DocType: model.KnowledgeDocTypeGuide, Question: "q", Description: ptr("d"), UserID: 42,
				FlowID: ptr(int64(7)), TaskID: ptr(int64(8)), SubtaskID: ptr(int64(9)),
				GuideType: ptr(model.KnowledgeGuideTypePentest), PartSize: 100, TotalSize: 200, Manual: true,
			},
		},
		{
			name:        "the content is left out when not asked for",
			row:         makeRow("doc-1", "text", `{"doc_type":"answer","answer_type":"code"}`),
			hideContent: true,
			want:        model.KnowledgeDocument{DocType: model.KnowledgeDocTypeAnswer, AnswerType: ptr(model.KnowledgeAnswerTypeCode)},
		},
		{
			name: "an answer keeps its answer type",
			row:  makeRow("doc-1", "text", `{"doc_type":"answer","answer_type":"vulnerability"}`),
			want: model.KnowledgeDocument{DocType: model.KnowledgeDocTypeAnswer, AnswerType: ptr(model.KnowledgeAnswerTypeVulnerability)},
		},
		{
			name: "code keeps its language and nothing absent is invented",
			row:  makeRow("doc-1", "text", `{"doc_type":"code","code_lang":"python"}`),
			want: model.KnowledgeDocument{DocType: model.KnowledgeDocTypeCode, CodeLang: ptr("python")},
		},
		{
			name: "an unknown doc type reads as an answer",
			row:  makeRow("doc-1", "text", `{"doc_type":"memory"}`),
			want: model.KnowledgeDocument{DocType: model.KnowledgeDocTypeAnswer},
		},
		{
			name: "a row without metadata reads as a bare answer",
			row:  database.GetKnowledgeDocumentRow{ID: "doc-1", Document: "text"},
			want: model.KnowledgeDocument{DocType: model.KnowledgeDocTypeAnswer},
		},
		{
			name: "unreadable metadata reads as a bare answer",
			row:  makeRow("doc-1", "text", `not-json`),
			want: model.KnowledgeDocument{DocType: model.KnowledgeDocTypeAnswer},
		},
		{
			name: "a guide type stored in another case reads as the enum value",
			row:  makeRow("doc-1", "text", `{"doc_type":"guide","guide_type":"Install"}`),
			want: model.KnowledgeDocument{DocType: model.KnowledgeDocTypeGuide, GuideType: ptr(model.KnowledgeGuideTypeInstall)},
		},
		{
			name: "a guide type outside the enum reads as other",
			row:  makeRow("doc-1", "text", `{"doc_type":"guide","guide_type":"sorcery"}`),
			want: model.KnowledgeDocument{DocType: model.KnowledgeDocTypeGuide, GuideType: ptr(model.KnowledgeGuideTypeOther)},
		},
		{
			name: "an answer type stored in another case reads as the enum value",
			row:  makeRow("doc-1", "text", `{"doc_type":"answer","answer_type":"Tool"}`),
			want: model.KnowledgeDocument{DocType: model.KnowledgeDocTypeAnswer, AnswerType: ptr(model.KnowledgeAnswerTypeTool)},
		},
		{
			name: "an answer type outside the enum reads as other",
			row:  makeRow("doc-1", "text", `{"doc_type":"answer","answer_type":"rumour"}`),
			want: model.KnowledgeDocument{DocType: model.KnowledgeDocTypeAnswer, AnswerType: ptr(model.KnowledgeAnswerTypeOther)},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			db := &mockDB{listAll: func() ([]database.ListAllKnowledgeDocumentsRow, error) {
				return []database.ListAllKnowledgeDocumentsRow{database.ListAllKnowledgeDocumentsRow(tc.row)}, nil
			}}

			docs, err := NewKnowledgeStore(db, nil, nil, nil, 0).ListDocuments(t.Context(), nil, !tc.hideContent)
			require.NoError(t, err)

			want := tc.want
			want.ID = "doc-1"
			if !tc.hideContent {
				want.Content = "text"
			}
			assert.Equal(t, []*model.KnowledgeDocument{&want}, docs)
		})
	}
}

func TestKnowledge_ApplyGoFilters_KeepsDocumentsMatchingEveryFilter(t *testing.T) {
	pentest := &model.KnowledgeDocument{DocType: model.KnowledgeDocTypeGuide, GuideType: ptr(model.KnowledgeGuideTypePentest)}
	install := &model.KnowledgeDocument{DocType: model.KnowledgeDocTypeGuide, GuideType: ptr(model.KnowledgeGuideTypeInstall)}
	answer := &model.KnowledgeDocument{DocType: model.KnowledgeDocTypeAnswer, AnswerType: ptr(model.KnowledgeAnswerTypeVulnerability)}
	code := &model.KnowledgeDocument{DocType: model.KnowledgeDocTypeCode, CodeLang: ptr("python")}
	manual := &model.KnowledgeDocument{DocType: model.KnowledgeDocTypeAnswer, Manual: true}
	all := []*model.KnowledgeDocument{pentest, install, answer, code, manual}

	for _, tc := range []struct {
		name   string
		filter *model.KnowledgeFilter
		want   []*model.KnowledgeDocument
	}{
		{name: "no filter keeps everything", filter: nil, want: all},
		{name: "an empty filter keeps everything", filter: &model.KnowledgeFilter{}, want: all},
		{
			name:   "one doc type",
			filter: &model.KnowledgeFilter{DocTypes: []model.KnowledgeDocType{model.KnowledgeDocTypeGuide}},
			want:   []*model.KnowledgeDocument{pentest, install},
		},
		{
			name:   "several doc types",
			filter: &model.KnowledgeFilter{DocTypes: []model.KnowledgeDocType{model.KnowledgeDocTypeGuide, model.KnowledgeDocTypeCode}},
			want:   []*model.KnowledgeDocument{pentest, install, code},
		},
		{
			name:   "a guide type drops other guides and documents without one",
			filter: &model.KnowledgeFilter{GuideTypes: []model.KnowledgeGuideType{model.KnowledgeGuideTypePentest}},
			want:   []*model.KnowledgeDocument{pentest},
		},
		{
			name:   "an answer type",
			filter: &model.KnowledgeFilter{AnswerTypes: []model.KnowledgeAnswerType{model.KnowledgeAnswerTypeVulnerability}},
			want:   []*model.KnowledgeDocument{answer},
		},
		{name: "a code language", filter: &model.KnowledgeFilter{CodeLangs: []string{"python"}}, want: []*model.KnowledgeDocument{code}},
		{name: "manual documents only", filter: &model.KnowledgeFilter{Manual: ptr(true)}, want: []*model.KnowledgeDocument{manual}},
		{
			name:   "generated documents only",
			filter: &model.KnowledgeFilter{Manual: ptr(false)},
			want:   []*model.KnowledgeDocument{pentest, install, answer, code},
		},
		{
			name:   "filters combine",
			filter: &model.KnowledgeFilter{DocTypes: []model.KnowledgeDocType{model.KnowledgeDocTypeAnswer}, Manual: ptr(false)},
			want:   []*model.KnowledgeDocument{answer},
		},
		{name: "nothing matches", filter: &model.KnowledgeFilter{CodeLangs: []string{"rust"}}, want: []*model.KnowledgeDocument{}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, applyGoFilters(all, tc.filter))
		})
	}
}

func TestKnowledge_PassesSearchFilter_KeepsDocumentsMatchingEveryFilter(t *testing.T) {
	guide := &model.KnowledgeDocument{DocType: model.KnowledgeDocTypeGuide, GuideType: ptr(model.KnowledgeGuideTypePentest), FlowID: ptr(int64(10))}
	answer := &model.KnowledgeDocument{DocType: model.KnowledgeDocTypeAnswer, AnswerType: ptr(model.KnowledgeAnswerTypeVulnerability), Manual: true}
	code := &model.KnowledgeDocument{DocType: model.KnowledgeDocTypeCode, CodeLang: ptr("python")}

	docTypes := func(types ...model.KnowledgeDocType) *model.KnowledgeFilter {
		return &model.KnowledgeFilter{DocTypes: types}
	}
	guideTypes := func(types ...model.KnowledgeGuideType) *model.KnowledgeFilter {
		return &model.KnowledgeFilter{GuideTypes: types}
	}
	answerTypes := func(types ...model.KnowledgeAnswerType) *model.KnowledgeFilter {
		return &model.KnowledgeFilter{AnswerTypes: types}
	}

	for _, tc := range []struct {
		name   string
		doc    *model.KnowledgeDocument
		filter *model.KnowledgeFilter
		want   bool
	}{
		{"no filter passes", guide, nil, true},
		{"an empty filter passes", answer, &model.KnowledgeFilter{}, true},
		{"a matching doc type passes", guide, docTypes(model.KnowledgeDocTypeGuide), true},
		{"another doc type blocks", guide, docTypes(model.KnowledgeDocTypeCode), false},
		{"one of several doc types passes", guide, docTypes(model.KnowledgeDocTypeGuide, model.KnowledgeDocTypeCode), true},
		{"a matching guide type passes", guide, guideTypes(model.KnowledgeGuideTypePentest), true},
		{"another guide type blocks", guide, guideTypes(model.KnowledgeGuideTypeInstall), false},
		{"a document without a guide type is blocked by one", answer, guideTypes(model.KnowledgeGuideTypePentest), false},
		{"a matching answer type passes", answer, answerTypes(model.KnowledgeAnswerTypeVulnerability), true},
		{"another answer type blocks", answer, answerTypes(model.KnowledgeAnswerTypeCode), false},
		{"a document without an answer type is blocked by one", code, answerTypes(model.KnowledgeAnswerTypeCode), false},
		{"a matching code language passes", code, &model.KnowledgeFilter{CodeLangs: []string{"python"}}, true},
		{"another code language blocks", code, &model.KnowledgeFilter{CodeLangs: []string{"go"}}, false},
		{"a document without a code language is blocked by one", guide, &model.KnowledgeFilter{CodeLangs: []string{"go"}}, false},
		{"manual only passes a manual document", answer, &model.KnowledgeFilter{Manual: ptr(true)}, true},
		{"manual only blocks a generated document", guide, &model.KnowledgeFilter{Manual: ptr(true)}, false},
		{"generated only passes a generated document", guide, &model.KnowledgeFilter{Manual: ptr(false)}, true},
		{"generated only blocks a manual document", answer, &model.KnowledgeFilter{Manual: ptr(false)}, false},
		{"the same flow passes", guide, &model.KnowledgeFilter{FlowID: ptr(int64(10))}, true},
		{"another flow blocks", guide, &model.KnowledgeFilter{FlowID: ptr(int64(99))}, false},
		{"a document without a flow is blocked by a flow", answer, &model.KnowledgeFilter{FlowID: ptr(int64(10))}, false},
		{
			"filters that all match pass", guide,
			&model.KnowledgeFilter{
				DocTypes:   []model.KnowledgeDocType{model.KnowledgeDocTypeGuide},
				GuideTypes: []model.KnowledgeGuideType{model.KnowledgeGuideTypePentest},
				FlowID:     ptr(int64(10)),
			},
			true,
		},
		{
			"one mismatching filter blocks the rest", guide,
			&model.KnowledgeFilter{DocTypes: []model.KnowledgeDocType{model.KnowledgeDocTypeGuide}, FlowID: ptr(int64(99))},
			false,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, passesSearchFilter(tc.doc, tc.filter))
		})
	}
}

// Rows are keyed by entry: userID 0 drives SearchDocuments, any other SearchUserDocuments.
func TestKnowledge_DoSearch_AsksTheScopedQueryAndKeepsWhatPassesTheFilter(t *testing.T) {
	type hit struct {
		ID    string
		Score float64
	}

	for _, tc := range []struct {
		name         string
		userID       int64
		maxBytes     int
		filter       *model.KnowledgeFilter
		limit        int
		vectors      [][]float32
		found        []database.SearchKnowledgeDocumentsRow
		wantUserID   sql.NullString
		wantLim      int32
		wantEmbedded string
		wantVector   string
		want         []hit
	}{
		{
			name:  "everyone's documents come back with their ids and scores",
			limit: 7,
			found: []database.SearchKnowledgeDocumentsRow{
				makeSearchRow("uuid-1", `{"doc_type":"answer"}`, 0.95),
				makeSearchRow("uuid-2", `{"doc_type":"guide"}`, 0.80),
			},
			wantLim:      7,
			wantEmbedded: "test query",
			wantVector:   "[0.1,0.2,0.3]",
			want:         []hit{{"uuid-1", 0.95}, {"uuid-2", 0.80}},
		},
		{name: "a limit of zero asks for the default", limit: 0, wantLim: 10, wantEmbedded: "test query", wantVector: "[0.1,0.2,0.3]"},
		{name: "a query past the embedding limit is cut before embedding", maxBytes: 4, limit: 5, wantLim: 5, wantEmbedded: "test", wantVector: "[0.1,0.2,0.3]"},
		{name: "an empty query vector is asked as an empty literal", vectors: [][]float32{{}}, limit: 5, wantLim: 5, wantEmbedded: "test query", wantVector: "[]"},
		{
			name:   "a doc type filter drops what the query found",
			filter: &model.KnowledgeFilter{DocTypes: []model.KnowledgeDocType{model.KnowledgeDocTypeGuide}},
			limit:  10,
			found: []database.SearchKnowledgeDocumentsRow{
				makeSearchRow("g1", `{"doc_type":"guide"}`, 0.9),
				makeSearchRow("a1", `{"doc_type":"answer"}`, 0.8),
			},
			wantLim:      10,
			wantEmbedded: "test query",
			wantVector:   "[0.1,0.2,0.3]",
			want:         []hit{{"g1", 0.9}},
		},
		{
			name:   "a flow filter drops other flows",
			filter: &model.KnowledgeFilter{FlowID: ptr(int64(10))},
			limit:  10,
			found: []database.SearchKnowledgeDocumentsRow{
				makeSearchRow("f10", `{"doc_type":"answer","flow_id":10}`, 0.9),
				makeSearchRow("f20", `{"doc_type":"answer","flow_id":20}`, 0.8),
			},
			wantLim:      10,
			wantEmbedded: "test query",
			wantVector:   "[0.1,0.2,0.3]",
			want:         []hit{{"f10", 0.9}},
		},
		{
			name:         "a user's search is scoped to that user",
			userID:       42,
			limit:        3,
			found:        []database.SearchKnowledgeDocumentsRow{makeSearchRow("uuid-u1", `{"doc_type":"answer","user_id":42}`, 0.88)},
			wantUserID:   sql.NullString{String: "42", Valid: true},
			wantLim:      3,
			wantEmbedded: "test query",
			wantVector:   "[0.1,0.2,0.3]",
			want:         []hit{{"uuid-u1", 0.88}},
		},
		{
			name:   "a user's search applies the filter too",
			userID: 42,
			filter: &model.KnowledgeFilter{CodeLangs: []string{"go"}},
			limit:  10,
			found: []database.SearchKnowledgeDocumentsRow{
				makeSearchRow("c1", `{"doc_type":"code","code_lang":"go"}`, 0.9),
				makeSearchRow("c2", `{"doc_type":"code","code_lang":"python"}`, 0.7),
			},
			wantUserID:   sql.NullString{String: "42", Valid: true},
			wantLim:      10,
			wantEmbedded: "test query",
			wantVector:   "[0.1,0.2,0.3]",
			want:         []hit{{"c1", 0.9}},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var asked []database.SearchUserKnowledgeDocumentsParams
			db := &mockDB{
				searchKnowledge: func(arg database.SearchKnowledgeDocumentsParams) ([]database.SearchKnowledgeDocumentsRow, error) {
					asked = append(asked, database.SearchUserKnowledgeDocumentsParams{Embedding: arg.Embedding, MaxDistance: arg.MaxDistance, Lim: arg.Lim})
					return tc.found, nil
				},
				searchUserKnowledge: func(arg database.SearchUserKnowledgeDocumentsParams) ([]database.SearchUserKnowledgeDocumentsRow, error) {
					asked = append(asked, arg)
					rows := make([]database.SearchUserKnowledgeDocumentsRow, 0, len(tc.found))
					for _, row := range tc.found {
						rows = append(rows, database.SearchUserKnowledgeDocumentsRow(row))
					}
					return rows, nil
				},
			}
			embedder := &mockEmbedder{vectors: tc.vectors}
			ks := NewKnowledgeStore(db, nil, embedder, nil, tc.maxBytes)

			var results []*model.KnowledgeDocumentWithScore
			var err error
			if tc.userID == 0 {
				results, err = ks.SearchDocuments(t.Context(), "test query", tc.filter, tc.limit)
			} else {
				results, err = ks.SearchUserDocuments(t.Context(), tc.userID, "test query", tc.filter, tc.limit)
			}
			require.NoError(t, err)

			require.Len(t, asked, 1)
			assert.Equal(t, tc.wantUserID, asked[0].UserID, "the admin query carries no user, a user's query always carries theirs")
			assert.Equal(t, tc.wantLim, asked[0].Lim)
			assert.InDelta(t, 0.8, asked[0].MaxDistance, 1e-6, "a similarity threshold of 0.2 is a cosine distance below 0.8")
			assert.Equal(t, tc.wantVector, asked[0].Embedding)
			assert.Equal(t, []string{tc.wantEmbedded}, embedder.embedded)

			var got []hit
			for _, result := range results {
				got = append(got, hit{result.Document.ID, result.Score})
				assert.Equal(t, "content of "+result.Document.ID, result.Document.Content, "a match is read with its content")
			}
			assert.Equal(t, tc.want, got)
		})
	}
}

// Rows are keyed by entry: userID 0 drives SearchDocuments, any other SearchUserDocuments.
func TestKnowledge_DoSearch_FailsWhenAnyStepFails(t *testing.T) {
	errEmbed := errors.New("embedding failed")
	errQuery := errors.New("query failed")

	for _, tc := range []struct {
		name     string
		userID   int64
		embedder embeddings.Embedder
		wantIs   error
		wantText string
	}{
		{name: "no embedder is configured", embedder: nil, wantText: "embedding provider is not available"},
		{name: "the embedder is unavailable", embedder: &mockEmbedder{unavailable: true}, wantText: "embedding provider is not available"},
		{name: "the query cannot be embedded", embedder: &mockEmbedder{err: errEmbed}, wantIs: errEmbed, wantText: "embed query"},
		{name: "the embedder returns no vector", embedder: &mockEmbedder{vectors: [][]float32{}}, wantText: "embedder returned no vectors for query"},
		{name: "the admin query fails", embedder: &mockEmbedder{}, wantIs: errQuery, wantText: "similarity search (admin)"},
		{name: "a user's query fails", userID: 42, embedder: &mockEmbedder{}, wantIs: errQuery, wantText: "similarity search (user 42)"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			db := &mockDB{
				searchKnowledge: func(database.SearchKnowledgeDocumentsParams) ([]database.SearchKnowledgeDocumentsRow, error) {
					return nil, errQuery
				},
				searchUserKnowledge: func(database.SearchUserKnowledgeDocumentsParams) ([]database.SearchUserKnowledgeDocumentsRow, error) {
					return nil, errQuery
				},
			}
			ks := NewKnowledgeStore(db, nil, tc.embedder, nil, 0)

			var err error
			if tc.userID == 0 {
				_, err = ks.SearchDocuments(t.Context(), "q", nil, 5)
			} else {
				_, err = ks.SearchUserDocuments(t.Context(), tc.userID, "q", nil, 5)
			}

			assert.ErrorContains(t, err, tc.wantText)
			if tc.wantIs != nil {
				assert.ErrorIs(t, err, tc.wantIs)
			}
		})
	}
}

func TestKnowledge_ListDocuments_ReadsEveryDocumentOrOneFlowThenFilters(t *testing.T) {
	errList := errors.New("list failed")

	for _, tc := range []struct {
		name     string
		filter   *model.KnowledgeFilter
		rows     []database.GetKnowledgeDocumentRow
		err      error
		wantFlow sql.NullString
		wantIDs  []string
		wantErr  string
	}{
		{
			name:    "without a flow every document is listed",
			rows:    []database.GetKnowledgeDocumentRow{makeRow("u1", "text", `{"doc_type":"answer"}`)},
			wantIDs: []string{"u1"},
		},
		{
			name:     "a flow filter lists that flow",
			filter:   &model.KnowledgeFilter{FlowID: ptr(int64(99))},
			rows:     []database.GetKnowledgeDocumentRow{makeRow("f1", "text", `{"doc_type":"answer","flow_id":99}`)},
			wantFlow: sql.NullString{String: "99", Valid: true},
			wantIDs:  []string{"f1"},
		},
		{
			name:   "the doc type filter applies after the query",
			filter: &model.KnowledgeFilter{DocTypes: []model.KnowledgeDocType{model.KnowledgeDocTypeGuide}},
			rows: []database.GetKnowledgeDocumentRow{
				makeRow("g1", "", `{"doc_type":"guide"}`),
				makeRow("a1", "", `{"doc_type":"answer"}`),
			},
			wantIDs: []string{"g1"},
		},
		{name: "a failed list is reported", err: errList, wantErr: "knowledge: list all"},
		{
			name:     "a failed flow list is reported",
			filter:   &model.KnowledgeFilter{FlowID: ptr(int64(99))},
			err:      errList,
			wantFlow: sql.NullString{String: "99", Valid: true},
			wantErr:  "knowledge: list by flow",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var askedFlow sql.NullString
			db := &mockDB{
				listAll: func() ([]database.ListAllKnowledgeDocumentsRow, error) {
					var rows []database.ListAllKnowledgeDocumentsRow
					for _, row := range tc.rows {
						rows = append(rows, database.ListAllKnowledgeDocumentsRow(row))
					}
					return rows, tc.err
				},
				listFlow: func(flowID sql.NullString) ([]database.ListFlowKnowledgeDocumentsRow, error) {
					askedFlow = flowID
					var rows []database.ListFlowKnowledgeDocumentsRow
					for _, row := range tc.rows {
						rows = append(rows, database.ListFlowKnowledgeDocumentsRow(row))
					}
					return rows, tc.err
				},
			}

			docs, err := NewKnowledgeStore(db, nil, nil, nil, 0).ListDocuments(t.Context(), tc.filter, false)

			assert.Equal(t, tc.wantFlow, askedFlow)
			if tc.wantErr != "" {
				assert.ErrorIs(t, err, errList)
				assert.ErrorContains(t, err, tc.wantErr)
				return
			}
			require.NoError(t, err)
			var ids []string
			for _, doc := range docs {
				ids = append(ids, doc.ID)
			}
			assert.Equal(t, tc.wantIDs, ids)
		})
	}
}

func TestKnowledge_ListUserDocuments_ReadsOnlyTheUsersDocuments(t *testing.T) {
	errList := errors.New("list failed")

	for _, tc := range []struct {
		name     string
		filter   *model.KnowledgeFilter
		rows     []database.GetKnowledgeDocumentRow
		err      error
		wantUser sql.NullString
		wantFlow sql.NullString
		wantIDs  []string
		wantErr  string
	}{
		{
			name:     "without a flow the user's own query is asked",
			rows:     []database.GetKnowledgeDocumentRow{makeRow("own", "", `{"doc_type":"answer","user_id":5}`)},
			wantUser: sql.NullString{String: "5", Valid: true},
			wantIDs:  []string{"own"},
		},
		{
			name:   "a flow filter keeps only documents the user owns",
			filter: &model.KnowledgeFilter{FlowID: ptr(int64(1))},
			rows: []database.GetKnowledgeDocumentRow{
				makeRow("own", "", `{"doc_type":"guide","user_id":5}`),
				makeRow("other", "", `{"doc_type":"guide","user_id":9}`),
				makeRow("ownerless", "", `{"doc_type":"answer"}`),
			},
			wantFlow: sql.NullString{String: "1", Valid: true},
			wantIDs:  []string{"own"},
		},
		{name: "a failed list is reported", err: errList, wantUser: sql.NullString{String: "5", Valid: true}, wantErr: "knowledge: list user docs"},
		{
			name:     "a failed flow list is reported",
			filter:   &model.KnowledgeFilter{FlowID: ptr(int64(1))},
			err:      errList,
			wantFlow: sql.NullString{String: "1", Valid: true},
			wantErr:  "knowledge: list by flow (user)",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var askedUser, askedFlow sql.NullString
			db := &mockDB{
				listUser: func(userID sql.NullString) ([]database.ListUserKnowledgeDocumentsRow, error) {
					askedUser = userID
					var rows []database.ListUserKnowledgeDocumentsRow
					for _, row := range tc.rows {
						rows = append(rows, database.ListUserKnowledgeDocumentsRow(row))
					}
					return rows, tc.err
				},
				listFlow: func(flowID sql.NullString) ([]database.ListFlowKnowledgeDocumentsRow, error) {
					askedFlow = flowID
					var rows []database.ListFlowKnowledgeDocumentsRow
					for _, row := range tc.rows {
						rows = append(rows, database.ListFlowKnowledgeDocumentsRow(row))
					}
					return rows, tc.err
				},
			}

			docs, err := NewKnowledgeStore(db, nil, nil, nil, 0).ListUserDocuments(t.Context(), 5, tc.filter, true)

			assert.Equal(t, tc.wantUser, askedUser)
			assert.Equal(t, tc.wantFlow, askedFlow)
			if tc.wantErr != "" {
				assert.ErrorIs(t, err, errList)
				assert.ErrorContains(t, err, tc.wantErr)
				return
			}
			require.NoError(t, err)
			var ids []string
			for _, doc := range docs {
				ids = append(ids, doc.ID)
			}
			assert.Equal(t, tc.wantIDs, ids)
		})
	}
}

func TestKnowledge_GetDocument_ReadsTheWholeDocument(t *testing.T) {
	t.Run("the row is read with its content", func(t *testing.T) {
		db := &mockDB{getKnowledge: func(uuid string) (database.GetKnowledgeDocumentRow, error) {
			return makeRow(uuid, "content", `{"doc_type":"code","code_lang":"go"}`), nil
		}}

		doc, err := NewKnowledgeStore(db, nil, nil, nil, 0).GetDocument(t.Context(), "abc123")

		require.NoError(t, err)
		assert.Equal(t, &model.KnowledgeDocument{
			ID: "abc123", Content: "content", DocType: model.KnowledgeDocTypeCode, CodeLang: ptr("go"),
		}, doc)
	})

	t.Run("a missing document is reported", func(t *testing.T) {
		db := &mockDB{getKnowledge: func(string) (database.GetKnowledgeDocumentRow, error) {
			return database.GetKnowledgeDocumentRow{}, sql.ErrNoRows
		}}

		_, err := NewKnowledgeStore(db, nil, nil, nil, 0).GetDocument(t.Context(), "xyz")

		assert.ErrorIs(t, err, sql.ErrNoRows)
		assert.ErrorContains(t, err, "knowledge: get document xyz")
	})
}

func TestKnowledge_GetUserDocument_ReadsOnlyTheUsersDocument(t *testing.T) {
	t.Run("the uuid and the user reach the query", func(t *testing.T) {
		var asked database.GetUserKnowledgeDocumentParams
		db := &mockDB{getUserKnowledge: func(arg database.GetUserKnowledgeDocumentParams) (database.GetUserKnowledgeDocumentRow, error) {
			asked = arg
			return database.GetUserKnowledgeDocumentRow(makeRow(arg.Uuid, "doc", `{"doc_type":"guide","guide_type":"use"}`)), nil
		}}

		doc, err := NewKnowledgeStore(db, nil, nil, nil, 0).GetUserDocument(t.Context(), 7, "doc-uuid")

		require.NoError(t, err)
		assert.Equal(t, database.GetUserKnowledgeDocumentParams{Uuid: "doc-uuid", UserID: sql.NullString{String: "7", Valid: true}}, asked)
		assert.Equal(t, &model.KnowledgeDocument{
			ID: "doc-uuid", Content: "doc", DocType: model.KnowledgeDocTypeGuide, GuideType: ptr(model.KnowledgeGuideTypeUse),
		}, doc)
	})

	t.Run("another user's document is not found", func(t *testing.T) {
		db := &mockDB{getUserKnowledge: func(database.GetUserKnowledgeDocumentParams) (database.GetUserKnowledgeDocumentRow, error) {
			return database.GetUserKnowledgeDocumentRow{}, sql.ErrNoRows
		}}

		_, err := NewKnowledgeStore(db, nil, nil, nil, 0).GetUserDocument(t.Context(), 7, "not-my-doc")

		assert.ErrorIs(t, err, sql.ErrNoRows)
		assert.ErrorContains(t, err, "knowledge: get user document not-my-doc")
	})
}

func TestKnowledge_DeleteDocument_DeletesAndAnnouncesOnlyWhatItFound(t *testing.T) {
	errDelete := errors.New("delete failed")

	for _, tc := range []struct {
		name       string
		getErr     error
		deleteErr  error
		wantDelete bool
		wantErr    error
	}{
		{name: "the document is deleted and announced", wantDelete: true},
		{name: "a missing document is neither deleted nor announced", getErr: sql.ErrNoRows, wantErr: sql.ErrNoRows},
		{name: "a failed delete is not announced", deleteErr: errDelete, wantDelete: true, wantErr: errDelete},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var deleted []sql.NullString
			db := &mockDB{
				getKnowledge: func(uuid string) (database.GetKnowledgeDocumentRow, error) {
					return makeRow(uuid, "content", `{"doc_type":"answer"}`), tc.getErr
				},
				deleteKnowledge: func(uuid sql.NullString) error {
					deleted = append(deleted, uuid)
					return tc.deleteErr
				},
			}
			pub := &mockPublisher{}

			err := NewKnowledgeStore(db, nil, nil, newPublisherFactory(pub), 0).DeleteDocument(t.Context(), 10, "doc1")

			if tc.wantDelete {
				assert.Equal(t, []sql.NullString{{String: "doc1", Valid: true}}, deleted)
			} else {
				assert.Empty(t, deleted)
			}
			if tc.wantErr != nil {
				assert.ErrorIs(t, err, tc.wantErr)
				assert.Empty(t, pub.deletedDocs)
				return
			}
			require.NoError(t, err)
			require.Len(t, pub.deletedDocs, 1)
			assert.Equal(t, "doc1", pub.deletedDocs[0].ID)
			assert.Equal(t, int64(10), pub.userID, "the event is scoped to the caller")
		})
	}
}

func TestKnowledge_DeleteUserDocument_DeletesOnlyTheUsersDocument(t *testing.T) {
	errDelete := errors.New("constraint violation")

	for _, tc := range []struct {
		name       string
		getErr     error
		deleteErr  error
		wantDelete bool
		wantErr    error
	}{
		{name: "the user's document is deleted and announced", wantDelete: true},
		{name: "another user's document is neither deleted nor announced", getErr: sql.ErrNoRows, wantErr: sql.ErrNoRows},
		{name: "a failed delete is not announced", deleteErr: errDelete, wantDelete: true, wantErr: errDelete},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var askedGet database.GetUserKnowledgeDocumentParams
			var deleted []database.DeleteUserKnowledgeDocumentParams
			db := &mockDB{
				getUserKnowledge: func(arg database.GetUserKnowledgeDocumentParams) (database.GetUserKnowledgeDocumentRow, error) {
					askedGet = arg
					return database.GetUserKnowledgeDocumentRow(makeRow(arg.Uuid, "doc", `{"doc_type":"answer"}`)), tc.getErr
				},
				deleteUserKnowledge: func(arg database.DeleteUserKnowledgeDocumentParams) error {
					deleted = append(deleted, arg)
					return tc.deleteErr
				},
			}
			pub := &mockPublisher{}

			err := NewKnowledgeStore(db, nil, nil, newPublisherFactory(pub), 0).DeleteUserDocument(t.Context(), 3, "my-doc")

			owner := sql.NullString{String: "3", Valid: true}
			assert.Equal(t, database.GetUserKnowledgeDocumentParams{Uuid: "my-doc", UserID: owner}, askedGet)
			if tc.wantDelete {
				assert.Equal(t, []database.DeleteUserKnowledgeDocumentParams{
					{Uuid: sql.NullString{String: "my-doc", Valid: true}, UserID: owner},
				}, deleted)
			} else {
				assert.Empty(t, deleted)
			}
			if tc.wantErr != nil {
				assert.ErrorIs(t, err, tc.wantErr)
				assert.Empty(t, pub.deletedDocs)
				return
			}
			require.NoError(t, err)
			require.Len(t, pub.deletedDocs, 1)
			assert.Equal(t, "my-doc", pub.deletedDocs[0].ID)
		})
	}
}

func TestKnowledge_CreateDocument_StoresAManualDocumentAsSentAndAnnouncesIt(t *testing.T) {
	for _, tc := range []struct {
		name         string
		maxBytes     int
		input        model.CreateKnowledgeDocumentInput
		wantEmbedded string
		wantMeta     knowledgeMeta
		want         model.KnowledgeDocument
	}{
		{
			name: "code with a description",
			input: model.CreateKnowledgeDocumentInput{
				DocType: model.KnowledgeDocTypeCode, Content: "  func main() {}\n", Question: "how to main?",
				Description: ptr("a Go main"), CodeLang: ptr("go"),
			},
			wantEmbedded: "func main() {}",
			wantMeta: knowledgeMeta{
				DocType: "code", UserID: 11, Question: "how to main?", Description: "a Go main", CodeLang: "go",
				PartSize: 17, TotalSize: 17, Manual: true,
			},
			want: model.KnowledgeDocument{
				DocType: model.KnowledgeDocTypeCode, Content: "  func main() {}\n", Question: "how to main?",
				Description: ptr("a Go main"), UserID: 11, CodeLang: ptr("go"), PartSize: 17, TotalSize: 17, Manual: true,
			},
		},
		{
			name:         "an answer leaves the optional fields unset",
			input:        model.CreateKnowledgeDocumentInput{DocType: model.KnowledgeDocTypeAnswer, Content: "c", Question: "q", AnswerType: ptr(model.KnowledgeAnswerTypeOther)},
			wantEmbedded: "c",
			wantMeta:     knowledgeMeta{DocType: "answer", UserID: 11, Question: "q", AnswerType: "other", PartSize: 1, TotalSize: 1, Manual: true},
			want: model.KnowledgeDocument{
				DocType: model.KnowledgeDocTypeAnswer, Content: "c", Question: "q", UserID: 11,
				AnswerType: ptr(model.KnowledgeAnswerTypeOther), PartSize: 1, TotalSize: 1, Manual: true,
			},
		},
		{
			name:         "a guide keeps its guide type",
			input:        model.CreateKnowledgeDocumentInput{DocType: model.KnowledgeDocTypeGuide, Content: "steps", Question: "q", GuideType: ptr(model.KnowledgeGuideTypeInstall)},
			wantEmbedded: "steps",
			wantMeta:     knowledgeMeta{DocType: "guide", UserID: 11, Question: "q", GuideType: "install", PartSize: 5, TotalSize: 5, Manual: true},
			want: model.KnowledgeDocument{
				DocType: model.KnowledgeDocTypeGuide, Content: "steps", Question: "q", UserID: 11,
				GuideType: ptr(model.KnowledgeGuideTypeInstall), PartSize: 5, TotalSize: 5, Manual: true,
			},
		},
		{
			name:         "content past the embedding limit is embedded cut and stored whole",
			maxBytes:     4,
			input:        model.CreateKnowledgeDocumentInput{DocType: model.KnowledgeDocTypeAnswer, Content: "  padded  ", Question: "q", AnswerType: ptr(model.KnowledgeAnswerTypeTool)},
			wantEmbedded: "padd",
			wantMeta:     knowledgeMeta{DocType: "answer", UserID: 11, Question: "q", AnswerType: "tool", PartSize: 10, TotalSize: 10, Manual: true},
			want: model.KnowledgeDocument{
				DocType: model.KnowledgeDocTypeAnswer, Content: "  padded  ", Question: "q", UserID: 11,
				AnswerType: ptr(model.KnowledgeAnswerTypeTool), PartSize: 10, TotalSize: 10, Manual: true,
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var stored []database.InsertKnowledgeDocumentParams
			db := &mockDB{insertKnowledge: func(arg database.InsertKnowledgeDocumentParams) (string, error) {
				stored = append(stored, arg)
				return "new-uuid", nil
			}}
			embedder := &mockEmbedder{}
			pub := &mockPublisher{}

			doc, err := NewKnowledgeStore(db, nil, embedder, newPublisherFactory(pub), tc.maxBytes).CreateDocument(t.Context(), 11, tc.input)
			require.NoError(t, err)

			require.Len(t, stored, 1)
			assert.Equal(t, sql.NullString{String: tc.want.Content, Valid: true}, stored[0].Document)
			assert.Equal(t, "[0.1,0.2,0.3]", stored[0].Embedding)
			assert.Equal(t, tc.wantMeta, parseMeta(string(stored[0].Cmetadata)))
			assert.Equal(t, []string{tc.wantEmbedded}, embedder.embedded)

			want := tc.want
			want.ID = "new-uuid"
			assert.Equal(t, &want, doc)
			assert.Equal(t, []*model.KnowledgeDocument{doc}, pub.createdDocs)
			assert.Equal(t, int64(11), pub.userID)
		})
	}
}

func TestKnowledge_CreateDocument_RefusesWithoutAnnouncing(t *testing.T) {
	errEmbed := errors.New("embedding failed")
	errInsert := errors.New("constraint error")
	answer := model.CreateKnowledgeDocumentInput{
		DocType: model.KnowledgeDocTypeAnswer, Content: "c", Question: "q", AnswerType: ptr(model.KnowledgeAnswerTypeOther),
	}

	for _, tc := range []struct {
		name       string
		embedder   embeddings.Embedder
		insertErr  error
		input      model.CreateKnowledgeDocumentInput
		wantIs     error
		wantText   string
		wantInsert bool
	}{
		{name: "no embedder is configured", embedder: nil, input: answer, wantText: "embedding provider is not available"},
		{name: "the embedder is unavailable", embedder: &mockEmbedder{unavailable: true}, input: answer, wantText: "embedding provider is not available"},
		{
			name:     "the content cannot be embedded",
			embedder: &mockEmbedder{err: errEmbed},
			input:    model.CreateKnowledgeDocumentInput{DocType: model.KnowledgeDocTypeGuide, Content: "c", Question: "q", GuideType: ptr(model.KnowledgeGuideTypeUse)},
			wantIs:   errEmbed,
			wantText: "knowledge: compute embedding",
		},
		{name: "the embedder returns no vector", embedder: &mockEmbedder{vectors: [][]float32{}}, input: answer, wantText: "knowledge: embedder returned no vectors"},
		{name: "the insert fails", embedder: &mockEmbedder{}, insertErr: errInsert, input: answer, wantIs: errInsert, wantText: "knowledge: create document", wantInsert: true},
		{
			name:     "an answer without an answer type",
			embedder: &mockEmbedder{},
			input:    model.CreateKnowledgeDocumentInput{DocType: model.KnowledgeDocTypeAnswer, Content: "c", Question: "q"},
			wantIs:   ErrInvalidDocument,
			wantText: "answer document requires answer type",
		},
		{
			name:     "a guide without a guide type",
			embedder: &mockEmbedder{},
			input:    model.CreateKnowledgeDocumentInput{DocType: model.KnowledgeDocTypeGuide, Content: "c", Question: "q"},
			wantIs:   ErrInvalidDocument,
			wantText: "guide document requires guide type",
		},
		{
			name:     "code without a language",
			embedder: &mockEmbedder{},
			input:    model.CreateKnowledgeDocumentInput{DocType: model.KnowledgeDocTypeCode, Content: "c", Question: "q"},
			wantIs:   ErrInvalidDocument,
			wantText: "code document requires code language",
		},
		{
			name:     "code whose language is only spaces",
			embedder: &mockEmbedder{},
			input:    model.CreateKnowledgeDocumentInput{DocType: model.KnowledgeDocTypeCode, Content: "c", Question: "q", CodeLang: ptr("   ")},
			wantIs:   ErrInvalidDocument,
			wantText: "code document requires code language",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			inserted := false
			db := &mockDB{insertKnowledge: func(database.InsertKnowledgeDocumentParams) (string, error) {
				inserted = true
				return "id", tc.insertErr
			}}
			pub := &mockPublisher{}

			_, err := NewKnowledgeStore(db, nil, tc.embedder, newPublisherFactory(pub), 0).CreateDocument(t.Context(), 11, tc.input)

			assert.ErrorContains(t, err, tc.wantText)
			if tc.wantIs != nil {
				assert.ErrorIs(t, err, tc.wantIs)
			}
			assert.Equal(t, tc.wantInsert, inserted)
			assert.Empty(t, pub.createdDocs, "subscribers were told about a document that was refused")
		})
	}
}

// Rows are keyed by entry: UpdateDocument unless userScoped, then UpdateUserDocument. The caller is always user 20.
func TestKnowledge_DoUpdate_MergesTheInputIntoTheStoredDocument(t *testing.T) {
	for _, tc := range []struct {
		name         string
		userScoped   bool
		maxBytes     int
		content      string
		cmetadata    string
		input        model.UpdateKnowledgeDocumentInput
		wantEmbedded string
		wantMeta     knowledgeMeta
	}{
		{
			name:      "the question and description are replaced and everything else is kept",
			content:   "old content",
			cmetadata: `{"doc_type":"guide","guide_type":"pentest","question":"original","description":"old","flow_id":1,"task_id":2,"subtask_id":3}`,
			input:     model.UpdateKnowledgeDocumentInput{Content: "new content", Question: ptr("new q"), Description: ptr("new")},
			wantMeta: knowledgeMeta{
				DocType: "guide", GuideType: "pentest", Question: "new q", Description: "new",
				FlowID: ptr(int64(1)), TaskID: ptr(int64(2)), SubtaskID: ptr(int64(3)), PartSize: 11, TotalSize: 11,
			},
		},
		{
			name:         "the content is written as it was sent, the line break it ends with included",
			content:      "old",
			cmetadata:    `{"doc_type":"guide","guide_type":"pentest","question":"q"}`,
			input:        model.UpdateKnowledgeDocumentInput{Content: "  # Title\n\ntext\n"},
			wantEmbedded: "# Title\n\ntext",
			wantMeta:     knowledgeMeta{DocType: "guide", GuideType: "pentest", Question: "q", PartSize: 16, TotalSize: 16},
		},
		{
			name:         "the sizes of a stored part move by the bytes of the content as it was sent",
			content:      "old",
			cmetadata:    `{"doc_type":"guide","guide_type":"pentest","question":"q","part_size":3,"total_size":9}`,
			input:        model.UpdateKnowledgeDocumentInput{Content: "  # Title\n\ntext\n"},
			wantEmbedded: "# Title\n\ntext",
			wantMeta:     knowledgeMeta{DocType: "guide", GuideType: "pentest", Question: "q", PartSize: 16, TotalSize: 22},
		},
		{
			name:      "guide to answer drops the guide type and takes the answer type",
			content:   "old",
			cmetadata: `{"doc_type":"guide","guide_type":"pentest","question":"q"}`,
			input: model.UpdateKnowledgeDocumentInput{
				Content: "new content", DocType: ptr(model.KnowledgeDocTypeAnswer), AnswerType: ptr(model.KnowledgeAnswerTypeVulnerability),
			},
			wantMeta: knowledgeMeta{DocType: "answer", AnswerType: "vulnerability", Question: "q", PartSize: 11, TotalSize: 11},
		},
		{
			name:      "answer to code drops the answer type and takes the language",
			content:   "old",
			cmetadata: `{"doc_type":"answer","answer_type":"vulnerability","question":"q"}`,
			input:     model.UpdateKnowledgeDocumentInput{Content: "code here", DocType: ptr(model.KnowledgeDocTypeCode), CodeLang: ptr("python")},
			wantMeta:  knowledgeMeta{DocType: "code", CodeLang: "python", Question: "q", PartSize: 9, TotalSize: 9},
		},
		{
			name:      "code to guide drops the language and takes the guide type",
			content:   "old",
			cmetadata: `{"doc_type":"code","code_lang":"go","question":"q"}`,
			input: model.UpdateKnowledgeDocumentInput{
				Content: "guide text", DocType: ptr(model.KnowledgeDocTypeGuide), GuideType: ptr(model.KnowledgeGuideTypePentest),
			},
			wantMeta: knowledgeMeta{DocType: "guide", GuideType: "pentest", Question: "q", PartSize: 10, TotalSize: 10},
		},
		{
			name:      "the same doc type keeps the stored sub-type",
			content:   "old",
			cmetadata: `{"doc_type":"guide","guide_type":"pentest","question":"q"}`,
			input:     model.UpdateKnowledgeDocumentInput{Content: "updated guide", DocType: ptr(model.KnowledgeDocTypeGuide)},
			wantMeta:  knowledgeMeta{DocType: "guide", GuideType: "pentest", Question: "q", PartSize: 13, TotalSize: 13},
		},
		{
			name:      "no doc type keeps the stored sub-type",
			content:   "old",
			cmetadata: `{"doc_type":"code","code_lang":"rust","question":"q"}`,
			input:     model.UpdateKnowledgeDocumentInput{Content: "updated code"},
			wantMeta:  knowledgeMeta{DocType: "code", CodeLang: "rust", Question: "q", PartSize: 12, TotalSize: 12},
		},
		{
			name:      "the same doc type with a new sub-type takes it",
			content:   "old",
			cmetadata: `{"doc_type":"answer","answer_type":"vulnerability","question":"q"}`,
			input: model.UpdateKnowledgeDocumentInput{
				Content: "updated", DocType: ptr(model.KnowledgeDocTypeAnswer), AnswerType: ptr(model.KnowledgeAnswerTypeCode),
			},
			wantMeta: knowledgeMeta{DocType: "answer", AnswerType: "code", Question: "q", PartSize: 7, TotalSize: 7},
		},
		{
			name:      "a single chunk takes the new length as both sizes",
			content:   "0123456789",
			cmetadata: `{"doc_type":"answer","answer_type":"other","part_size":10,"total_size":10}`,
			input:     model.UpdateKnowledgeDocumentInput{Content: "hello"},
			wantMeta:  knowledgeMeta{DocType: "answer", AnswerType: "other", PartSize: 5, TotalSize: 5},
		},
		{
			name:      "a shrinking chunk shrinks the whole document by the same amount",
			content:   string(make([]byte, 100)),
			cmetadata: `{"doc_type":"guide","guide_type":"use","part_size":100,"total_size":300}`,
			input:     model.UpdateKnowledgeDocumentInput{Content: string(make([]byte, 80))},
			wantMeta:  knowledgeMeta{DocType: "guide", GuideType: "use", PartSize: 80, TotalSize: 280},
		},
		{
			name:      "a growing chunk grows the whole document by the same amount",
			content:   string(make([]byte, 50)),
			cmetadata: `{"doc_type":"code","code_lang":"go","part_size":50,"total_size":150}`,
			input:     model.UpdateKnowledgeDocumentInput{Content: string(make([]byte, 70))},
			wantMeta:  knowledgeMeta{DocType: "code", CodeLang: "go", PartSize: 70, TotalSize: 170},
		},
		{
			name:      "a document stored without sizes takes the new length",
			content:   "old",
			cmetadata: `{"doc_type":"answer","answer_type":"other"}`,
			input:     model.UpdateKnowledgeDocumentInput{Content: "new content"},
			wantMeta:  knowledgeMeta{DocType: "answer", AnswerType: "other", PartSize: 11, TotalSize: 11},
		},
		{
			name:      "the stored owner is kept when someone else updates it",
			content:   "old",
			cmetadata: `{"doc_type":"answer","answer_type":"other","user_id":99,"question":"q"}`,
			input:     model.UpdateKnowledgeDocumentInput{Content: "new"},
			wantMeta:  knowledgeMeta{DocType: "answer", AnswerType: "other", UserID: 99, Question: "q", PartSize: 3, TotalSize: 3},
		},
		{
			name:         "content past the embedding limit is embedded cut and stored whole",
			maxBytes:     3,
			content:      "old",
			cmetadata:    `{"doc_type":"answer","answer_type":"other"}`,
			input:        model.UpdateKnowledgeDocumentInput{Content: "new content"},
			wantEmbedded: "new",
			wantMeta:     knowledgeMeta{DocType: "answer", AnswerType: "other", PartSize: 11, TotalSize: 11},
		},
		{
			name:       "a user's update reads the document through the ownership check",
			userScoped: true,
			content:    "old",
			cmetadata:  `{"doc_type":"answer","answer_type":"tool","question":"q","user_id":20}`,
			input:      model.UpdateKnowledgeDocumentInput{Content: "updated"},
			wantMeta:   knowledgeMeta{DocType: "answer", AnswerType: "tool", Question: "q", UserID: 20, PartSize: 7, TotalSize: 7},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var fetched []any
			var written []database.UpdateKnowledgeDocumentParams
			db := &mockDB{
				getKnowledge: func(uuid string) (database.GetKnowledgeDocumentRow, error) {
					fetched = append(fetched, uuid)
					return makeRow(uuid, tc.content, tc.cmetadata), nil
				},
				getUserKnowledge: func(arg database.GetUserKnowledgeDocumentParams) (database.GetUserKnowledgeDocumentRow, error) {
					fetched = append(fetched, arg)
					return database.GetUserKnowledgeDocumentRow(makeRow(arg.Uuid, tc.content, tc.cmetadata)), nil
				},
				updateKnowledge: func(arg database.UpdateKnowledgeDocumentParams) (database.UpdateKnowledgeDocumentRow, error) {
					written = append(written, arg)
					return database.UpdateKnowledgeDocumentRow{
						ID:        arg.Uuid.String,
						Document:  arg.Document.String,
						Cmetadata: sql.NullString{String: string(arg.Cmetadata.RawMessage), Valid: arg.Cmetadata.Valid},
					}, nil
				},
			}
			embedder := &mockEmbedder{}
			pub := &mockPublisher{}
			ks := NewKnowledgeStore(db, nil, embedder, newPublisherFactory(pub), tc.maxBytes)

			var doc *model.KnowledgeDocument
			var err error
			if tc.userScoped {
				doc, err = ks.UpdateUserDocument(t.Context(), 20, "doc-id", tc.input)
			} else {
				doc, err = ks.UpdateDocument(t.Context(), 20, "doc-id", tc.input)
			}
			require.NoError(t, err)

			if tc.userScoped {
				assert.Equal(t, []any{database.GetUserKnowledgeDocumentParams{Uuid: "doc-id", UserID: sql.NullString{String: "20", Valid: true}}}, fetched)
			} else {
				assert.Equal(t, []any{"doc-id"}, fetched)
			}
			require.Len(t, written, 1)
			assert.Equal(t, sql.NullString{String: "doc-id", Valid: true}, written[0].Uuid)
			assert.Equal(t, sql.NullString{String: tc.input.Content, Valid: true}, written[0].Document)
			assert.Equal(t, "[0.1,0.2,0.3]", written[0].Embedding)
			assert.True(t, written[0].Cmetadata.Valid, "an invalid metadata value would be written as NULL")
			assert.Equal(t, tc.wantMeta, parseMeta(string(written[0].Cmetadata.RawMessage)))
			wantEmbedded := tc.input.Content
			if tc.wantEmbedded != "" {
				wantEmbedded = tc.wantEmbedded
			}
			assert.Equal(t, []string{wantEmbedded}, embedder.embedded)

			assert.Equal(t, "doc-id", doc.ID)
			assert.Equal(t, tc.input.Content, doc.Content)
			assert.Equal(t, tc.wantMeta, metaFromDoc(doc), "the document handed back is the one written")
			assert.Equal(t, []*model.KnowledgeDocument{doc}, pub.updatedDocs)
			assert.Equal(t, int64(20), pub.userID, "the event is scoped to the caller, not the owner")
		})
	}
}

// Rows are keyed by entry: UpdateDocument unless userScoped, then UpdateUserDocument.
func TestKnowledge_DoUpdate_RefusesWithoutAnnouncing(t *testing.T) {
	errEmbed := errors.New("embedding failed")
	errWrite := errors.New("constraint error")
	guide := `{"doc_type":"guide","guide_type":"pentest","question":"q"}`

	for _, tc := range []struct {
		name       string
		userScoped bool
		getErr     error
		writeErr   error
		embedder   embeddings.Embedder
		cmetadata  string
		input      model.UpdateKnowledgeDocumentInput
		wantIs     error
		wantText   string
		wantWrite  bool
	}{
		{name: "no embedder is configured", embedder: nil, cmetadata: guide, wantText: "embedding provider is not available"},
		{name: "the embedder is unavailable", embedder: &mockEmbedder{unavailable: true}, cmetadata: guide, wantText: "embedding provider is not available"},
		{name: "the content cannot be embedded", embedder: &mockEmbedder{err: errEmbed}, cmetadata: guide, wantIs: errEmbed, wantText: "knowledge: compute embedding"},
		{name: "the embedder returns no vector", embedder: &mockEmbedder{vectors: [][]float32{}}, cmetadata: guide, wantText: "knowledge: embedder returned no vectors"},
		{name: "the write fails", embedder: &mockEmbedder{}, writeErr: errWrite, cmetadata: guide, wantIs: errWrite, wantText: "knowledge: update document doc-id", wantWrite: true},
		{
			name:      "a type change that leaves no sub-type",
			embedder:  &mockEmbedder{},
			cmetadata: `{"doc_type":"guide","guide_type":"install","question":"q"}`,
			input:     model.UpdateKnowledgeDocumentInput{DocType: ptr(model.KnowledgeDocTypeAnswer)},
			wantIs:    ErrInvalidDocument,
			wantText:  "answer document requires answer type",
		},
		{name: "a missing document", embedder: &mockEmbedder{}, getErr: sql.ErrNoRows, cmetadata: guide, wantIs: sql.ErrNoRows, wantText: "knowledge: get document doc-id"},
		{
			name:       "another user's document",
			userScoped: true,
			embedder:   &mockEmbedder{},
			getErr:     sql.ErrNoRows,
			cmetadata:  guide,
			wantIs:     sql.ErrNoRows,
			wantText:   "knowledge: get user document doc-id",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			written := false
			db := &mockDB{
				getKnowledge: func(uuid string) (database.GetKnowledgeDocumentRow, error) {
					return makeRow(uuid, "old", tc.cmetadata), tc.getErr
				},
				getUserKnowledge: func(arg database.GetUserKnowledgeDocumentParams) (database.GetUserKnowledgeDocumentRow, error) {
					return database.GetUserKnowledgeDocumentRow(makeRow(arg.Uuid, "old", tc.cmetadata)), tc.getErr
				},
				updateKnowledge: func(database.UpdateKnowledgeDocumentParams) (database.UpdateKnowledgeDocumentRow, error) {
					written = true
					return database.UpdateKnowledgeDocumentRow{}, tc.writeErr
				},
			}
			pub := &mockPublisher{}
			ks := NewKnowledgeStore(db, nil, tc.embedder, newPublisherFactory(pub), 0)
			input := tc.input
			input.Content = "new"

			var err error
			if tc.userScoped {
				_, err = ks.UpdateUserDocument(t.Context(), 20, "doc-id", input)
			} else {
				_, err = ks.UpdateDocument(t.Context(), 20, "doc-id", input)
			}

			assert.ErrorContains(t, err, tc.wantText)
			if tc.wantIs != nil {
				assert.ErrorIs(t, err, tc.wantIs)
			}
			assert.Equal(t, tc.wantWrite, written)
			assert.Empty(t, pub.updatedDocs)
		})
	}
}

// Rows are keyed by entry: RenameDocument unless userScoped, then RenameUserDocument. No embedder is configured.
func TestKnowledge_DoRename_ReplacesOnlyTheQuestion(t *testing.T) {
	for _, tc := range []struct {
		name       string
		userScoped bool
		cmetadata  string
		question   string
		wantMeta   knowledgeMeta
	}{
		{
			name: "every other stored field is kept and the question is trimmed",
			cmetadata: `{"doc_type":"answer","answer_type":"vulnerability","question":"original","user_id":7,"manual":true,` +
				`"flow_id":1,"task_id":2,"subtask_id":3,"description":"d","part_size":17,"total_size":17}`,
			question: "  RENAMED  ",
			wantMeta: knowledgeMeta{
				DocType: "answer", AnswerType: "vulnerability", Question: "RENAMED", UserID: 7, Manual: true,
				FlowID: ptr(int64(1)), TaskID: ptr(int64(2)), SubtaskID: ptr(int64(3)), Description: "d", PartSize: 17, TotalSize: 17,
			},
		},
		{
			name:       "a user's rename reads the document through the ownership check",
			userScoped: true,
			cmetadata:  `{"doc_type":"answer","question":"old","user_id":30}`,
			question:   "new title",
			wantMeta:   knowledgeMeta{DocType: "answer", Question: "new title", UserID: 30},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var fetched []any
			var written []database.UpdateKnowledgeDocumentMetadataParams
			db := &mockDB{
				getKnowledge: func(uuid string) (database.GetKnowledgeDocumentRow, error) {
					fetched = append(fetched, uuid)
					return makeRow(uuid, "PRESERVED CONTENT", tc.cmetadata), nil
				},
				getUserKnowledge: func(arg database.GetUserKnowledgeDocumentParams) (database.GetUserKnowledgeDocumentRow, error) {
					fetched = append(fetched, arg)
					return database.GetUserKnowledgeDocumentRow(makeRow(arg.Uuid, "PRESERVED CONTENT", tc.cmetadata)), nil
				},
				updateKnowledgeMeta: func(arg database.UpdateKnowledgeDocumentMetadataParams) (database.UpdateKnowledgeDocumentMetadataRow, error) {
					written = append(written, arg)
					return database.UpdateKnowledgeDocumentMetadataRow{
						ID:        arg.Uuid.String,
						Document:  "PRESERVED CONTENT",
						Cmetadata: sql.NullString{String: string(arg.Cmetadata.RawMessage), Valid: arg.Cmetadata.Valid},
					}, nil
				},
			}
			pub := &mockPublisher{}
			ks := NewKnowledgeStore(db, nil, nil, newPublisherFactory(pub), 0)

			var doc *model.KnowledgeDocument
			var err error
			if tc.userScoped {
				doc, err = ks.RenameUserDocument(t.Context(), 30, "doc-id", tc.question)
			} else {
				doc, err = ks.RenameDocument(t.Context(), 30, "doc-id", tc.question)
			}
			require.NoError(t, err, "a rename must not need an embedder")

			if tc.userScoped {
				assert.Equal(t, []any{database.GetUserKnowledgeDocumentParams{Uuid: "doc-id", UserID: sql.NullString{String: "30", Valid: true}}}, fetched)
			} else {
				assert.Equal(t, []any{"doc-id"}, fetched)
			}
			require.Len(t, written, 1)
			assert.Equal(t, sql.NullString{String: "doc-id", Valid: true}, written[0].Uuid)
			assert.True(t, written[0].Cmetadata.Valid, "an invalid metadata value would be written as NULL")
			assert.Equal(t, tc.wantMeta, parseMeta(string(written[0].Cmetadata.RawMessage)))

			assert.Equal(t, "PRESERVED CONTENT", doc.Content)
			assert.Equal(t, tc.wantMeta.Question, doc.Question)
			assert.Equal(t, []*model.KnowledgeDocument{doc}, pub.updatedDocs)
			assert.Equal(t, int64(30), pub.userID)
		})
	}
}

// Rows are keyed by entry: RenameDocument unless userScoped, then RenameUserDocument.
func TestKnowledge_DoRename_RefusesWithoutAnnouncing(t *testing.T) {
	errWrite := errors.New("write failed")

	for _, tc := range []struct {
		name       string
		userScoped bool
		getErr     error
		writeErr   error
		wantIs     error
		wantText   string
		wantWrite  bool
	}{
		{name: "a missing document", getErr: sql.ErrNoRows, wantIs: sql.ErrNoRows, wantText: "knowledge: get document doc-id"},
		{name: "another user's document", userScoped: true, getErr: sql.ErrNoRows, wantIs: sql.ErrNoRows, wantText: "knowledge: get user document doc-id"},
		{name: "the write fails", writeErr: errWrite, wantIs: errWrite, wantText: "knowledge: rename document doc-id", wantWrite: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			written := false
			db := &mockDB{
				getKnowledge: func(uuid string) (database.GetKnowledgeDocumentRow, error) {
					return makeRow(uuid, "kept", `{"doc_type":"answer","question":"old"}`), tc.getErr
				},
				getUserKnowledge: func(arg database.GetUserKnowledgeDocumentParams) (database.GetUserKnowledgeDocumentRow, error) {
					return database.GetUserKnowledgeDocumentRow(makeRow(arg.Uuid, "kept", `{"doc_type":"answer","question":"old"}`)), tc.getErr
				},
				updateKnowledgeMeta: func(database.UpdateKnowledgeDocumentMetadataParams) (database.UpdateKnowledgeDocumentMetadataRow, error) {
					written = true
					return database.UpdateKnowledgeDocumentMetadataRow{}, tc.writeErr
				},
			}
			pub := &mockPublisher{}
			ks := NewKnowledgeStore(db, nil, nil, newPublisherFactory(pub), 0)

			var err error
			if tc.userScoped {
				_, err = ks.RenameUserDocument(t.Context(), 30, "doc-id", "evil rename")
			} else {
				_, err = ks.RenameDocument(t.Context(), 30, "doc-id", "evil rename")
			}

			assert.ErrorIs(t, err, tc.wantIs)
			assert.ErrorContains(t, err, tc.wantText)
			assert.Equal(t, tc.wantWrite, written)
			assert.Empty(t, pub.updatedDocs)
		})
	}
}
