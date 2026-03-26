package tools

import (
	"context"

	"pentagi/pkg/database"
)

type AgentContextKey int

var agentContextKey AgentContextKey

type agentContext struct {
	ParentAgentType  database.MsgchainType `json:"parent_agent_type"`
	CurrentAgentType database.MsgchainType `json:"current_agent_type"`
}

func GetAgentContext(ctx context.Context) (agentContext, bool) {
	agentCtx, ok := ctx.Value(agentContextKey).(agentContext)
	return agentCtx, ok
}

func PutAgentContext(ctx context.Context, agent database.MsgchainType) context.Context {
	agentCtx, ok := GetAgentContext(ctx)
	if !ok {
		agentCtx.ParentAgentType = agent
		agentCtx.CurrentAgentType = agent
	} else {
		agentCtx.ParentAgentType = agentCtx.CurrentAgentType
		agentCtx.CurrentAgentType = agent
	}

	return context.WithValue(ctx, agentContextKey, agentCtx)
}
