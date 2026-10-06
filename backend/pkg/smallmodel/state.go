package smallmodel

import (
	"fmt"
	"sort"
	"strings"
	"sync"
)

// State is the structured working memory of a pentest task. It is built by
// deterministic code from tool output, never by the small model, and is placed
// at the top of the prompt on every turn. Anything that must survive chain
// compaction lives here, not in the conversation history.
type State struct {
	Objective   string
	Facts       []string
	FailedTries []string
	NextStep    string
}

// ToBlock serializes the state into the fixed block that heads the prompt. The
// labels are stable so a small model learns the shape; an empty state yields an
// empty string so nothing is injected before there is anything to say.
func (s State) ToBlock() string {
	if s.Objective == "" && len(s.Facts) == 0 && len(s.FailedTries) == 0 && s.NextStep == "" {
		return ""
	}

	var b strings.Builder
	if s.Objective != "" {
		fmt.Fprintf(&b, "OBJECTIVE: %s\n", s.Objective)
	}
	b.WriteString("ESTABLISHED FACTS:\n")
	for _, f := range s.Facts {
		fmt.Fprintf(&b, "  - %s\n", f)
	}
	b.WriteString("FAILED ATTEMPTS:\n")
	for _, f := range s.FailedTries {
		fmt.Fprintf(&b, "  - %s\n", f)
	}
	if s.NextStep != "" {
		fmt.Fprintf(&b, "NEXT STEP: %s\n", s.NextStep)
	}
	return strings.TrimRight(b.String(), "\n")
}

// Key identifies the task a State belongs to. Subtasks of the same task share
// the task-level state so facts found in one carry into the next.
type Key struct {
	FlowID int64
	TaskID int64
}

// Store holds per-task state and is safe for concurrent use: tool handlers on
// different workers update it while the agent loop reads it.
type Store struct {
	mu     sync.RWMutex
	states map[Key]*State
}

// NewStore returns an empty Store.
func NewStore() *Store {
	return &Store{states: make(map[Key]*State)}
}

func (st *Store) stateLocked(k Key) *State {
	s, ok := st.states[k]
	if !ok {
		s = &State{}
		st.states[k] = s
	}
	return s
}

// Block returns the serialized state for a key, or an empty string if none.
func (st *Store) Block(k Key) string {
	st.mu.RLock()
	defer st.mu.RUnlock()
	if s, ok := st.states[k]; ok {
		return s.ToBlock()
	}
	return ""
}

// SetObjective records the task objective.
func (st *Store) SetObjective(k Key, objective string) {
	st.mu.Lock()
	defer st.mu.Unlock()
	st.stateLocked(k).Objective = strings.TrimSpace(objective)
}

// SetNextStep records the next planned step.
func (st *Store) SetNextStep(k Key, next string) {
	st.mu.Lock()
	defer st.mu.Unlock()
	st.stateLocked(k).NextStep = strings.TrimSpace(next)
}

// AddFacts appends facts, de-duplicating against what is already recorded.
func (st *Store) AddFacts(k Key, facts ...string) {
	st.mu.Lock()
	defer st.mu.Unlock()
	s := st.stateLocked(k)
	s.Facts = appendUnique(s.Facts, facts)
}

// AddFailedTry records a failed attempt, de-duplicating.
func (st *Store) AddFailedTry(k Key, try string) {
	st.mu.Lock()
	defer st.mu.Unlock()
	s := st.stateLocked(k)
	s.FailedTries = appendUnique(s.FailedTries, []string{try})
}

// Reset drops the state for a key, used when a task is torn down.
func (st *Store) Reset(k Key) {
	st.mu.Lock()
	defer st.mu.Unlock()
	delete(st.states, k)
}

func appendUnique(existing, incoming []string) []string {
	seen := make(map[string]struct{}, len(existing))
	for _, e := range existing {
		seen[e] = struct{}{}
	}
	for _, in := range incoming {
		in = strings.TrimSpace(in)
		if in == "" {
			continue
		}
		if _, dup := seen[in]; dup {
			continue
		}
		seen[in] = struct{}{}
		existing = append(existing, in)
	}
	return existing
}

// MergeFacts is a convenience for feeding deterministic extraction output into
// the store in a stable order.
func (st *Store) MergeFacts(k Key, facts []string) {
	sorted := make([]string, len(facts))
	copy(sorted, facts)
	sort.Strings(sorted)
	st.AddFacts(k, sorted...)
}
