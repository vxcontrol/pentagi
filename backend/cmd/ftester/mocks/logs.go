package mocks

import (
	"context"
	"encoding/json"

	"pentagi/cmd/ftester/terminal"
	"pentagi/pkg/database"
	"pentagi/pkg/tools"
)

type ProxyProviders interface {
	GetScreenshotProvider() tools.ScreenshotProvider
	GetAgentLogProvider() tools.AgentLogProvider
	GetMsgLogProvider() tools.MsgLogProvider
	GetSearchLogProvider() tools.SearchLogProvider
	GetTermLogProvider() tools.TermLogProvider
	GetVectorStoreLogProvider() tools.VectorStoreLogProvider
}

// proxyProviders contains all the proxy implementations for various providers
type proxyProviders struct {
	screenshot     *proxyScreenshotProvider
	agentLog       *proxyAgentLogProvider
	msgLog         *proxyMsgLogProvider
	searchLog      *proxySearchLogProvider
	termLog        *proxyTermLogProvider
	vectorStoreLog *proxyVectorStoreLogProvider
}

// NewProxyProviders creates a new set of proxy providers
func NewProxyProviders() ProxyProviders {
	return &proxyProviders{
		screenshot:     &proxyScreenshotProvider{},
		agentLog:       &proxyAgentLogProvider{},
		msgLog:         &proxyMsgLogProvider{},
		searchLog:      &proxySearchLogProvider{},
		termLog:        &proxyTermLogProvider{},
		vectorStoreLog: &proxyVectorStoreLogProvider{},
	}
}

func (p *proxyProviders) GetScreenshotProvider() tools.ScreenshotProvider {
	return p.screenshot
}

func (p *proxyProviders) GetAgentLogProvider() tools.AgentLogProvider {
	return p.agentLog
}

func (p *proxyProviders) GetMsgLogProvider() tools.MsgLogProvider {
	return p.msgLog
}

func (p *proxyProviders) GetSearchLogProvider() tools.SearchLogProvider {
	return p.searchLog
}

func (p *proxyProviders) GetTermLogProvider() tools.TermLogProvider {
	return p.termLog
}

func (p *proxyProviders) GetVectorStoreLogProvider() tools.VectorStoreLogProvider {
	return p.vectorStoreLog
}

// proxyScreenshotProvider is a proxy implementation of ScreenshotProvider
type proxyScreenshotProvider struct{}

// PutScreenshot implements the ScreenshotProvider interface
func (p *proxyScreenshotProvider) PutScreenshot(ctx context.Context, name, url string) (int64, error) {
	terminal.PrintInfo("Screenshot saved:")
	terminal.PrintKeyValue("Name", name)
	terminal.PrintKeyValue("URL", url)
	return 0, nil
}

// proxyAgentLogProvider is a proxy implementation of AgentLogProvider
type proxyAgentLogProvider struct{}

// PutLog implements the AgentLogProvider interface
func (p *proxyAgentLogProvider) PutLog(
	ctx context.Context,
	initiator database.MsgchainType,
	executor database.MsgchainType,
	task string,
	result string,
	taskID *int64,
	subtaskID *int64,
) (int64, error) {
	terminal.PrintInfo("Agent log saved:")
	terminal.PrintKeyValue("Initiator", string(initiator))
	terminal.PrintKeyValue("Executor", string(executor))
	terminal.PrintKeyValue("Task", task)

	if taskID != nil {
		terminal.PrintKeyValueFormat("Task ID", "%d", *taskID)
	}
	if subtaskID != nil {
		terminal.PrintKeyValueFormat("Subtask ID", "%d", *subtaskID)
	}

	if len(result) > 0 {
		terminal.PrintResultWithKey("Result", result)
	}

	return 0, nil
}

// proxyMsgLogProvider is a proxy implementation of MsgLogProvider
type proxyMsgLogProvider struct{}

// PutMsg implements the MsgLogProvider interface
func (p *proxyMsgLogProvider) PutMsg(
	ctx context.Context,
	msgType database.MsglogType,
	taskID, subtaskID *int64,
	msg string,
) (int64, error) {
	terminal.PrintInfo("Message logged:")
	terminal.PrintKeyValue("Type", string(msgType))

	if taskID != nil {
		terminal.PrintKeyValueFormat("Task ID", "%d", *taskID)
	}
	if subtaskID != nil {
		terminal.PrintKeyValueFormat("Subtask ID", "%d", *subtaskID)
	}

	if len(msg) > 0 {
		terminal.PrintResultWithKey("Message", msg)
	}

	return 0, nil
}

// UpdateMsgResult implements the MsgLogProvider interface
func (p *proxyMsgLogProvider) UpdateMsgResult(
	ctx context.Context,
	msgID int64,
	result string,
	resultFormat database.MsglogResultFormat,
) error {
	terminal.PrintInfo("Message result updated:")
	terminal.PrintKeyValueFormat("Message ID", "%d", msgID)
	terminal.PrintKeyValue("Format", string(resultFormat))

	if len(result) > 0 {
		terminal.PrintResultWithKey("Result", result)
	}

	return nil
}

// proxySearchLogProvider is a proxy implementation of SearchLogProvider
type proxySearchLogProvider struct{}

// PutLog implements the SearchLogProvider interface
func (p *proxySearchLogProvider) PutLog(
	ctx context.Context,
	initiator database.MsgchainType,
	executor database.MsgchainType,
	engine database.SearchengineType,
	query string,
	result string,
	taskID *int64,
	subtaskID *int64,
) (int64, error) {
	terminal.PrintInfo("Search log saved:")
	terminal.PrintKeyValue("Initiator", string(initiator))
	terminal.PrintKeyValue("Executor", string(executor))
	terminal.PrintKeyValue("Engine", string(engine))
	terminal.PrintKeyValue("Query", query)

	if taskID != nil {
		terminal.PrintKeyValueFormat("Task ID", "%d", *taskID)
	}
	if subtaskID != nil {
		terminal.PrintKeyValueFormat("Subtask ID", "%d", *subtaskID)
	}

	if len(result) > 0 {
		terminal.PrintResultWithKey("Search Result", result)
	}

	return 0, nil
}

// proxyTermLogProvider is a proxy implementation of TermLogProvider
type proxyTermLogProvider struct{}

// PutMsg implements the TermLogProvider interface
func (p *proxyTermLogProvider) PutMsg(
	ctx context.Context,
	msgType database.TermlogType,
	msg string,
	containerID int64,
) (int64, error) {
	terminal.PrintInfo("Terminal log saved:")
	terminal.PrintKeyValue("Type", string(msgType))
	terminal.PrintKeyValueFormat("Container ID", "%d", containerID)

	if len(msg) > 0 {
		terminal.PrintResultWithKey("Terminal Output", msg)
	}

	return 0, nil
}

// proxyVectorStoreLogProvider is a proxy implementation of VectorStoreLogProvider
type proxyVectorStoreLogProvider struct{}

// PutLog implements the VectorStoreLogProvider interface
func (p *proxyVectorStoreLogProvider) PutLog(
	ctx context.Context,
	initiator database.MsgchainType,
	executor database.MsgchainType,
	filter json.RawMessage,
	query string,
	action database.VecstoreActionType,
	result string,
	taskID *int64,
	subtaskID *int64,
) (int64, error) {
	terminal.PrintInfo("Vector store log saved:")
	terminal.PrintKeyValue("Initiator", string(initiator))
	terminal.PrintKeyValue("Executor", string(executor))
	terminal.PrintKeyValue("Action", string(action))
	terminal.PrintKeyValue("Query", query)

	if taskID != nil {
		terminal.PrintKeyValueFormat("Task ID", "%d", *taskID)
	}
	if subtaskID != nil {
		terminal.PrintKeyValueFormat("Subtask ID", "%d", *subtaskID)
	}

	if len(result) > 0 {
		terminal.PrintResultWithKey("Vector Store Result", result)
	}

	return 0, nil
}
