package subscriptions

import (
	"context"
	"sync"
	"time"

	"pentagi/pkg/database"
	"pentagi/pkg/graph/model"
	"pentagi/pkg/providers/pconfig"
)

const (
	defChannelLen  = 50
	defSendTimeout = 5 * time.Second
)

type SubscriptionsController interface {
	NewFlowSubscriber(userID, flowID int64) FlowSubscriber
	NewFlowPublisher(userID, flowID int64) FlowPublisher
	NewResourceSubscriber(userID int64) ResourceSubscriber
	NewResourcePublisher(userID int64) ResourcePublisher
	NewProviderSubscriber(userID int64) ProviderSubscriber
	NewProviderPublisher(userID int64) ProviderPublisher
	NewAPITokenSubscriber(userID int64) APITokenSubscriber
	NewAPITokenPublisher(userID int64) APITokenPublisher
	NewSettingsSubscriber(userID int64) SettingsSubscriber
	NewSettingsPublisher(userID int64) SettingsPublisher
	NewFlowTemplateSubscriber(userID int64) FlowTemplateSubscriber
	NewFlowTemplatePublisher(userID int64) FlowTemplatePublisher
}

type UserContext interface {
	GetUserID() int64
	SetUserID(userID int64)
}

type FlowContext interface {
	GetFlowID() int64
	SetFlowID(flowID int64)
}

type FlowSubscriber interface {
	FlowCreatedAdmin(ctx context.Context) (<-chan *model.Flow, error)
	FlowCreated(ctx context.Context) (<-chan *model.Flow, error)
	FlowDeletedAdmin(ctx context.Context) (<-chan *model.Flow, error)
	FlowDeleted(ctx context.Context) (<-chan *model.Flow, error)
	FlowUpdatedAdmin(ctx context.Context) (<-chan *model.Flow, error)
	FlowUpdated(ctx context.Context) (<-chan *model.Flow, error)
	TaskCreated(ctx context.Context) (<-chan *model.Task, error)
	TaskUpdated(ctx context.Context) (<-chan *model.Task, error)
	AssistantCreated(ctx context.Context) (<-chan *model.Assistant, error)
	AssistantUpdated(ctx context.Context) (<-chan *model.Assistant, error)
	AssistantDeleted(ctx context.Context) (<-chan *model.Assistant, error)
	FlowFileAdded(ctx context.Context) (<-chan *model.FlowFile, error)
	FlowFileUpdated(ctx context.Context) (<-chan *model.FlowFile, error)
	FlowFileDeleted(ctx context.Context) (<-chan *model.FlowFile, error)
	ScreenshotAdded(ctx context.Context) (<-chan *model.Screenshot, error)
	TerminalLogAdded(ctx context.Context) (<-chan *model.TerminalLog, error)
	MessageLogAdded(ctx context.Context) (<-chan *model.MessageLog, error)
	MessageLogUpdated(ctx context.Context) (<-chan *model.MessageLog, error)
	AgentLogAdded(ctx context.Context) (<-chan *model.AgentLog, error)
	SearchLogAdded(ctx context.Context) (<-chan *model.SearchLog, error)
	VectorStoreLogAdded(ctx context.Context) (<-chan *model.VectorStoreLog, error)
	AssistantLogAdded(ctx context.Context) (<-chan *model.AssistantLog, error)
	AssistantLogUpdated(ctx context.Context) (<-chan *model.AssistantLog, error)
	FlowContext
	UserContext
}

type ProviderSubscriber interface {
	ProviderCreated(ctx context.Context) (<-chan *model.ProviderConfig, error)
	ProviderUpdated(ctx context.Context) (<-chan *model.ProviderConfig, error)
	ProviderDeleted(ctx context.Context) (<-chan *model.ProviderConfig, error)
	UserContext
}

type APITokenSubscriber interface {
	APITokenCreated(ctx context.Context) (<-chan *model.APIToken, error)
	APITokenUpdated(ctx context.Context) (<-chan *model.APIToken, error)
	APITokenDeleted(ctx context.Context) (<-chan *model.APIToken, error)
	UserContext
}

type SettingsSubscriber interface {
	SettingsUserUpdated(ctx context.Context) (<-chan *model.UserPreferences, error)
	UserContext
}

type FlowTemplateSubscriber interface {
	FlowTemplateCreated(ctx context.Context) (<-chan *model.FlowTemplate, error)
	FlowTemplateUpdated(ctx context.Context) (<-chan *model.FlowTemplate, error)
	FlowTemplateDeleted(ctx context.Context) (<-chan *model.FlowTemplate, error)
	UserContext
}

type ResourceSubscriber interface {
	ResourceAdded(ctx context.Context) (<-chan *model.UserResource, error)
	ResourceUpdated(ctx context.Context) (<-chan *model.UserResource, error)
	ResourceDeleted(ctx context.Context) (<-chan *model.UserResource, error)
	ResourceAddedAdmin(ctx context.Context) (<-chan *model.UserResource, error)
	ResourceUpdatedAdmin(ctx context.Context) (<-chan *model.UserResource, error)
	ResourceDeletedAdmin(ctx context.Context) (<-chan *model.UserResource, error)
	UserContext
}

type FlowPublisher interface {
	FlowCreated(ctx context.Context, flow database.Flow, terms []database.Container)
	FlowDeleted(ctx context.Context, flow database.Flow, terms []database.Container)
	FlowUpdated(ctx context.Context, flow database.Flow, terms []database.Container)
	TaskCreated(ctx context.Context, task database.Task, subtasks []database.Subtask)
	TaskUpdated(ctx context.Context, task database.Task, subtasks []database.Subtask)
	AssistantCreated(ctx context.Context, assistant database.Assistant)
	AssistantUpdated(ctx context.Context, assistant database.Assistant)
	AssistantDeleted(ctx context.Context, assistant database.Assistant)
	FlowFileAdded(ctx context.Context, file *model.FlowFile)
	FlowFileUpdated(ctx context.Context, file *model.FlowFile)
	FlowFileDeleted(ctx context.Context, file *model.FlowFile)
	ScreenshotAdded(ctx context.Context, screenshot database.Screenshot)
	TerminalLogAdded(ctx context.Context, terminalLog database.Termlog)
	MessageLogAdded(ctx context.Context, messageLog database.Msglog)
	MessageLogUpdated(ctx context.Context, messageLog database.Msglog)
	AgentLogAdded(ctx context.Context, agentLog database.Agentlog)
	SearchLogAdded(ctx context.Context, searchLog database.Searchlog)
	VectorStoreLogAdded(ctx context.Context, vectorStoreLog database.Vecstorelog)
	AssistantLogAdded(ctx context.Context, assistantLog database.Assistantlog)
	AssistantLogUpdated(ctx context.Context, assistantLog database.Assistantlog, appendPart bool)
	FlowContext
	UserContext
}

type ProviderPublisher interface {
	ProviderCreated(ctx context.Context, provider database.Provider, cfg *pconfig.ProviderConfig)
	ProviderUpdated(ctx context.Context, provider database.Provider, cfg *pconfig.ProviderConfig)
	ProviderDeleted(ctx context.Context, provider database.Provider, cfg *pconfig.ProviderConfig)
	UserContext
}

type APITokenPublisher interface {
	APITokenCreated(ctx context.Context, apiToken database.APITokenWithSecret)
	APITokenUpdated(ctx context.Context, apiToken database.ApiToken)
	APITokenDeleted(ctx context.Context, apiToken database.ApiToken)
	UserContext
}

type SettingsPublisher interface {
	SettingsUserUpdated(ctx context.Context, userPreferences database.UserPreference)
	UserContext
}

type FlowTemplatePublisher interface {
	FlowTemplateCreated(ctx context.Context, template database.FlowTemplate)
	FlowTemplateUpdated(ctx context.Context, template database.FlowTemplate)
	FlowTemplateDeleted(ctx context.Context, template database.FlowTemplate)
	UserContext
}

type ResourcePublisher interface {
	ResourceAdded(ctx context.Context, resource *model.UserResource)
	ResourceUpdated(ctx context.Context, resource *model.UserResource)
	ResourceDeleted(ctx context.Context, resource *model.UserResource)
	UserContext
}

type controller struct {
	flowCreatedAdmin     Channel[*model.Flow]
	flowCreated          Channel[*model.Flow]
	flowDeletedAdmin     Channel[*model.Flow]
	flowDeleted          Channel[*model.Flow]
	flowUpdatedAdmin     Channel[*model.Flow]
	flowUpdated          Channel[*model.Flow]
	taskCreated          Channel[*model.Task]
	taskUpdated          Channel[*model.Task]
	assistantCreated     Channel[*model.Assistant]
	assistantUpdated     Channel[*model.Assistant]
	assistantDeleted     Channel[*model.Assistant]
	flowFileAdded        Channel[*model.FlowFile]
	flowFileUpdated      Channel[*model.FlowFile]
	flowFileDeleted      Channel[*model.FlowFile]
	screenshotAdded      Channel[*model.Screenshot]
	terminalLogAdded     Channel[*model.TerminalLog]
	messageLogAdded      Channel[*model.MessageLog]
	messageLogUpdated    Channel[*model.MessageLog]
	agentLogAdded        Channel[*model.AgentLog]
	searchLogAdded       Channel[*model.SearchLog]
	vecStoreLogAdded     Channel[*model.VectorStoreLog]
	assistantLogAdded    Channel[*model.AssistantLog]
	assistantLogUpdated  Channel[*model.AssistantLog]
	providerCreated      Channel[*model.ProviderConfig]
	providerUpdated      Channel[*model.ProviderConfig]
	providerDeleted      Channel[*model.ProviderConfig]
	apiTokenCreated      Channel[*model.APIToken]
	apiTokenUpdated      Channel[*model.APIToken]
	apiTokenDeleted      Channel[*model.APIToken]
	settingsUserUpdated  Channel[*model.UserPreferences]
	flowTemplateCreated  Channel[*model.FlowTemplate]
	flowTemplateUpdated  Channel[*model.FlowTemplate]
	flowTemplateDeleted  Channel[*model.FlowTemplate]
	resourceAdded        Channel[*model.UserResource]
	resourceUpdated      Channel[*model.UserResource]
	resourceDeleted      Channel[*model.UserResource]
	resourceAddedAdmin   Channel[*model.UserResource]
	resourceUpdatedAdmin Channel[*model.UserResource]
	resourceDeletedAdmin Channel[*model.UserResource]
}

func NewSubscriptionsController() SubscriptionsController {
	return &controller{
		flowCreatedAdmin:     NewChannel[*model.Flow](),
		flowCreated:          NewChannel[*model.Flow](),
		flowDeletedAdmin:     NewChannel[*model.Flow](),
		flowDeleted:          NewChannel[*model.Flow](),
		flowUpdatedAdmin:     NewChannel[*model.Flow](),
		flowUpdated:          NewChannel[*model.Flow](),
		taskCreated:          NewChannel[*model.Task](),
		taskUpdated:          NewChannel[*model.Task](),
		assistantCreated:     NewChannel[*model.Assistant](),
		assistantUpdated:     NewChannel[*model.Assistant](),
		assistantDeleted:     NewChannel[*model.Assistant](),
		flowFileAdded:        NewChannel[*model.FlowFile](),
		flowFileUpdated:      NewChannel[*model.FlowFile](),
		flowFileDeleted:      NewChannel[*model.FlowFile](),
		screenshotAdded:      NewChannel[*model.Screenshot](),
		terminalLogAdded:     NewChannel[*model.TerminalLog](),
		messageLogAdded:      NewChannel[*model.MessageLog](),
		messageLogUpdated:    NewChannel[*model.MessageLog](),
		agentLogAdded:        NewChannel[*model.AgentLog](),
		searchLogAdded:       NewChannel[*model.SearchLog](),
		vecStoreLogAdded:     NewChannel[*model.VectorStoreLog](),
		assistantLogAdded:    NewChannel[*model.AssistantLog](),
		assistantLogUpdated:  NewChannel[*model.AssistantLog](),
		providerCreated:      NewChannel[*model.ProviderConfig](),
		providerUpdated:      NewChannel[*model.ProviderConfig](),
		providerDeleted:      NewChannel[*model.ProviderConfig](),
		apiTokenCreated:      NewChannel[*model.APIToken](),
		apiTokenUpdated:      NewChannel[*model.APIToken](),
		apiTokenDeleted:      NewChannel[*model.APIToken](),
		settingsUserUpdated:  NewChannel[*model.UserPreferences](),
		flowTemplateCreated:  NewChannel[*model.FlowTemplate](),
		flowTemplateUpdated:  NewChannel[*model.FlowTemplate](),
		flowTemplateDeleted:  NewChannel[*model.FlowTemplate](),
		resourceAdded:        NewChannel[*model.UserResource](),
		resourceUpdated:      NewChannel[*model.UserResource](),
		resourceDeleted:      NewChannel[*model.UserResource](),
		resourceAddedAdmin:   NewChannel[*model.UserResource](),
		resourceUpdatedAdmin: NewChannel[*model.UserResource](),
		resourceDeletedAdmin: NewChannel[*model.UserResource](),
	}
}

func (s *controller) NewFlowPublisher(userID, flowID int64) FlowPublisher {
	return &flowPublisher{
		userID: userID,
		flowID: flowID,
		ctrl:   s,
	}
}

func (s *controller) NewFlowSubscriber(userID, flowID int64) FlowSubscriber {
	return &flowSubscriber{
		userID: userID,
		flowID: flowID,
		ctrl:   s,
	}
}

func (s *controller) NewResourcePublisher(userID int64) ResourcePublisher {
	return &resourcePublisher{
		userID: userID,
		ctrl:   s,
	}
}

func (s *controller) NewResourceSubscriber(userID int64) ResourceSubscriber {
	return &resourceSubscriber{
		userID: userID,
		ctrl:   s,
	}
}

func (s *controller) NewProviderPublisher(userID int64) ProviderPublisher {
	return &providerPublisher{userID: userID, ctrl: s}
}

func (s *controller) NewProviderSubscriber(userID int64) ProviderSubscriber {
	return &providerSubscriber{userID: userID, ctrl: s}
}

func (s *controller) NewAPITokenPublisher(userID int64) APITokenPublisher {
	return &apiTokenPublisher{userID: userID, ctrl: s}
}

func (s *controller) NewAPITokenSubscriber(userID int64) APITokenSubscriber {
	return &apiTokenSubscriber{userID: userID, ctrl: s}
}

func (s *controller) NewSettingsPublisher(userID int64) SettingsPublisher {
	return &settingsPublisher{userID: userID, ctrl: s}
}

func (s *controller) NewSettingsSubscriber(userID int64) SettingsSubscriber {
	return &settingsSubscriber{userID: userID, ctrl: s}
}

func (s *controller) NewFlowTemplatePublisher(userID int64) FlowTemplatePublisher {
	return &flowTemplatePublisher{userID: userID, ctrl: s}
}

func (s *controller) NewFlowTemplateSubscriber(userID int64) FlowTemplateSubscriber {
	return &flowTemplateSubscriber{userID: userID, ctrl: s}
}

type Channel[T any] interface {
	Subscribe(ctx context.Context, id int64) <-chan T
	Publish(ctx context.Context, id int64, data T)
	Broadcast(ctx context.Context, data T)
}

func NewChannel[T any]() Channel[T] {
	return &channel[T]{
		mx:   &sync.RWMutex{},
		subs: make(map[int64][]chan T),
	}
}

type channel[T any] struct {
	mx   *sync.RWMutex
	subs map[int64][]chan T
}

func (c *channel[T]) Subscribe(ctx context.Context, id int64) <-chan T {
	c.mx.Lock()
	defer c.mx.Unlock()

	ch := make(chan T, defChannelLen)
	c.subs[id] = append(c.subs[id], ch)

	go func() {
		<-ctx.Done()

		c.mx.Lock()
		defer c.mx.Unlock()

		if subs, ok := c.subs[id]; ok {
			for i, sub := range subs {
				if sub == ch {
					c.subs[id] = append(subs[:i], subs[i+1:]...)
					break
				}
			}
		}

		if len(c.subs[id]) == 0 {
			delete(c.subs, id)
		}

		close(ch)
	}()

	return ch
}

func (c *channel[T]) Publish(ctx context.Context, id int64, data T) {
	c.mx.RLock()
	defer c.mx.RUnlock()

	for _, ch := range c.subs[id] {
		select {
		case ch <- data:
		case <-ctx.Done():
			return
		}
	}
}

func (c *channel[T]) Broadcast(ctx context.Context, data T) {
	c.mx.RLock()
	defer c.mx.RUnlock()

	for _, subs := range c.subs {
		for _, ch := range subs {
			select {
			case ch <- data:
			case <-ctx.Done():
				return
			}
		}
	}
}
