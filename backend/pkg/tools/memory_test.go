package tools

import (
	"context"
	"encoding/json"
	"errors"
	"reflect"
	"strings"
	"testing"

	"pentagi/pkg/database"

	"github.com/vxcontrol/langchaingo/schema"
)

const memoryFlowID = 42

func memoryToolOver(t *testing.T, conn *fakeVectorConn, log *recordingVectorStoreLog) Tool {
	t.Helper()

	store := newFakeVectorStore(t, conn, &fakeEmbedder{err: errors.New("search_in_memory has no documents to embed")})

	return NewMemoryTool(memoryFlowID, hostReplacer{}, store, log)
}

// Metadata numbers are float64: the store's JSON metadata column scans back into that type.
func memoryFact(content string, score float32, metadata map[string]any) schema.Document {
	return schema.Document{PageContent: content, Score: score, Metadata: metadata}
}

func memoryArgs(questions []string, ids map[string]any) json.RawMessage {
	args := map[string]any{"questions": questions, "message": "Looking for what the flow already knows."}
	for key, value := range ids {
		args[key] = value
	}
	return mustJSON(args)
}

func TestMemory_Handle_RefusesACallItCannotServe(t *testing.T) {
	t.Parallel()

	args := memoryArgs([]string{"which ports are open on the gateway?"}, nil)

	t.Run("a tool without a vector store", func(t *testing.T) {
		t.Parallel()

		tool := NewMemoryTool(memoryFlowID, hostReplacer{}, nil, &recordingVectorStoreLog{})

		if tool.IsAvailable() {
			t.Error("a memory tool without a vector store reports itself available, so agents are offered search_in_memory")
		}
		if _, err := tool.Handle(t.Context(), SearchInMemoryToolName, args); err == nil || err.Error() != "memory is not available" {
			t.Errorf("err = %v, want memory is not available", err)
		}
	})

	for name, tc := range map[string]struct {
		tool string
		args json.RawMessage
		want string
	}{
		"a tool name it does not serve": {
			tool: SearchGuideToolName,
			args: args,
			want: "unknown tool: search_guide",
		},
		"arguments cut off mid-list": {
			tool: SearchInMemoryToolName,
			args: json.RawMessage(`{"questions": ["which ports are open on the gateway?"`),
			want: "failed to unmarshal search_in_memory search in memory action arguments",
		},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			conn := &fakeVectorConn{}
			tool := memoryToolOver(t, conn, &recordingVectorStoreLog{})

			if !tool.IsAvailable() {
				t.Fatal("a memory tool over a vector store reports itself unavailable")
			}

			result, err := tool.Handle(t.Context(), tc.tool, tc.args)
			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Errorf("err = %v, want it to contain %q", err, tc.want)
			}
			if result != "" {
				t.Errorf("a refused call answered %q", result)
			}
			if len(conn.queries) != 0 {
				t.Errorf("a refused call searched the store %d times", len(conn.searches()))
			}
		})
	}
}

func TestMemory_Handle_ScopesEverySearchToTheFlowsMemory(t *testing.T) {
	t.Parallel()

	questions := []string{"which ports are open on the gateway?", "which credentials worked on ssh?"}

	for name, tc := range map[string]struct {
		ids  map[string]any
		want map[string]string
	}{
		"no task or subtask": {
			want: map[string]string{"flow_id": "42", "doc_type": "memory"},
		},
		"a task and a subtask": {
			ids:  map[string]any{"task_id": 7, "subtask_id": 9},
			want: map[string]string{"flow_id": "42", "doc_type": "memory", "task_id": "7", "subtask_id": "9"},
		},
		"a task only": {
			ids:  map[string]any{"task_id": 7},
			want: map[string]string{"flow_id": "42", "doc_type": "memory", "task_id": "7"},
		},
		"ids of zero": {
			ids:  map[string]any{"task_id": 0, "subtask_id": 0},
			want: map[string]string{"flow_id": "42", "doc_type": "memory"},
		},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			conn := &fakeVectorConn{results: []vectorResult{
				{docs: []schema.Document{memoryFact("22/tcp open ssh", 0.9, map[string]any{"tool_name": "terminal"})}},
				{docs: []schema.Document{memoryFact("root:toor accepted", 0.8, map[string]any{"tool_name": "terminal"})}},
			}}

			if _, err := memoryToolOver(t, conn, &recordingVectorStoreLog{}).Handle(
				t.Context(), SearchInMemoryToolName, memoryArgs(questions, tc.ids),
			); err != nil {
				t.Fatalf("Handle: %v", err)
			}

			if len(conn.searches()) != len(questions) {
				t.Fatalf("%d searches for %d questions that each found a fact: %+v", len(conn.searches()), len(questions), conn.searches())
			}
			for i, search := range conn.searches() {
				if search.question != questions[i] {
					t.Errorf("search %d asked %q, want %q", i, search.question, questions[i])
				}
				if !reflect.DeepEqual(search.filters, tc.want) {
					t.Errorf("search %d filtered on %v, want %v", i, search.filters, tc.want)
				}
				if search.limit != 3 {
					t.Errorf("search %d asked for %d facts, want 3", i, search.limit)
				}
				if search.minScore != 0.2 {
					t.Errorf("search %d kept facts scoring above %v, want 0.2", i, search.minScore)
				}
			}
		})
	}
}

// The widened search drops only the task and subtask ids; its limit and threshold stay.
func TestMemory_Handle_FallsBackToTheWholeFlowWhenTheTaskHasNoFacts(t *testing.T) {
	t.Parallel()

	question := "which credentials worked on ssh?"
	conn := &fakeVectorConn{results: []vectorResult{
		{},
		{docs: []schema.Document{memoryFact("root:toor accepted", 0.8, map[string]any{"task_id": 3.0, "tool_name": "terminal"})}},
	}}

	result, err := memoryToolOver(t, conn, &recordingVectorStoreLog{}).Handle(
		t.Context(), SearchInMemoryToolName, memoryArgs([]string{question}, map[string]any{"task_id": 7, "subtask_id": 9}),
	)
	if err != nil {
		t.Fatalf("Handle: %v", err)
	}

	want := []vectorSearch{
		{
			question: question,
			filters:  map[string]string{"flow_id": "42", "doc_type": "memory", "task_id": "7", "subtask_id": "9"},
			limit:    3,
			minScore: 0.2,
		},
		{question: question, filters: map[string]string{"flow_id": "42", "doc_type": "memory"}, limit: 3, minScore: 0.2},
	}
	if !reflect.DeepEqual(conn.searches(), want) {
		t.Errorf("searches:\n%+v\nwant the subtask's, then the same search over the whole flow:\n%+v", conn.searches(), want)
	}
	if !strings.Contains(result, "## Task ID 3\n\n## Tool Name 'terminal'\n\n## Content\n\nroot:toor accepted\n") {
		t.Errorf("the fact found across the flow is not in the answer:\n%s", result)
	}
}

func TestMemory_Handle_KeepsTheFactsOfTheQuestionsWhoseSearchSucceeded(t *testing.T) {
	t.Parallel()

	questions := []string{"which ports are open on the gateway?", "which credentials worked on ssh?"}
	found := memoryFact("root:toor accepted", 0.8, map[string]any{"tool_name": "terminal"})
	lost := errors.New("conn closed")
	flow := map[string]string{"flow_id": "42", "doc_type": "memory"}
	task := map[string]string{"flow_id": "42", "doc_type": "memory", "task_id": "7"}

	for name, tc := range map[string]struct {
		ids      map[string]any
		results  []vectorResult
		searches []vectorSearch
	}{
		"the search of the first question fails": {
			results: []vectorResult{{err: lost}, {docs: []schema.Document{found}}},
			searches: []vectorSearch{
				{question: questions[0], filters: flow, limit: 3, minScore: 0.2},
				{question: questions[1], filters: flow, limit: 3, minScore: 0.2},
			},
		},
		"the whole-flow search of the first question fails": {
			ids:     map[string]any{"task_id": 7},
			results: []vectorResult{{}, {err: lost}, {docs: []schema.Document{found}}},
			searches: []vectorSearch{
				{question: questions[0], filters: task, limit: 3, minScore: 0.2},
				{question: questions[0], filters: flow, limit: 3, minScore: 0.2},
				{question: questions[1], filters: task, limit: 3, minScore: 0.2},
			},
		},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			conn := &fakeVectorConn{results: tc.results}

			result, err := memoryToolOver(t, conn, &recordingVectorStoreLog{}).Handle(
				t.Context(), SearchInMemoryToolName, memoryArgs(questions, tc.ids),
			)
			if err != nil {
				t.Fatalf("one failed search failed the call: %v", err)
			}

			if !reflect.DeepEqual(conn.searches(), tc.searches) {
				t.Fatalf("searches:\n%+v\nwant the second question searched after the first one's failure:\n%+v", conn.searches(), tc.searches)
			}
			if !strings.Contains(result, "## Content\n\nroot:toor accepted\n") {
				t.Errorf("the second question's fact is not in the answer:\n%s", result)
			}
		})
	}
}

// With a task or subtask id, every question, not only the first, gets its own whole-flow search.
func TestMemory_Handle_AnswersNothingFoundWhenNoFactMatches(t *testing.T) {
	t.Parallel()

	questions := []string{"which ports are open on the gateway?", "which credentials worked on ssh?"}
	flow := map[string]string{"flow_id": "42", "doc_type": "memory"}
	task := map[string]string{"flow_id": "42", "doc_type": "memory", "task_id": "7"}

	for name, tc := range map[string]struct {
		ids      map[string]any
		searches []vectorSearch
	}{
		"searched across the flow": {
			searches: []vectorSearch{
				{question: questions[0], filters: flow, limit: 3, minScore: 0.2},
				{question: questions[1], filters: flow, limit: 3, minScore: 0.2},
			},
		},
		"searched within a task, then across": {
			ids: map[string]any{"task_id": 7},
			searches: []vectorSearch{
				{question: questions[0], filters: task, limit: 3, minScore: 0.2},
				{question: questions[0], filters: flow, limit: 3, minScore: 0.2},
				{question: questions[1], filters: task, limit: 3, minScore: 0.2},
				{question: questions[1], filters: flow, limit: 3, minScore: 0.2},
			},
		},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			conn := &fakeVectorConn{}

			result, err := memoryToolOver(t, conn, &recordingVectorStoreLog{}).Handle(
				t.Context(), SearchInMemoryToolName, memoryArgs(questions, tc.ids),
			)
			if err != nil {
				t.Fatalf("Handle: %v", err)
			}

			if result != "nothing found in memory store by this question" {
				t.Errorf("answer = %q", result)
			}
			if !reflect.DeepEqual(conn.searches(), tc.searches) {
				t.Errorf("searches:\n%+v\nwant:\n%+v", conn.searches(), tc.searches)
			}
		})
	}
}

func TestMemory_Handle_AnswersWithTheBestFactsOfAllQuestions(t *testing.T) {
	t.Parallel()

	ssh := map[string]any{
		"task_id": 7.0, "subtask_id": 9.0, "tool_name": "terminal", "tool_description": "Runs a command in the sandbox.",
	}
	conn := &fakeVectorConn{results: []vectorResult{
		{docs: []schema.Document{
			memoryFact("22/tcp open ssh", 0.9, ssh),
			memoryFact("robots.txt lists /admin", 0.5, map[string]any{"tool_name": "browser"}),
		}},
		{docs: []schema.Document{
			memoryFact("22/tcp open ssh", 0.7, ssh),
			memoryFact("OpenSSH 7.2p2 is affected by CVE-2016-6210", 0.8, map[string]any{"task_id": 7.0, "tool_name": "search"}),
			memoryFact("the banner names nginx", 0.3, map[string]any{"tool_name": "terminal"}),
		}},
	}}

	result, err := memoryToolOver(t, conn, &recordingVectorStoreLog{}).Handle(
		t.Context(), SearchInMemoryToolName,
		memoryArgs([]string{"which ports are open on the gateway?", "is the ssh daemon vulnerable?"}, nil),
	)
	if err != nil {
		t.Fatalf("Handle: %v", err)
	}

	want := `# Retrieved Memory Fact 1 Match score: 0.900000

## Task ID 7

## Subtask ID 9

## Tool Name 'terminal'

## Tool Description

Runs a command in the sandbox.

## Content

22/tcp open ssh
---------------------------
# Retrieved Memory Fact 2 Match score: 0.800000

## Task ID 7

## Tool Name 'search'

## Content

OpenSSH 7.2p2 is affected by CVE-2016-6210
---------------------------
# Retrieved Memory Fact 3 Match score: 0.500000

## Tool Name 'browser'

## Content

robots.txt lists /admin
---------------------------
`
	if result != want {
		t.Errorf("answer:\n%s\nwant:\n%s", result, want)
	}
}

func TestMemory_Handle_LogsTheRetrievalAnonymized(t *testing.T) {
	t.Parallel()

	conn := &fakeVectorConn{results: []vectorResult{
		{docs: []schema.Document{memoryFact("22/tcp open ssh", 0.9, map[string]any{"tool_name": "terminal"})}},
		{docs: []schema.Document{memoryFact("root:toor accepted", 0.8, map[string]any{"tool_name": "terminal"})}},
	}}
	log := &recordingVectorStoreLog{}
	ctx := PutAgentContext(PutAgentContext(t.Context(), database.MsgchainTypePrimaryAgent), database.MsgchainTypeMemorist)

	result, err := memoryToolOver(t, conn, log).Handle(
		ctx, SearchInMemoryToolName,
		memoryArgs([]string{"what listens on " + secretHost + "?", "which credentials worked?"}, map[string]any{"task_id": 7}),
	)
	if err != nil {
		t.Fatalf("Handle: %v", err)
	}

	if len(log.entries) != 1 {
		t.Fatalf("%d vector store log entries, want 1", len(log.entries))
	}
	entry := log.entries[0]

	if entry.initiator != database.MsgchainTypePrimaryAgent || entry.executor != database.MsgchainTypeMemorist {
		t.Errorf("logged as %s -> %s, want primary_agent -> memorist", entry.initiator, entry.executor)
	}
	if entry.filter != `{"doc_type":"memory","flow_id":"42","task_id":"7"}` {
		t.Errorf("logged filter %s", entry.filter)
	}
	if want := "what listens on <REDACTED>?\n--------------------------------\nwhich credentials worked?"; entry.query != want {
		t.Errorf("logged query %q, want %q", entry.query, want)
	}
	if entry.action != database.VecstoreActionTypeRetrieve {
		t.Errorf("logged action %s, want retrieve", entry.action)
	}
	if entry.result != result {
		t.Errorf("logged result %q, but the agent was answered %q", entry.result, result)
	}
	if entry.taskID == nil || *entry.taskID != 7 || entry.subtaskID != nil {
		t.Errorf("logged task %v subtask %v, want task 7 and no subtask", entry.taskID, entry.subtaskID)
	}
}

func TestMemory_Handle_AnswersWhetherOrNotTheRetrievalIsLogged(t *testing.T) {
	t.Parallel()

	for name, tc := range map[string]struct {
		ctx     func(context.Context) context.Context
		logErr  error
		entries int
	}{
		"no agent to attribute it to": {
			ctx:     func(ctx context.Context) context.Context { return ctx },
			entries: 0,
		},
		"the log write fails": {
			ctx: func(ctx context.Context) context.Context {
				return PutAgentContext(ctx, database.MsgchainTypeMemorist)
			},
			logErr:  errors.New("connection reset by peer"),
			entries: 1,
		},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			conn := &fakeVectorConn{results: []vectorResult{
				{docs: []schema.Document{memoryFact("22/tcp open ssh", 0.9, map[string]any{"tool_name": "terminal"})}},
			}}
			log := &recordingVectorStoreLog{err: tc.logErr}

			result, err := memoryToolOver(t, conn, log).Handle(
				tc.ctx(t.Context()), SearchInMemoryToolName, memoryArgs([]string{"which ports are open?"}, nil),
			)
			if err != nil {
				t.Fatalf("the retrieval log decided the call: %v", err)
			}

			if !strings.Contains(result, "## Content\n\n22/tcp open ssh\n") {
				t.Errorf("the fact is not in the answer:\n%s", result)
			}
			if len(log.entries) != tc.entries {
				t.Errorf("%d vector store log entries, want %d", len(log.entries), tc.entries)
			}
		})
	}
}
