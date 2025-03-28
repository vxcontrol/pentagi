package graph

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.
// Code generated by github.com/99designs/gqlgen version v0.17.45

import (
	"context"
	"errors"
	"fmt"
	"pentagi/pkg/controller"
	"pentagi/pkg/database"
	"pentagi/pkg/database/converter"
	"pentagi/pkg/graph/model"
	"pentagi/pkg/providers/provider"
	"pentagi/pkg/templates"

	"github.com/sirupsen/logrus"
)

// CreateFlow is the resolver for the createFlow field.
func (r *mutationResolver) CreateFlow(ctx context.Context, modelProvider string, input string) (*model.Flow, error) {
	uid, _, err := validatePermission(ctx, "flows.create")
	if err != nil {
		return nil, err
	}

	r.Logger.WithFields(logrus.Fields{
		"uid":   uid,
		"model": modelProvider,
		"input": input,
	}).Debug("create flow")

	if modelProvider == "" {
		return nil, fmt.Errorf("model provider is required")
	}

	if input == "" {
		return nil, fmt.Errorf("user input is required")
	}

	prvtype := provider.ProviderType(modelProvider)
	if _, err := r.ProvidersCtrl.Get(prvtype); err != nil {
		return nil, err
	}

	fw, err := r.Controller.CreateFlow(ctx, uid, input, prvtype, nil)
	if err != nil {
		return nil, err
	}

	flow, err := r.DB.GetUserFlow(ctx, database.GetUserFlowParams{
		ID:     fw.GetFlowID(),
		UserID: uid,
	})
	if err != nil {
		return nil, err
	}

	containers, err := r.DB.GetUserFlowContainers(ctx, database.GetUserFlowContainersParams{
		FlowID: flow.ID,
		UserID: uid,
	})
	if err != nil {
		return nil, err
	}

	r.Subscriptions.NewFlowPublisher(flow.UserID, flow.ID).FlowCreated(ctx, flow, containers)

	return converter.ConvertFlow(flow, containers), nil
}

// PutUserInput is the resolver for the putUserInput field.
func (r *mutationResolver) PutUserInput(ctx context.Context, flowID int64, input string) (model.ResultType, error) {
	uid, err := validatePermissionWithFlowID(ctx, "flows.edit", flowID, r.DB)
	if err != nil {
		return model.ResultTypeError, err
	}

	r.Logger.WithFields(logrus.Fields{
		"uid":  uid,
		"flow": flowID,
	}).Debug("put user input")

	fw, err := r.Controller.GetFlow(ctx, flowID)
	if err != nil {
		return model.ResultTypeError, err
	}

	if err := fw.PutInput(ctx, input); err != nil {
		return model.ResultTypeError, err
	}

	return model.ResultTypeSuccess, nil
}

// FinishFlow is the resolver for the finishFlow field.
func (r *mutationResolver) FinishFlow(ctx context.Context, flowID int64) (*model.Flow, error) {
	uid, err := validatePermissionWithFlowID(ctx, "flows.edit", flowID, r.DB)
	if err != nil {
		return nil, err
	}

	r.Logger.WithFields(logrus.Fields{
		"uid":  uid,
		"flow": flowID,
	}).Debug("finish flow")

	fw, err := r.Controller.GetFlow(ctx, flowID)
	if err != nil {
		return nil, err
	}

	if err := fw.Finish(ctx); err != nil {
		return nil, err
	}

	flow, err := r.DB.GetFlow(ctx, fw.GetFlowID())
	if err != nil {
		return nil, err
	}

	containers, err := r.DB.GetUserFlowContainers(ctx, database.GetUserFlowContainersParams{
		FlowID: flow.ID,
		UserID: uid,
	})
	if err != nil {
		return nil, err
	}

	r.Subscriptions.NewFlowPublisher(flow.UserID, flow.ID).FlowUpdated(ctx, flow, containers)

	return converter.ConvertFlow(flow, containers), nil
}

// DeleteFlow is the resolver for the deleteFlow field.
func (r *mutationResolver) DeleteFlow(ctx context.Context, flowID int64) (model.ResultType, error) {
	uid, err := validatePermissionWithFlowID(ctx, "flows.edit", flowID, r.DB)
	if err != nil {
		return model.ResultTypeError, err
	}

	r.Logger.WithFields(logrus.Fields{
		"uid":  uid,
		"flow": flowID,
	}).Debug("finish flow")

	if fw, err := r.Controller.GetFlow(ctx, flowID); err == nil {
		if err := fw.Finish(ctx); err != nil {
			return model.ResultTypeError, err
		}
	} else if !errors.Is(err, controller.ErrFlowNotFound) {
		return model.ResultTypeError, err
	}

	flow, err := r.DB.GetFlow(ctx, flowID)
	if err != nil {
		return model.ResultTypeError, err
	}

	containers, err := r.DB.GetUserFlowContainers(ctx, database.GetUserFlowContainersParams{
		FlowID: flow.ID,
		UserID: uid,
	})
	if err != nil {
		return model.ResultTypeError, err
	}

	if _, err := r.DB.DeleteFlow(ctx, flowID); err != nil {
		return model.ResultTypeError, err
	}

	publisher := r.Subscriptions.NewFlowPublisher(flow.UserID, flow.ID)
	publisher.FlowUpdated(ctx, flow, containers)
	publisher.FlowDeleted(ctx, flow, containers)

	return model.ResultTypeSuccess, nil
}

// UpdatePrompt is the resolver for the updatePrompt field.
func (r *mutationResolver) UpdatePrompt(ctx context.Context, promptType string, prompt string) (model.ResultType, error) {
	uid, _, err := validatePermission(ctx, "prompts.edit")
	if err != nil {
		return model.ResultTypeError, err
	}

	r.Logger.WithFields(logrus.Fields{
		"uid":        uid,
		"promptType": promptType,
	}).Debug("get prompt")

	_, err = r.DB.UpdateUserTypePrompt(ctx, database.UpdateUserTypePromptParams{
		Prompt: prompt,
		Type:   promptType,
		UserID: uid,
	})
	if err != nil {
		return model.ResultTypeError, err
	}

	return model.ResultTypeSuccess, nil
}

// ResetPrompt is the resolver for the resetPrompt field.
func (r *mutationResolver) ResetPrompt(ctx context.Context, promptType string) (model.ResultType, error) {
	uid, _, err := validatePermission(ctx, "prompts.edit")
	if err != nil {
		return model.ResultTypeError, err
	}

	r.Logger.WithFields(logrus.Fields{
		"uid":        uid,
		"promptType": promptType,
	}).Debug("get prompt")

	prompt, err := r.DefaultPrompter.GetTemplate(templates.PromptType(promptType))
	if err != nil {
		return model.ResultTypeError, err
	}

	_, err = r.DB.UpdateUserTypePrompt(ctx, database.UpdateUserTypePromptParams{
		Prompt: prompt,
		Type:   promptType,
		UserID: uid,
	})
	if err != nil {
		return model.ResultTypeError, err
	}

	return model.ResultTypeSuccess, nil
}

// Providers is the resolver for the providers field.
func (r *queryResolver) Providers(ctx context.Context) ([]string, error) {
	_, _, err := validatePermission(ctx, "providers.view")
	if err != nil {
		return nil, err
	}

	return r.ProvidersCtrl.ListStrings(), nil
}

// Prompts is the resolver for the prompts field.
func (r *queryResolver) Prompts(ctx context.Context) ([]*model.Prompt, error) {
	uid, _, err := validatePermission(ctx, "prompts.view")
	if err != nil {
		return nil, err
	}

	r.Logger.WithFields(logrus.Fields{
		"uid": uid,
	}).Debug("get prompts")

	prompts, err := r.DB.GetUserPrompts(ctx, uid)
	if err != nil {
		return nil, err
	}

	return converter.ConvertPrompts(prompts), nil
}

// Prompt is the resolver for the prompt field.
func (r *queryResolver) Prompt(ctx context.Context, promptType string) (string, error) {
	uid, _, err := validatePermission(ctx, "prompts.view")
	if err != nil {
		return "", err
	}

	r.Logger.WithFields(logrus.Fields{
		"uid":        uid,
		"promptType": promptType,
	}).Debug("get prompt")

	prompt, err := r.DB.GetUserTypePrompt(ctx, database.GetUserTypePromptParams{
		Type:   promptType,
		UserID: uid,
	})
	if err != nil {
		return "", err
	}

	return prompt.Prompt, nil
}

// Flows is the resolver for the flows field.
func (r *queryResolver) Flows(ctx context.Context) ([]*model.Flow, error) {
	uid, admin, err := validatePermission(ctx, "flows.view")
	if err != nil {
		return nil, err
	}

	r.Logger.WithFields(logrus.Fields{
		"uid": uid,
	}).Debug("get term logs")

	var (
		flows      []database.Flow
		containers []database.Container
	)

	if admin {
		flows, err = r.DB.GetFlows(ctx)
	} else {
		flows, err = r.DB.GetUserFlows(ctx, uid)
	}
	if err != nil {
		return nil, err
	}

	_, admin, err = validatePermission(ctx, "containers.view")
	if err == nil {
		if admin {
			containers, err = r.DB.GetContainers(ctx)
		} else {
			containers, err = r.DB.GetUserContainers(ctx, uid)
		}
	}

	return converter.ConvertFlows(flows, containers), nil
}

// Flow is the resolver for the flow field.
func (r *queryResolver) Flow(ctx context.Context, flowID int64) (*model.Flow, error) {
	uid, admin, err := validatePermission(ctx, "flows.view")
	if err != nil {
		return nil, err
	}

	r.Logger.WithFields(logrus.Fields{
		"uid": uid,
	}).Debug("get flow")

	var (
		flow       database.Flow
		containers []database.Container
	)

	if admin {
		flow, err = r.DB.GetFlow(ctx, flowID)
	} else {
		flow, err = r.DB.GetUserFlow(ctx, database.GetUserFlowParams{
			ID:     flowID,
			UserID: uid,
		})
	}
	if err != nil {
		return nil, err
	}

	_, admin, err = validatePermission(ctx, "containers.view")
	if err == nil {
		if admin {
			containers, err = r.DB.GetFlowContainers(ctx, flowID)
		} else {
			containers, err = r.DB.GetUserFlowContainers(ctx, database.GetUserFlowContainersParams{
				FlowID: flow.ID,
				UserID: uid,
			})
		}
	}

	return converter.ConvertFlow(flow, containers), nil
}

// Tasks is the resolver for the tasks field.
func (r *queryResolver) Tasks(ctx context.Context, flowID int64) ([]*model.Task, error) {
	uid, err := validatePermissionWithFlowID(ctx, "tasks.view", flowID, r.DB)
	if err != nil {
		return nil, err
	}

	r.Logger.WithFields(logrus.Fields{
		"uid":  uid,
		"flow": flowID,
	}).Debug("get tasks")

	tasks, err := r.DB.GetFlowTasks(ctx, flowID)
	if err != nil {
		return nil, err
	}

	var subtasks []database.Subtask
	_, admin, err := validatePermission(ctx, "subtasks.view")
	if err == nil {
		if admin {
			subtasks, err = r.DB.GetFlowSubtasks(ctx, flowID)
		} else {
			subtasks, err = r.DB.GetUserFlowSubtasks(ctx, database.GetUserFlowSubtasksParams{
				FlowID: flowID,
				UserID: uid,
			})
		}
	}

	return converter.ConvertTasks(tasks, subtasks), nil
}

// Screenshots is the resolver for the screenshots field.
func (r *queryResolver) Screenshots(ctx context.Context, flowID int64) ([]*model.Screenshot, error) {
	uid, err := validatePermissionWithFlowID(ctx, "screenshots.view", flowID, r.DB)
	if err != nil {
		return nil, err
	}

	r.Logger.WithFields(logrus.Fields{
		"uid":  uid,
		"flow": flowID,
	}).Debug("get screenshots")

	screenshots, err := r.DB.GetFlowScreenshots(ctx, flowID)
	if err != nil {
		return nil, err
	}

	return converter.ConvertScreenshots(screenshots), nil
}

// TerminalLogs is the resolver for the terminalLogs field.
func (r *queryResolver) TerminalLogs(ctx context.Context, flowID int64) ([]*model.TerminalLog, error) {
	uid, err := validatePermissionWithFlowID(ctx, "termlogs.view", flowID, r.DB)
	if err != nil {
		return nil, err
	}

	r.Logger.WithFields(logrus.Fields{
		"uid":  uid,
		"flow": flowID,
	}).Debug("get term logs")

	logs, err := r.DB.GetFlowTermLogs(ctx, flowID)
	if err != nil {
		return nil, err
	}

	return converter.ConvertTerminalLogs(logs, flowID), nil
}

// MessageLogs is the resolver for the messageLogs field.
func (r *queryResolver) MessageLogs(ctx context.Context, flowID int64) ([]*model.MessageLog, error) {
	uid, err := validatePermissionWithFlowID(ctx, "msglogs.view", flowID, r.DB)
	if err != nil {
		return nil, err
	}

	r.Logger.WithFields(logrus.Fields{
		"uid":  uid,
		"flow": flowID,
	}).Debug("get msg logs")

	logs, err := r.DB.GetFlowMsgLogs(ctx, flowID)
	if err != nil {
		return nil, err
	}

	return converter.ConvertMessageLogs(logs), nil
}

// AgentLogs is the resolver for the agentLogs field.
func (r *queryResolver) AgentLogs(ctx context.Context, flowID int64) ([]*model.AgentLog, error) {
	uid, err := validatePermissionWithFlowID(ctx, "agentlogs.view", flowID, r.DB)
	if err != nil {
		return nil, err
	}

	r.Logger.WithFields(logrus.Fields{
		"uid":  uid,
		"flow": flowID,
	}).Debug("get agent logs")

	logs, err := r.DB.GetFlowAgentLogs(ctx, flowID)
	if err != nil {
		return nil, err
	}

	return converter.ConvertAgentLogs(logs), nil
}

// SearchLogs is the resolver for the searchLogs field.
func (r *queryResolver) SearchLogs(ctx context.Context, flowID int64) ([]*model.SearchLog, error) {
	uid, err := validatePermissionWithFlowID(ctx, "searchlogs.view", flowID, r.DB)
	if err != nil {
		return nil, err
	}

	r.Logger.WithFields(logrus.Fields{
		"uid":  uid,
		"flow": flowID,
	}).Debug("get search logs")

	logs, err := r.DB.GetFlowSearchLogs(ctx, flowID)
	if err != nil {
		return nil, err
	}

	return converter.ConvertSearchLogs(logs), nil
}

// VectorStoreLogs is the resolver for the vectorStoreLogs field.
func (r *queryResolver) VectorStoreLogs(ctx context.Context, flowID int64) ([]*model.VectorStoreLog, error) {
	uid, err := validatePermissionWithFlowID(ctx, "vecstorelogs.view", flowID, r.DB)
	if err != nil {
		return nil, err
	}

	r.Logger.WithFields(logrus.Fields{
		"uid":  uid,
		"flow": flowID,
	}).Debug("get vector store logs")

	logs, err := r.DB.GetFlowVectorStoreLogs(ctx, flowID)
	if err != nil {
		return nil, err
	}

	return converter.ConvertVectorStoreLogs(logs), nil
}

// FlowCreated is the resolver for the flowCreated field.
func (r *subscriptionResolver) FlowCreated(ctx context.Context) (<-chan *model.Flow, error) {
	uid, admin, err := validatePermission(ctx, "flows.subscribe")
	if err != nil {
		return nil, err
	}

	subscriber := r.Subscriptions.NewFlowSubscriber(uid, 0)
	if admin {
		return subscriber.FlowCreatedAdmin(ctx)
	}

	return subscriber.FlowCreated(ctx)
}

// FlowDeleted is the resolver for the flowDeleted field.
func (r *subscriptionResolver) FlowDeleted(ctx context.Context) (<-chan *model.Flow, error) {
	uid, admin, err := validatePermission(ctx, "flows.subscribe")
	if err != nil {
		return nil, err
	}

	subscriber := r.Subscriptions.NewFlowSubscriber(uid, 0)
	if admin {
		return subscriber.FlowDeletedAdmin(ctx)
	}

	return subscriber.FlowDeleted(ctx)
}

// FlowUpdated is the resolver for the flowUpdated field.
func (r *subscriptionResolver) FlowUpdated(ctx context.Context, flowID int64) (<-chan *model.Flow, error) {
	uid, err := validatePermissionWithFlowID(ctx, "flows.subscribe", flowID, r.DB)
	if err != nil {
		return nil, err
	}

	return r.Subscriptions.NewFlowSubscriber(uid, flowID).FlowUpdated(ctx)
}

// TaskCreated is the resolver for the taskCreated field.
func (r *subscriptionResolver) TaskCreated(ctx context.Context, flowID int64) (<-chan *model.Task, error) {
	uid, err := validatePermissionWithFlowID(ctx, "tasks.subscribe", flowID, r.DB)
	if err != nil {
		return nil, err
	}

	return r.Subscriptions.NewFlowSubscriber(uid, flowID).TaskCreated(ctx)
}

// TaskUpdated is the resolver for the taskUpdated field.
func (r *subscriptionResolver) TaskUpdated(ctx context.Context, flowID int64) (<-chan *model.Task, error) {
	uid, err := validatePermissionWithFlowID(ctx, "tasks.subscribe", flowID, r.DB)
	if err != nil {
		return nil, err
	}

	return r.Subscriptions.NewFlowSubscriber(uid, flowID).TaskUpdated(ctx)
}

// ScreenshotAdded is the resolver for the screenshotAdded field.
func (r *subscriptionResolver) ScreenshotAdded(ctx context.Context, flowID int64) (<-chan *model.Screenshot, error) {
	uid, err := validatePermissionWithFlowID(ctx, "screenshots.subscribe", flowID, r.DB)
	if err != nil {
		return nil, err
	}

	return r.Subscriptions.NewFlowSubscriber(uid, flowID).ScreenshotAdded(ctx)
}

// TerminalLogAdded is the resolver for the terminalLogAdded field.
func (r *subscriptionResolver) TerminalLogAdded(ctx context.Context, flowID int64) (<-chan *model.TerminalLog, error) {
	uid, err := validatePermissionWithFlowID(ctx, "termlogs.subscribe", flowID, r.DB)
	if err != nil {
		return nil, err
	}

	return r.Subscriptions.NewFlowSubscriber(uid, flowID).TerminalLogAdded(ctx)
}

// MessageLogAdded is the resolver for the messageLogAdded field.
func (r *subscriptionResolver) MessageLogAdded(ctx context.Context, flowID int64) (<-chan *model.MessageLog, error) {
	uid, err := validatePermissionWithFlowID(ctx, "msglogs.subscribe", flowID, r.DB)
	if err != nil {
		return nil, err
	}

	return r.Subscriptions.NewFlowSubscriber(uid, flowID).MessageLogAdded(ctx)
}

// MessageLogUpdated is the resolver for the messageLogUpdated field.
func (r *subscriptionResolver) MessageLogUpdated(ctx context.Context, flowID int64) (<-chan *model.MessageLog, error) {
	uid, err := validatePermissionWithFlowID(ctx, "msglogs.subscribe", flowID, r.DB)
	if err != nil {
		return nil, err
	}

	return r.Subscriptions.NewFlowSubscriber(uid, flowID).MessageLogUpdated(ctx)
}

// AgentLogAdded is the resolver for the agentLogAdded field.
func (r *subscriptionResolver) AgentLogAdded(ctx context.Context, flowID int64) (<-chan *model.AgentLog, error) {
	uid, err := validatePermissionWithFlowID(ctx, "agentlogs.subscribe", flowID, r.DB)
	if err != nil {
		return nil, err
	}

	return r.Subscriptions.NewFlowSubscriber(uid, flowID).AgentLogAdded(ctx)
}

// SearchLogAdded is the resolver for the searchLogAdded field.
func (r *subscriptionResolver) SearchLogAdded(ctx context.Context, flowID int64) (<-chan *model.SearchLog, error) {
	uid, err := validatePermissionWithFlowID(ctx, "searchlogs.subscribe", flowID, r.DB)
	if err != nil {
		return nil, err
	}

	return r.Subscriptions.NewFlowSubscriber(uid, flowID).SearchLogAdded(ctx)
}

// VectorStoreLogAdded is the resolver for the vectorStoreLogAdded field.
func (r *subscriptionResolver) VectorStoreLogAdded(ctx context.Context, flowID int64) (<-chan *model.VectorStoreLog, error) {
	uid, err := validatePermissionWithFlowID(ctx, "vecstorelogs.subscribe", flowID, r.DB)
	if err != nil {
		return nil, err
	}

	return r.Subscriptions.NewFlowSubscriber(uid, flowID).VectorStoreLogAdded(ctx)
}

// Mutation returns MutationResolver implementation.
func (r *Resolver) Mutation() MutationResolver { return &mutationResolver{r} }

// Query returns QueryResolver implementation.
func (r *Resolver) Query() QueryResolver { return &queryResolver{r} }

// Subscription returns SubscriptionResolver implementation.
func (r *Resolver) Subscription() SubscriptionResolver { return &subscriptionResolver{r} }

type mutationResolver struct{ *Resolver }
type queryResolver struct{ *Resolver }
type subscriptionResolver struct{ *Resolver }
