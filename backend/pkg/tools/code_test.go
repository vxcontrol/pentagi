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

// codeDoor is the code tool of flow 2, task 3 and subtask 4, over the rig's doubles.
func codeDoor(rig *knowledgeRig, replacer anonymizer.Replacer, maxEmbeddingBytes int) Tool {
	taskID, subtaskID := int64(3), int64(4)
	return NewCodeTool(1, 2, &taskID, &subtaskID, replacer, rig.store, rig.doorEmbedder, rig.db,
		maxEmbeddingBytes, rig.log, rig.known)
}

func TestCode_Handle_StoresASampleUnderABoundedQuestionAndDescription(t *testing.T) {
	const (
		question    = "how do I list open ports?"
		description = "A short summary."
	)

	for name, tc := range map[string]struct {
		question        string
		description     string
		wantQuestion    string
		wantDescription string
	}{
		"an ordinary question": {
			question: question, description: description, wantQuestion: question, wantDescription: description,
		},
		"an empty question": {
			question: "", description: description, wantQuestion: description, wantDescription: description,
		},
		"a question of spaces": {
			question: "   ", description: description, wantQuestion: description, wantDescription: description,
		},
		"a question of tabs and newlines": {
			question: "\t\n", description: description, wantQuestion: description, wantDescription: description,
		},
		"an oversize question": {
			question: strings.Repeat("я", limits.MaxQuestionLen+50), description: description,
			wantQuestion: strings.Repeat("я", limits.MaxQuestionLen), wantDescription: description,
		},
		"an oversize description": {
			question: question, description: strings.Repeat("я", limits.MaxDescriptionLen+50),
			wantQuestion: question, wantDescription: strings.Repeat("я", limits.MaxDescriptionLen),
		},
	} {
		t.Run(name, func(t *testing.T) {
			rig := newKnowledgeRig(t)

			rig.handle(t, codeDoor(rig, identityReplacer{}, 0), StoreCodeToolName, mustJSON(map[string]any{
				"code": "print('x')", "description": tc.description, "explanation": "e",
				"lang": "python", "message": "m", "question": tc.question,
			}))

			for field, want := range map[string]string{"question": tc.wantQuestion, "description": tc.wantDescription} {
				if got, _ := rig.db.meta[field].(string); got != want {
					t.Errorf("stored the %s %.60q (%d runes), want %.60q (%d runes)",
						field, got, utf8.RuneCountInString(got), want, utf8.RuneCountInString(want))
				}
			}
		})
	}
}

func TestCode_Handle_StoresASampleWhereItsSearchLooks(t *testing.T) {
	for name, tc := range map[string]struct {
		lang  any
		limit int
		want  string
	}{
		"no language":      {lang: nil, want: "text"},
		"a blank language": {lang: "  ", want: "text"},
		"a named language": {lang: "python", want: "python"},
		"a named language within the embedding limit": {lang: "python", limit: 64, want: "python"},
	} {
		t.Run(name, func(t *testing.T) {
			rig := newKnowledgeRig(t)
			door := codeDoor(rig, identityReplacer{}, tc.limit)

			store := map[string]any{"code": "print('x')", "description": "d", "explanation": "e", "message": "m", "question": "q"}
			if tc.lang != nil {
				store["lang"] = tc.lang
			}
			rig.handle(t, door, StoreCodeToolName, mustJSON(store))
			rig.handle(t, door, SearchCodeToolName, mustJSON(map[string]any{
				"questions": []string{"q"}, "message": "m", "lang": tc.want,
			}))

			filters := rig.conn.search(t, 0).filters
			if want := map[string]string{"doc_type": "code", "code_lang": tc.want}; !maps.Equal(filters, want) {
				t.Fatalf("search_code filtered on %v, want %v", filters, want)
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

func TestCode_Handle_LogsOnlyTheAnonymizedForm(t *testing.T) {
	const stored = "e\n\n```python\nconnect('<REDACTED>')\n```"

	for name, tc := range map[string]struct {
		tool       string
		args       map[string]any
		found      []schema.Document
		wantFilter string
		wantQuery  string
		wantResult string
		wantAction database.VecstoreActionType
	}{
		"storing a sample": {
			tool: StoreCodeToolName,
			args: map[string]any{
				"code": "connect('" + secretHost + "')", "description": "A short summary.",
				"explanation": "e", "lang": "python", "message": "m",
				"question": "How do I reach " + secretHost + "?",
			},
			wantFilter: `{"code_lang":"python","doc_type":"code","subtask_id":4,"task_id":3}`,
			wantQuery:  "How do I reach <REDACTED>?",
			wantResult: stored,
			wantAction: database.VecstoreActionTypeStore,
		},
		"searching for samples": {
			tool: SearchCodeToolName,
			args: map[string]any{
				"questions": []string{"How do I reach " + secretHost + "?", "Which client library connects?"},
				"message":   "m", "lang": "python",
			},
			found: []schema.Document{{
				PageContent: stored, Score: 0.875,
				Metadata: map[string]any{"question": "How do I reach <REDACTED>?", "description": "A short summary."},
			}},
			wantFilter: `{"code_lang":"python","doc_type":"code"}`,
			wantQuery:  "How do I reach <REDACTED>?\n--------------------------------\nWhich client library connects?",
			wantResult: "# Document 1 Match score: 0.875000\n\n" +
				"## Original Code Question\n\nHow do I reach <REDACTED>?\n\n" +
				"## Original Code Description\n\nA short summary.\n\n" +
				"## Content\n\n" + stored + "\n\n",
			wantAction: database.VecstoreActionTypeRetrieve,
		},
	} {
		t.Run(name, func(t *testing.T) {
			rig := newKnowledgeRig(t)
			rig.conn.results = []vectorResult{{docs: tc.found}}

			rig.handle(t, codeDoor(rig, hostReplacer{}, 0), tc.tool, mustJSON(tc.args))

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

func TestCode_Handle_ReportsWhatTheStoreFound(t *testing.T) {
	sample := func(content string, score float32) schema.Document {
		return schema.Document{PageContent: content, Score: score, Metadata: map[string]any{
			"question": "how is " + content + " found?", "description": content + " in short",
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
				{docs: []schema.Document{sample("A", 0.5), sample("B", 0.875)}},
				{docs: []schema.Document{sample("A", 0.75), sample("C", 0.625), sample("D", 0.25)}},
			},
			want: "# Document 1 Match score: 0.875000\n\n## Original Code Question\n\nhow is B found?\n\n" +
				"## Original Code Description\n\nB in short\n\n## Content\n\nB\n\n" +
				"# Document 2 Match score: 0.750000\n\n## Original Code Question\n\nhow is A found?\n\n" +
				"## Original Code Description\n\nA in short\n\n## Content\n\nA\n\n" +
				"# Document 3 Match score: 0.625000\n\n## Original Code Question\n\nhow is C found?\n\n" +
				"## Original Code Description\n\nC in short\n\n## Content\n\nC\n\n",
		},
		"a query the store fails": {
			questions: []string{"first", "second"},
			results: []vectorResult{
				{err: errors.New("connection reset by peer")},
				{docs: []schema.Document{sample("A", 0.75)}},
			},
			want: "# Document 1 Match score: 0.750000\n\n## Original Code Question\n\nhow is A found?\n\n" +
				"## Original Code Description\n\nA in short\n\n## Content\n\nA\n\n",
		},
		"nothing found": {
			questions: []string{"first"},
			results:   []vectorResult{{}},
			want:      "nothing found in code samples store and you need to store it after figure out this case",
		},
	} {
		t.Run(name, func(t *testing.T) {
			rig := newKnowledgeRig(t)
			rig.conn.results = tc.results

			got := rig.handle(t, codeDoor(rig, identityReplacer{}, 0), SearchCodeToolName, mustJSON(map[string]any{
				"questions": tc.questions, "message": "m", "lang": "python",
			}))

			if got != tc.want {
				t.Errorf("search_code answered\n%s\nwant\n%s", got, tc.want)
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

func TestCode_Handle_StoresOnlyTheAnonymizedForm(t *testing.T) {
	const (
		sample      = "e\n\n```python\nconnect('<REDACTED>')\n```"
		question    = "How do I reach <REDACTED>?"
		description = "Connects to <REDACTED>."
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

			rig.handle(t, codeDoor(rig, hostReplacer{}, tc.limit), StoreCodeToolName, mustJSON(map[string]any{
				"code": "connect('" + secretHost + "')", "description": "Connects to " + secretHost + ".",
				"explanation": "e", "lang": "python", "message": "m",
				"question": "How do I reach " + secretHost + "?",
			}))

			if viaStore := len(rig.conn.added) > 0; viaStore != tc.viaStore {
				t.Errorf("written through the vector store: %v, want %v", viaStore, tc.viaStore)
			}
			_, stored, meta := rig.stored(t)
			if stored != sample || meta["question"] != question || meta["description"] != description {
				t.Errorf("stored %q under the question %q described %q, want %q under %q described %q",
					stored, meta["question"], meta["description"], sample, question, description)
			}
			if !slices.Equal(rig.embedder.embedded, []string{sample}) {
				t.Errorf("sent %q to the embedder, want %q", rig.embedder.embedded, sample)
			}
			doc := rig.announced(t)
			if doc.Content != sample || doc.Question != question || deref(doc.Description) != description {
				t.Errorf("announced %q under the question %q described %q, want %q under %q described %q",
					doc.Content, doc.Question, deref(doc.Description), sample, question, description)
			}
		})
	}
}

func TestCode_Handle_StoresTheWholeSampleAndEmbedsWhatFits(t *testing.T) {
	const (
		fence  = "e\n\n```python\n"
		sample = fence + "print('x')\n```"
	)

	for name, tc := range map[string]struct {
		code       string
		limit      int
		noEmbedder bool
		viaStore   bool
		want       string
		embedded   string // not checked when empty
	}{
		"a sample within the embedding limit": {
			code: "print('x')", limit: 64, viaStore: true, want: sample, embedded: sample,
		},
		"a sample beyond the embedding limit": {code: "print('x')", limit: 8, want: sample, embedded: "e\n\n```py"},
		"no embedding limit":                  {code: "print('x')", limit: 0, want: sample, embedded: sample},
		"no embedder of its own": {
			code: "print('x')", limit: 8, noEmbedder: true, viaStore: true, want: sample, embedded: sample,
		},
		"a sample beyond the content limit": {
			code: strings.Repeat("я", limits.MaxContentLen+1),
			want: fence + strings.Repeat("я", limits.MaxContentLen-utf8.RuneCountInString(fence)),
		},
	} {
		t.Run(name, func(t *testing.T) {
			rig := newKnowledgeRig(t)
			if tc.noEmbedder {
				rig.doorEmbedder = nil
			}

			result := rig.handle(t, codeDoor(rig, identityReplacer{}, tc.limit), StoreCodeToolName, mustJSON(map[string]any{
				"code": tc.code, "description": "A short summary.", "explanation": "e",
				"lang": "python", "message": "m", "question": "q",
			}))

			if result != "code sample stored successfully" {
				t.Errorf("store_code answered %q", result)
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
			if doc.DocType != model.KnowledgeDocTypeCode || deref(doc.CodeLang) != "python" ||
				deref(doc.Description) != "A short summary." {
				t.Errorf("announced as a %q in %q described %q",
					doc.DocType, deref(doc.CodeLang), deref(doc.Description))
			}
			if deref(doc.FlowID) != 2 || deref(doc.TaskID) != 3 || deref(doc.SubtaskID) != 4 {
				t.Errorf("announced as from flow %d, task %d, subtask %d, want 2, 3, 4",
					deref(doc.FlowID), deref(doc.TaskID), deref(doc.SubtaskID))
			}
		})
	}
}

func TestCode_Handle_RefusesWhatItCannotServe(t *testing.T) {
	store := mustJSON(map[string]any{
		"code": "print('x')", "description": "d", "explanation": "e", "lang": "python", "message": "m", "question": "q",
	})

	for name, tc := range map[string]struct {
		tool    string
		args    json.RawMessage
		limit   int
		prepare func(rig *knowledgeRig)
		want    string
	}{
		"no vector store": {
			tool: StoreCodeToolName, args: store,
			prepare: func(rig *knowledgeRig) { rig.store = nil },
			want:    "code is not available",
		},
		"an unknown tool": {tool: SearchAnswerToolName, args: store, want: "unknown tool: search_answer"},
		"malformed search arguments": {
			tool: SearchCodeToolName, args: json.RawMessage(`{"questions":`),
			want: "failed to unmarshal search_code search code action arguments: unexpected end of JSON input",
		},
		"malformed store arguments": {
			tool: StoreCodeToolName, args: json.RawMessage(`{"code":`),
			want: "failed to unmarshal store_code store code action arguments: unexpected end of JSON input",
		},
		"a write the database refuses": {
			tool: StoreCodeToolName, args: store,
			prepare: func(rig *knowledgeRig) { rig.db.err = errors.New("connection reset by peer") },
			want: "failed to store code sample: " +
				"failed to insert document with pre-computed embedding: connection reset by peer",
		},
		"a write the vector store refuses": {
			tool: StoreCodeToolName, args: store, limit: 64,
			prepare: func(rig *knowledgeRig) { rig.conn.addErr = errors.New("connection reset by peer") },
			want:    "failed to store code sample: connection reset by peer",
		},
	} {
		t.Run(name, func(t *testing.T) {
			rig := newKnowledgeRig(t)
			if tc.prepare != nil {
				tc.prepare(rig)
			}
			door := codeDoor(rig, identityReplacer{}, tc.limit)

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
