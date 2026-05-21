package models

import (
	"encoding/json"
	"fmt"

	gqlmodel "pentagi/pkg/graph/model"
)

// knowledgeRawMeta mirrors the flat JSON stored in langchain_pg_embedding.cmetadata.
// Used only within the models package for REST → gorm conversion.
type knowledgeRawMeta struct {
	DocType     string `json:"doc_type"`
	UserID      int64  `json:"user_id"`
	FlowID      *int64 `json:"flow_id"`
	TaskID      *int64 `json:"task_id"`
	SubtaskID   *int64 `json:"subtask_id"`
	Question    string `json:"question"`
	Description string `json:"description"`
	GuideType   string `json:"guide_type"`
	AnswerType  string `json:"answer_type"`
	CodeLang    string `json:"code_lang"`
	PartSize    int    `json:"part_size"`
	TotalSize   int    `json:"total_size"`
	Manual      bool   `json:"manual"`
}

func knowledgeMetaFromJSON(raw string) knowledgeRawMeta {
	var m knowledgeRawMeta
	if raw != "" && raw != "{}" {
		_ = json.Unmarshal([]byte(raw), &m)
	}
	return m
}

func knowledgeMetaToEntry(id, content string, m knowledgeRawMeta) KnowledgeDocEntry {
	entry := KnowledgeDocEntry{
		ID:        id,
		DocType:   KnowledgeDocType(m.DocType),
		Content:   content,
		Question:  m.Question,
		UserID:    m.UserID,
		FlowID:    m.FlowID,
		TaskID:    m.TaskID,
		SubtaskID: m.SubtaskID,
		PartSize:  m.PartSize,
		TotalSize: m.TotalSize,
		Manual:    m.Manual,
	}
	if m.Description != "" {
		d := m.Description
		entry.Description = &d
	}
	if m.GuideType != "" {
		gt := KnowledgeGuideType(m.GuideType)
		entry.GuideType = &gt
	}
	if m.AnswerType != "" {
		at := KnowledgeAnswerType(m.AnswerType)
		entry.AnswerType = &at
	}
	if m.CodeLang != "" {
		cl := m.CodeLang
		entry.CodeLang = &cl
	}
	return entry
}

// ==================== Enum types =============================================

// KnowledgeDocType mirrors model.KnowledgeDocType for the REST layer.
type KnowledgeDocType string

const (
	KnowledgeDocTypeAnswer KnowledgeDocType = "answer"
	KnowledgeDocTypeGuide  KnowledgeDocType = "guide"
	KnowledgeDocTypeCode   KnowledgeDocType = "code"
)

// Valid implements IValid.
func (t KnowledgeDocType) Valid() error {
	switch t {
	case KnowledgeDocTypeAnswer, KnowledgeDocTypeGuide, KnowledgeDocTypeCode:
		return nil
	default:
		return fmt.Errorf("invalid KnowledgeDocType: %q (expected answer|guide|code)", t)
	}
}

// KnowledgeGuideType mirrors model.KnowledgeGuideType for the REST layer.
type KnowledgeGuideType string

const (
	KnowledgeGuideTypeInstall     KnowledgeGuideType = "install"
	KnowledgeGuideTypeConfigure   KnowledgeGuideType = "configure"
	KnowledgeGuideTypeUse         KnowledgeGuideType = "use"
	KnowledgeGuideTypePentest     KnowledgeGuideType = "pentest"
	KnowledgeGuideTypeDevelopment KnowledgeGuideType = "development"
	KnowledgeGuideTypeOther       KnowledgeGuideType = "other"
)

// Valid implements IValid.
func (t KnowledgeGuideType) Valid() error {
	switch t {
	case KnowledgeGuideTypeInstall,
		KnowledgeGuideTypeConfigure,
		KnowledgeGuideTypeUse,
		KnowledgeGuideTypePentest,
		KnowledgeGuideTypeDevelopment,
		KnowledgeGuideTypeOther:
		return nil
	default:
		return fmt.Errorf("invalid KnowledgeGuideType: %q", t)
	}
}

// KnowledgeAnswerType mirrors model.KnowledgeAnswerType for the REST layer.
type KnowledgeAnswerType string

const (
	KnowledgeAnswerTypeGuide         KnowledgeAnswerType = "guide"
	KnowledgeAnswerTypeVulnerability KnowledgeAnswerType = "vulnerability"
	KnowledgeAnswerTypeCode          KnowledgeAnswerType = "code"
	KnowledgeAnswerTypeTool          KnowledgeAnswerType = "tool"
	KnowledgeAnswerTypeOther         KnowledgeAnswerType = "other"
)

// Valid implements IValid.
func (t KnowledgeAnswerType) Valid() error {
	switch t {
	case KnowledgeAnswerTypeGuide,
		KnowledgeAnswerTypeVulnerability,
		KnowledgeAnswerTypeCode,
		KnowledgeAnswerTypeTool,
		KnowledgeAnswerTypeOther:
		return nil
	default:
		return fmt.Errorf("invalid KnowledgeAnswerType: %q", t)
	}
}

// ==================== Response models =========================================

// KnowledgeDocEntry is the REST response representation of a knowledge document.
//
//nolint:lll
type KnowledgeDocEntry struct {
	ID          string               `json:"id"`
	DocType     KnowledgeDocType     `json:"doc_type"`
	Content     string               `json:"content"`
	Question    string               `json:"question"`
	Description *string              `json:"description,omitempty"`
	UserID      int64                `json:"user_id"`
	FlowID      *int64               `json:"flow_id,omitempty"`
	TaskID      *int64               `json:"task_id,omitempty"`
	SubtaskID   *int64               `json:"subtask_id,omitempty"`
	GuideType   *KnowledgeGuideType  `json:"guide_type,omitempty"`
	AnswerType  *KnowledgeAnswerType `json:"answer_type,omitempty"`
	CodeLang    *string              `json:"code_lang,omitempty"`
	PartSize    int                  `json:"part_size"`
	TotalSize   int                  `json:"total_size"`
	Manual      bool                 `json:"manual"`
}

// KnowledgeDocList is the REST list response (Total matches rdb.TableQuery return type).
type KnowledgeDocList struct {
	Items []KnowledgeDocEntry `json:"items"`
	Total uint64              `json:"total"`
}

// KnowledgeEmbeddingRow is a gorm scan target for langchain_pg_embedding rows.
// Only the columns that the knowledge API exposes are selected (no embedding vector).
type KnowledgeEmbeddingRow struct {
	ID        string `gorm:"column:id"`
	Document  string `gorm:"column:document"`
	Cmetadata string `gorm:"column:cmetadata"`
}

// KnowledgeDocWithScore wraps a document with its semantic similarity score.
type KnowledgeDocWithScore struct {
	Score    float64           `json:"score"`
	Document KnowledgeDocEntry `json:"document"`
}

// KnowledgeSearchResult is the REST search response.
type KnowledgeSearchResult struct {
	Items []KnowledgeDocWithScore `json:"items"`
	Total int                     `json:"total"`
}

// ==================== Query / request models =================================

// KnowledgeListQuery holds query-string parameters for the list endpoint.
//
//nolint:lll
type KnowledgeListQuery struct {
	WithContent bool                  `form:"with_content" json:"with_content"`
	DocTypes    []KnowledgeDocType    `form:"doc_types[]" json:"doc_types,omitempty"`
	GuideTypes  []KnowledgeGuideType  `form:"guide_types[]" json:"guide_types,omitempty"`
	AnswerTypes []KnowledgeAnswerType `form:"answer_types[]" json:"answer_types,omitempty"`
	CodeLangs   []string              `form:"code_langs[]" json:"code_langs,omitempty"`
	FlowID      *int64                `form:"flow_id" json:"flow_id,omitempty"`
	Manual      *bool                 `form:"manual" json:"manual,omitempty"`
}

// Valid validates all enum values supplied in the list query.
func (q KnowledgeListQuery) Valid() error {
	for _, dt := range q.DocTypes {
		if err := dt.Valid(); err != nil {
			return err
		}
	}
	for _, gt := range q.GuideTypes {
		if err := gt.Valid(); err != nil {
			return err
		}
	}
	for _, at := range q.AnswerTypes {
		if err := at.Valid(); err != nil {
			return err
		}
	}
	return nil
}

// CreateKnowledgeDocRequest is the POST body for creating a knowledge document.
//
//nolint:lll
type CreateKnowledgeDocRequest struct {
	DocType     KnowledgeDocType     `json:"doc_type" validate:"required,valid"`
	Content     string               `json:"content" validate:"required,min=1,max=65536"`
	Question    string               `json:"question" validate:"required,min=1,max=2048"`
	Description *string              `json:"description,omitempty" validate:"omitempty,max=1000"`
	GuideType   *KnowledgeGuideType  `json:"guide_type,omitempty" validate:"omitempty,valid"`
	AnswerType  *KnowledgeAnswerType `json:"answer_type,omitempty" validate:"omitempty,valid"`
	CodeLang    *string              `json:"code_lang,omitempty" validate:"omitempty,max=100"`
}

// Valid implements IValid.
func (r CreateKnowledgeDocRequest) Valid() error {
	if err := r.DocType.Valid(); err != nil {
		return err
	}
	if r.GuideType != nil {
		if err := r.GuideType.Valid(); err != nil {
			return err
		}
	}
	if r.AnswerType != nil {
		if err := r.AnswerType.Valid(); err != nil {
			return err
		}
	}
	return validate.Struct(r)
}

// UpdateKnowledgeDocRequest is the PUT body for updating a knowledge document.
// content is mandatory (triggers re-embedding); all other fields are optional overrides.
//
//nolint:lll
type UpdateKnowledgeDocRequest struct {
	Content     string               `json:"content" validate:"required,min=1,max=65536"`
	Question    *string              `json:"question,omitempty" validate:"omitempty,min=1,max=2048"`
	Description *string              `json:"description,omitempty" validate:"omitempty,max=1000"`
	GuideType   *KnowledgeGuideType  `json:"guide_type,omitempty" validate:"omitempty,valid"`
	AnswerType  *KnowledgeAnswerType `json:"answer_type,omitempty" validate:"omitempty,valid"`
	CodeLang    *string              `json:"code_lang,omitempty" validate:"omitempty,max=100"`
}

// Valid implements IValid.
func (r UpdateKnowledgeDocRequest) Valid() error {
	if r.GuideType != nil {
		if err := r.GuideType.Valid(); err != nil {
			return err
		}
	}
	if r.AnswerType != nil {
		if err := r.AnswerType.Valid(); err != nil {
			return err
		}
	}
	return validate.Struct(r)
}

// KnowledgeSearchRequest is the POST body for semantic search.
//
//nolint:lll
type KnowledgeSearchRequest struct {
	Query       string                `json:"query" validate:"required,min=1,max=2048"`
	Limit       int                   `json:"limit,omitempty" validate:"omitempty,min=1,max=100"`
	DocTypes    []KnowledgeDocType    `json:"doc_types,omitempty" validate:"omitempty,dive,valid"`
	GuideTypes  []KnowledgeGuideType  `json:"guide_types,omitempty" validate:"omitempty,dive,valid"`
	AnswerTypes []KnowledgeAnswerType `json:"answer_types,omitempty" validate:"omitempty,dive,valid"`
	CodeLangs   []string              `json:"code_langs,omitempty" validate:"omitempty,dive,max=100"`
	FlowID      *int64                `json:"flow_id,omitempty"`
	Manual      *bool                 `json:"manual,omitempty"`
}

// Valid implements IValid.
func (r KnowledgeSearchRequest) Valid() error {
	for _, dt := range r.DocTypes {
		if err := dt.Valid(); err != nil {
			return err
		}
	}
	for _, gt := range r.GuideTypes {
		if err := gt.Valid(); err != nil {
			return err
		}
	}
	for _, at := range r.AnswerTypes {
		if err := at.Valid(); err != nil {
			return err
		}
	}
	return validate.Struct(r)
}

// ==================== Converters (REST ↔ GraphQL model) ======================

// KnowledgeDocFromGQL converts a GraphQL KnowledgeDocument to a REST KnowledgeDocEntry.
func KnowledgeDocFromGQL(doc *gqlmodel.KnowledgeDocument) KnowledgeDocEntry {
	entry := KnowledgeDocEntry{
		ID:        doc.ID,
		DocType:   KnowledgeDocType(doc.DocType),
		Content:   doc.Content,
		Question:  doc.Question,
		UserID:    doc.UserID,
		FlowID:    doc.FlowID,
		TaskID:    doc.TaskID,
		SubtaskID: doc.SubtaskID,
		PartSize:  doc.PartSize,
		TotalSize: doc.TotalSize,
		Manual:    doc.Manual,
	}
	if doc.Description != nil {
		entry.Description = doc.Description
	}
	if doc.GuideType != nil {
		gt := KnowledgeGuideType(*doc.GuideType)
		entry.GuideType = &gt
	}
	if doc.AnswerType != nil {
		at := KnowledgeAnswerType(*doc.AnswerType)
		entry.AnswerType = &at
	}
	if doc.CodeLang != nil {
		entry.CodeLang = doc.CodeLang
	}
	return entry
}

// KnowledgeDocListFromGQL converts a slice of GraphQL documents to a REST list response.
func KnowledgeDocListFromGQL(docs []*gqlmodel.KnowledgeDocument) KnowledgeDocList {
	items := make([]KnowledgeDocEntry, 0, len(docs))
	for _, d := range docs {
		items = append(items, KnowledgeDocFromGQL(d))
	}
	return KnowledgeDocList{Items: items, Total: uint64(len(items))}
}

// KnowledgeDocListFromRows converts raw gorm scan rows to a REST list response.
func KnowledgeDocListFromRows(rows []KnowledgeEmbeddingRow, total uint64, withContent bool) KnowledgeDocList {
	items := make([]KnowledgeDocEntry, 0, len(rows))
	for _, r := range rows {
		items = append(items, KnowledgeDocEntryFromRow(r, withContent))
	}
	return KnowledgeDocList{Items: items, Total: total}
}

// KnowledgeDocEntryFromRow converts a single gorm scan row to a REST KnowledgeDocEntry.
func KnowledgeDocEntryFromRow(r KnowledgeEmbeddingRow, withContent bool) KnowledgeDocEntry {
	meta := knowledgeMetaFromJSON(r.Cmetadata)
	content := r.Document
	if !withContent {
		content = ""
	}
	return knowledgeMetaToEntry(r.ID, content, meta)
}

// KnowledgeSearchResultFromGQL converts GraphQL search results to the REST response.
func KnowledgeSearchResultFromGQL(results []*gqlmodel.KnowledgeDocumentWithScore) KnowledgeSearchResult {
	items := make([]KnowledgeDocWithScore, 0, len(results))
	for _, r := range results {
		items = append(items, KnowledgeDocWithScore{
			Score:    r.Score,
			Document: KnowledgeDocFromGQL(r.Document),
		})
	}
	return KnowledgeSearchResult{Items: items, Total: len(items)}
}

// CreateRequestToGQL converts a REST create request to the GraphQL input type.
func (r CreateKnowledgeDocRequest) ToGQL() gqlmodel.CreateKnowledgeDocumentInput {
	input := gqlmodel.CreateKnowledgeDocumentInput{
		DocType:     gqlmodel.KnowledgeDocType(r.DocType),
		Content:     r.Content,
		Question:    r.Question,
		Description: r.Description,
		CodeLang:    r.CodeLang,
	}
	if r.GuideType != nil {
		gt := gqlmodel.KnowledgeGuideType(*r.GuideType)
		input.GuideType = &gt
	}
	if r.AnswerType != nil {
		at := gqlmodel.KnowledgeAnswerType(*r.AnswerType)
		input.AnswerType = &at
	}
	return input
}

// UpdateRequestToGQL converts a REST update request to the GraphQL input type.
func (r UpdateKnowledgeDocRequest) ToGQL() gqlmodel.UpdateKnowledgeDocumentInput {
	input := gqlmodel.UpdateKnowledgeDocumentInput{
		Content:     r.Content,
		Question:    r.Question,
		Description: r.Description,
		CodeLang:    r.CodeLang,
	}
	if r.GuideType != nil {
		gt := gqlmodel.KnowledgeGuideType(*r.GuideType)
		input.GuideType = &gt
	}
	if r.AnswerType != nil {
		at := gqlmodel.KnowledgeAnswerType(*r.AnswerType)
		input.AnswerType = &at
	}
	return input
}

// ListQueryToGQLFilter converts a REST list query to a GraphQL KnowledgeFilter.
func (q KnowledgeListQuery) ToGQLFilter() *gqlmodel.KnowledgeFilter {
	f := &gqlmodel.KnowledgeFilter{
		FlowID: q.FlowID,
		Manual: q.Manual,
	}
	for _, dt := range q.DocTypes {
		f.DocTypes = append(f.DocTypes, gqlmodel.KnowledgeDocType(dt))
	}
	for _, gt := range q.GuideTypes {
		f.GuideTypes = append(f.GuideTypes, gqlmodel.KnowledgeGuideType(gt))
	}
	for _, at := range q.AnswerTypes {
		f.AnswerTypes = append(f.AnswerTypes, gqlmodel.KnowledgeAnswerType(at))
	}
	f.CodeLangs = append(f.CodeLangs, q.CodeLangs...)
	if f.FlowID == nil && f.Manual == nil && len(f.DocTypes) == 0 &&
		len(f.GuideTypes) == 0 && len(f.AnswerTypes) == 0 && len(f.CodeLangs) == 0 {
		return nil
	}
	return f
}

// SearchRequestToGQLFilter converts a REST search request to a GraphQL KnowledgeFilter.
func (r KnowledgeSearchRequest) ToGQLFilter() *gqlmodel.KnowledgeFilter {
	f := &gqlmodel.KnowledgeFilter{
		FlowID: r.FlowID,
		Manual: r.Manual,
	}
	for _, dt := range r.DocTypes {
		f.DocTypes = append(f.DocTypes, gqlmodel.KnowledgeDocType(dt))
	}
	for _, gt := range r.GuideTypes {
		f.GuideTypes = append(f.GuideTypes, gqlmodel.KnowledgeGuideType(gt))
	}
	for _, at := range r.AnswerTypes {
		f.AnswerTypes = append(f.AnswerTypes, gqlmodel.KnowledgeAnswerType(at))
	}
	f.CodeLangs = append(f.CodeLangs, r.CodeLangs...)
	if f.FlowID == nil && f.Manual == nil && len(f.DocTypes) == 0 &&
		len(f.GuideTypes) == 0 && len(f.AnswerTypes) == 0 && len(f.CodeLangs) == 0 {
		return nil
	}
	return f
}
