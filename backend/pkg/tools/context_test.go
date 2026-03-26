package tools

import (
	"context"
	"testing"

	"pentagi/pkg/database"
)

func TestGetAgentContextEmpty(t *testing.T) {
	t.Parallel()

	ctx := t.Context()
	_, ok := GetAgentContext(ctx)
	if ok {
		t.Error("GetAgentContext() on empty context should return false")
	}
}

func TestPutAgentContextFirst(t *testing.T) {
	t.Parallel()

	ctx := t.Context()
	agent := database.MsgchainTypePrimaryAgent

	ctx = PutAgentContext(ctx, agent)
	agentCtx, ok := GetAgentContext(ctx)
	if !ok {
		t.Fatal("GetAgentContext() should return true after PutAgentContext")
	}
	if agentCtx.ParentAgentType != agent {
		t.Errorf("ParentAgentType = %q, want %q", agentCtx.ParentAgentType, agent)
	}
	if agentCtx.CurrentAgentType != agent {
		t.Errorf("CurrentAgentType = %q, want %q", agentCtx.CurrentAgentType, agent)
	}
}

func TestPutAgentContextChaining(t *testing.T) {
	t.Parallel()

	ctx := t.Context()
	first := database.MsgchainTypePrimaryAgent
	second := database.MsgchainTypeSearcher

	ctx = PutAgentContext(ctx, first)
	ctx = PutAgentContext(ctx, second)

	agentCtx, ok := GetAgentContext(ctx)
	if !ok {
		t.Fatal("GetAgentContext() should return true")
	}
	if agentCtx.ParentAgentType != first {
		t.Errorf("ParentAgentType = %q, want %q (first agent should become parent)", agentCtx.ParentAgentType, first)
	}
	if agentCtx.CurrentAgentType != second {
		t.Errorf("CurrentAgentType = %q, want %q", agentCtx.CurrentAgentType, second)
	}
}

func TestPutAgentContextTripleChaining(t *testing.T) {
	t.Parallel()

	ctx := t.Context()
	first := database.MsgchainTypePrimaryAgent
	second := database.MsgchainTypeSearcher
	third := database.MsgchainTypePentester

	ctx = PutAgentContext(ctx, first)
	ctx = PutAgentContext(ctx, second)
	ctx = PutAgentContext(ctx, third)

	agentCtx, ok := GetAgentContext(ctx)
	if !ok {
		t.Fatal("GetAgentContext() should return true")
	}
	// After triple chaining: parent = second (promoted from current), current = third
	if agentCtx.ParentAgentType != second {
		t.Errorf("ParentAgentType = %q, want %q (previous current should become parent)", agentCtx.ParentAgentType, second)
	}
	if agentCtx.CurrentAgentType != third {
		t.Errorf("CurrentAgentType = %q, want %q", agentCtx.CurrentAgentType, third)
	}
}

func TestPutAgentContextIsolation(t *testing.T) {
	t.Parallel()

	ctx := t.Context()
	agent := database.MsgchainTypeCoder

	newCtx := PutAgentContext(ctx, agent)

	// Original context should not be affected
	_, ok := GetAgentContext(ctx)
	if ok {
		t.Error("original context should not contain agent context")
	}

	// New context should have the agent
	agentCtx, ok := GetAgentContext(newCtx)
	if !ok {
		t.Fatal("new context should contain agent context")
	}
	if agentCtx.CurrentAgentType != agent {
		t.Errorf("CurrentAgentType = %q, want %q", agentCtx.CurrentAgentType, agent)
	}
}

func TestPutAgentContextDoesNotMutatePreviousDerivedContext(t *testing.T) {
	t.Parallel()

	baseCtx := t.Context()
	first := database.MsgchainTypePrimaryAgent
	second := database.MsgchainTypeSearcher

	ctx1 := PutAgentContext(baseCtx, first)
	ctx2 := PutAgentContext(ctx1, second)

	agentCtx1, ok := GetAgentContext(ctx1)
	if !ok {
		t.Fatal("ctx1 should contain agent context")
	}
	if agentCtx1.ParentAgentType != first || agentCtx1.CurrentAgentType != first {
		t.Fatalf("ctx1 changed unexpectedly: parent=%q current=%q", agentCtx1.ParentAgentType, agentCtx1.CurrentAgentType)
	}

	agentCtx2, ok := GetAgentContext(ctx2)
	if !ok {
		t.Fatal("ctx2 should contain agent context")
	}
	if agentCtx2.ParentAgentType != first || agentCtx2.CurrentAgentType != second {
		t.Fatalf("ctx2 mismatch: parent=%q current=%q", agentCtx2.ParentAgentType, agentCtx2.CurrentAgentType)
	}
}

func TestGetAgentContextIgnoresOtherContextValues(t *testing.T) {
	t.Parallel()

	type foreignKey string
	ctx := context.WithValue(t.Context(), foreignKey("k"), "v")
	_, ok := GetAgentContext(ctx)
	if ok {
		t.Error("GetAgentContext() should ignore unrelated context values")
	}
}
