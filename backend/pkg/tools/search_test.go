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

// searchDoor is the search tool of flow 2, task 3 and subtask 4, over the rig's doubles.
func searchDoor(rig *knowledgeRig, replacer anonymizer.Replacer, maxEmbeddingBytes int) Tool {
	taskID, subtaskID := int64(3), int64(4)
	return NewSearchTool(1, 2, &taskID, &subtaskID, replacer, rig.store, rig.doorEmbedder, rig.db,
		maxEmbeddingBytes, rig.log, rig.known)
}

func TestSearch_Handle_StoresAnAnswerUnderABoundedNonBlankQuestion(t *testing.T) {
	for name, tc := range map[string]struct {
		question string
		want     string
	}{
		"an ordinary question": {
			question: "how do I check the kernel version?", want: "how do I check the kernel version?",
		},
		"an empty question":               {question: "", want: "The answer body."},
		"a question of spaces":            {question: "   ", want: "The answer body."},
		"a question of tabs and newlines": {question: "\t\n", want: "The answer body."},
		"an oversize question": {
			question: strings.Repeat("я", limits.MaxQuestionLen+50), want: strings.Repeat("я", limits.MaxQuestionLen),
		},
	} {
		t.Run(name, func(t *testing.T) {
			rig := newKnowledgeRig(t)

			rig.handle(t, searchDoor(rig, identityReplacer{}, 0), StoreAnswerToolName, mustJSON(map[string]any{
				"answer": "The answer body.", "message": "m", "question": tc.question, "type": "tool",
			}))

			if got, _ := rig.db.meta["question"].(string); got != tc.want {
				t.Fatalf("stored under the question %.60q (%d runes), want %.60q (%d runes)",
					got, utf8.RuneCountInString(got), tc.want, utf8.RuneCountInString(tc.want))
			}
		})
	}
}

func TestSearch_Handle_StoresAnAnswerWhereItsSearchLooks(t *testing.T) {
	for name, tc := range map[string]struct {
		store map[string]any
		limit int
		want  string
	}{
		"no type":      {store: map[string]any{"answer": "a", "message": "m", "question": "q"}, want: "other"},
		"a blank type": {store: map[string]any{"answer": "a", "message": "m", "question": "q", "type": "  "}, want: "other"},
		"a named type": {store: map[string]any{"answer": "a", "message": "m", "question": "q", "type": "tool"}, want: "tool"},
		"a named type within the embedding limit": {
			store: map[string]any{"answer": "a", "message": "m", "question": "q", "type": "tool"}, limit: 64, want: "tool",
		},
	} {
		t.Run(name, func(t *testing.T) {
			rig := newKnowledgeRig(t)
			door := searchDoor(rig, identityReplacer{}, tc.limit)

			rig.handle(t, door, StoreAnswerToolName, mustJSON(tc.store))
			rig.handle(t, door, SearchAnswerToolName, mustJSON(map[string]any{
				"questions": []string{"q"}, "message": "m", "type": tc.want,
			}))

			filters := rig.conn.search(t, 0).filters
			if want := map[string]string{"doc_type": "answer", "answer_type": tc.want}; !maps.Equal(filters, want) {
				t.Fatalf("search_answer filtered on %v, want %v", filters, want)
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

func TestSearch_Handle_LogsOnlyTheAnonymizedForm(t *testing.T) {
	for name, tc := range map[string]struct {
		tool       string
		args       map[string]any
		found      []schema.Document
		wantFilter string
		wantQuery  string
		wantResult string
		wantAction database.VecstoreActionType
	}{
		"storing an answer": {
			tool: StoreAnswerToolName,
			args: map[string]any{
				"answer": "The service answered from " + secretHost + ".", "message": "m",
				"question": "What answered on " + secretHost + "?", "type": "tool",
			},
			wantFilter: `{"answer_type":"tool","doc_type":"answer","subtask_id":4,"task_id":3}`,
			wantQuery:  "What answered on <REDACTED>?",
			wantResult: "The service answered from <REDACTED>.",
			wantAction: database.VecstoreActionTypeStore,
		},
		"searching for answers": {
			tool: SearchAnswerToolName,
			args: map[string]any{
				"questions": []string{"What answered on " + secretHost + "?", "Which service runs there?"},
				"message":   "m", "type": "tool",
			},
			found: []schema.Document{{
				PageContent: "The service answered from <REDACTED>.", Score: 0.875,
				Metadata: map[string]any{"answer_type": "tool", "question": "What answered on <REDACTED>?"},
			}},
			wantFilter: `{"answer_type":"tool","doc_type":"answer"}`,
			wantQuery:  "What answered on <REDACTED>?\n--------------------------------\nWhich service runs there?",
			wantResult: "# Document 1 Search Score: 0.875000\n\n## Original Answer Type: tool\n\n" +
				"## Original Search Question\n\nWhat answered on <REDACTED>?\n\n" +
				"## Content\n\nThe service answered from <REDACTED>.\n\n",
			wantAction: database.VecstoreActionTypeRetrieve,
		},
	} {
		t.Run(name, func(t *testing.T) {
			rig := newKnowledgeRig(t)
			rig.conn.results = []vectorResult{{docs: tc.found}}

			rig.handle(t, searchDoor(rig, hostReplacer{}, 0), tc.tool, mustJSON(tc.args))

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

func TestSearch_Handle_ReportsWhatTheStoreFound(t *testing.T) {
	answer := func(content string, score float32) schema.Document {
		return schema.Document{PageContent: content, Score: score, Metadata: map[string]any{
			"answer_type": "tool", "question": "how is " + content + " found?",
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
				{docs: []schema.Document{answer("A", 0.5), answer("B", 0.875)}},
				{docs: []schema.Document{answer("A", 0.75), answer("C", 0.625), answer("D", 0.25)}},
			},
			want: "# Document 1 Search Score: 0.875000\n\n## Original Answer Type: tool\n\n" +
				"## Original Search Question\n\nhow is B found?\n\n## Content\n\nB\n\n" +
				"# Document 2 Search Score: 0.750000\n\n## Original Answer Type: tool\n\n" +
				"## Original Search Question\n\nhow is A found?\n\n## Content\n\nA\n\n" +
				"# Document 3 Search Score: 0.625000\n\n## Original Answer Type: tool\n\n" +
				"## Original Search Question\n\nhow is C found?\n\n## Content\n\nC\n\n",
		},
		"a query the store fails": {
			questions: []string{"first", "second"},
			results: []vectorResult{
				{err: errors.New("connection reset by peer")},
				{docs: []schema.Document{answer("A", 0.75)}},
			},
			want: "# Document 1 Search Score: 0.750000\n\n## Original Answer Type: tool\n\n" +
				"## Original Search Question\n\nhow is A found?\n\n## Content\n\nA\n\n",
		},
		"nothing found": {
			questions: []string{"first"},
			results:   []vectorResult{{}},
			want:      "nothing found in answer store and you need to store it after figure out this case",
		},
	} {
		t.Run(name, func(t *testing.T) {
			rig := newKnowledgeRig(t)
			rig.conn.results = tc.results

			got := rig.handle(t, searchDoor(rig, identityReplacer{}, 0), SearchAnswerToolName, mustJSON(map[string]any{
				"questions": tc.questions, "message": "m", "type": "tool",
			}))

			if got != tc.want {
				t.Errorf("search_answer answered\n%s\nwant\n%s", got, tc.want)
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

func TestSearch_Handle_StoresOnlyTheAnonymizedForm(t *testing.T) {
	const (
		answer   = "The service answered from <REDACTED>."
		question = "What answered on <REDACTED>?"
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

			rig.handle(t, searchDoor(rig, hostReplacer{}, tc.limit), StoreAnswerToolName, mustJSON(map[string]any{
				"answer": "The service answered from " + secretHost + ".", "message": "m",
				"question": "What answered on " + secretHost + "?", "type": "tool",
			}))

			if viaStore := len(rig.conn.added) > 0; viaStore != tc.viaStore {
				t.Errorf("written through the vector store: %v, want %v", viaStore, tc.viaStore)
			}
			_, stored, meta := rig.stored(t)
			if stored != answer || meta["question"] != question {
				t.Errorf("stored %q under the question %q, want %q under %q", stored, meta["question"], answer, question)
			}
			if !slices.Equal(rig.embedder.embedded, []string{answer}) {
				t.Errorf("sent %q to the embedder, want %q", rig.embedder.embedded, answer)
			}
			doc := rig.announced(t)
			if doc.Content != answer || doc.Question != question {
				t.Errorf("announced %q under the question %q, want %q under %q", doc.Content, doc.Question, answer, question)
			}
		})
	}
}

func TestSearch_Handle_StoresTheWholeAnswerAndEmbedsWhatFits(t *testing.T) {
	const answer = "The answer body."

	for name, tc := range map[string]struct {
		answer     string
		limit      int
		noEmbedder bool
		viaStore   bool
		want       string
		embedded   string // not checked when empty
	}{
		"an answer within the embedding limit": {
			answer: answer, limit: 64, viaStore: true, want: answer, embedded: answer,
		},
		"an answer beyond the embedding limit": {answer: answer, limit: 8, want: answer, embedded: "The answ"},
		"no embedding limit":                   {answer: answer, limit: 0, want: answer, embedded: answer},
		"no embedder of its own": {
			answer: answer, limit: 8, noEmbedder: true, viaStore: true, want: answer, embedded: answer,
		},
		"an answer beyond the content limit": {
			answer: strings.Repeat("я", limits.MaxContentLen+1), want: strings.Repeat("я", limits.MaxContentLen),
		},
	} {
		t.Run(name, func(t *testing.T) {
			rig := newKnowledgeRig(t)
			if tc.noEmbedder {
				rig.doorEmbedder = nil
			}

			result := rig.handle(t, searchDoor(rig, identityReplacer{}, tc.limit), StoreAnswerToolName, mustJSON(map[string]any{
				"answer": tc.answer, "message": "m", "question": "q", "type": "tool",
			}))

			if result != "answer for question stored successfully" {
				t.Errorf("store_answer answered %q", result)
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
			if doc.DocType != model.KnowledgeDocTypeAnswer || deref(doc.AnswerType) != "tool" {
				t.Errorf("announced as a %q of type %q", doc.DocType, deref(doc.AnswerType))
			}
			if deref(doc.FlowID) != 2 || deref(doc.TaskID) != 3 || deref(doc.SubtaskID) != 4 {
				t.Errorf("announced as from flow %d, task %d, subtask %d, want 2, 3, 4",
					deref(doc.FlowID), deref(doc.TaskID), deref(doc.SubtaskID))
			}
		})
	}
}

func TestSearch_Handle_RefusesWhatItCannotServe(t *testing.T) {
	store := mustJSON(map[string]any{"answer": "The answer body.", "message": "m", "question": "q", "type": "tool"})

	for name, tc := range map[string]struct {
		tool    string
		args    json.RawMessage
		limit   int
		prepare func(rig *knowledgeRig)
		want    string
	}{
		"no vector store": {
			tool: StoreAnswerToolName, args: store,
			prepare: func(rig *knowledgeRig) { rig.store = nil },
			want:    "pgvector store is not initialized",
		},
		"an unknown tool": {tool: SearchGuideToolName, args: store, want: "unknown tool: search_guide"},
		"malformed search arguments": {
			tool: SearchAnswerToolName, args: json.RawMessage(`{"questions":`),
			want: "failed to unmarshal search_answer search answer action arguments: unexpected end of JSON input",
		},
		"malformed store arguments": {
			tool: StoreAnswerToolName, args: json.RawMessage(`{"answer":`),
			want: "failed to unmarshal store_answer store answer action arguments: unexpected end of JSON input",
		},
		"a write the database refuses": {
			tool: StoreAnswerToolName, args: store,
			prepare: func(rig *knowledgeRig) { rig.db.err = errors.New("connection reset by peer") },
			want: "failed to store answer for question: " +
				"failed to insert document with pre-computed embedding: connection reset by peer",
		},
		"a write the vector store refuses": {
			tool: StoreAnswerToolName, args: store, limit: 64,
			prepare: func(rig *knowledgeRig) { rig.conn.addErr = errors.New("connection reset by peer") },
			want:    "failed to store answer for question: connection reset by peer",
		},
	} {
		t.Run(name, func(t *testing.T) {
			rig := newKnowledgeRig(t)
			if tc.prepare != nil {
				tc.prepare(rig)
			}
			door := searchDoor(rig, identityReplacer{}, tc.limit)

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
