package tools

import (
	"reflect"
	"testing"

	"github.com/vxcontrol/langchaingo/schema"
)

// Metadata is compared too: search_in_memory takes each fact's task and subtask headers from it.
func TestMemoryUtils_MergeAndDeduplicateDocs_KeepsTheBestCopyOfEachFactInScoreOrder(t *testing.T) {
	t.Parallel()

	doc := func(content string, score float32, metadata map[string]any) schema.Document {
		return schema.Document{PageContent: content, Score: score, Metadata: metadata}
	}

	cases := map[string]struct {
		docs    []schema.Document
		maxDocs int
		want    []schema.Document
	}{
		"no documents": {
			docs:    []schema.Document{},
			maxDocs: 10,
		},
		"distinct contents are all kept, best score first": {
			docs: []schema.Document{
				doc("content1", 0.3, map[string]any{"id": 1}),
				doc("content2", 0.9, map[string]any{"id": 2}),
				doc("content3", 0.1, map[string]any{"id": 3}),
				doc("content4", 0.7, map[string]any{"id": 4}),
				doc("content5", 0.5, map[string]any{"id": 5}),
			},
			maxDocs: 10,
			want: []schema.Document{
				doc("content2", 0.9, map[string]any{"id": 2}),
				doc("content4", 0.7, map[string]any{"id": 4}),
				doc("content5", 0.5, map[string]any{"id": 5}),
				doc("content1", 0.3, map[string]any{"id": 1}),
				doc("content3", 0.1, map[string]any{"id": 3}),
			},
		},
		"a repeated content keeps its best-scored copy and that copy's metadata": {
			docs: []schema.Document{
				doc("duplicate content", 0.5, map[string]any{"id": 1}),
				doc("unique content", 0.8, map[string]any{"id": 2}),
				doc("duplicate content", 0.9, map[string]any{"id": 3}),
				doc("another unique", 0.7, map[string]any{"id": 4}),
				doc("duplicate content", 0.3, map[string]any{"id": 5}),
			},
			maxDocs: 10,
			want: []schema.Document{
				doc("duplicate content", 0.9, map[string]any{"id": 3}),
				doc("unique content", 0.8, map[string]any{"id": 2}),
				doc("another unique", 0.7, map[string]any{"id": 4}),
			},
		},
		"more distinct contents than the limit": {
			docs: []schema.Document{
				doc("content1", 0.9, map[string]any{}),
				doc("content2", 0.8, map[string]any{}),
				doc("content3", 0.7, map[string]any{}),
				doc("content4", 0.6, map[string]any{}),
				doc("content5", 0.5, map[string]any{}),
				doc("content6", 0.4, map[string]any{}),
				doc("content7", 0.3, map[string]any{}),
			},
			maxDocs: 3,
			want: []schema.Document{
				doc("content1", 0.9, map[string]any{}),
				doc("content2", 0.8, map[string]any{}),
				doc("content3", 0.7, map[string]any{}),
			},
		},
		"a limit of zero": {
			docs: []schema.Document{
				doc("content1", 0.9, map[string]any{}),
				doc("content2", 0.8, map[string]any{}),
			},
			maxDocs: 0,
		},
		"facts of several questions overlapping": {
			docs: []schema.Document{
				doc("result A", 0.85, map[string]any{"query": 1}),
				doc("result B", 0.75, map[string]any{"query": 1}),
				doc("result C", 0.65, map[string]any{"query": 1}),
				doc("result A", 0.90, map[string]any{"query": 2}),
				doc("result D", 0.80, map[string]any{"query": 2}),
				doc("result E", 0.70, map[string]any{"query": 2}),
				doc("result B", 0.60, map[string]any{"query": 3}),
				doc("result F", 0.88, map[string]any{"query": 3}),
				doc("result C", 0.72, map[string]any{"query": 3}),
			},
			maxDocs: 5,
			want: []schema.Document{
				doc("result A", 0.90, map[string]any{"query": 2}),
				doc("result F", 0.88, map[string]any{"query": 3}),
				doc("result D", 0.80, map[string]any{"query": 2}),
				doc("result B", 0.75, map[string]any{"query": 1}),
				doc("result C", 0.72, map[string]any{"query": 3}),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := MergeAndDeduplicateDocs(tc.docs, tc.maxDocs)

			if len(got) != len(tc.want) {
				t.Fatalf("got %d documents, want %d: %v", len(got), len(tc.want), got)
			}
			for i, want := range tc.want {
				if got[i].PageContent != want.PageContent || got[i].Score != want.Score ||
					!reflect.DeepEqual(got[i].Metadata, want.Metadata) {
					t.Errorf("document %d = %q %v %v, want %q %v %v", i,
						got[i].PageContent, got[i].Score, got[i].Metadata,
						want.PageContent, want.Score, want.Metadata)
				}
			}
		})
	}
}
