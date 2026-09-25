package tools

import (
	"encoding/json"
	"errors"
	"fmt"
	"maps"
	"slices"
	"strings"
	"testing"
	"unicode/utf8"

	"pentagi/pkg/database"
	"pentagi/pkg/database/knowledge/limits"
	"pentagi/pkg/graph/model"

	"github.com/vxcontrol/cloud/anonymizer"
	"github.com/vxcontrol/langchaingo/schema"
)

// guideDoor is the guide tool of flow 2, task 3 and subtask 4, over the rig's doubles.
func guideDoor(rig *knowledgeRig, replacer anonymizer.Replacer, maxEmbeddingBytes int) Tool {
	taskID, subtaskID := int64(3), int64(4)
	return NewGuideTool(1, 2, &taskID, &subtaskID, replacer, rig.store, rig.doorEmbedder, rig.db,
		maxEmbeddingBytes, rig.log, rig.known)
}

func TestGuide_Handle_StoresAGuideUnderABoundedNonBlankQuestion(t *testing.T) {
	for name, tc := range map[string]struct {
		question string
		want     string
	}{
		"an ordinary question": {
			question: "how do I check the kernel version?", want: "how do I check the kernel version?",
		},
		"an empty question":               {question: "", want: "The guide body."},
		"a question of spaces":            {question: "   ", want: "The guide body."},
		"a question of tabs and newlines": {question: "\t\n", want: "The guide body."},
		"an oversize question": {
			question: strings.Repeat("я", limits.MaxQuestionLen+50), want: strings.Repeat("я", limits.MaxQuestionLen),
		},
	} {
		t.Run(name, func(t *testing.T) {
			rig := newKnowledgeRig(t)

			rig.handle(t, guideDoor(rig, identityReplacer{}, 0), StoreGuideToolName, mustJSON(map[string]any{
				"guide": "The guide body.", "message": "m", "question": tc.question, "type": "use",
			}))

			if got, _ := rig.db.meta["question"].(string); got != tc.want {
				t.Fatalf("stored under the question %.60q (%d runes), want %.60q (%d runes)",
					got, utf8.RuneCountInString(got), tc.want, utf8.RuneCountInString(tc.want))
			}
		})
	}
}

func TestGuide_Handle_StoresAGuideWhereItsSearchLooks(t *testing.T) {
	for name, tc := range map[string]struct {
		store map[string]any
		limit int
		want  string
	}{
		"no type":      {store: map[string]any{"guide": "g", "message": "m", "question": "q"}, want: "other"},
		"a blank type": {store: map[string]any{"guide": "g", "message": "m", "question": "q", "type": "\t"}, want: "other"},
		"a named type": {store: map[string]any{"guide": "g", "message": "m", "question": "q", "type": "use"}, want: "use"},
		"a named type within the embedding limit": {
			store: map[string]any{"guide": "g", "message": "m", "question": "q", "type": "use"}, limit: 64, want: "use",
		},
	} {
		t.Run(name, func(t *testing.T) {
			rig := newKnowledgeRig(t)
			door := guideDoor(rig, identityReplacer{}, tc.limit)

			rig.handle(t, door, StoreGuideToolName, mustJSON(tc.store))
			rig.handle(t, door, SearchGuideToolName, mustJSON(map[string]any{
				"questions": []string{"q"}, "message": "m", "type": tc.want,
			}))

			filters := rig.conn.search(t, 0).filters
			if want := map[string]string{"doc_type": "guide", "guide_type": tc.want}; !maps.Equal(filters, want) {
				t.Fatalf("search_guide filtered on %v, want %v", filters, want)
			}
			_, _, meta := rig.stored(t)
			for key, value := range filters {
				if stored := fmt.Sprint(meta[key]); stored != value {
					t.Errorf("stored with %s %q, but its search looks for %q and never finds it", key, stored, value)
				}
			}
		})
	}
}

func TestGuide_Handle_LogsOnlyTheAnonymizedForm(t *testing.T) {
	const stored = "Question:\nHow do I configure <REDACTED>?\n\nGuide:\nPoint the agent at <REDACTED>."

	for name, tc := range map[string]struct {
		tool       string
		args       map[string]any
		found      []schema.Document
		wantFilter string
		wantQuery  string
		wantResult string
		wantAction database.VecstoreActionType
	}{
		"storing a guide": {
			tool: StoreGuideToolName,
			args: map[string]any{
				"guide": "Point the agent at " + secretHost + ".", "message": "m",
				"question": "How do I configure " + secretHost + "?", "type": "use",
			},
			wantFilter: `{"doc_type":"guide","guide_type":"use","subtask_id":4,"task_id":3}`,
			wantQuery:  "How do I configure <REDACTED>?",
			wantResult: stored,
			wantAction: database.VecstoreActionTypeStore,
		},
		"searching for guides": {
			tool: SearchGuideToolName,
			args: map[string]any{
				"questions": []string{"How do I configure " + secretHost + "?", "Which settings matter?"},
				"message":   "m", "type": "use",
			},
			found: []schema.Document{{
				PageContent: stored, Score: 0.875,
				Metadata: map[string]any{"guide_type": "use", "question": "How do I configure <REDACTED>?"},
			}},
			wantFilter: `{"doc_type":"guide","guide_type":"use"}`,
			wantQuery:  "How do I configure <REDACTED>?\n--------------------------------\nWhich settings matter?",
			wantResult: "# Document 1 Match score: 0.875000\n\n## Original Guide Type: use\n\n" +
				"## Original Guide Question\n\nHow do I configure <REDACTED>?\n\n" +
				"## Content\n\n" + stored + "\n\n",
			wantAction: database.VecstoreActionTypeRetrieve,
		},
	} {
		t.Run(name, func(t *testing.T) {
			rig := newKnowledgeRig(t)
			rig.conn.results = []vectorResult{{docs: tc.found}}

			rig.handle(t, guideDoor(rig, hostReplacer{}, 0), tc.tool, mustJSON(tc.args))

			if len(rig.log.entries) != 1 {
				t.Fatalf("%d entries reached the vector store log, want 1", len(rig.log.entries))
			}
			entry := rig.log.entries[0]
			if entry.filter != tc.wantFilter {
				t.Errorf("logged the filter %s, want %s", entry.filter, tc.wantFilter)
			}
			if entry.query != tc.wantQuery {
				t.Errorf("the vector store tab renders the query verbatim, and it was logged as %q, want %q",
					entry.query, tc.wantQuery)
			}
			if entry.result != tc.wantResult {
				t.Errorf("logged the result %q, want %q", entry.result, tc.wantResult)
			}
			if entry.action != tc.wantAction {
				t.Errorf("logged as a %q, want %q", entry.action, tc.wantAction)
			}
		})
	}
}

func TestGuide_Handle_ReportsWhatTheStoreFound(t *testing.T) {
	guide := func(content string, score float32) schema.Document {
		return schema.Document{PageContent: content, Score: score, Metadata: map[string]any{
			"guide_type": "use", "question": "how is " + content + " found?",
		}}
	}

	for name, tc := range map[string]struct {
		questions []string
		results   []vectorResult
		want      string
	}{
		"documents found by two queries": {
			questions: []string{"first", "second"},
			results: []vectorResult{
				{docs: []schema.Document{guide("A", 0.5), guide("B", 0.875)}},
				{docs: []schema.Document{guide("A", 0.75), guide("C", 0.625), guide("D", 0.25)}},
			},
			want: "# Document 1 Match score: 0.875000\n\n## Original Guide Type: use\n\n" +
				"## Original Guide Question\n\nhow is B found?\n\n## Content\n\nB\n\n" +
				"# Document 2 Match score: 0.750000\n\n## Original Guide Type: use\n\n" +
				"## Original Guide Question\n\nhow is A found?\n\n## Content\n\nA\n\n" +
				"# Document 3 Match score: 0.625000\n\n## Original Guide Type: use\n\n" +
				"## Original Guide Question\n\nhow is C found?\n\n## Content\n\nC\n\n",
		},
		"a query the store fails": {
			questions: []string{"first", "second"},
			results: []vectorResult{
				{err: errors.New("connection reset by peer")},
				{docs: []schema.Document{guide("A", 0.75)}},
			},
			want: "# Document 1 Match score: 0.750000\n\n## Original Guide Type: use\n\n" +
				"## Original Guide Question\n\nhow is A found?\n\n## Content\n\nA\n\n",
		},
		"nothing found": {
			questions: []string{"first"},
			results:   []vectorResult{{}},
			want:      "nothing found in guide store and you need to store it after figure out this case",
		},
	} {
		t.Run(name, func(t *testing.T) {
			rig := newKnowledgeRig(t)
			rig.conn.results = tc.results

			got := rig.handle(t, guideDoor(rig, identityReplacer{}, 0), SearchGuideToolName, mustJSON(map[string]any{
				"questions": tc.questions, "message": "m", "type": "use",
			}))

			if got != tc.want {
				t.Errorf("search_guide answered\n%s\nwant\n%s", got, tc.want)
			}
			if len(rig.conn.queries) != len(tc.questions) {
				t.Errorf("%d similarity queries for %d questions", len(rig.conn.queries), len(tc.questions))
			}
			for i := range rig.conn.queries {
				if search := rig.conn.search(t, i); search.limit != 3 || search.minScore != 0.2 {
					t.Errorf("similarity query %d asked for %d documents scoring at least %v, want 3 scoring at least 0.2",
						i+1, search.limit, search.minScore)
				}
			}
		})
	}
}

// The database path embeds from its own anonymized copy of the guide, so both paths are checked.
func TestGuide_Handle_StoresOnlyTheAnonymizedForm(t *testing.T) {
	const (
		guide    = "Question:\nHow do I configure <REDACTED>?\n\nGuide:\nPoint the agent at <REDACTED>."
		question = "How do I configure <REDACTED>?"
	)

	for name, tc := range map[string]struct {
		limit    int
		viaStore bool
	}{
		"written through the vector store": {limit: 4096, viaStore: true},
		"written straight to the database": {limit: 0},
	} {
		t.Run(name, func(t *testing.T) {
			rig := newKnowledgeRig(t)

			rig.handle(t, guideDoor(rig, hostReplacer{}, tc.limit), StoreGuideToolName, mustJSON(map[string]any{
				"guide": "Point the agent at " + secretHost + ".", "message": "m",
				"question": "How do I configure " + secretHost + "?", "type": "use",
			}))

			if viaStore := len(rig.conn.added) > 0; viaStore != tc.viaStore {
				t.Errorf("written through the vector store: %v, want %v", viaStore, tc.viaStore)
			}
			_, stored, meta := rig.stored(t)
			if stored != guide || meta["question"] != question {
				t.Errorf("stored %q under the question %q, want %q under %q", stored, meta["question"], guide, question)
			}
			if !slices.Equal(rig.embedder.embedded, []string{guide}) {
				t.Errorf("sent %q to the embedder, want %q", rig.embedder.embedded, guide)
			}
			doc := rig.announced(t)
			if doc.Content != guide || doc.Question != question {
				t.Errorf("announced %q under the question %q, want %q under %q", doc.Content, doc.Question, guide, question)
			}
		})
	}
}

func TestGuide_Handle_StoresTheWholeGuideAndEmbedsWhatFits(t *testing.T) {
	const (
		prefix = "Question:\nq\n\nGuide:\n"
		guide  = "The guide body."
	)

	for name, tc := range map[string]struct {
		guide      string
		limit      int
		noEmbedder bool
		viaStore   bool
		want       string
		embedded   string // not checked when empty
	}{
		"a guide within the embedding limit": {
			guide: guide, limit: 64, viaStore: true, want: prefix + guide, embedded: prefix + guide,
		},
		"a guide beyond the embedding limit": {
			guide: guide, limit: 24, want: prefix + guide, embedded: prefix + "The ",
		},
		"no embedding limit": {guide: guide, limit: 0, want: prefix + guide, embedded: prefix + guide},
		"no embedder of its own": {
			guide: guide, limit: 24, noEmbedder: true, viaStore: true, want: prefix + guide, embedded: prefix + guide,
		},
		"a guide beyond the content limit": {
			guide: strings.Repeat("я", limits.MaxContentLen+1),
			want:  prefix + strings.Repeat("я", limits.MaxContentLen-utf8.RuneCountInString(prefix)),
		},
	} {
		t.Run(name, func(t *testing.T) {
			rig := newKnowledgeRig(t)
			if tc.noEmbedder {
				rig.doorEmbedder = nil
			}

			result := rig.handle(t, guideDoor(rig, identityReplacer{}, tc.limit), StoreGuideToolName, mustJSON(map[string]any{
				"guide": tc.guide, "message": "m", "question": "q", "type": "use",
			}))

			if result != "guide stored successfully" {
				t.Errorf("store_guide answered %q", result)
			}
			if viaStore := len(rig.conn.added) > 0; viaStore != tc.viaStore {
				t.Errorf("written through the vector store: %v, want %v", viaStore, tc.viaStore)
			}
			id, stored, _ := rig.stored(t)
			if stored != tc.want {
				t.Errorf("stored %.60q (%d runes), want %.60q (%d runes)",
					stored, utf8.RuneCountInString(stored), tc.want, utf8.RuneCountInString(tc.want))
			}
			if tc.embedded != "" && !slices.Equal(rig.embedder.embedded, []string{tc.embedded}) {
				t.Errorf("embedded %q, want %q", rig.embedder.embedded, tc.embedded)
			}

			doc := rig.announced(t)
			if doc.ID != id || doc.Content != stored {
				t.Errorf("announced %q with %d runes, not the stored %q", doc.ID, utf8.RuneCountInString(doc.Content), id)
			}
			if doc.DocType != model.KnowledgeDocTypeGuide || deref(doc.GuideType) != "use" {
				t.Errorf("announced as a %q of type %q", doc.DocType, deref(doc.GuideType))
			}
			if deref(doc.FlowID) != 2 || deref(doc.TaskID) != 3 || deref(doc.SubtaskID) != 4 {
				t.Errorf("announced as from flow %d, task %d, subtask %d, want 2, 3, 4",
					deref(doc.FlowID), deref(doc.TaskID), deref(doc.SubtaskID))
			}
		})
	}
}

func TestGuide_Handle_RefusesWhatItCannotServe(t *testing.T) {
	store := mustJSON(map[string]any{"guide": "The guide body.", "message": "m", "question": "q", "type": "use"})

	for name, tc := range map[string]struct {
		tool    string
		args    json.RawMessage
		limit   int
		prepare func(rig *knowledgeRig)
		want    string
	}{
		"no vector store": {
			tool: StoreGuideToolName, args: store,
			prepare: func(rig *knowledgeRig) { rig.store = nil },
			want:    "guide is not available",
		},
		"an unknown tool": {tool: SearchAnswerToolName, args: store, want: "unknown tool: search_answer"},
		"malformed search arguments": {
			tool: SearchGuideToolName, args: json.RawMessage(`{"questions":`),
			want: "failed to unmarshal search_guide search guide action arguments: unexpected end of JSON input",
		},
		"malformed store arguments": {
			tool: StoreGuideToolName, args: json.RawMessage(`{"guide":`),
			want: "failed to unmarshal store_guide store guide action arguments: unexpected end of JSON input",
		},
		"a write the database refuses": {
			tool: StoreGuideToolName, args: store,
			prepare: func(rig *knowledgeRig) { rig.db.err = errors.New("connection reset by peer") },
			want:    "failed to store guide: failed to insert document with pre-computed embedding: connection reset by peer",
		},
		"a write the vector store refuses": {
			tool: StoreGuideToolName, args: store, limit: 64,
			prepare: func(rig *knowledgeRig) { rig.conn.addErr = errors.New("connection reset by peer") },
			want:    "failed to store guide: connection reset by peer",
		},
	} {
		t.Run(name, func(t *testing.T) {
			rig := newKnowledgeRig(t)
			if tc.prepare != nil {
				tc.prepare(rig)
			}
			door := guideDoor(rig, identityReplacer{}, tc.limit)

			result, err := door.Handle(PutAgentContext(t.Context(), database.MsgchainTypeSearcher), tc.tool, tc.args)

			if err == nil || err.Error() != tc.want {
				t.Fatalf("answered %q with the error %v, want the error %q", result, err, tc.want)
			}
			if door.IsAvailable() != (rig.store != nil) {
				t.Errorf("IsAvailable is %v with a store %v", door.IsAvailable(), rig.store)
			}
			if len(rig.known.created) != 0 || len(rig.log.entries) != 0 {
				t.Errorf("a refused call still announced %d documents and logged %d entries",
					len(rig.known.created), len(rig.log.entries))
			}
		})
	}
}
