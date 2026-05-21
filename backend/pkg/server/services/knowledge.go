package services

import (
	"errors"
	"net/http"
	"strings"

	knowledgepkg "pentagi/pkg/database/knowledge"
	gqlmodel "pentagi/pkg/graph/model"
	"pentagi/pkg/server/auth"
	"pentagi/pkg/server/logger"
	"pentagi/pkg/server/models"
	"pentagi/pkg/server/rdb"
	"pentagi/pkg/server/response"

	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
)

// knowledgeSQLMappers maps rdb filter field names to the corresponding SQL
// expressions against langchain_pg_embedding / its cmetadata JSON column.
// Fields ending in _id are treated as numeric by rdb (operator = instead of like),
// so numeric JSON fields use custom function mappers that cast cmetadata values.
var knowledgeSQLMappers = map[string]any{
	// Primary key exposed as "id" so TableQuery default ordering still works via SetOrders override.
	"id": "langchain_pg_embedding.uuid::text",

	// Text fields from cmetadata — rdb wraps them in LOWER(…::text) automatically.
	"doc_type":    "(langchain_pg_embedding.cmetadata ->> 'doc_type')",
	"question":    "(langchain_pg_embedding.cmetadata ->> 'question')",
	"description": "(langchain_pg_embedding.cmetadata ->> 'description')",
	"guide_type":  "(langchain_pg_embedding.cmetadata ->> 'guide_type')",
	"answer_type": "(langchain_pg_embedding.cmetadata ->> 'answer_type')",
	"code_lang":   "(langchain_pg_embedding.cmetadata ->> 'code_lang')",

	// Searchable concat of common text fields for generic "data" filter.
	"data": "((langchain_pg_embedding.cmetadata ->> 'doc_type') || ' ' || " +
		"COALESCE(langchain_pg_embedding.cmetadata ->> 'question', ''))",

	// Numeric ID fields: stored as JSON numbers; cast to bigint before comparison.
	"user_id": func(q *rdb.TableQuery, db *gorm.DB, value any) *gorm.DB {
		return applyKnowledgeIntFilter(db, "user_id", value)
	},
	"flow_id": func(q *rdb.TableQuery, db *gorm.DB, value any) *gorm.DB {
		return applyKnowledgeIntFilter(db, "flow_id", value)
	},
	"task_id": func(q *rdb.TableQuery, db *gorm.DB, value any) *gorm.DB {
		return applyKnowledgeIntFilter(db, "task_id", value)
	},
	"subtask_id": func(q *rdb.TableQuery, db *gorm.DB, value any) *gorm.DB {
		return applyKnowledgeIntFilter(db, "subtask_id", value)
	},

	// Boolean field: 'true' / anything else (JSON booleans serialise as 'true'/'false').
	"manual": func(q *rdb.TableQuery, db *gorm.DB, value any) *gorm.DB {
		if b, ok := value.(bool); ok {
			if b {
				return db.Where("(langchain_pg_embedding.cmetadata ->> 'manual') = 'true'")
			}
			return db.Where("(langchain_pg_embedding.cmetadata ->> 'manual') IS DISTINCT FROM 'true'")
		}
		return db
	},
}

// applyKnowledgeIntFilter adds a bigint equality / IN filter against a cmetadata integer field.
func applyKnowledgeIntFilter(db *gorm.DB, field string, value any) *gorm.DB {
	expr := "(langchain_pg_embedding.cmetadata ->> '" + field + "')::bigint"
	switch v := value.(type) {
	case float64:
		return db.Where(expr+" = ?", int64(v))
	case []any:
		var ids []int64
		for _, vi := range v {
			if f, ok := vi.(float64); ok {
				ids = append(ids, int64(f))
			}
		}
		if len(ids) > 0 {
			return db.Where(expr+" IN (?)", ids)
		}
	}
	return db
}

// knowledgeGrouped is the response shape for grouped list queries.
type knowledgeGrouped struct {
	Grouped []string `json:"grouped"`
	Total   uint64   `json:"total"`
}

// KnowledgeService exposes the knowledge (pgvector) store via a REST API
// that mirrors the GraphQL knowledge operations.
//
// Authorization model:
//   - knowledge.admin  → read/write any document (no user_id filtering)
//   - knowledge.view   → list/get own documents (filtered by user_id)
//   - knowledge.create → create a new document
//   - knowledge.edit   → update own document (admin: any)
//   - knowledge.delete → delete own document (admin: any)
//   - knowledge.search → semantic search (own docs for regular users)
type KnowledgeService struct {
	db    *gorm.DB
	store knowledgepkg.KnowledgeStore
}

// NewKnowledgeService creates a KnowledgeService.
// db is the gorm connection used for paginated list queries via rdb.TableQuery.
// store may be nil when the embedding provider is not configured; in that
// case embedding-dependent endpoints return 503.
func NewKnowledgeService(db *gorm.DB, store knowledgepkg.KnowledgeStore) *KnowledgeService {
	return &KnowledgeService{db: db, store: store}
}

func isKnowledgeAdmin(c *gin.Context) bool {
	return auth.LookupPerm(c.GetStringSlice("prm"), "knowledge.admin")
}

// ---- ListDocuments ----------------------------------------------------------

// ListDocuments returns a paginated, sortable, filterable list of knowledge documents.
// It uses the standard rdb.TableQuery protocol (page, pageSize, sort[], filters[])
// plus an additional `with_content` boolean query parameter.
//
// Filterable fields:
//   - id            – UUID (exact match)
//   - doc_type      – answer | guide | code  (text, like/=)
//   - question      – question text          (text, like)
//   - description   – description text       (text, like)
//   - guide_type    – guide sub-type         (text, like/=)
//   - answer_type   – answer sub-type        (text, like/=)
//   - code_lang     – code language          (text, like/=)
//   - manual        – boolean flag           (bool, =)
//   - user_id       – owner user ID          (int64, =/in)
//   - flow_id       – originating flow ID    (int64, =/in)
//   - task_id       – originating task ID    (int64, =/in)
//   - subtask_id    – originating subtask ID (int64, =/in)
//   - data          – full-text across doc_type + question (text, like)
//
// @Summary List knowledge documents
// @Tags Knowledge
// @Produce json
// @Security BearerAuth
// @Param request query rdb.TableQuery true "pagination / sort / filter params"
// @Param with_content query bool false "include document text in the response"
// @Success 200 {object} response.successResp{data=models.KnowledgeDocList}
// @Failure 400 {object} response.errorResp "invalid query parameters"
// @Failure 403 {object} response.errorResp "not permitted"
// @Failure 500 {object} response.errorResp "internal error"
// @Router /knowledge/ [get]
func (s *KnowledgeService) ListDocuments(c *gin.Context) {
	uid := int64(c.GetUint64("uid"))
	admin := isKnowledgeAdmin(c)

	var query rdb.TableQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		logger.FromContext(c).WithError(err).Error("error binding list query")
		response.Error(c, response.ErrKnowledgeInvalidRequest, err)
		return
	}

	withContent := c.Query("with_content") == "true" || c.Query("with_content") == "1"

	if err := query.Init("langchain_pg_embedding", knowledgeSQLMappers); err != nil {
		// Init fails when the requested group field is not a sortable string column
		// (e.g. a custom function mapper like user_id).
		logger.FromContext(c).WithError(err).Error("error initialising knowledge query")
		response.Error(c, response.ErrKnowledgeInvalidRequest, err)
		return
	}

	// Base scope: JOIN the collection table, restrict to 'langchain', exclude memory.
	baseScope := func(db *gorm.DB) *gorm.DB {
		return db.
			Joins("JOIN langchain_pg_collection ON langchain_pg_embedding.collection_id = langchain_pg_collection.uuid").
			Where("langchain_pg_collection.name = ?", "langchain").
			Where("COALESCE(langchain_pg_embedding.cmetadata ->> 'doc_type', '') NOT IN (?)", []string{"memory"})
	}

	// Authorization scope: admins see all documents; regular users see only their own.
	authScope := func(db *gorm.DB) *gorm.DB {
		if admin {
			return db
		}
		return db.Where("(langchain_pg_embedding.cmetadata ->> 'user_id')::bigint = ?", uid)
	}

	// ---- Grouped query -------------------------------------------------------
	// When a group field is requested the response is a list of distinct values
	// rather than full document rows (mirrors the pattern used in other services).
	if query.Group != "" {
		if _, ok := knowledgeSQLMappers[query.Group]; !ok {
			logger.FromContext(c).Errorf("group field %q not found in knowledge mappers", query.Group)
			response.Error(c, response.ErrKnowledgeInvalidRequest, errors.New("group field not found"))
			return
		}

		var resp knowledgeGrouped
		var err error
		if resp.Total, err = query.QueryGrouped(s.db, &resp.Grouped, baseScope, authScope); err != nil {
			logger.FromContext(c).WithError(err).Error("error querying knowledge documents grouped")
			response.Error(c, response.ErrInternal, err)
			return
		}

		response.Success(c, http.StatusOK, resp)
		return
	}

	// ---- Paginated / sorted query --------------------------------------------

	// Override the default "ORDER BY id DESC" — the PK column is uuid, not id.
	query.SetOrders([]func(*gorm.DB) *gorm.DB{
		func(db *gorm.DB) *gorm.DB {
			return db.Order("langchain_pg_embedding.uuid DESC")
		},
	})

	// SELECT only the columns the API exposes; never return the embedding vector.
	query.SetFind(func(out any) func(*gorm.DB) *gorm.DB {
		docCol := "'' AS document"
		if withContent {
			docCol = "COALESCE(langchain_pg_embedding.document, '') AS document"
		}
		sel := "langchain_pg_embedding.uuid::text AS id, " +
			docCol + ", " +
			"COALESCE(langchain_pg_embedding.cmetadata::text, '{}') AS cmetadata"
		return func(db *gorm.DB) *gorm.DB {
			return db.Select(sel).Find(out)
		}
	})

	var rows []models.KnowledgeEmbeddingRow
	total, err := query.Query(s.db, &rows, baseScope, authScope)
	if err != nil {
		logger.FromContext(c).WithError(err).Error("error querying knowledge documents")
		response.Error(c, response.ErrInternal, err)
		return
	}

	response.Success(c, http.StatusOK, models.KnowledgeDocListFromRows(rows, total, withContent))
}

// ---- GetDocument ------------------------------------------------------------

// GetDocument returns a single knowledge document by its UUID.
//
// @Summary Get knowledge document
// @Tags Knowledge
// @Produce json
// @Security BearerAuth
// @Param id path string true "Document UUID"
// @Success 200 {object} response.successResp{data=models.KnowledgeDocEntry}
// @Failure 403 {object} response.errorResp "not permitted or wrong owner"
// @Failure 404 {object} response.errorResp "document not found"
// @Failure 500 {object} response.errorResp "internal error"
// @Router /knowledge/{id} [get]
func (s *KnowledgeService) GetDocument(c *gin.Context) {
	uid := int64(c.GetUint64("uid"))
	admin := isKnowledgeAdmin(c)
	id := c.Param("id")

	if id == "" {
		response.Error(c, response.ErrKnowledgeInvalidRequest, errors.New("document id is required"))
		return
	}

	ctx := c.Request.Context()

	var (
		doc *gqlmodel.KnowledgeDocument
		err error
	)
	if admin {
		doc, err = s.store.GetDocument(ctx, id)
	} else {
		doc, err = s.store.GetUserDocument(ctx, uid, id)
	}
	if err != nil {
		logger.FromContext(c).WithError(err).Errorf("error getting knowledge document %s", id)
		response.Error(c, response.ErrKnowledgeNotFound, err)
		return
	}

	response.Success(c, http.StatusOK, models.KnowledgeDocFromGQL(doc))
}

// ---- SearchDocuments --------------------------------------------------------

// SearchDocuments performs a semantic similarity search over the knowledge store.
//
// @Summary Semantic search in knowledge base
// @Tags Knowledge
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param json body models.KnowledgeSearchRequest true "Search request"
// @Success 200 {object} response.successResp{data=models.KnowledgeSearchResult}
// @Failure 400 {object} response.errorResp "invalid request"
// @Failure 403 {object} response.errorResp "not permitted"
// @Failure 503 {object} response.errorResp "embedding provider not configured"
// @Failure 500 {object} response.errorResp "internal error"
// @Router /knowledge/search [post]
func (s *KnowledgeService) SearchDocuments(c *gin.Context) {
	uid := int64(c.GetUint64("uid"))
	admin := isKnowledgeAdmin(c)

	var req models.KnowledgeSearchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.FromContext(c).WithError(err).Error("error binding search request")
		response.Error(c, response.ErrKnowledgeInvalidRequest, err)
		return
	}
	if err := req.Valid(); err != nil {
		logger.FromContext(c).WithError(err).Error("invalid search request")
		response.Error(c, response.ErrKnowledgeInvalidRequest, err)
		return
	}

	ctx := c.Request.Context()
	filter := req.ToGQLFilter()

	var (
		results []*gqlmodel.KnowledgeDocumentWithScore
		err     error
	)
	if admin {
		results, err = s.store.SearchDocuments(ctx, req.Query, filter, req.Limit)
	} else {
		results, err = s.store.SearchUserDocuments(ctx, uid, req.Query, filter, req.Limit)
	}
	if err != nil {
		if isKnowledgeStoreUnavailable(err) {
			response.Error(c, response.ErrKnowledgeStoreUnavail, err)
			return
		}
		logger.FromContext(c).WithError(err).Error("error searching knowledge documents")
		response.Error(c, response.ErrInternal, err)
		return
	}

	response.Success(c, http.StatusOK, models.KnowledgeSearchResultFromGQL(results))
}

// ---- CreateDocument ---------------------------------------------------------

// CreateDocument creates a new knowledge document and computes its embedding.
//
// @Summary Create knowledge document
// @Tags Knowledge
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param json body models.CreateKnowledgeDocRequest true "Create request"
// @Success 201 {object} response.successResp{data=models.KnowledgeDocEntry}
// @Failure 400 {object} response.errorResp "invalid request"
// @Failure 403 {object} response.errorResp "not permitted"
// @Failure 503 {object} response.errorResp "embedding provider not configured"
// @Failure 500 {object} response.errorResp "internal error"
// @Router /knowledge/ [post]
func (s *KnowledgeService) CreateDocument(c *gin.Context) {
	uid := int64(c.GetUint64("uid"))

	var req models.CreateKnowledgeDocRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.FromContext(c).WithError(err).Error("error binding create request")
		response.Error(c, response.ErrKnowledgeInvalidRequest, err)
		return
	}
	if err := req.Valid(); err != nil {
		logger.FromContext(c).WithError(err).Error("invalid create request")
		response.Error(c, response.ErrKnowledgeInvalidRequest, err)
		return
	}

	doc, err := s.store.CreateDocument(c.Request.Context(), uid, req.ToGQL())
	if err != nil {
		if isKnowledgeStoreUnavailable(err) {
			response.Error(c, response.ErrKnowledgeStoreUnavail, err)
			return
		}
		logger.FromContext(c).WithError(err).Error("error creating knowledge document")
		response.Error(c, response.ErrInternal, err)
		return
	}

	response.Success(c, http.StatusCreated, models.KnowledgeDocFromGQL(doc))
}

// ---- UpdateDocument ---------------------------------------------------------

// UpdateDocument updates a knowledge document, re-computing its embedding.
// Admin can update any document; regular users can only update their own.
//
// @Summary Update knowledge document
// @Tags Knowledge
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Document UUID"
// @Param json body models.UpdateKnowledgeDocRequest true "Update request"
// @Success 200 {object} response.successResp{data=models.KnowledgeDocEntry}
// @Failure 400 {object} response.errorResp "invalid request"
// @Failure 403 {object} response.errorResp "not permitted or wrong owner"
// @Failure 404 {object} response.errorResp "document not found"
// @Failure 503 {object} response.errorResp "embedding provider not configured"
// @Failure 500 {object} response.errorResp "internal error"
// @Router /knowledge/{id} [put]
func (s *KnowledgeService) UpdateDocument(c *gin.Context) {
	uid := int64(c.GetUint64("uid"))
	admin := isKnowledgeAdmin(c)
	id := c.Param("id")

	if id == "" {
		response.Error(c, response.ErrKnowledgeInvalidRequest, errors.New("document id is required"))
		return
	}

	var req models.UpdateKnowledgeDocRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.FromContext(c).WithError(err).Error("error binding update request")
		response.Error(c, response.ErrKnowledgeInvalidRequest, err)
		return
	}
	if err := req.Valid(); err != nil {
		logger.FromContext(c).WithError(err).Error("invalid update request")
		response.Error(c, response.ErrKnowledgeInvalidRequest, err)
		return
	}

	ctx := c.Request.Context()
	input := req.ToGQL()

	var (
		doc *gqlmodel.KnowledgeDocument
		err error
	)
	if admin {
		doc, err = s.store.UpdateDocument(ctx, uid, id, input)
	} else {
		doc, err = s.store.UpdateUserDocument(ctx, uid, id, input)
	}
	if err != nil {
		if isKnowledgeStoreUnavailable(err) {
			response.Error(c, response.ErrKnowledgeStoreUnavail, err)
			return
		}
		if isKnowledgeNotFound(err) {
			response.Error(c, response.ErrKnowledgeNotFound, err)
			return
		}
		logger.FromContext(c).WithError(err).Errorf("error updating knowledge document %s", id)
		response.Error(c, response.ErrInternal, err)
		return
	}

	response.Success(c, http.StatusOK, models.KnowledgeDocFromGQL(doc))
}

// ---- DeleteDocument ---------------------------------------------------------

// DeleteDocument deletes a knowledge document by UUID.
// Admin can delete any document; regular users can only delete their own.
//
// @Summary Delete knowledge document
// @Tags Knowledge
// @Produce json
// @Security BearerAuth
// @Param id path string true "Document UUID"
// @Success 200 {object} response.successResp "document deleted"
// @Failure 403 {object} response.errorResp "not permitted or wrong owner"
// @Failure 404 {object} response.errorResp "document not found"
// @Failure 500 {object} response.errorResp "internal error"
// @Router /knowledge/{id} [delete]
func (s *KnowledgeService) DeleteDocument(c *gin.Context) {
	uid := int64(c.GetUint64("uid"))
	admin := isKnowledgeAdmin(c)
	id := c.Param("id")

	if id == "" {
		response.Error(c, response.ErrKnowledgeInvalidRequest, errors.New("document id is required"))
		return
	}

	ctx := c.Request.Context()

	var err error
	if admin {
		err = s.store.DeleteDocument(ctx, uid, id)
	} else {
		err = s.store.DeleteUserDocument(ctx, uid, id)
	}
	if err != nil {
		if isKnowledgeNotFound(err) {
			response.Error(c, response.ErrKnowledgeNotFound, err)
			return
		}
		logger.FromContext(c).WithError(err).Errorf("error deleting knowledge document %s", id)
		response.Error(c, response.ErrInternal, err)
		return
	}

	response.Success(c, http.StatusOK, gin.H{"message": "knowledge document deleted successfully"})
}

// ---- error classification ---------------------------------------------------

// isKnowledgeStoreUnavailable returns true when the error originates from a
// missing embedding provider (store or embedder not configured).
func isKnowledgeStoreUnavailable(err error) bool {
	return err != nil && strings.Contains(err.Error(), "knowledge: embedding provider")
}

// isKnowledgeNotFound returns true when the underlying DB returned no rows,
// which means the document does not exist or does not belong to the caller.
func isKnowledgeNotFound(err error) bool {
	return err != nil && (strings.Contains(err.Error(), "no rows") ||
		strings.Contains(err.Error(), "not found"))
}
