package tools

import (
	"context"
	"testing"

	"pentagi/pkg/database"
)

func TestGetAgentContextEmpty(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	_, ok := GetAgentContext(ctx)
	if ok {
		t.Error("GetAgentContext() on empty context should return false")
	}
}

func TestPutAgentContextFirst(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
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

	ctx := context.Background()
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

	ctx := context.Background()
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

	ctx := context.Background()
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
