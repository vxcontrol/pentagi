package mocks

import (
	"context"
	"encoding/json"
	"fmt"

	"pentagi/pkg/database"
	"pentagi/pkg/graph/model"
	"pentagi/pkg/terminal"
	"pentagi/pkg/tools"
)

type ProxyProviders interface {
	GetScreenshotProvider() tools.ScreenshotProvider
	GetAgentLogProvider() tools.AgentLogProvider
	GetMsgLogProvider() tools.MsgLogProvider
	GetSearchLogProvider() tools.SearchLogProvider
	GetTermLogProvider() tools.TermLogProvider
	GetVectorStoreLogProvider() tools.VectorStoreLogProvider
	GetKnowledgeProvider() tools.KnowledgeProvider
}

// proxyProviders contains all the proxy implementations for various providers
type proxyProviders struct {
	screenshot        *proxyScreenshotProvider
	agentLog          *proxyAgentLogProvider
	msgLog            *proxyMsgLogProvider
	searchLog         *proxySearchLogProvider
	termLog           *proxyTermLogProvider
	vectorStoreLog    *proxyVectorStoreLogProvider
	knowledgeProvider *proxyKnowledgeProvider
}

// NewProxyProviders creates a new set of proxy providers
func NewProxyProviders() ProxyProviders {
	return &proxyProviders{
		screenshot:        &proxyScreenshotProvider{},
		agentLog:          &proxyAgentLogProvider{},
		msgLog:            &proxyMsgLogProvider{},
		searchLog:         &proxySearchLogProvider{},
		termLog:           &proxyTermLogProvider{},
		vectorStoreLog:    &proxyVectorStoreLogProvider{},
		knowledgeProvider: &proxyKnowledgeProvider{},
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

func (p *proxyProviders) GetKnowledgeProvider() tools.KnowledgeProvider {
	return p.knowledgeProvider
}

// proxyScreenshotProvider is a proxy implementation of ScreenshotProvider
type proxyScreenshotProvider struct{}

// PutScreenshot implements the ScreenshotProvider interface
func (p *proxyScreenshotProvider) PutScreenshot(ctx context.Context, name, url string, taskID, subtaskID *int64) (int64, error) {
	terminal.PrintInfo("Screenshot saved:")
	terminal.PrintKeyValue("Name", name)
	terminal.PrintKeyValue("URL", url)

	if taskID != nil {
		terminal.PrintKeyValueFormat("Task ID", "%d", *taskID)
	}
	if subtaskID != nil {
		terminal.PrintKeyValueFormat("Subtask ID", "%d", *subtaskID)
	}

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
	streamID int64, // unsupported for now
	thinking, msg string,
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
	streamID int64, // unsupported for now
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
	taskID, subtaskID *int64,
) (int64, error) {
	terminal.PrintInfo("Terminal log saved:")
	terminal.PrintKeyValue("Type", string(msgType))
	terminal.PrintKeyValueFormat("Container ID", "%d", containerID)

	if taskID != nil {
		terminal.PrintKeyValueFormat("Task ID", "%d", *taskID)
	}
	if subtaskID != nil {
		terminal.PrintKeyValueFormat("Subtask ID", "%d", *subtaskID)
	}

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

// proxyKnowledgeProvider is a proxy implementation of KnowledgeProvider
type proxyKnowledgeProvider struct{}

// KnowledgeDocumentCreated implements the KnowledgeProvider interface
func (p *proxyKnowledgeProvider) KnowledgeDocumentCreated(ctx context.Context, doc *model.KnowledgeDocument) {
	terminal.PrintInfo("Knowledge document created:")
	terminal.PrintKeyValue("ID", doc.ID)
	terminal.PrintKeyValue("Type", string(doc.DocType))
	terminal.PrintKeyValue("Content", doc.Content)
	terminal.PrintKeyValue("Question", doc.Question)
	if doc.Description != nil {
		terminal.PrintKeyValue("Description", *doc.Description)
	}
	if doc.FlowID != nil {
		terminal.PrintKeyValueFormat("Flow ID", "%d", *doc.FlowID)
	}
	if doc.TaskID != nil {
		terminal.PrintKeyValueFormat("Task ID", "%d", *doc.TaskID)
	}
	if doc.SubtaskID != nil {
		terminal.PrintKeyValueFormat("Subtask ID", "%d", *doc.SubtaskID)
	}
	if doc.GuideType != nil {
		terminal.PrintKeyValue("Guide Type", string(*doc.GuideType))
	}
	if doc.AnswerType != nil {
		terminal.PrintKeyValue("Answer Type", string(*doc.AnswerType))
	}
	if doc.CodeLang != nil {
		terminal.PrintKeyValue("Code Lang", *doc.CodeLang)
	}
	terminal.PrintKeyValueFormat("Part Size", "%d", doc.PartSize)
	terminal.PrintKeyValueFormat("Total Size", "%d", doc.TotalSize)
	terminal.PrintKeyValue("Manual", fmt.Sprintf("%t", doc.Manual))
	terminal.PrintKeyValueFormat("User ID", "%d", doc.UserID)
}
