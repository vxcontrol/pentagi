package controller

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"slices"
	"sync"
	"sync/atomic"
	"time"

	"pentagi/pkg/cast"
	"pentagi/pkg/config"
	"pentagi/pkg/database"
	"pentagi/pkg/docker"
	"pentagi/pkg/flowfiles"
	"pentagi/pkg/graph/model"
	"pentagi/pkg/graph/subscriptions"
	obs "pentagi/pkg/observability"
	"pentagi/pkg/observability/langfuse"
	"pentagi/pkg/providers"
	"pentagi/pkg/providers/pconfig"
	"pentagi/pkg/providers/provider"
	"pentagi/pkg/resources"
	"pentagi/pkg/tools"

	"github.com/moby/moby/client"
	"github.com/sirupsen/logrus"
)

const stopTaskTimeout = 60 * time.Second

type FlowWorker interface {
	GetFlowID() int64
	GetUserID() int64
	GetTitle() string
	GetContext() *FlowContext
	GetStatus(ctx context.Context) (database.FlowStatus, error)
	SetStatus(ctx context.Context, status database.FlowStatus) error
	AddAssistant(ctx context.Context, aw AssistantWorker) error
	GetAssistant(ctx context.Context, assistantID int64) (AssistantWorker, error)
	DeleteAssistant(ctx context.Context, assistantID int64) error
	ListAssistants(ctx context.Context) []AssistantWorker
	ListTasks(ctx context.Context) []TaskWorker
	PutInput(ctx context.Context, input string, prv provider.Provider, resources []database.UserResource) error
	PutResources(ctx context.Context, resources []database.UserResource) error
	Finish(ctx context.Context) error
	Stop(ctx context.Context) error
	Report(ctx context.Context, budget time.Duration) error
	Rename(ctx context.Context, title string) error
	WaitTaskCompletion(ctx context.Context) error
	InvalidateTaskSubtasks(ctx context.Context, taskID int64, subtaskIDs []int64)
}

type flowWorker struct {
	tc        TaskController
	wg        *sync.WaitGroup
	cfg       *config.Config
	aws       map[int64]AssistantWorker
	awsMX     *sync.Mutex
	ctx       context.Context
	cancel    context.CancelFunc
	taskMX    *sync.Mutex
	taskST    context.CancelFunc
	taskWG    *sync.WaitGroup
	reportWG  sync.WaitGroup
	reportMX  sync.Mutex
	reportST  context.CancelFunc
	finalized atomic.Bool
	taskCMX   sync.Mutex
	taskCCH   chan struct{}
	input     chan flowInput
	flowCtx   *FlowContext
	docker    docker.DockerClient
	logger    *logrus.Entry
}

type newFlowWorkerCtx struct {
	userID    int64
	input     string
	dryRun    bool
	prvname   provider.ProviderName
	prvtype   provider.ProviderType
	functions *tools.Functions
	resources []database.UserResource

	flowWorkerCtx
}

type flowWorkerCtx struct {
	db     database.Querier
	cfg    *config.Config
	docker docker.DockerClient
	provs  providers.ProviderController
	subs   subscriptions.SubscriptionsController

	flowProviderControllers
}

type flowProviderControllers struct {
	mlc  MsgLogController
	aslc AssistantLogController
	alc  AgentLogController
	slc  SearchLogController
	tlc  TermLogController
	vslc VectorStoreLogController
	tclc ToolCallLogController
	sc   ScreenshotController
}

type flowProviderWorkers struct {
	mlw  FlowMsgLogWorker
	alw  FlowAgentLogWorker
	slw  FlowSearchLogWorker
	tlw  FlowTermLogWorker
	vslw FlowVectorStoreLogWorker
	tclw FlowToolCallLogWorker
	sw   FlowScreenshotWorker
}

const flowInputTimeout = 1 * time.Second

// ErrInputAccepted reports that the worker took the input but had not finished
// with it before the CALLER's deadline. The work continues, so the caller must
// observe the outcome through the flow's status rather than send it again.
// A worker that merely outlives flowInputTimeout returns nil, as it always has.
var ErrInputAccepted = errors.New("input accepted, still being processed")

// ErrInputNotAccepted reports the opposite: the worker never took the input, so
// nothing was started and sending it again is safe. Telling these two apart is
// the whole point -- a caller that cannot duplicates the action or drops it.
var ErrInputNotAccepted = errors.New("input was not accepted")

type flowInput struct {
	input string
	reply *inputReply
}

func reserveFlow(ctx context.Context, fwc newFlowWorkerCtx) (database.Flow, error) {
	ctx, span := obs.Observer.NewSpan(ctx, obs.SpanKindInternal, "controller.reserveFlow")
	defer span.End()

	flow, err := fwc.db.CreateFlow(ctx, database.CreateFlowParams{
		Title:              "untitled",
		Status:             database.FlowStatusCreated,
		Model:              "unknown",
		ModelProviderName:  fwc.prvname.String(),
		ModelProviderType:  database.ProviderType(fwc.prvtype),
		Language:           "English",
		ToolCallIDTemplate: cast.ToolCallIDTemplate,
		Functions:          []byte("{}"),
		UserID:             fwc.userID,
	})
	if err != nil {
		obs.LogErrorOrCancel(logrus.WithContext(ctx), err, "failed to create flow in DB")
		return database.Flow{}, fmt.Errorf("failed to create flow in DB: %w", err)
	}

	logrus.WithContext(ctx).WithFields(logrus.Fields{
		"flow_id":       flow.ID,
		"user_id":       fwc.userID,
		"provider_name": fwc.prvname.String(),
		"provider_type": fwc.prvtype.String(),
	}).Info("flow created in DB")

	return flow, nil
}

func buildFlowWorker(
	ctx context.Context, flow database.Flow, fwc newFlowWorkerCtx, commit func() error,
) (FlowWorker, error) {
	ctx, span := obs.Observer.NewSpan(ctx, obs.SpanKindInternal, "controller.buildFlowWorker")
	defer span.End()

	logger := logrus.WithContext(ctx).WithFields(logrus.Fields{
		"flow_id":       flow.ID,
		"user_id":       fwc.userID,
		"provider_name": fwc.prvname.String(),
		"provider_type": fwc.prvtype.String(),
	})

	user, err := fwc.db.GetUser(ctx, fwc.userID)
	if err != nil {
		logger.WithError(err).Error("failed to get user")
		return nil, fmt.Errorf("failed to get user %d: %w", fwc.userID, err)
	}

	ctx, observation := obs.Observer.NewObservation(ctx,
		langfuse.WithObservationTraceContext(
			langfuse.WithTraceName(fmt.Sprintf("%s%d flow worker", fwc.cfg.TenantLabel(), flow.ID)),
			langfuse.WithTraceUserID(tenantUserID(fwc.cfg, user.Mail)),
			langfuse.WithTraceTags(tenantTags(fwc.cfg, "controller", "flow")),
			langfuse.WithTraceInput(fwc.input),
			langfuse.WithTraceSessionID(fwc.cfg.ScopedName(fmt.Sprintf("flow-%d", flow.ID))),
			langfuse.WithTraceMetadata(tenantMeta(fwc.cfg, langfuse.Metadata{
				"flow_id":       flow.ID,
				"user_id":       fwc.userID,
				"user_email":    user.Mail,
				"user_name":     user.Name,
				"user_hash":     user.Hash,
				"user_role":     user.RoleName,
				"provider_name": fwc.prvname.String(),
				"provider_type": fwc.prvtype.String(),
			})),
		),
	)
	flowSpan := observation.Span(langfuse.WithSpanName("prepare flow worker"))
	ctx, _ = flowSpan.Observation(ctx)

	prompter, err := newUserPrompter(ctx, fwc.db, fwc.userID)
	if err != nil {
		return nil, wrapErrorEndSpan(ctx, flowSpan, "failed to build user prompter", err)
	}
	executor, err := tools.NewFlowToolsExecutor(fwc.db, fwc.cfg, fwc.docker, fwc.functions, fwc.userID, flow.ID)
	if err != nil {
		return nil, wrapErrorEndSpan(ctx, flowSpan, "failed to create flow tools executor", err)
	}
	flowProvider, err := fwc.provs.NewFlowProvider(
		ctx, fwc.prvname, prompter, executor, flow.ID, fwc.userID, fwc.cfg.AskUser, fwc.input,
	)
	if err != nil {
		return nil, wrapErrorEndSpan(ctx, flowSpan, "failed to get flow provider", err)
	}

	functionsBlob, err := json.Marshal(fwc.functions)
	if err != nil {
		return nil, wrapErrorEndSpan(ctx, flowSpan, "failed to marshal functions", err)
	}

	flow, err = fwc.db.UpdateFlow(ctx, database.UpdateFlowParams{
		Title:              flowProvider.Title(),
		Model:              flowProvider.Model(pconfig.OptionsTypePrimaryAgent),
		Language:           flowProvider.Language(),
		ToolCallIDTemplate: flowProvider.ToolCallIDTemplate(),
		Functions:          functionsBlob,
		TraceID:            database.StringToNullString(observation.TraceID()),
		ID:                 flow.ID,
	})
	if err != nil {
		return nil, wrapErrorEndSpan(ctx, flowSpan, "failed to update flow in DB", err)
	}

	pub := fwc.subs.NewFlowPublisher(fwc.userID, flow.ID)
	workers, err := newFlowProviderWorkers(ctx, flow.ID, &fwc.flowProviderControllers, pub)
	if err != nil {
		return nil, wrapErrorEndSpan(ctx, flowSpan, "failed to create flow provider workers", err)
	}

	flowProvider.SetAgentLogProvider(workers.alw)
	flowProvider.SetMsgLogProvider(workers.mlw)

	executor.SetImage(flowProvider.Image())
	executor.SetEmbedder(flowProvider.Embedder())
	executor.SetScreenshotProvider(workers.sw)
	executor.SetAgentLogProvider(workers.alw)
	executor.SetMsgLogProvider(workers.mlw)
	executor.SetSearchLogProvider(workers.slw)
	executor.SetTermLogProvider(workers.tlw)
	executor.SetVectorStoreLogProvider(workers.vslw)
	executor.SetToolCallLogProvider(workers.tclw)
	executor.SetKnowledgeProvider(pub)
	executor.SetGraphitiClient(fwc.provs.GraphitiClient())

	flowCtx := &FlowContext{
		DB:         fwc.db,
		UserID:     fwc.userID,
		FlowID:     flow.ID,
		TraceID:    observation.TraceID(),
		Executor:   executor,
		Provider:   flowProvider,
		Publisher:  pub,
		MsgLog:     workers.mlw,
		TermLog:    workers.tlw,
		Screenshot: workers.sw,
	}
	ctx, cancel := context.WithCancel(context.Background())
	ctx, _ = obs.Observer.NewObservation(ctx, langfuse.WithObservationTraceID(observation.TraceID()))
	fw := &flowWorker{
		tc:      NewTaskController(flowCtx),
		wg:      &sync.WaitGroup{},
		aws:     make(map[int64]AssistantWorker),
		awsMX:   &sync.Mutex{},
		cfg:     fwc.cfg,
		ctx:     ctx,
		cancel:  cancel,
		taskMX:  &sync.Mutex{},
		taskST:  func() {},
		taskWG:  &sync.WaitGroup{},
		taskCCH: make(chan struct{}),
		input:   make(chan flowInput),
		flowCtx: flowCtx,
		docker:  fwc.docker,
		logger: logrus.WithFields(logrus.Fields{
			"flow_id":   flow.ID,
			"user_id":   fwc.userID,
			"trace_id":  observation.TraceID(),
			"component": "worker",
		}),
	}

	if err := executor.Prepare(ctx); err != nil {
		return nil, wrapErrorEndSpan(ctx, flowSpan, "failed to prepare flow resources", err)
	}

	if err := commit(); err != nil {
		if relErr := executor.Release(ctx); relErr != nil {
			logger.WithError(relErr).Warn("failed to release the sandbox of a flow closed while it was preparing")
		}
		cancel()
		flowSpan.End(langfuse.WithSpanStatus("flow closed while preparing"))

		return nil, err
	}

	containers, err := fwc.db.GetFlowContainers(ctx, flow.ID)
	if err != nil {
		return nil, wrapErrorEndSpan(ctx, flowSpan, "failed to get flow containers", err)
	}

	pub.FlowUpdated(ctx, flow, containers)

	fw.wg.Add(1)
	go fw.worker()

	if !fwc.dryRun {
		if err := fw.PutInput(ctx, fwc.input, nil, fwc.resources); err != nil {
			return nil, wrapErrorEndSpan(ctx, flowSpan, "failed to run flow worker", err)
		}
	}

	flowSpan.End(langfuse.WithSpanStatus("flow worker started"))

	return fw, nil
}

func LoadFlowWorker(ctx context.Context, flow database.Flow, fwc flowWorkerCtx) (FlowWorker, error) {
	ctx, span := obs.Observer.NewSpan(ctx, obs.SpanKindInternal, "controller.LoadFlowWorker")
	defer span.End()

	switch flow.Status {
	case database.FlowStatusRunning, database.FlowStatusWaiting:
	default:
		return nil, fmt.Errorf("flow %d has status %s: loading aborted: %w", flow.ID, flow.Status, ErrNothingToLoad)
	}

	logger := logrus.WithContext(ctx).WithFields(logrus.Fields{
		"flow_id":       flow.ID,
		"user_id":       flow.UserID,
		"provider_name": flow.ModelProviderName,
		"provider_type": flow.ModelProviderType,
	})

	container, err := fwc.db.GetFlowPrimaryContainer(ctx, flow.ID)
	if err != nil {
		logger.WithError(err).Error("failed to get flow primary container")
		return nil, fmt.Errorf("failed to get flow primary container: %w", err)
	}

	logger.Info("flow loaded from DB")

	user, err := fwc.db.GetUser(ctx, flow.UserID)
	if err != nil {
		logger.WithError(err).Error("failed to get user")
		return nil, fmt.Errorf("failed to get user %d: %w", flow.UserID, err)
	}

	ctx, observation := obs.Observer.NewObservation(ctx,
		langfuse.WithObservationTraceID(flow.TraceID.String),
		langfuse.WithObservationTraceContext(
			langfuse.WithTraceName(fmt.Sprintf("%s%d flow worker", fwc.cfg.TenantLabel(), flow.ID)),
			langfuse.WithTraceUserID(tenantUserID(fwc.cfg, user.Mail)),
			langfuse.WithTraceTags(tenantTags(fwc.cfg, "controller", "flow")),
			langfuse.WithTraceSessionID(fwc.cfg.ScopedName(fmt.Sprintf("flow-%d", flow.ID))),
			langfuse.WithTraceMetadata(tenantMeta(fwc.cfg, langfuse.Metadata{
				"flow_id":       flow.ID,
				"user_id":       flow.UserID,
				"user_email":    user.Mail,
				"user_name":     user.Name,
				"user_hash":     user.Hash,
				"user_role":     user.RoleName,
				"provider_name": flow.ModelProviderName,
				"provider_type": flow.ModelProviderType,
			})),
		),
	)
	flowSpan := observation.Span(langfuse.WithSpanName("prepare flow worker"))
	ctx, _ = flowSpan.Observation(ctx)

	functions := &tools.Functions{}
	if err := json.Unmarshal(flow.Functions, functions); err != nil {
		return nil, wrapErrorEndSpan(ctx, flowSpan, "failed to unmarshal functions", err)
	}

	prompter, err := newUserPrompter(ctx, fwc.db, flow.UserID)
	if err != nil {
		return nil, wrapErrorEndSpan(ctx, flowSpan, "failed to build user prompter", err)
	}
	executor, err := tools.NewFlowToolsExecutor(fwc.db, fwc.cfg, fwc.docker, functions, flow.UserID, flow.ID)
	if err != nil {
		return nil, wrapErrorEndSpan(ctx, flowSpan, "failed to create flow tools executor", err)
	}
	flowProvider, err := fwc.provs.LoadFlowProvider(
		ctx, provider.ProviderName(flow.ModelProviderName),
		prompter, executor, flow.ID, flow.UserID, fwc.cfg.AskUser,
		container.Image, flow.Language, flow.Title, flow.ToolCallIDTemplate,
	)
	if err != nil {
		return nil, wrapErrorEndSpan(ctx, flowSpan, "failed to get flow provider", err)
	}

	pub := fwc.subs.NewFlowPublisher(flow.UserID, flow.ID)
	workers, err := newFlowProviderWorkers(ctx, flow.ID, &fwc.flowProviderControllers, pub)
	if err != nil {
		return nil, wrapErrorEndSpan(ctx, flowSpan, "failed to create flow provider workers", err)
	}

	flowProvider.SetAgentLogProvider(workers.alw)
	flowProvider.SetMsgLogProvider(workers.mlw)

	executor.SetImage(flowProvider.Image())
	executor.SetEmbedder(flowProvider.Embedder())
	executor.SetScreenshotProvider(workers.sw)
	executor.SetAgentLogProvider(workers.alw)
	executor.SetMsgLogProvider(workers.mlw)
	executor.SetSearchLogProvider(workers.slw)
	executor.SetTermLogProvider(workers.tlw)
	executor.SetVectorStoreLogProvider(workers.vslw)
	executor.SetToolCallLogProvider(workers.tclw)
	executor.SetKnowledgeProvider(pub)
	executor.SetGraphitiClient(fwc.provs.GraphitiClient())

	flowCtx := &FlowContext{
		DB:         fwc.db,
		UserID:     flow.UserID,
		FlowID:     flow.ID,
		TraceID:    observation.TraceID(),
		Executor:   executor,
		Provider:   flowProvider,
		Publisher:  pub,
		MsgLog:     workers.mlw,
		TermLog:    workers.tlw,
		Screenshot: workers.sw,
	}
	ctx, cancel := context.WithCancel(context.Background())
	ctx, _ = obs.Observer.NewObservation(ctx, langfuse.WithObservationTraceID(observation.TraceID()))
	fw := &flowWorker{
		tc:      NewTaskController(flowCtx),
		wg:      &sync.WaitGroup{},
		aws:     make(map[int64]AssistantWorker),
		awsMX:   &sync.Mutex{},
		cfg:     fwc.cfg,
		ctx:     ctx,
		cancel:  cancel,
		taskMX:  &sync.Mutex{},
		taskST:  func() {},
		taskWG:  &sync.WaitGroup{},
		taskCCH: make(chan struct{}),
		input:   make(chan flowInput),
		flowCtx: flowCtx,
		docker:  fwc.docker,
		logger: logrus.WithFields(logrus.Fields{
			"flow_id":   flow.ID,
			"user_id":   flow.UserID,
			"trace_id":  observation.TraceID(),
			"component": "worker",
		}),
	}

	if err := executor.Prepare(ctx); err != nil {
		return nil, wrapErrorEndSpan(ctx, flowSpan, "failed to prepare flow resources", err)
	}

	containers, err := fwc.db.GetFlowContainers(ctx, flow.ID)
	if err != nil {
		return nil, wrapErrorEndSpan(ctx, flowSpan, "failed to get flow containers", err)
	}

	if err := fw.tc.LoadTasks(ctx, flow.ID, fw); err != nil && !errors.Is(err, ErrNothingToLoad) {
		return nil, wrapErrorEndSpan(ctx, flowSpan, "failed to load tasks", err)
	}

	assistants, err := fwc.db.GetFlowAssistants(ctx, flow.ID)
	if err != nil {
		return nil, wrapErrorEndSpan(ctx, flowSpan, "failed to get flow assistants", err)
	}

	awc := assistantWorkerCtx{
		userID:        flow.UserID,
		flowID:        flow.ID,
		prompter:      prompter,
		fw:            fw,
		flowWorkerCtx: fwc,
	}
	for _, assistant := range assistants {
		aw, err := LoadAssistantWorker(ctx, assistant, awc)
		if err != nil {
			if errors.Is(err, ErrNothingToLoad) {
				continue
			}
			// One unloadable assistant must not take its flow down with it.
			// Aborting here would leave the flow absent from the controller's
			// map while its row stays alive in the DB, which makes the whole
			// flow permanently unreachable ("flow not found" on every action) —
			// the exact failure this used to produce when an assistant pointed
			// at a renamed or deleted provider. The assistant simply stays
			// unloaded and returns on the next start once its cause is fixed.
			logger.WithError(err).Errorf("failed to load assistant %d, skipping it", assistant.ID)
			continue
		}
		if err := fw.AddAssistant(ctx, aw); err != nil {
			return nil, wrapErrorEndSpan(ctx, flowSpan, "failed to add assistant worker", err)
		}
	}

	fw.flowCtx.Publisher.FlowUpdated(ctx, flow, containers)

	fw.wg.Add(1)
	go fw.worker()

	flowSpan.End(langfuse.WithSpanStatus("flow worker restored"))

	return fw, nil
}

func (fw *flowWorker) GetFlowID() int64 {
	return fw.flowCtx.FlowID
}

func (fw *flowWorker) GetUserID() int64 {
	return fw.flowCtx.UserID
}

func (fw *flowWorker) GetTitle() string {
	if fw.flowCtx.Provider != nil {
		return fw.flowCtx.Provider.Title()
	}
	return ""
}

func (fw *flowWorker) takeStatus(status database.FlowStatus) bool {
	switch status {
	case database.FlowStatusFinished, database.FlowStatusFailed:
		fw.finalized.Store(true)

		return true
	default:
		return !fw.finalized.Load()
	}
}

func (fw *flowWorker) GetContext() *FlowContext {
	return fw.flowCtx
}

func (fw *flowWorker) GetStatus(ctx context.Context) (database.FlowStatus, error) {
	flow, err := fw.flowCtx.DB.GetUserFlow(ctx, database.GetUserFlowParams{
		UserID: fw.flowCtx.UserID,
		ID:     fw.flowCtx.FlowID,
	})
	if err != nil {
		return database.FlowStatusFailed, err
	}

	return flow.Status, nil
}

func (fw *flowWorker) SetStatus(ctx context.Context, status database.FlowStatus) error {
	if !fw.takeStatus(status) {
		return nil
	}

	flow, err := fw.flowCtx.DB.UpdateFlowStatus(ctx, database.UpdateFlowStatusParams{
		Status: status,
		ID:     fw.flowCtx.FlowID,
	})
	if err != nil {
		return fmt.Errorf("failed to set flow %d status: %w", fw.flowCtx.FlowID, err)
	}

	containers, err := fw.flowCtx.DB.GetFlowContainers(ctx, fw.flowCtx.FlowID)
	if err != nil {
		return fmt.Errorf("failed to get flow %d containers: %w", fw.flowCtx.FlowID, err)
	}

	fw.flowCtx.Publisher.FlowUpdated(ctx, flow, containers)

	return nil
}

// InvalidateTaskSubtasks drops stale workers after direct DB deletion,
// preventing delayed ErrNoRows failures.
func (fw *flowWorker) InvalidateTaskSubtasks(ctx context.Context, taskID int64, subtaskIDs []int64) {
	task, err := fw.tc.GetTask(ctx, taskID)
	if err != nil {
		return
	}

	task.InvalidateSubtasks(subtaskIDs)
}

func (fw *flowWorker) AddAssistant(ctx context.Context, aw AssistantWorker) error {
	fw.awsMX.Lock()
	defer fw.awsMX.Unlock()

	if fw.ctx.Err() != nil {
		return fmt.Errorf("flow %d: %w", fw.flowCtx.FlowID, ErrFlowAlreadyStopped)
	}

	if taw, ok := fw.aws[aw.GetAssistantID()]; ok {
		if taw == aw {
			return nil
		}

		if err := taw.Finish(ctx); err != nil {
			return fmt.Errorf("failed to finish assistant %d: %w", aw.GetAssistantID(), err)
		}
	}

	fw.aws[aw.GetAssistantID()] = aw

	return nil
}

func (fw *flowWorker) GetAssistant(ctx context.Context, assistantID int64) (AssistantWorker, error) {
	fw.awsMX.Lock()
	defer fw.awsMX.Unlock()

	if aw, ok := fw.aws[assistantID]; ok {
		return aw, nil
	}

	return nil, fmt.Errorf("assistant %d not found", assistantID)
}

func (fw *flowWorker) DeleteAssistant(ctx context.Context, assistantID int64) error {
	fw.awsMX.Lock()
	defer fw.awsMX.Unlock()

	aw, ok := fw.aws[assistantID]
	if ok {
		if err := aw.Finish(ctx); err != nil {
			return fmt.Errorf("failed to finish assistant %d: %w", assistantID, err)
		}

		delete(fw.aws, assistantID)
	}

	if assistant, err := fw.flowCtx.DB.DeleteAssistant(ctx, assistantID); err != nil {
		return fmt.Errorf("failed to delete assistant %d: %w", assistantID, err)
	} else {
		fw.flowCtx.Publisher.AssistantDeleted(ctx, assistant)
	}

	return nil
}

func (fw *flowWorker) ListAssistants(ctx context.Context) []AssistantWorker {
	fw.awsMX.Lock()
	defer fw.awsMX.Unlock()

	assistants := make([]AssistantWorker, 0, len(fw.aws))
	for _, aw := range fw.aws {
		assistants = append(assistants, aw)
	}

	slices.SortFunc(assistants, func(a, b AssistantWorker) int {
		return int(a.GetAssistantID() - b.GetAssistantID())
	})

	return assistants
}

func (fw *flowWorker) ListTasks(ctx context.Context) []TaskWorker {
	return fw.tc.ListTasks(ctx)
}

func (fw *flowWorker) PutInput(
	ctx context.Context,
	input string,
	prv provider.Provider,
	resources []database.UserResource,
) error {
	ctx, span := obs.Observer.NewSpan(ctx, obs.SpanKindInternal, "controller.flowWorker.PutInput")
	defer span.End()

	if err := fw.ctx.Err(); err != nil {
		return fmt.Errorf("flow %d stopped: %w", fw.flowCtx.FlowID, err)
	}

	if err := fw.switchProvider(ctx, prv); err != nil {
		return fmt.Errorf("failed to switch provider: %w", err)
	}

	if err := fw.PutResources(ctx, resources); err != nil {
		fw.logger.WithError(err).Warn("failed to copy resources before user input")
	}

	flin := flowInput{input: input, reply: newInputReply()}
	select {
	case <-fw.ctx.Done():
		return fmt.Errorf("flow %d stopped: %w: %w", fw.flowCtx.FlowID, ErrInputNotAccepted, fw.ctx.Err())
	case <-ctx.Done():
		return fmt.Errorf("flow %d input processing timeout: %w: %w",
			fw.flowCtx.FlowID, ErrInputNotAccepted, ctx.Err())
	case fw.input <- flin:
		timer := time.NewTimer(flowInputTimeout)
		defer timer.Stop()

		// The worker has the input. Whether it answers in time decides only what
		// this caller learns, not whether the work happens -- so the timer
		// expiring stays the ordinary quiet success it has always been, and only
		// the caller's own deadline needs a name, because that is the one that
		// reaches an HTTP client deciding whether to send the input again.
		var stopErr error
		select {
		case err := <-flin.reply.done:
			return err
		case <-timer.C:
		case <-fw.ctx.Done():
			stopErr = fmt.Errorf("flow %d stopped: %w", fw.flowCtx.FlowID, fw.ctx.Err())
		case <-ctx.Done():
			stopErr = fmt.Errorf("flow %d input processing timeout: %w: %w",
				fw.flowCtx.FlowID, ErrInputAccepted, ctx.Err())
		}

		if isDelivered, err := flin.reply.abandon(); isDelivered {
			return err
		}

		return stopErr
	}
}

func (fw *flowWorker) PutResources(ctx context.Context, dbResources []database.UserResource) error {
	ctx, span := obs.Observer.NewSpan(ctx, obs.SpanKindInternal, "controller.flowWorker.PutResources")
	defer span.End()

	addedPaths, err := fw.copyResourcesToFS(dbResources)
	if err != nil {
		return err
	}
	if len(addedPaths) == 0 {
		return nil
	}

	fw.pushResourcesToContainer(ctx, addedPaths)
	fw.publishResourceFileEvents(ctx, addedPaths)
	return nil
}

// copyResourcesToFS copies user resource blobs into the flow resources directory on disk.
// Returns relative paths of newly written files (skips files already present).
func (fw *flowWorker) copyResourcesToFS(dbResources []database.UserResource) ([]string, error) {
	if len(dbResources) == 0 {
		return nil, nil
	}

	refs := make([]flowfiles.ResourceRef, 0, len(dbResources))
	for _, r := range dbResources {
		refs = append(refs, flowfiles.ResourceRef{
			Hash:        r.Hash,
			VirtualPath: r.Path,
			Name:        r.Name,
			IsDir:       r.IsDir,
		})
	}

	storeDir := resources.ResourcesDir(fw.cfg.DataDir)
	return flowfiles.CopyResourcesToFlow(fw.cfg.DataDir, storeDir, uint64(fw.flowCtx.FlowID), refs, false)
}

// pushResourcesToContainer pushes newly added resource files into the running primary container.
// Each file is sent individually so partial failures are non-fatal.
func (fw *flowWorker) pushResourcesToContainer(ctx context.Context, addedPaths []string) {
	if fw.docker == nil {
		return
	}
	containerName := tools.PrimaryTerminalName(fw.cfg.TenantPrefix(), fw.flowCtx.FlowID)
	running, _ := fw.docker.IsContainerRunning(ctx, containerName)
	if !running {
		return
	}

	resourcesDir := flowfiles.FlowResourcesDir(fw.cfg.DataDir, uint64(fw.flowCtx.FlowID))
	for _, relPath := range addedPaths {
		fsRelPath := relPath[len(flowfiles.ResourcesDirName)+1:]
		absPath := resourcesDir + "/" + fsRelPath

		pr, pw := io.Pipe()
		errCh := make(chan error, 1)
		go func() {
			errCh <- flowfiles.WriteSingleFileTar(pw, absPath, flowfiles.ResourcesDirName+"/"+fsRelPath)
		}()

		copyErr := fw.docker.CopyToContainer(ctx, containerName, docker.WorkFolderPathInContainer, pr,
			client.CopyToContainerOptions{AllowOverwriteDirWithFile: true})
		pr.Close()
		writeErr := <-errCh

		if copyErr != nil || writeErr != nil {
			fw.logger.WithFields(logrus.Fields{
				"path":      relPath,
				"copy_err":  copyErr,
				"write_err": writeErr,
			}).Warn("failed to push resource file to container; will be synced on restart")
		}
	}
}

// publishResourceFileEvents emits flowFileAdded subscription events for newly written resource files.
func (fw *flowWorker) publishResourceFileEvents(ctx context.Context, addedPaths []string) {
	if len(addedPaths) == 0 {
		return
	}

	resourcesDir := flowfiles.FlowResourcesDir(fw.cfg.DataDir, uint64(fw.flowCtx.FlowID))
	pub := fw.flowCtx.Publisher
	for _, relPath := range addedPaths {
		fsRelPath := relPath[len(flowfiles.ResourcesDirName)+1:]
		absPath := resourcesDir + "/" + fsRelPath

		file := &model.FlowFile{
			ID:         flowfiles.ID(relPath),
			Name:       flowfiles.BaseName(relPath),
			Path:       relPath,
			IsDir:      false,
			ModifiedAt: time.Now(),
		}
		if info, err := os.Lstat(absPath); err == nil {
			file.Size = int(info.Size())
			file.ModifiedAt = info.ModTime()
		}

		pub.FlowFileAdded(ctx, file)
	}
}

func (fw *flowWorker) Finish(ctx context.Context) error {
	ctx, span := obs.Observer.NewSpan(ctx, obs.SpanKindInternal, "controller.flowWorker.Finish")
	defer span.End()

	if err := fw.finish(); err != nil {
		return err
	}

	for _, task := range fw.tc.ListTasks(ctx) {
		if !task.IsCompleted() {
			if err := task.Finish(ctx); err != nil {
				return fmt.Errorf("failed to finish task %d: %w", task.GetTaskID(), err)
			}
		}
	}

	if err := fw.SetStatus(ctx, database.FlowStatusFinished); err != nil {
		return fmt.Errorf("failed to set flow %d status: %w", fw.flowCtx.FlowID, err)
	}

	fw.awsMX.Lock()
	defer fw.awsMX.Unlock()

	for _, aw := range fw.aws {
		if err := aw.Finish(ctx); err != nil {
			return fmt.Errorf("failed to finish assistant %d: %w", aw.GetAssistantID(), err)
		}
	}

	if err := fw.flowCtx.Executor.Release(ctx); err != nil {
		return fmt.Errorf("failed to release flow %d resources: %w", fw.flowCtx.FlowID, err)
	}

	return nil
}

func (fw *flowWorker) Stop(ctx context.Context) error {
	ctx, span := obs.Observer.NewSpan(ctx, obs.SpanKindInternal, "controller.flowWorker.Stop")
	defer span.End()

	fw.taskMX.Lock()
	defer fw.taskMX.Unlock()

	fw.taskST()
	fw.stopReports()
	fw.killFlowCommands(ctx)
	defer fw.killFlowCommands(ctx)

	done := make(chan struct{})
	timer := time.NewTimer(stopTaskTimeout)
	defer timer.Stop()

	go func() {
		fw.taskWG.Wait()
		close(done)
	}()

	select {
	case <-timer.C:
		return fmt.Errorf("task stop timeout")
	case <-done:
		return nil
	}
}

func (fw *flowWorker) killFlowCommands(ctx context.Context) {
	if fw.docker == nil {
		return
	}

	containerName := tools.PrimaryTerminalName(fw.cfg.TenantPrefix(), fw.flowCtx.FlowID)
	if err := fw.docker.KillFlowCommands(ctx, containerName); err != nil {
		fw.logger.WithError(err).Warn("failed to stop the commands left running in the sandbox")
	}
}

const (
	// maxReportAttempts is how many times the write-up is asked for before the
	// task is left unreported.
	maxReportAttempts = 3
	// reportRetryDelay is multiplied by the attempt number, so the three
	// attempts span roughly a minute -- long enough to outlast a gateway
	// restart, short enough that a stopped run is not held open by it.
	reportRetryDelay = 20 * time.Second
	// defaultReportAllowance bounds a report whose caller named no budget. The
	// three attempts and their backoff all live inside it.
	defaultReportAllowance = 30 * time.Minute
	// reportDrainTimeout bounds the wait for the run a report displaces, so an
	// unresponsive task delays the write-up instead of withholding it.
	reportDrainTimeout = 30 * time.Second
)

// ErrNothingToReport is a report asked for on a flow whose tasks are all
// finished. The caller's mistake, not a fault.
var ErrNothingToReport = errors.New("flow has no unfinished task to report on")

func reportAllowance(budget time.Duration) time.Duration {
	if budget > 0 {
		return budget
	}

	return defaultReportAllowance
}

// waitForDisplacedTask drains the run whose slot a report is taking.
func (fw *flowWorker) waitForDisplacedTask() {
	drained := make(chan struct{})
	go func() {
		defer close(drained)
		fw.taskWG.Wait()
	}()

	select {
	case <-drained:
	case <-time.After(reportDrainTimeout):
		fw.logger.Warn("the displaced run did not finish in time; reporting anyway")
	}
}

// Report asks for the write-up of this flow's open task.
//
// budget zero means the reporter's ordinary allowance; a caller that supplies
// one bounds the run by it.
func (fw *flowWorker) Report(ctx context.Context, budget time.Duration) error {
	ctx, span := obs.Observer.NewSpan(ctx, obs.SpanKindInternal, "controller.flowWorker.Report")
	defer span.End()

	if err := fw.ctx.Err(); err != nil {
		return fmt.Errorf("flow %d is stopped: %w", fw.flowCtx.FlowID, err)
	}

	task := reportableTask(fw.tc.ListTasks(ctx))
	if task == nil {
		return fmt.Errorf("%w: flow %d", ErrNothingToReport, fw.flowCtx.FlowID)
	}

	// Registered before the goroutine exists, so a Finish arriving in between
	// still waits for it. Its own group rather than taskWG: Stop holds taskMX
	// across the whole of taskWG.Wait, and runReport's first act is to take
	// that mutex, so a report counted there before it starts would stall Stop
	// for its full timeout.
	fw.reportWG.Add(1)

	// Detached, like every other way a task starts here. A report is a model run
	// with a budget measured in minutes, and the caller is an HTTP request:
	// holding it open for the whole write-up would time out at the first proxy
	// in front of this server. Everything that can fail synchronously -- the
	// flow being stopped, there being nothing to report on -- is checked above,
	// so the caller still gets a real error for the cases it can act on.
	go fw.runReport(task, budget)

	return nil
}

// runReport is the body of Report, on its own goroutine.
func (fw *flowWorker) runReport(task TaskWorker, budget time.Duration) {
	defer fw.reportWG.Done()

	// The same locking as runTask: a report is a model run against this flow's
	// container, so it takes the task slot and it is interruptible by Stop.
	//
	// The run being displaced is drained before the slot is taken, the way Stop
	// drains it. Without that, its unwinding writes TaskStatusWaiting on a
	// background context and lands on top of the report's own status, leaving
	// the task reading waiting for the whole write-up.
	fw.taskMX.Lock()
	fw.taskST()
	fw.taskMX.Unlock()

	fw.waitForDisplacedTask()

	fw.taskMX.Lock()
	runCtx, taskST := context.WithCancel(fw.ctx)
	fw.taskST = taskST
	fw.taskMX.Unlock()

	defer taskST()

	fw.taskWG.Add(1)
	defer fw.taskWG.Done()
	defer fw.signalTaskComplete()

	_, observation := obs.Observer.NewObservation(fw.ctx)
	obsSpan := observation.Span(
		langfuse.WithSpanName(fmt.Sprintf("report task %d: %s", task.GetTaskID(), task.GetTitle())),
		langfuse.WithSpanMetadata(langfuse.Metadata{"task_id": task.GetTaskID()}),
	)
	runCtx, _ = obsSpan.Observation(runCtx)

	// The write-up is the most valuable thing a flow produces, and the one piece
	// of work with nothing after it to compensate for its loss: every subtask can
	// be re-planned, the report cannot be re-derived once the flow is finished.
	// The chain already retries each model turn; these attempts are spaced far
	// enough apart to outlive a gateway restart, which is the failure those
	// inner retries cannot cover.
	// Detached from runCtx and bounded once, outside the loop: the write-up is
	// what the whole call exists to produce, so an input arriving while it runs
	// must not discard it, and a caller's timeout is the allowance for the
	// request rather than for each attempt.
	reportCtx, cancelReport := context.WithTimeout(context.WithoutCancel(runCtx), reportAllowance(budget))
	defer cancelReport()

	fw.reportMX.Lock()
	fw.reportST = cancelReport
	fw.reportMX.Unlock()

	var err error
retry:
	for attempt := 1; attempt <= maxReportAttempts; attempt++ {
		if err = task.Report(reportCtx); err == nil {
			break
		}

		// A cancelled report was not lost, it was stopped; and a task that is
		// already written up has nothing more to give. Repeating either fights
		// the operator or spends a model run for nothing.
		if errors.Is(err, context.Canceled) || errors.Is(err, ErrTaskAlreadyCompleted) || reportCtx.Err() != nil {
			break
		}

		if attempt == maxReportAttempts {
			break
		}

		fw.logger.WithError(err).WithFields(logrus.Fields{
			"task_id": task.GetTaskID(),
			"attempt": attempt,
		}).Warn("the write-up failed; trying again rather than losing it")

		select {
		case <-reportCtx.Done():
			break retry
		case <-time.After(time.Duration(attempt) * reportRetryDelay):
		}
	}

	if err != nil {
		entry := fw.logger.WithError(err).WithField("task_id", task.GetTaskID())
		if errors.Is(err, context.Canceled) {
			obsSpan.End(langfuse.WithSpanStatus(err.Error()), langfuse.WithSpanLevel(langfuse.ObservationLevelWarning))
			entry.Warn("the write-up was stopped before it finished")

			return
		}

		obsSpan.End(langfuse.WithSpanStatus(err.Error()), langfuse.WithSpanLevel(langfuse.ObservationLevelError))
		entry.Error("failed to write up the task on request")
		fw.reportWriteUpFailure(task, err)

		return
	}

	result, _ := task.GetResult(fw.ctx)
	obsSpan.End(langfuse.WithSpanOutput(result), langfuse.WithSpanStatus("success"))
}

func (fw *flowWorker) reportWriteUpFailure(task TaskWorker, err error) {
	if errors.Is(err, ErrTaskAlreadyCompleted) {
		return
	}

	_, putErr := fw.flowCtx.MsgLog.PutTaskMsg(context.WithoutCancel(fw.ctx), database.MsglogTypeReport,
		task.GetTaskID(), "", fmt.Sprintf("The report could not be written: %s", err))
	if putErr != nil {
		fw.logger.WithError(putErr).Warn("failed to report why the write-up did not arrive")
	}
}

// reportableTask picks the task a report would be about: the open one, running
// or waiting alike.
//
// A waiting task is deliberately accepted -- that is the whole case the
// endpoint exists for. Last rather than first, because a flow accumulates tasks
// and the one an operator means is the one they were just watching.
func reportableTask(tasks []TaskWorker) TaskWorker {
	for idx := len(tasks) - 1; idx >= 0; idx-- {
		if !tasks[idx].IsCompleted() {
			return tasks[idx]
		}
	}

	return nil
}

func (fw *flowWorker) Rename(ctx context.Context, title string) error {
	fw.flowCtx.Provider.SetTitle(title)

	flow, err := fw.flowCtx.DB.UpdateFlowTitle(ctx, database.UpdateFlowTitleParams{
		ID:    fw.flowCtx.FlowID,
		Title: title,
	})
	if err != nil {
		return fmt.Errorf("failed to rename flow %d: %w", fw.flowCtx.FlowID, err)
	}

	containers, err := fw.flowCtx.DB.GetFlowContainers(ctx, fw.flowCtx.FlowID)
	if err != nil {
		return fmt.Errorf("failed to get flow %d containers: %w", fw.flowCtx.FlowID, err)
	}

	fw.flowCtx.Publisher.FlowUpdated(ctx, flow, containers)

	return nil
}

// switchProvider performs runtime provider switch for the flow.
//
// This is the single place where a running flow picks up a provider change. A
// rename or deletion of a user provider only rewrites the DB reference (see
// flowController.reassignFlowsProvider); the in-memory instance is refreshed
// here, on the next user input, or rebuilt from the DB row on the next start.
//
// Deciding whether anything changed is delegated to SetProvider, which compares
// the raw configuration and not just the provider name — a user provider may be
// named exactly like a built-in one, so the name alone cannot tell an override
// apart from the default it shadows.
//
// Note on tool_call_id_template: it is resolved once, for a single model of the
// provider configuration. A provider that routes different models to different
// upstream backends (an OpenRouter-style gateway) may therefore need different
// templates per agent, and this single value can be wrong for some of them.
// Fixing that properly means keeping a per-model template registry, which is out
// of scope here; if real users hit it, this is the place to start.
func (fw *flowWorker) switchProvider(ctx context.Context, prv provider.Provider) error {
	ctx, span := obs.Observer.NewSpan(ctx, obs.SpanKindInternal, "controller.flowWorker.switchProvider")
	defer span.End()

	if prv == nil {
		return nil // no provider to switch to
	}

	logger := fw.logger.WithFields(logrus.Fields{
		"new_provider_name": prv.Name().String(),
		"new_provider_type": prv.Type().String(),
	})

	changed, tcIDTemplate, err := fw.flowCtx.Provider.SetProvider(ctx, prv)
	if err != nil {
		logger.WithError(err).Error("failed to set provider")
		return fmt.Errorf("failed to set provider: %w", err)
	}

	if !changed {
		logger.Debug("provider is the same, skipping switch")
		return nil
	}

	logger.Info("switching flow provider")

	// Every persisted value is taken from prv (and the template SetProvider
	// resolved for it) rather than re-read from the shared flow provider, so a
	// concurrent switch cannot interleave into a mixed-provider row.
	flow, err := fw.flowCtx.DB.UpdateFlowProvider(ctx, database.UpdateFlowProviderParams{
		ModelProviderName:  prv.Name().String(),
		ModelProviderType:  database.ProviderType(prv.Type()),
		ToolCallIDTemplate: tcIDTemplate,
		Model:              prv.Model(pconfig.OptionsTypePrimaryAgent),
		ID:                 fw.flowCtx.FlowID,
	})
	if err != nil {
		logger.WithError(err).Error("failed to update flow provider in DB")
		return fmt.Errorf("failed to update flow provider in DB: %w", err)
	}

	logger.WithFields(logrus.Fields{
		"new_tool_call_id_template": tcIDTemplate,
		"new_model":                 prv.Model(pconfig.OptionsTypePrimaryAgent),
	}).Info("provider switched successfully")

	if containers, err := fw.flowCtx.DB.GetFlowContainers(ctx, fw.flowCtx.FlowID); err == nil {
		fw.flowCtx.Publisher.FlowUpdated(ctx, flow, containers)
	}

	return nil
}

func (fw *flowWorker) finish() error {
	if err := fw.ctx.Err(); err != nil {
		if errors.Is(err, context.Canceled) {
			return nil
		}
		return fmt.Errorf("flow %d stop failed: %w", fw.flowCtx.FlowID, err)
	}

	fw.cancel()
	fw.stopReports()
	fw.wg.Wait()
	fw.waitForReports()

	return nil
}

func (fw *flowWorker) stopReports() {
	fw.reportMX.Lock()
	defer fw.reportMX.Unlock()

	if fw.reportST != nil {
		fw.reportST()
	}
}

// waitForReports drains an on-demand report that is still unwinding. Its run
// context is already cancelled by the caller, so this is bounded by how long
// the write-up takes to notice rather than by its own budget.
func (fw *flowWorker) waitForReports() {
	drained := make(chan struct{})
	go func() {
		defer close(drained)
		fw.reportWG.Wait()
	}()

	select {
	case <-drained:
	case <-time.After(reportDrainTimeout):
		fw.logger.Warn("a report was still running when the flow finished")
	}
}

// signalTaskComplete broadcasts task completion to all goroutines currently
// blocked in WaitTaskCompletion. It replaces the shared channel so that future
// callers block on a fresh channel until the next task finishes.
func (fw *flowWorker) signalTaskComplete() {
	fw.taskCMX.Lock()
	old := fw.taskCCH
	fw.taskCCH = make(chan struct{})
	fw.taskCMX.Unlock()
	close(old)
}

// WaitTaskCompletion blocks until the currently running task completes,
// the supplied context expires, or the flow worker itself is stopped.
// Multiple concurrent callers are all unblocked at once when a task finishes.
func (fw *flowWorker) WaitTaskCompletion(ctx context.Context) error {
	fw.taskCMX.Lock()
	ch := fw.taskCCH
	fw.taskCMX.Unlock()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-ch:
		return nil
	case <-fw.ctx.Done():
		return nil
	}
}

func (fw *flowWorker) worker() {
	defer fw.wg.Done()

	_, observation := obs.Observer.NewObservation(fw.ctx)

	getLogger := func(input string, task TaskWorker) *logrus.Entry {
		logger := fw.logger.WithField("input", input)
		if task != nil {
			logger = logger.WithFields(logrus.Fields{
				"task_id":       task.GetTaskID(),
				"task_complete": task.IsCompleted(),
				"task_waiting":  task.IsWaiting(),
				"task_title":    task.GetTitle(),
				"trace_id":      observation.TraceID(),
			})
		}
		return logger
	}

	// continue incomplete tasks after loading
	for _, task := range fw.tc.ListTasks(fw.ctx) {
		if !task.IsCompleted() && !task.IsWaiting() {
			input := "continue after loading"
			spanName := fmt.Sprintf("continue task %d: %s", task.GetTaskID(), task.GetTitle())
			if err := fw.reportFailure(task, input, fw.runTask(spanName, input, task)); err != nil {
				if errors.Is(err, context.Canceled) {
					getLogger(input, task).Info("flow are going to be stopped by user")
					return
				} else {
					getLogger(input, task).WithError(err).Error("failed to continue task")

					// anyway there need to set flow status to Waiting new user input even an error happened
					_ = fw.SetStatus(fw.ctx, database.FlowStatusWaiting)
				}
			} else {
				getLogger(input, task).Info("task continued successfully")
			}
		}
	}

	// process user input in regular job
	for {
		var flin flowInput
		select {
		case <-fw.ctx.Done():
			return
		case flin = <-fw.input:
		}

		if task, err := fw.processInput(flin); err != nil {
			if errors.Is(err, context.Canceled) {
				getLogger(flin.input, task).Info("flow are going to be stopped by user")
				return
			} else {
				getLogger(flin.input, task).WithError(err).Error("failed to process input")

				// anyway there need to set flow status to Waiting new user input even an error happened
				_ = fw.SetStatus(fw.ctx, database.FlowStatusWaiting)
			}
		} else {
			getLogger(flin.input, task).Info("user input processed")
		}
	}
}

func (fw *flowWorker) processInput(flin flowInput) (TaskWorker, error) {
	for _, task := range fw.tc.ListTasks(fw.ctx) {
		if !task.IsCompleted() && task.IsWaiting() {
			if err := task.PutInput(fw.ctx, flin.input); err != nil {
				err = fmt.Errorf("failed to process input to task %d: %w", task.GetTaskID(), err)
				if flin.reply.deliver(err) {
					return nil, err
				}
				return nil, fw.reportFailure(task, flin.input, err)
			}

			flin.reply.deliver(nil)
			return task, fw.reportFailure(task, flin.input, fw.runTask("put input to task and run", flin.input, task))
		}
	}

	// anyway there need to set flow status to Running to disable user input
	_ = fw.SetStatus(fw.ctx, database.FlowStatusRunning)

	// Pre-create the per-task cancellable context BEFORE calling CreateTask.
	// GenerateSubtasks (an LLM call) runs synchronously inside CreateTask and may take
	// many seconds. Without this, a concurrent Stop() would invoke a no-op taskST and
	// find taskWG at zero—reporting success while the generator is still running.
	fw.taskMX.Lock()
	fw.taskST()
	ctx, taskST := context.WithCancel(fw.ctx)
	fw.taskST = taskST
	fw.taskMX.Unlock()

	defer taskST()

	fw.taskWG.Add(1)
	defer fw.taskWG.Done()
	defer fw.signalTaskComplete()

	task, err := fw.tc.CreateTask(ctx, flin.input, fw)
	if err != nil {
		if errors.Is(err, context.Canceled) && fw.ctx.Err() == nil {
			// CreateTask was cancelled by Stop() — not a fatal flow error.
			// Keep the worker alive and return the flow to Waiting state.
			flin.reply.deliver(nil)
			_ = fw.SetStatus(fw.ctx, database.FlowStatusWaiting)
			return nil, nil
		}
		err = fmt.Errorf("failed to create task for flow %d: %w", fw.flowCtx.FlowID, err)
		if flin.reply.deliver(err) {
			return nil, err
		}
		return nil, fw.reportFailure(nil, flin.input, err)
	}

	flin.reply.deliver(nil)
	spanName := fmt.Sprintf("perform task %d: %s", task.GetTaskID(), task.GetTitle())
	return task, fw.reportFailure(task, flin.input, fw.execTask(ctx, spanName, flin.input, task))
}

func (fw *flowWorker) reportFailure(task TaskWorker, input string, err error) error {
	if err == nil || errors.Is(err, context.Canceled) {
		return err
	}

	var putErr error
	if task == nil {
		_, putErr = fw.flowCtx.MsgLog.PutFlowMsgResult(fw.ctx, database.MsglogTypeReport, "",
			fmt.Sprintf("The task could not start: %s", err), input, database.MsglogResultFormatPlain)
	} else {
		_, putErr = fw.flowCtx.MsgLog.PutTaskMsg(fw.ctx, database.MsglogTypeReport, task.GetTaskID(), "",
			fmt.Sprintf("The task stopped on an error: %s", err))
	}
	if putErr != nil {
		fw.logger.WithError(putErr).Warn("failed to report why the task did not go on")
	}

	return err
}

// runTask creates a fresh per-task cancellable context and runs an already-created task.
// Use this for tasks that were previously created and are being resumed (e.g. after waiting).
func (fw *flowWorker) runTask(spanName, input string, task TaskWorker) error {
	fw.taskMX.Lock()
	fw.taskST()
	ctx, taskST := context.WithCancel(fw.ctx)
	fw.taskST = taskST
	fw.taskMX.Unlock()

	defer taskST()

	fw.taskWG.Add(1)
	defer fw.taskWG.Done()
	defer fw.signalTaskComplete()

	return fw.execTask(ctx, spanName, input, task)
}

// execTask executes a task using an already-prepared context and cancel function.
func (fw *flowWorker) execTask(ctx context.Context, spanName, input string, task TaskWorker) error {
	_, observation := obs.Observer.NewObservation(fw.ctx)
	span := observation.Span(
		langfuse.WithSpanName(spanName),
		langfuse.WithSpanInput(input),
		langfuse.WithSpanMetadata(langfuse.Metadata{
			"task_id": task.GetTaskID(),
		}),
	)

	ctx, _ = span.Observation(ctx)

	if err := task.Run(ctx); err != nil {
		// if task is stopped by user and it's not finished yet
		if errors.Is(err, context.Canceled) && fw.ctx.Err() == nil {
			span.End(
				langfuse.WithSpanStatus("stopped"),
				langfuse.WithSpanLevel(langfuse.ObservationLevelWarning),
			)
			return nil
		}
		span.End(
			langfuse.WithSpanStatus(err.Error()),
			langfuse.WithSpanLevel(langfuse.ObservationLevelError),
		)
		return fmt.Errorf("failed to run task %d: %w", task.GetTaskID(), err)
	}

	result, _ := task.GetResult(fw.ctx)
	status, _ := task.GetStatus(fw.ctx)
	if status == database.TaskStatusFailed {
		span.End(
			langfuse.WithSpanOutput(result),
			langfuse.WithSpanStatus("failed"),
			langfuse.WithSpanLevel(langfuse.ObservationLevelWarning),
		)
	} else {
		span.End(
			langfuse.WithSpanOutput(result),
			langfuse.WithSpanStatus("success"),
		)
	}

	return nil
}

func newFlowProviderWorkers(
	ctx context.Context,
	flowID int64,
	cnts *flowProviderControllers,
	pub subscriptions.FlowPublisher,
) (*flowProviderWorkers, error) {
	alw, err := cnts.alc.NewFlowAgentLog(ctx, flowID, pub)
	if err != nil {
		return nil, fmt.Errorf("failed to create flow agent log: %w", err)
	}

	mlw, err := cnts.mlc.NewFlowMsgLog(ctx, flowID, pub)
	if err != nil {
		return nil, fmt.Errorf("failed to create flow msg log: %w", err)
	}

	slw, err := cnts.slc.NewFlowSearchLog(ctx, flowID, pub)
	if err != nil {
		return nil, fmt.Errorf("failed to create flow search log: %w", err)
	}

	tlw, err := cnts.tlc.NewFlowTermLog(ctx, flowID, pub)
	if err != nil {
		return nil, fmt.Errorf("failed to create flow term log: %w", err)
	}

	vslw, err := cnts.vslc.NewFlowVectorStoreLog(ctx, flowID, pub)
	if err != nil {
		return nil, fmt.Errorf("failed to create flow vector store log: %w", err)
	}

	tclw, err := cnts.tclc.NewFlowToolCallLog(ctx, flowID, pub)
	if err != nil {
		return nil, fmt.Errorf("failed to create flow tool call log: %w", err)
	}

	sw, err := cnts.sc.NewFlowScreenshot(ctx, flowID, pub)
	if err != nil {
		return nil, fmt.Errorf("failed to create flow screenshot: %w", err)
	}

	return &flowProviderWorkers{
		mlw:  mlw,
		alw:  alw,
		slw:  slw,
		tlw:  tlw,
		vslw: vslw,
		tclw: tclw,
		sw:   sw,
	}, nil
}

func getFlowProviderWorkers(
	ctx context.Context,
	flowID int64,
	cnts *flowProviderControllers,
) (*flowProviderWorkers, error) {
	alw, err := cnts.alc.GetFlowAgentLog(ctx, flowID)
	if err != nil {
		return nil, fmt.Errorf("failed to get flow agent log: %w", err)
	}

	mlw, err := cnts.mlc.GetFlowMsgLog(ctx, flowID)
	if err != nil {
		return nil, fmt.Errorf("failed to get flow msg log: %w", err)
	}

	slw, err := cnts.slc.GetFlowSearchLog(ctx, flowID)
	if err != nil {
		return nil, fmt.Errorf("failed to get flow search log: %w", err)
	}

	tlw, err := cnts.tlc.GetFlowTermLog(ctx, flowID)
	if err != nil {
		return nil, fmt.Errorf("failed to get flow term log: %w", err)
	}

	vslw, err := cnts.vslc.GetFlowVectorStoreLog(ctx, flowID)
	if err != nil {
		return nil, fmt.Errorf("failed to get flow vector store log: %w", err)
	}

	tclw, err := cnts.tclc.GetFlowToolCallLog(ctx, flowID)
	if err != nil {
		return nil, fmt.Errorf("failed to get flow tool call log: %w", err)
	}

	sw, err := cnts.sc.GetFlowScreenshot(ctx, flowID)
	if err != nil {
		return nil, fmt.Errorf("failed to get flow screenshot: %w", err)
	}

	return &flowProviderWorkers{
		mlw:  mlw,
		alw:  alw,
		slw:  slw,
		tlw:  tlw,
		vslw: vslw,
		tclw: tclw,
		sw:   sw,
	}, nil
}
