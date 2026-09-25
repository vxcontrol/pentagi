package models

import (
	"fmt"
	"reflect"
	"strings"
	"testing"

	"pentagi/pkg/database/knowledge/limits"

	"github.com/stretchr/testify/assert"
)

// Subtests are keyed by field and class of text: every bounded text field meets every class.
func TestKnowledge_Valid_RefusesBlankOrOverlongText(t *testing.T) {
	t.Parallel()

	fields := []struct {
		name      string
		namespace string
		emptyTag  string
		limit     int
		build     func(text string) IValid
	}{
		{"create question", "CreateKnowledgeDocRequest.Question", "required", 2048, func(text string) IValid {
			return CreateKnowledgeDocRequest{DocType: KnowledgeDocTypeAnswer, Content: "content", Question: text}
		}},
		{"update question", "UpdateKnowledgeDocRequest.Question", "notblank", 2048, func(text string) IValid {
			return UpdateKnowledgeDocRequest{Content: "content", Question: &text}
		}},
		{"create content", "CreateKnowledgeDocRequest.Content", "required", 65536, func(text string) IValid {
			return CreateKnowledgeDocRequest{DocType: KnowledgeDocTypeAnswer, Content: text, Question: "question"}
		}},
		{"update content", "UpdateKnowledgeDocRequest.Content", "required", 65536, func(text string) IValid {
			return UpdateKnowledgeDocRequest{Content: text}
		}},
		{"search query", "KnowledgeSearchRequest.Query", "required", 2048, func(text string) IValid {
			return KnowledgeSearchRequest{Query: text}
		}},
	}

	texts := []struct {
		name string
		text func(limit int) string
		tag  string
	}{
		{"ordinary text", func(int) string { return "open ports" }, ""},
		{"empty", func(int) string { return "" }, "required"},
		{"spaces only", func(int) string { return "   " }, "notblank"},
		{"tab and newline only", func(int) string { return "\t\n " }, "notblank"},
		{"at the length limit", func(limit int) string { return strings.Repeat("a", limit) }, ""},
		{"one over the length limit", func(limit int) string { return strings.Repeat("a", limit+1) }, "max"},
		{"multibyte at the length limit", func(limit int) string { return strings.Repeat("я", limit) }, ""},
		{"multibyte one over the length limit", func(limit int) string { return strings.Repeat("я", limit+1) }, "max"},
	}

	for _, field := range fields {
		for _, text := range texts {
			t.Run(field.name+" "+text.name, func(t *testing.T) {
				t.Parallel()

				want := text.tag
				if want == "required" {
					want = field.emptyTag
				}
				if want != "" {
					want = field.namespace + ":" + want
				}

				assert.Equal(t, want, modelsRefusal(field.build(text.text(field.limit)).Valid()))
			})
		}
	}

	t.Run("update without a question", func(t *testing.T) {
		t.Parallel()

		assert.NoError(t, UpdateKnowledgeDocRequest{Content: "content"}.Valid())
	})
}

func TestKnowledge_Valid_BoundsTheSearchLimit(t *testing.T) {
	t.Parallel()

	modelsCheckValid(t, []modelsValidCase{
		{"an omitted limit leaves it to the store", KnowledgeSearchRequest{Query: "open ports"}, ""},
		{"a limit at the max", KnowledgeSearchRequest{Query: "open ports", Limit: 100}, ""},
		{"a limit one over the max", KnowledgeSearchRequest{Query: "open ports", Limit: 101}, "KnowledgeSearchRequest.Limit:max"},
		{"a limit below zero", KnowledgeSearchRequest{Query: "open ports", Limit: -1}, "KnowledgeSearchRequest.Limit:min"},
	})
}

func TestKnowledge_KnowledgeSearchRequest_DeclaresTheSharedLimits(t *testing.T) {
	t.Parallel()

	requestType := reflect.TypeFor[KnowledgeSearchRequest]()
	for field, want := range map[string]string{
		"Query": fmt.Sprintf("max=%d", limits.MaxQuestionLen),
		"Limit": fmt.Sprintf("max=%d", limits.MaxSearchLimit),
	} {
		declared, ok := requestType.FieldByName(field)
		if assert.True(t, ok, "KnowledgeSearchRequest has no %s field", field) {
			assert.Contains(t, strings.Split(declared.Tag.Get("validate"), ","), want, field)
		}
	}
}

func TestKnowledge_KnowledgeMetaToEntry_HandsClientsAValueTheEnumDefines(t *testing.T) {
	t.Parallel()

	for stored, want := range map[string]KnowledgeGuideType{
		"Install": "install",
		"sorcery": "other",
		"  use ":  "use",
	} {
		entry := knowledgeMetaToEntry("doc-1", "content", knowledgeRawMeta{DocType: "guide", GuideType: stored})

		if assert.NotNil(t, entry.GuideType, "stored %q dropped the guide type entirely", stored) {
			assert.Equal(t, want, *entry.GuideType, "stored %q", stored)
		}
	}
}
