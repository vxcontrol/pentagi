package controller

import (
	"context"
	"errors"
	"slices"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"pentagi/pkg/config"
	"pentagi/pkg/database"
	"pentagi/pkg/docker"
	"pentagi/pkg/graph/subscriptions"
	"pentagi/pkg/graphiti"
	"pentagi/pkg/providers"
	"pentagi/pkg/providers/provider"
	"pentagi/pkg/templates"
	"pentagi/pkg/tools"

	"github.com/sirupsen/logrus"
)

// controllerReceive fails the test instead of hanging when ch stays silent.
func controllerReceive[T any](t *testing.T, ch <-chan T, what string) T {
	t.Helper()

	select {
	case v := <-ch:
		return v
	case <-time.After(10 * time.Second):
		t.Fatalf("timed out waiting for %s", what)
	}

	var zero T

	return zero
}

type noopFlowWorker struct {
	flowID int64
}

func (w *noopFlowWorker) GetFlowID() int64         { return w.flowID }
func (w *noopFlowWorker) GetUserID() int64         { return 1 }
func (w *noopFlowWorker) GetTitle() string         { return "noop" }
func (w *noopFlowWorker) GetContext() *FlowContext { return nil }
func (w *noopFlowWorker) GetStatus(context.Context) (database.FlowStatus, error) {
	return database.FlowStatusRunning, nil
}
func (w *noopFlowWorker) SetStatus(context.Context, database.FlowStatus) error { return nil }
func (w *noopFlowWorker) AddAssistant(context.Context, AssistantWorker) error  { return nil }
func (w *noopFlowWorker) GetAssistant(context.Context, int64) (AssistantWorker, error) {
	return nil, nil
}
func (w *noopFlowWorker) DeleteAssistant(context.Context, int64) error     { return nil }
func (w *noopFlowWorker) ListAssistants(context.Context) []AssistantWorker { return nil }
func (w *noopFlowWorker) ListTasks(context.Context) []TaskWorker           { return nil }
func (w *noopFlowWorker) PutInput(context.Context, string, provider.Provider, []database.UserResource) error {
	return nil
}
func (w *noopFlowWorker) PutResources(context.Context, []database.UserResource) error { return nil }
func (w *noopFlowWorker) Finish(context.Context) error                                { return nil }
func (w *noopFlowWorker) Stop(context.Context) error                                  { return nil }
func (w *noopFlowWorker) Rename(context.Context, string) error                        { return nil }
func (w *noopFlowWorker) Report(context.Context, time.Duration) error                 { return nil }
func (w *noopFlowWorker) WaitTaskCompletion(context.Context) error                    { return nil }
func (w *noopFlowWorker) InvalidateTaskSubtasks(context.Context, int64, []int64)      {}

var _ FlowWorker = (*noopFlowWorker)(nil)

// finishRecordingWorker records its Finish and the context it was handed, and answers err.
type finishRecordingWorker struct {
	noopFlowWorker

	finished atomic.Bool
	seen     chan error
	err      error
}

func (w *finishRecordingWorker) Finish(ctx context.Context) error {
	w.finished.Store(true)
	if w.seen != nil {
		w.seen <- ctx.Err()
	}

	return w.err
}

type cascadeFakePublisher struct {
	subscriptions.FlowPublisher

	mx               sync.Mutex
	flowCreated      []database.Flow
	assistantCreated []database.Assistant
	flowUpdated      []database.Flow
	flowTerms        [][]database.Container
	assistantUpdated []database.Assistant
	taskUpdated      []database.Task
	taskSubtasks     [][]database.Subtask
}

func (p *cascadeFakePublisher) FlowCreated(_ context.Context, flow database.Flow, _ []database.Container) {
	p.mx.Lock()
	defer p.mx.Unlock()
	p.flowCreated = append(p.flowCreated, flow)
}

func (p *cascadeFakePublisher) created() []database.Flow {
	p.mx.Lock()
	defer p.mx.Unlock()

	return slices.Clone(p.flowCreated)
}

func (p *cascadeFakePublisher) MessageLogAdded(context.Context, database.Msglog)               {}
func (p *cascadeFakePublisher) TerminalLogAdded(context.Context, database.Termlog)             {}
func (p *cascadeFakePublisher) TaskCreated(context.Context, database.Task, []database.Subtask) {}

func (p *cascadeFakePublisher) AssistantCreated(_ context.Context, assistant database.Assistant) {
	p.mx.Lock()
	defer p.mx.Unlock()
	p.assistantCreated = append(p.assistantCreated, assistant)
}

func (p *cascadeFakePublisher) assistants() ([]database.Assistant, []database.Assistant) {
	p.mx.Lock()
	defer p.mx.Unlock()

	return slices.Clone(p.assistantCreated), slices.Clone(p.assistantUpdated)
}

func (p *cascadeFakePublisher) FlowUpdated(_ context.Context, flow database.Flow, terms []database.Container) {
	p.mx.Lock()
	defer p.mx.Unlock()
	p.flowUpdated = append(p.flowUpdated, flow)
	p.flowTerms = append(p.flowTerms, terms)
}

func (p *cascadeFakePublisher) TaskUpdated(_ context.Context, task database.Task, subtasks []database.Subtask) {
	p.mx.Lock()
	defer p.mx.Unlock()
	p.taskUpdated = append(p.taskUpdated, task)
	p.taskSubtasks = append(p.taskSubtasks, subtasks)
}

func (p *cascadeFakePublisher) AssistantUpdated(_ context.Context, assistant database.Assistant) {
	p.mx.Lock()
	defer p.mx.Unlock()
	p.assistantUpdated = append(p.assistantUpdated, assistant)
}

// cascadeFakeSubscriptions hands every flow the same publisher, so a test reads one place.
type cascadeFakeSubscriptions struct {
	subscriptions.SubscriptionsController

	pub *cascadeFakePublisher
}

func (s *cascadeFakeSubscriptions) NewFlowPublisher(int64, int64) subscriptions.FlowPublisher {
	return s.pub
}

// finishFakeQuerier fails its calls once the context is done, the way pgx does.
type finishFakeQuerier struct {
	database.Querier

	flow     database.Flow
	flowErr  error
	tasks    []database.Task
	tasksErr error
	subtasks map[int64][]database.Subtask

	assistants        []database.Assistant
	containers        []database.Container
	containersErr     error
	containersErrOnce bool

	flowStatus         []database.UpdateFlowStatusParams
	taskStatus         []database.UpdateTaskStatusParams
	subtaskStatus      []database.UpdateSubtaskStatusParams
	assistantStatus    []database.UpdateAssistantStatusParams
	assistantStatusErr error
}

func (f *finishFakeQuerier) GetFlow(ctx context.Context, _ int64) (database.Flow, error) {
	if err := ctx.Err(); err != nil {
		return database.Flow{}, err
	}
	if f.flowErr != nil {
		return database.Flow{}, f.flowErr
	}

	return f.flow, nil
}

func (f *finishFakeQuerier) GetFlowTasks(ctx context.Context, _ int64) ([]database.Task, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if f.tasksErr != nil {
		return nil, f.tasksErr
	}

	return f.tasks, nil
}

func (f *finishFakeQuerier) GetTaskSubtasks(ctx context.Context, taskID int64) ([]database.Subtask, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	return slices.Clone(f.subtasks[taskID]), nil
}

func (f *finishFakeQuerier) UpdateFlowStatus(
	ctx context.Context, arg database.UpdateFlowStatusParams,
) (database.Flow, error) {
	if err := ctx.Err(); err != nil {
		return database.Flow{}, err
	}
	f.flowStatus = append(f.flowStatus, arg)
	updated := f.flow
	updated.Status = arg.Status

	return updated, nil
}

func (f *finishFakeQuerier) UpdateTaskStatus(
	ctx context.Context, arg database.UpdateTaskStatusParams,
) (database.Task, error) {
	if err := ctx.Err(); err != nil {
		return database.Task{}, err
	}
	f.taskStatus = append(f.taskStatus, arg)

	return database.Task{ID: arg.ID, Status: arg.Status}, nil
}

func (f *finishFakeQuerier) UpdateSubtaskStatus(
	ctx context.Context, arg database.UpdateSubtaskStatusParams,
) (database.Subtask, error) {
	if err := ctx.Err(); err != nil {
		return database.Subtask{}, err
	}
	f.subtaskStatus = append(f.subtaskStatus, arg)
	for _, subtasks := range f.subtasks {
		for i := range subtasks {
			if subtasks[i].ID == arg.ID {
				subtasks[i].Status = arg.Status
			}
		}
	}

	return database.Subtask{ID: arg.ID, Status: arg.Status}, nil
}

func (f *finishFakeQuerier) GetFlowContainers(ctx context.Context, _ int64) ([]database.Container, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if err := f.containersErr; err != nil {
		if f.containersErrOnce {
			f.containersErr = nil
		}
		return nil, err
	}

	return f.containers, nil
}

func (f *finishFakeQuerier) GetFlowAssistants(ctx context.Context, _ int64) ([]database.Assistant, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	return f.assistants, nil
}

func (f *finishFakeQuerier) CreateAssistant(
	ctx context.Context, arg database.CreateAssistantParams,
) (database.Assistant, error) {
	if err := ctx.Err(); err != nil {
		return database.Assistant{}, err
	}

	return database.Assistant{ID: reservedAssistantID, FlowID: arg.FlowID, Status: arg.Status}, nil
}

func (f *finishFakeQuerier) UpdateAssistantStatus(
	ctx context.Context, arg database.UpdateAssistantStatusParams,
) (database.Assistant, error) {
	if err := ctx.Err(); err != nil {
		return database.Assistant{}, err
	}
	f.assistantStatus = append(f.assistantStatus, arg)
	if f.assistantStatusErr != nil {
		return database.Assistant{}, f.assistantStatusErr
	}

	return database.Assistant{ID: arg.ID, Status: arg.Status}, nil
}

type finishFakeDocker struct {
	docker.DockerClient

	removed []string
	err     error
}

func (d *finishFakeDocker) RemoveContainer(ctx context.Context, containerID string, _ int64) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if d.err != nil {
		return d.err
	}
	d.removed = append(d.removed, containerID)

	return nil
}

func newFinishController(q database.Querier) (*flowController, *cascadeFakePublisher, *finishFakeDocker) {
	pub := &cascadeFakePublisher{}
	dkr := &finishFakeDocker{}

	return &flowController{
		db:     q,
		mx:     &sync.Mutex{},
		flows:  map[int64]*flowEntry{},
		subs:   &cascadeFakeSubscriptions{pub: pub},
		docker: dkr,
	}, pub, dkr
}

type callerFakeExecutor struct {
	tools.FlowToolsExecutor

	err      error
	released bool
}

func (e *callerFakeExecutor) Release(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	e.released = true

	return e.err
}

func newFinishableWorker(q database.Querier, ex tools.FlowToolsExecutor) *flowWorker {
	ctx, cancel := context.WithCancel(context.Background())
	flowCtx := &FlowContext{
		FlowID:    42,
		DB:        q,
		Executor:  ex,
		Publisher: &cascadeFakePublisher{},
	}

	return &flowWorker{
		wg:      &sync.WaitGroup{},
		ctx:     ctx,
		cancel:  cancel,
		aws:     map[int64]AssistantWorker{},
		awsMX:   &sync.Mutex{},
		tc:      NewTaskController(flowCtx),
		flowCtx: flowCtx,
		logger:  logrus.WithField("component", "test"),
	}
}

const (
	reservedFlowID      = int64(42)
	reservedAssistantID = int64(43)
)

// createFakeQuerier fails its writes once the context is done, the way pgx does.
type createFakeQuerier struct {
	database.Querier

	mx        sync.Mutex
	flow      database.Flow
	createErr error
	statuses  []database.FlowStatus
	messages  []database.CreateMsgLogParams
	results   []database.CreateResultMsgLogParams

	readDelay         time.Duration
	assistant         database.Assistant
	assistantErr      error
	assistantStatuses []database.AssistantStatus
	containers        []database.Container
}

func (f *createFakeQuerier) CreateFlow(ctx context.Context, arg database.CreateFlowParams) (database.Flow, error) {
	f.mx.Lock()
	defer f.mx.Unlock()

	if err := ctx.Err(); err != nil {
		return database.Flow{}, err
	}
	if f.createErr != nil {
		return database.Flow{}, f.createErr
	}
	f.flow = database.Flow{ID: reservedFlowID, UserID: arg.UserID, Status: arg.Status, Title: arg.Title}

	return f.flow, nil
}

func (f *createFakeQuerier) GetFlow(context.Context, int64) (database.Flow, error) {
	f.mx.Lock()
	delay, flow := f.readDelay, f.flow
	f.mx.Unlock()

	// The round trip is the window two callers can both walk through.
	time.Sleep(delay)

	return flow, nil
}

func (f *createFakeQuerier) UpdateFlowStatus(
	ctx context.Context, arg database.UpdateFlowStatusParams,
) (database.Flow, error) {
	f.mx.Lock()
	defer f.mx.Unlock()

	if err := ctx.Err(); err != nil {
		return database.Flow{}, err
	}
	f.statuses = append(f.statuses, arg.Status)
	f.flow.Status = arg.Status

	return f.flow, nil
}

func (f *createFakeQuerier) GetFlowContainers(ctx context.Context, _ int64) ([]database.Container, error) {
	f.mx.Lock()
	defer f.mx.Unlock()

	if err := ctx.Err(); err != nil {
		return nil, err
	}

	return slices.Clone(f.containers), nil
}

// GetFlowPrimaryContainer finds none, so a flow reloaded from the database fails to load.
func (f *createFakeQuerier) GetFlowPrimaryContainer(ctx context.Context, _ int64) (database.Container, error) {
	if err := ctx.Err(); err != nil {
		return database.Container{}, err
	}

	return database.Container{}, errors.New("the flow has no primary container")
}

func (f *createFakeQuerier) GetFlowTasks(context.Context, int64) ([]database.Task, error) {
	return nil, nil
}

func (f *createFakeQuerier) GetFlowAssistants(context.Context, int64) ([]database.Assistant, error) {
	return nil, nil
}

func (f *createFakeQuerier) GetTaskSubtasks(context.Context, int64) ([]database.Subtask, error) {
	return nil, nil
}

func (f *createFakeQuerier) UpdateTaskStatus(
	ctx context.Context, arg database.UpdateTaskStatusParams,
) (database.Task, error) {
	if err := ctx.Err(); err != nil {
		return database.Task{}, err
	}

	return database.Task{ID: arg.ID, Status: arg.Status}, nil
}

func (f *createFakeQuerier) CreateMsgLog(
	ctx context.Context, arg database.CreateMsgLogParams,
) (database.Msglog, error) {
	f.mx.Lock()
	defer f.mx.Unlock()

	if err := ctx.Err(); err != nil {
		return database.Msglog{}, err
	}
	f.messages = append(f.messages, arg)

	return database.Msglog{ID: int64(len(f.messages)), Type: arg.Type, Message: arg.Message}, nil
}

func (f *createFakeQuerier) CreateResultMsgLog(
	ctx context.Context, arg database.CreateResultMsgLogParams,
) (database.Msglog, error) {
	f.mx.Lock()
	defer f.mx.Unlock()

	if err := ctx.Err(); err != nil {
		return database.Msglog{}, err
	}
	f.results = append(f.results, arg)

	return database.Msglog{ID: int64(len(f.messages) + len(f.results)), Type: arg.Type, Message: arg.Message}, nil
}

func (f *createFakeQuerier) CreateAssistant(
	ctx context.Context, arg database.CreateAssistantParams,
) (database.Assistant, error) {
	f.mx.Lock()
	defer f.mx.Unlock()

	if err := ctx.Err(); err != nil {
		return database.Assistant{}, err
	}
	if f.assistantErr != nil {
		return database.Assistant{}, f.assistantErr
	}
	f.assistant = database.Assistant{ID: reservedAssistantID, FlowID: arg.FlowID, Status: arg.Status}

	return f.assistant, nil
}

func (f *createFakeQuerier) UpdateAssistantStatus(
	ctx context.Context, arg database.UpdateAssistantStatusParams,
) (database.Assistant, error) {
	f.mx.Lock()
	defer f.mx.Unlock()

	if err := ctx.Err(); err != nil {
		return database.Assistant{}, err
	}
	f.assistantStatuses = append(f.assistantStatuses, arg.Status)
	f.assistant.Status = arg.Status

	return f.assistant, nil
}

func (f *createFakeQuerier) recordedAssistantStatuses() []database.AssistantStatus {
	f.mx.Lock()
	defer f.mx.Unlock()

	return slices.Clone(f.assistantStatuses)
}

func (f *createFakeQuerier) recordedStatuses() []database.FlowStatus {
	f.mx.Lock()
	defer f.mx.Unlock()

	return slices.Clone(f.statuses)
}

func (f *createFakeQuerier) recordedMessages() []database.CreateMsgLogParams {
	f.mx.Lock()
	defer f.mx.Unlock()

	return slices.Clone(f.messages)
}

func (f *createFakeQuerier) recordedResults() []database.CreateResultMsgLogParams {
	f.mx.Lock()
	defer f.mx.Unlock()

	return slices.Clone(f.results)
}

const lifecycleAssistantID = int64(4242)

// lifecycleQuerier serves an assistant's build and load; updateErr comes with sqlc's zero row.
type lifecycleQuerier struct {
	database.Querier

	updateErr error
	statusErr error

	mx          sync.Mutex
	statusCalls []database.UpdateAssistantStatusParams
}

func (*lifecycleQuerier) GetUser(_ context.Context, id int64) (database.GetUserRow, error) {
	return database.GetUserRow{ID: id, Mail: "user@pentagi.com", Name: "user"}, nil
}

func (*lifecycleQuerier) GetFlowPrimaryContainer(_ context.Context, flowID int64) (database.Container, error) {
	return database.Container{ID: 1, FlowID: flowID, Image: "debian:latest"}, nil
}

func (*lifecycleQuerier) GetUserPrompts(context.Context, int64) ([]database.Prompt, error) {
	return nil, nil
}

func (*lifecycleQuerier) CreateAssistant(
	ctx context.Context, arg database.CreateAssistantParams,
) (database.Assistant, error) {
	if err := ctx.Err(); err != nil {
		return database.Assistant{}, err
	}

	return database.Assistant{
		ID: lifecycleAssistantID, FlowID: arg.FlowID, Status: database.AssistantStatusCreated, Title: arg.Title,
	}, nil
}

func (q *lifecycleQuerier) UpdateAssistant(
	ctx context.Context, arg database.UpdateAssistantParams,
) (database.Assistant, error) {
	if err := ctx.Err(); err != nil {
		return database.Assistant{}, err
	}
	if q.updateErr != nil {
		return database.Assistant{}, q.updateErr
	}

	return database.Assistant{ID: arg.ID, Status: database.AssistantStatusCreated, Title: arg.Title}, nil
}

func (*lifecycleQuerier) UpdateAssistantUseAgents(
	ctx context.Context, arg database.UpdateAssistantUseAgentsParams,
) (database.Assistant, error) {
	if err := ctx.Err(); err != nil {
		return database.Assistant{}, err
	}

	return database.Assistant{ID: arg.ID, UseAgents: arg.UseAgents}, nil
}

func (q *lifecycleQuerier) UpdateAssistantStatus(
	ctx context.Context, arg database.UpdateAssistantStatusParams,
) (database.Assistant, error) {
	if err := ctx.Err(); err != nil {
		return database.Assistant{}, err
	}
	if q.statusErr != nil {
		return database.Assistant{}, q.statusErr
	}

	q.mx.Lock()
	defer q.mx.Unlock()
	q.statusCalls = append(q.statusCalls, arg)

	return database.Assistant{ID: arg.ID, Status: arg.Status}, nil
}

func (q *lifecycleQuerier) recordedStatuses() []database.AssistantStatus {
	q.mx.Lock()
	defer q.mx.Unlock()

	statuses := make([]database.AssistantStatus, 0, len(q.statusCalls))
	for _, call := range q.statusCalls {
		statuses = append(statuses, call.Status)
	}

	return statuses
}

type lifecycleNote struct {
	msgType database.MsglogType
	msg     string
}

// lifecycleAssistantLog records messages, refusing them while err is set or the context is done.
type lifecycleAssistantLog struct {
	FlowAssistantLogWorker

	mx    sync.Mutex
	err   error
	notes []lifecycleNote
}

func (*lifecycleAssistantLog) StreamFlowAssistantMsg(ctx context.Context, _ *providers.StreamMessageChunk) error {
	return ctx.Err()
}

func (l *lifecycleAssistantLog) PutFlowAssistantMsg(
	ctx context.Context, msgType database.MsglogType, _, msg string,
) (int64, error) {
	l.mx.Lock()
	defer l.mx.Unlock()

	if err := ctx.Err(); err != nil {
		return 0, err
	}
	if l.err != nil {
		return 0, l.err
	}
	l.notes = append(l.notes, lifecycleNote{msgType: msgType, msg: msg})

	return int64(len(l.notes)), nil
}

func (l *lifecycleAssistantLog) recovers() {
	l.mx.Lock()
	defer l.mx.Unlock()

	l.err = nil
}

func (l *lifecycleAssistantLog) recorded() []lifecycleNote {
	l.mx.Lock()
	defer l.mx.Unlock()

	return slices.Clone(l.notes)
}

type lifecycleLogController struct {
	AssistantLogController

	log *lifecycleAssistantLog
}

func (c *lifecycleLogController) NewFlowAssistantLog(
	context.Context, int64, int64, subscriptions.FlowPublisher,
) (FlowAssistantLogWorker, error) {
	return c.log, nil
}

type lifecycleWorkerControllers struct {
	AgentLogController
	MsgLogController
	SearchLogController
	TermLogController
	VectorStoreLogController
	ToolCallLogController
	ScreenshotController
}

func (lifecycleWorkerControllers) GetFlowAgentLog(context.Context, int64) (FlowAgentLogWorker, error) {
	return nil, nil
}
func (lifecycleWorkerControllers) GetFlowMsgLog(context.Context, int64) (FlowMsgLogWorker, error) {
	return nil, nil
}
func (lifecycleWorkerControllers) GetFlowSearchLog(context.Context, int64) (FlowSearchLogWorker, error) {
	return nil, nil
}
func (lifecycleWorkerControllers) GetFlowTermLog(context.Context, int64) (FlowTermLogWorker, error) {
	return nil, nil
}
func (lifecycleWorkerControllers) GetFlowVectorStoreLog(context.Context, int64) (FlowVectorStoreLogWorker, error) {
	return nil, nil
}
func (lifecycleWorkerControllers) GetFlowToolCallLog(context.Context, int64) (FlowToolCallLogWorker, error) {
	return nil, nil
}
func (lifecycleWorkerControllers) GetFlowScreenshot(context.Context, int64) (FlowScreenshotWorker, error) {
	return nil, nil
}

// lifecycleProviders hands out provider, or with entered set blocks until release and fails.
type lifecycleProviders struct {
	providers.ProviderController

	provider providers.AssistantProvider
	entered  chan struct{}
	release  chan struct{}
}

func (p *lifecycleProviders) NewAssistantProvider(
	context.Context, provider.ProviderName, templates.Prompter, tools.FlowToolsExecutor,
	int64, int64, int64, string, string, providers.StreamMessageHandler,
) (providers.AssistantProvider, error) {
	if p.entered != nil {
		close(p.entered)
		<-p.release

		return nil, errors.New("the provider went away")
	}

	return p.provider, nil
}

func (p *lifecycleProviders) LoadAssistantProvider(
	context.Context, provider.ProviderName, templates.Prompter, tools.FlowToolsExecutor,
	int64, int64, int64, string, string, string, string, providers.StreamMessageHandler,
) (providers.AssistantProvider, error) {
	return p.provider, nil
}

func (p *lifecycleProviders) GraphitiClient() *graphiti.Client { return nil }

func lifecycleAssistantCtx(q database.Querier, provs providers.ProviderController) newAssistantWorkerCtx {
	return newAssistantWorkerCtx{
		userID:    7,
		flowID:    42,
		input:     "run the scan",
		prvname:   provider.ProviderName("custom"),
		prvtype:   provider.ProviderType("custom"),
		functions: &tools.Functions{},
		flowWorkerCtx: flowWorkerCtx{
			db:    q,
			cfg:   &config.Config{},
			provs: provs,
			subs:  &cascadeFakeSubscriptions{pub: &cascadeFakePublisher{}},
			flowProviderControllers: flowProviderControllers{
				aslc: &lifecycleLogController{log: &lifecycleAssistantLog{}},
				alc:  lifecycleWorkerControllers{},
				mlc:  lifecycleWorkerControllers{},
				slc:  lifecycleWorkerControllers{},
				tlc:  lifecycleWorkerControllers{},
				vslc: lifecycleWorkerControllers{},
				tclc: lifecycleWorkerControllers{},
				sc:   lifecycleWorkerControllers{},
			},
		},
	}
}
