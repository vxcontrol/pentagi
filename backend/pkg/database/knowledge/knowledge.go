package knowledge

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"slices"
	"strconv"
	"strings"

	"pentagi/pkg/database"
	"pentagi/pkg/graph/model"
	"pentagi/pkg/graph/subscriptions"
	"pentagi/pkg/providers/embeddings"

	"github.com/sqlc-dev/pqtype"
	"github.com/vxcontrol/langchaingo/schema"
	"github.com/vxcontrol/langchaingo/vectorstores"
)

const (
	collectionName         = "langchain"
	defaultSearchLimit     = 10
	defaultSearchThreshold = float32(0.2)
	manualDocumentFlag     = "manual"
)

// PublisherFactory creates a per-user KnowledgePublisher.
// Matches subscriptions.SubscriptionsController.NewKnowledgePublisher signature.
type PublisherFactory func(userID int64) subscriptions.KnowledgePublisher

// KnowledgeStore covers all GraphQL knowledge operations:
//   - Admin read methods have no user_id filtering.
//   - User-scoped read methods filter cmetadata by user_id.
//   - Write methods always record the caller's userID for event scoping and
//     are added to cmetadata; user-scoped delete/update also enforce ownership.
type KnowledgeStore interface {
	// Admin reads
	ListDocuments(ctx context.Context, filter *model.KnowledgeFilter, withContent bool) ([]*model.KnowledgeDocument, error)
	GetDocument(ctx context.Context, id string) (*model.KnowledgeDocument, error)
	SearchDocuments(ctx context.Context, query string, filter *model.KnowledgeFilter, limit int) ([]*model.KnowledgeDocumentWithScore, error)

	// User-scoped reads (filter by user_id in cmetadata)
	ListUserDocuments(ctx context.Context, userID int64, filter *model.KnowledgeFilter, withContent bool) ([]*model.KnowledgeDocument, error)
	GetUserDocument(ctx context.Context, userID int64, id string) (*model.KnowledgeDocument, error)
	SearchUserDocuments(ctx context.Context, userID int64, query string, filter *model.KnowledgeFilter, limit int) ([]*model.KnowledgeDocumentWithScore, error)

	// Writes
	CreateDocument(ctx context.Context, userID int64, input model.CreateKnowledgeDocumentInput) (*model.KnowledgeDocument, error)
	UpdateDocument(ctx context.Context, userID int64, id string, input model.UpdateKnowledgeDocumentInput) (*model.KnowledgeDocument, error)
	UpdateUserDocument(ctx context.Context, userID int64, id string, input model.UpdateKnowledgeDocumentInput) (*model.KnowledgeDocument, error)
	DeleteDocument(ctx context.Context, userID int64, id string) error
	DeleteUserDocument(ctx context.Context, userID int64, id string) error
}

type knowledgeStore struct {
	db       database.Querier
	store    vectorstores.VectorStore // may be nil when no embedder is configured
	embedder embeddings.Embedder      // used for computing new embeddings on create/update
	newKnp   PublisherFactory
}

// NewKnowledgeStore constructs a KnowledgeStore.
//   - store and embedder may be nil when the embedding provider is not configured;
//     in that case create/update/search return a descriptive error while
//     list/get/delete still work.
//   - newKnp is called with the acting user's ID on each write to create a
//     correctly scoped event publisher.
func NewKnowledgeStore(
	db database.Querier,
	store vectorstores.VectorStore,
	embedder embeddings.Embedder,
	newKnp PublisherFactory,
) KnowledgeStore {
	return &knowledgeStore{
		db:       db,
		store:    store,
		embedder: embedder,
		newKnp:   newKnp,
	}
}

// ---- cmetadata helpers -------------------------------------------------------

// knowledgeMeta mirrors the flat JSON stored in langchain_pg_embedding.cmetadata.
type knowledgeMeta struct {
	DocType     string `json:"doc_type,omitempty"`
	UserID      int64  `json:"user_id,omitempty"`
	FlowID      *int64 `json:"flow_id,omitempty"`
	TaskID      *int64 `json:"task_id,omitempty"`
	SubtaskID   *int64 `json:"subtask_id,omitempty"`
	Question    string `json:"question,omitempty"`
	Description string `json:"description,omitempty"`
	GuideType   string `json:"guide_type,omitempty"`
	AnswerType  string `json:"answer_type,omitempty"`
	CodeLang    string `json:"code_lang,omitempty"`
	PartSize    int    `json:"part_size,omitempty"`
	TotalSize   int    `json:"total_size,omitempty"`
	Manual      bool   `json:"manual,omitempty"`
}

func parseMeta(raw string) knowledgeMeta {
	var m knowledgeMeta
	if raw != "" {
		_ = json.Unmarshal([]byte(raw), &m)
	}
	return m
}

func nullStr(ns sql.NullString) string {
	if ns.Valid {
		return ns.String
	}
	return "{}"
}

func nsOf(s string) sql.NullString {
	return sql.NullString{String: s, Valid: true}
}

// formatVector converts a float32 slice into the PostgreSQL vector literal '[f1,f2,...]'.
func formatVector(v []float32) string {
	strs := make([]string, len(v))
	for i, f := range v {
		strs[i] = strconv.FormatFloat(float64(f), 'f', -1, 32)
	}
	return "[" + strings.Join(strs, ",") + "]"
}

// metaToJSON serialises a knowledgeMeta into a pqtype.NullRawMessage for sqlc.
func metaToJSON(m knowledgeMeta) (pqtype.NullRawMessage, error) {
	b, err := json.Marshal(m)
	if err != nil {
		return pqtype.NullRawMessage{}, err
	}
	return pqtype.NullRawMessage{RawMessage: b, Valid: true}, nil
}

// rowToModel converts a sqlc knowledge row into a GraphQL model.
func rowToModel(id, document, cmetadataRaw string, withContent bool) *model.KnowledgeDocument {
	meta := parseMeta(cmetadataRaw)
	content := document
	if !withContent {
		content = ""
	}
	return metaToModelDoc(id, content, meta)
}

func metaToModelDoc(id, content string, meta knowledgeMeta) *model.KnowledgeDocument {
	doc := &model.KnowledgeDocument{
		ID:        id,
		UserID:    meta.UserID,
		Content:   content,
		Question:  meta.Question,
		PartSize:  meta.PartSize,
		TotalSize: meta.TotalSize,
		Manual:    meta.Manual,
	}

	switch meta.DocType {
	case "answer":
		doc.DocType = model.KnowledgeDocTypeAnswer
	case "guide":
		doc.DocType = model.KnowledgeDocTypeGuide
	case "code":
		doc.DocType = model.KnowledgeDocTypeCode
	default:
		doc.DocType = model.KnowledgeDocTypeAnswer
	}

	if meta.FlowID != nil {
		doc.FlowID = meta.FlowID
	}
	if meta.TaskID != nil {
		doc.TaskID = meta.TaskID
	}
	if meta.SubtaskID != nil {
		doc.SubtaskID = meta.SubtaskID
	}
	if meta.Description != "" {
		d := meta.Description
		doc.Description = &d
	}
	if meta.GuideType != "" {
		gt := model.KnowledgeGuideType(meta.GuideType)
		doc.GuideType = &gt
	}
	if meta.AnswerType != "" {
		at := model.KnowledgeAnswerType(meta.AnswerType)
		doc.AnswerType = &at
	}
	if meta.CodeLang != "" {
		cl := meta.CodeLang
		doc.CodeLang = &cl
	}
	return doc
}

// applyGoFilters applies optional filters that cannot be expressed as single SQL predicates.
func applyGoFilters(docs []*model.KnowledgeDocument, filter *model.KnowledgeFilter) []*model.KnowledgeDocument {
	if filter == nil {
		return docs
	}
	result := make([]*model.KnowledgeDocument, 0, len(docs))
	for _, doc := range docs {
		if len(filter.DocTypes) > 0 && !slices.Contains(filter.DocTypes, doc.DocType) {
			continue
		}
		if len(filter.GuideTypes) > 0 && (doc.GuideType == nil || !slices.Contains(filter.GuideTypes, *doc.GuideType)) {
			continue
		}
		if len(filter.AnswerTypes) > 0 && (doc.AnswerType == nil || !slices.Contains(filter.AnswerTypes, *doc.AnswerType)) {
			continue
		}
		if len(filter.CodeLangs) > 0 && (doc.CodeLang == nil || !slices.Contains(filter.CodeLangs, *doc.CodeLang)) {
			continue
		}
		if filter.Manual != nil && doc.Manual != *filter.Manual {
			continue
		}
		result = append(result, doc)
	}
	return result
}

// buildSearchFilters converts a KnowledgeFilter to a pgvector filter map.
// Only single-value equality filters are expressed here; arrays are applied in Go.
func buildSearchFilters(userID int64, filter *model.KnowledgeFilter) map[string]any {
	f := make(map[string]any)
	if userID > 0 {
		f["user_id"] = strconv.FormatInt(userID, 10)
	}
	if filter == nil {
		return f
	}
	if len(filter.DocTypes) == 1 {
		f["doc_type"] = string(filter.DocTypes[0])
	}
	if len(filter.GuideTypes) == 1 {
		f["guide_type"] = string(filter.GuideTypes[0])
	}
	if len(filter.AnswerTypes) == 1 {
		f["answer_type"] = string(filter.AnswerTypes[0])
	}
	if len(filter.CodeLangs) == 1 {
		f["code_lang"] = filter.CodeLangs[0]
	}
	if filter.FlowID != nil {
		f["flow_id"] = strconv.FormatInt(*filter.FlowID, 10)
	}
	return f
}

func (ks *knowledgeStore) requireStore() error {
	if ks.store == nil {
		return fmt.Errorf("knowledge: embedding provider is not configured")
	}
	return nil
}

func (ks *knowledgeStore) requireEmbedder() error {
	if ks.embedder == nil || !ks.embedder.IsAvailable() {
		return fmt.Errorf("knowledge: embedding provider is not available")
	}
	return nil
}

// ---- ListDocuments (admin) --------------------------------------------------

func (ks *knowledgeStore) ListDocuments(ctx context.Context, filter *model.KnowledgeFilter, withContent bool) ([]*model.KnowledgeDocument, error) {
	var docs []*model.KnowledgeDocument

	if filter != nil && filter.FlowID != nil {
		rows, err := ks.db.ListFlowKnowledgeDocuments(ctx, nsOf(strconv.FormatInt(*filter.FlowID, 10)))
		if err != nil {
			return nil, fmt.Errorf("knowledge: list by flow: %w", err)
		}
		for _, r := range rows {
			docs = append(docs, rowToModel(r.ID, r.Document, nullStr(r.Cmetadata), withContent))
		}
	} else {
		rows, err := ks.db.ListAllKnowledgeDocuments(ctx)
		if err != nil {
			return nil, fmt.Errorf("knowledge: list all: %w", err)
		}
		for _, r := range rows {
			docs = append(docs, rowToModel(r.ID, r.Document, nullStr(r.Cmetadata), withContent))
		}
	}

	return applyGoFilters(docs, filter), nil
}

// ---- ListUserDocuments (user-scoped) ----------------------------------------

func (ks *knowledgeStore) ListUserDocuments(ctx context.Context, userID int64, filter *model.KnowledgeFilter, withContent bool) ([]*model.KnowledgeDocument, error) {
	userIDStr := strconv.FormatInt(userID, 10)
	var docs []*model.KnowledgeDocument

	if filter != nil && filter.FlowID != nil {
		// Flow-scoped listing; user_id check applied in Go for safety.
		rows, err := ks.db.ListFlowKnowledgeDocuments(ctx, nsOf(strconv.FormatInt(*filter.FlowID, 10)))
		if err != nil {
			return nil, fmt.Errorf("knowledge: list by flow (user): %w", err)
		}
		for _, r := range rows {
			meta := parseMeta(nullStr(r.Cmetadata))
			if strconv.FormatInt(meta.UserID, 10) != userIDStr {
				continue
			}
			docs = append(docs, rowToModel(r.ID, r.Document, nullStr(r.Cmetadata), withContent))
		}
	} else {
		rows, err := ks.db.ListUserKnowledgeDocuments(ctx, nsOf(userIDStr))
		if err != nil {
			return nil, fmt.Errorf("knowledge: list user docs: %w", err)
		}
		for _, r := range rows {
			docs = append(docs, rowToModel(r.ID, r.Document, nullStr(r.Cmetadata), withContent))
		}
	}

	return applyGoFilters(docs, filter), nil
}

// ---- GetDocument (admin) ----------------------------------------------------

func (ks *knowledgeStore) GetDocument(ctx context.Context, id string) (*model.KnowledgeDocument, error) {
	row, err := ks.db.GetKnowledgeDocument(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("knowledge: get document %s: %w", id, err)
	}
	return rowToModel(row.ID, row.Document, nullStr(row.Cmetadata), true), nil
}

// ---- GetUserDocument (user-scoped) ------------------------------------------

func (ks *knowledgeStore) GetUserDocument(ctx context.Context, userID int64, id string) (*model.KnowledgeDocument, error) {
	row, err := ks.db.GetUserKnowledgeDocument(ctx, database.GetUserKnowledgeDocumentParams{
		Uuid:      id,
		Cmetadata: nsOf(strconv.FormatInt(userID, 10)),
	})
	if err != nil {
		return nil, fmt.Errorf("knowledge: get user document %s: %w", id, err)
	}
	return rowToModel(row.ID, row.Document, nullStr(row.Cmetadata), true), nil
}

// ---- SearchDocuments (admin) ------------------------------------------------

func (ks *knowledgeStore) SearchDocuments(ctx context.Context, query string, filter *model.KnowledgeFilter, limit int) ([]*model.KnowledgeDocumentWithScore, error) {
	return ks.doSearch(ctx, 0, query, filter, limit)
}

// ---- SearchUserDocuments (user-scoped) --------------------------------------

func (ks *knowledgeStore) SearchUserDocuments(ctx context.Context, userID int64, query string, filter *model.KnowledgeFilter, limit int) ([]*model.KnowledgeDocumentWithScore, error) {
	return ks.doSearch(ctx, userID, query, filter, limit)
}

func (ks *knowledgeStore) doSearch(ctx context.Context, userID int64, query string, filter *model.KnowledgeFilter, limit int) ([]*model.KnowledgeDocumentWithScore, error) {
	if err := ks.requireStore(); err != nil {
		return nil, err
	}
	if limit <= 0 {
		limit = defaultSearchLimit
	}

	pgFilters := buildSearchFilters(userID, filter)
	opts := []vectorstores.Option{
		vectorstores.WithScoreThreshold(defaultSearchThreshold),
	}
	if len(pgFilters) > 0 {
		opts = append(opts, vectorstores.WithFilters(pgFilters))
	}

	schemaDocs, err := ks.store.SimilaritySearch(ctx, query, limit, opts...)
	if err != nil {
		return nil, fmt.Errorf("knowledge: similarity search: %w", err)
	}

	results := make([]*model.KnowledgeDocumentWithScore, 0, len(schemaDocs))
	for _, sd := range schemaDocs {
		doc := schemaDocToModel(sd)
		if !passesGoFilter(doc, filter) {
			continue
		}
		results = append(results, &model.KnowledgeDocumentWithScore{
			Score:    float64(sd.Score),
			Document: doc,
		})
	}
	return results, nil
}

func schemaDocToModel(sd schema.Document) *model.KnowledgeDocument {
	raw, _ := json.Marshal(sd.Metadata)
	meta := parseMeta(string(raw))
	return metaToModelDoc("", sd.PageContent, meta)
}

func passesGoFilter(doc *model.KnowledgeDocument, filter *model.KnowledgeFilter) bool {
	if filter == nil {
		return true
	}
	if len(filter.DocTypes) > 1 && !slices.Contains(filter.DocTypes, doc.DocType) {
		return false
	}
	if len(filter.GuideTypes) > 1 && (doc.GuideType == nil || !slices.Contains(filter.GuideTypes, *doc.GuideType)) {
		return false
	}
	if len(filter.AnswerTypes) > 1 && (doc.AnswerType == nil || !slices.Contains(filter.AnswerTypes, *doc.AnswerType)) {
		return false
	}
	if len(filter.CodeLangs) > 1 && (doc.CodeLang == nil || !slices.Contains(filter.CodeLangs, *doc.CodeLang)) {
		return false
	}
	if filter.Manual != nil && doc.Manual != *filter.Manual {
		return false
	}
	return true
}

// ---- CreateDocument ---------------------------------------------------------

func (ks *knowledgeStore) CreateDocument(ctx context.Context, userID int64, input model.CreateKnowledgeDocumentInput) (*model.KnowledgeDocument, error) {
	if err := ks.requireStore(); err != nil {
		return nil, err
	}

	meta := knowledgeMeta{
		DocType:  string(input.DocType),
		UserID:   userID,
		Question: input.Question,
		Manual:   true,
	}
	if input.Description != nil {
		meta.Description = *input.Description
	}
	if input.GuideType != nil {
		meta.GuideType = string(*input.GuideType)
	}
	if input.AnswerType != nil {
		meta.AnswerType = string(*input.AnswerType)
	}
	if input.CodeLang != nil {
		meta.CodeLang = *input.CodeLang
	}

	content := strings.TrimSpace(input.Content)
	meta.PartSize = len(content)
	meta.TotalSize = len(content)

	metaMap := metaToMap(meta)
	ids, err := ks.store.AddDocuments(ctx, []schema.Document{{
		PageContent: content,
		Metadata:    metaMap,
	}})
	if err != nil {
		return nil, fmt.Errorf("knowledge: create document: %w", err)
	}
	if len(ids) == 0 {
		return nil, fmt.Errorf("knowledge: create returned no IDs")
	}

	doc := metaToModelDoc(ids[0], content, meta)
	ks.newKnp(userID).KnowledgeDocumentCreated(ctx, doc)
	return doc, nil
}

// ---- UpdateDocument (admin) -------------------------------------------------

func (ks *knowledgeStore) UpdateDocument(ctx context.Context, userID int64, id string, input model.UpdateKnowledgeDocumentInput) (*model.KnowledgeDocument, error) {
	existing, err := ks.GetDocument(ctx, id)
	if err != nil {
		return nil, err
	}
	return ks.doUpdate(ctx, userID, id, existing, input)
}

// ---- UpdateUserDocument (user-scoped) ---------------------------------------

func (ks *knowledgeStore) UpdateUserDocument(ctx context.Context, userID int64, id string, input model.UpdateKnowledgeDocumentInput) (*model.KnowledgeDocument, error) {
	existing, err := ks.GetUserDocument(ctx, userID, id)
	if err != nil {
		return nil, err
	}
	return ks.doUpdate(ctx, userID, id, existing, input)
}

func (ks *knowledgeStore) doUpdate(ctx context.Context, userID int64, id string, existing *model.KnowledgeDocument, input model.UpdateKnowledgeDocumentInput) (*model.KnowledgeDocument, error) {
	if err := ks.requireEmbedder(); err != nil {
		return nil, err
	}

	// Build updated metadata from existing document.
	// Preserve the original owner (UserID) — the caller's identity (userID) is
	// used only for event scoping via the publisher, not stored in cmetadata.
	meta := knowledgeMeta{
		DocType:  string(existing.DocType),
		UserID:   existing.UserID,
		Manual:   existing.Manual,
		Question: existing.Question,
	}
	if existing.FlowID != nil {
		meta.FlowID = existing.FlowID
	}
	if existing.TaskID != nil {
		meta.TaskID = existing.TaskID
	}
	if existing.SubtaskID != nil {
		meta.SubtaskID = existing.SubtaskID
	}
	if existing.Description != nil {
		meta.Description = *existing.Description
	}
	if existing.GuideType != nil {
		meta.GuideType = string(*existing.GuideType)
	}
	if existing.AnswerType != nil {
		meta.AnswerType = string(*existing.AnswerType)
	}
	if existing.CodeLang != nil {
		meta.CodeLang = *existing.CodeLang
	}

	// Apply input fields.
	content := strings.TrimSpace(input.Content)
	if input.Question != nil {
		meta.Question = *input.Question
	}
	if input.Description != nil {
		meta.Description = *input.Description
	}
	if input.GuideType != nil {
		meta.GuideType = string(*input.GuideType)
	}
	if input.AnswerType != nil {
		meta.AnswerType = string(*input.AnswerType)
	}
	if input.CodeLang != nil {
		meta.CodeLang = *input.CodeLang
	}
	meta.PartSize = len(content)
	meta.TotalSize = len(content)

	// Compute new embedding.
	vecs, err := ks.embedder.EmbedDocuments(ctx, []string{content})
	if err != nil {
		return nil, fmt.Errorf("knowledge: compute embedding: %w", err)
	}
	if len(vecs) == 0 {
		return nil, fmt.Errorf("knowledge: embedder returned no vectors")
	}

	cmJSON, err := metaToJSON(meta)
	if err != nil {
		return nil, fmt.Errorf("knowledge: marshal cmetadata: %w", err)
	}

	row, err := ks.db.UpdateKnowledgeDocument(ctx, database.UpdateKnowledgeDocumentParams{
		Column1: nsOf(id),              // uuid::text = $1
		Column2: formatVector(vecs[0]), // embedding = $2::vector
		Column3: nsOf(content),         // document  = $3
		Column4: cmJSON,                // cmetadata = $4::json
	})
	if err != nil {
		return nil, fmt.Errorf("knowledge: update document %s: %w", id, err)
	}

	doc := rowToModel(row.ID, row.Document, nullStr(row.Cmetadata), true)
	ks.newKnp(userID).KnowledgeDocumentUpdated(ctx, doc)
	return doc, nil
}

// ---- DeleteDocument (admin) -------------------------------------------------

func (ks *knowledgeStore) DeleteDocument(ctx context.Context, userID int64, id string) error {
	doc, err := ks.GetDocument(ctx, id)
	if err != nil {
		return err
	}
	if err := ks.db.DeleteKnowledgeDocument(ctx, nsOf(id)); err != nil {
		return fmt.Errorf("knowledge: delete document %s: %w", id, err)
	}
	ks.newKnp(userID).KnowledgeDocumentDeleted(ctx, doc)
	return nil
}

// ---- DeleteUserDocument (user-scoped) ---------------------------------------

func (ks *knowledgeStore) DeleteUserDocument(ctx context.Context, userID int64, id string) error {
	doc, err := ks.GetUserDocument(ctx, userID, id)
	if err != nil {
		return err
	}
	if err := ks.db.DeleteUserKnowledgeDocument(ctx, database.DeleteUserKnowledgeDocumentParams{
		Column1: nsOf(id),
		Column2: nsOf(strconv.FormatInt(userID, 10)),
	}); err != nil {
		return fmt.Errorf("knowledge: delete user document %s: %w", id, err)
	}
	ks.newKnp(userID).KnowledgeDocumentDeleted(ctx, doc)
	return nil
}

// ---- helpers ----------------------------------------------------------------

// metaToMap converts knowledgeMeta to the map[string]any format used by
// langchaingo's schema.Document.Metadata / pgvector cmetadata.
func metaToMap(m knowledgeMeta) map[string]any {
	mp := map[string]any{
		"doc_type":         m.DocType,
		"user_id":          m.UserID,
		"question":         m.Question,
		"part_size":        m.PartSize,
		"total_size":       m.TotalSize,
		manualDocumentFlag: m.Manual,
	}
	if m.Description != "" {
		mp["description"] = m.Description
	}
	if m.GuideType != "" {
		mp["guide_type"] = m.GuideType
	}
	if m.AnswerType != "" {
		mp["answer_type"] = m.AnswerType
	}
	if m.CodeLang != "" {
		mp["code_lang"] = m.CodeLang
	}
	if m.FlowID != nil {
		mp["flow_id"] = *m.FlowID
	}
	if m.TaskID != nil {
		mp["task_id"] = *m.TaskID
	}
	if m.SubtaskID != nil {
		mp["subtask_id"] = *m.SubtaskID
	}
	return mp
}
