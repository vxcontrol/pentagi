package smallmodel

import (
	"strings"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestState_ToBlock_EmptyStateYieldsEmptyString(t *testing.T) {
	assert.Equal(t, "", State{}.ToBlock())
}

func TestState_ToBlock_RendersAllSections(t *testing.T) {
	s := State{
		Objective:   "authorized test on 10.0.0.5",
		Facts:       []string{"port 8443 open, Tomcat 9.0.31", "/manager reachable, needs auth"},
		FailedTries: []string{"default creds on /manager -> 401"},
		NextStep:    "check known CVEs for Tomcat 9.0.31",
	}
	block := s.ToBlock()

	assert.Contains(t, block, "OBJECTIVE: authorized test on 10.0.0.5")
	assert.Contains(t, block, "  - port 8443 open, Tomcat 9.0.31")
	assert.Contains(t, block, "FAILED ATTEMPTS:")
	assert.Contains(t, block, "  - default creds on /manager -> 401")
	assert.Contains(t, block, "NEXT STEP: check known CVEs for Tomcat 9.0.31")
	assert.False(t, strings.HasSuffix(block, "\n"))
}

func TestStore_AddFacts_DeduplicatesAndSkipsBlanks(t *testing.T) {
	st := NewStore()
	k := Key{FlowID: 1, TaskID: 2}

	st.AddFacts(k, "port 22 open", "port 22 open", "  ", "port 80 open")
	block := st.Block(k)

	assert.Equal(t, 1, strings.Count(block, "port 22 open"))
	assert.Contains(t, block, "port 80 open")
}

func TestStore_Block_IsEmptyForUnknownKey(t *testing.T) {
	st := NewStore()
	assert.Equal(t, "", st.Block(Key{FlowID: 9, TaskID: 9}))
}

func TestStore_Reset_DropsState(t *testing.T) {
	st := NewStore()
	k := Key{FlowID: 1, TaskID: 1}
	st.SetObjective(k, "x")
	st.Reset(k)
	assert.Equal(t, "", st.Block(k))
}

func TestStore_ConcurrentWritesAreSafe(t *testing.T) {
	st := NewStore()
	k := Key{FlowID: 1, TaskID: 1}

	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			st.AddFacts(k, string(rune('a'+i%26))+"-fact")
		}(i)
	}
	wg.Wait()

	assert.NotEmpty(t, st.Block(k))
}
