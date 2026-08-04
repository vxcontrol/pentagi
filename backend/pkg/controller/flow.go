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
	Rename(ctx context.Context, title string) error
	WaitTaskCompletion(ctx context.Context) error
	InvalidateTaskSubtasks(ctx context.Context, taskID int64, subtaskIDs []int64)
}

type flowWorker struct {
	tc      TaskController
	wg      *sync.WaitGroup
	cfg     *config.Config
	aws     map[int64]AssistantWorker
	awsMX   *sync.Mutex
	ctx     context.Context
	cancel  context.CancelFunc
	taskMX  *sync.Mutex
	taskST  context.CancelFunc
	taskWG  *sync.WaitGroup
	taskCMX sync.Mutex
	taskCCH chan struct{}
	input   chan flowInput
	flowCtx *FlowContext
	docker  docker.DockerClient
	logger  *logrus.Entry
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

type flowInput struct {
	input string
	done  chan error
}

func NewFlowWorker(
	ctx context.Context,
	fwc newFlowWorkerCtx,
) (_ FlowWorker, err error) {
	ctx, span := obs.Observer.NewSpan(ctx, obs.SpanKindInternal, "controller.NewFlowWorker")
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
		return nil, fmt.Errorf("failed to create flow in DB: %w", err)
	}

	logger := logrus.WithContext(ctx).WithFields(logrus.Fields{
		"flow_id":       flow.ID,
		"user_id":       fwc.userID,
		"provider_name": fwc.prvname.String(),
		"provider_type": fwc.prvtype.String(),
	})
	logger.Info("flow created in DB")

	// Held separately because `flow` is reassigned below by UpdateFlow, which — like every sqlc query —
	// returns a zero-valued row alongside an error, and a cleanup aimed at id 0 deletes nothing.
	flowID := flow.ID

	// DeleteFlow is a soft delete and the listings filter on deleted_at, so this is what keeps a flow
	// the caller was told it never got out of the UI. Disarmed once the worker goroutine owns the flow —
	// from there a failure is the worker's to unwind, not ours.
	cleanupFlow := true
	defer func() {
		if err == nil || !cleanupFlow {
			return
		}

		// The caller's context is usually already cancelled by whatever failed.
		if _, cerr := fwc.db.DeleteFlow(context.WithoutCancel(ctx), flowID); cerr != nil {
			logger.WithError(cerr).Error("failed to drop the flow left behind by a failed start")
		}
	}()

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

	containers, err := fwc.db.GetFlowContainers(ctx, flow.ID)
	if err != nil {
		return nil, wrapErrorEndSpan(ctx, flowSpan, "failed to get flow containers", err)
	}

	fw.flowCtx.Publisher.FlowCreated(ctx, flow, containers)

	cleanupFlow = false

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

	if err := fw.switchProvider(ctx, prv); err != nil {
		return fmt.Errorf("failed to switch provider: %w", err)
	}

	if err := fw.PutResources(ctx, resources); err != nil {
		fw.logger.WithError(err).Warn("failed to copy resources before user input")
	}

	flin := flowInput{input: input, done: make(chan error, 1)}
	select {
	case <-fw.ctx.Done():
		close(flin.done)
		return fmt.Errorf("flow %d stopped: %w", fw.flowCtx.FlowID, fw.ctx.Err())
	case <-ctx.Done():
		close(flin.done)
		return fmt.Errorf("flow %d input processing timeout: %w", fw.flowCtx.FlowID, ctx.Err())
	case fw.input <- flin:
		timer := time.NewTimer(flowInputTimeout)
		defer timer.Stop()

		select {
		case err := <-flin.done:
			return err
		case <-timer.C:
			return nil // no early error
		case <-fw.ctx.Done():
			return fmt.Errorf("flow %d stopped: %w", fw.flowCtx.FlowID, fw.ctx.Err())
		case <-ctx.Done():
			return fmt.Errorf("flow %d input processing timeout: %w", fw.flowCtx.FlowID, ctx.Err())
		}
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

	if err := fw.SetStatus(ctx, database.FlowStatusFinished); err != nil {
		return fmt.Errorf("failed to set flow %d status: %w", fw.flowCtx.FlowID, err)
	}

	return nil
}

func (fw *flowWorker) Stop(ctx context.Context) error {
	ctx, span := obs.Observer.NewSpan(ctx, obs.SpanKindInternal, "controller.flowWorker.Stop")
	defer span.End()

	fw.taskMX.Lock()
	defer fw.taskMX.Unlock()

	fw.taskST()
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
	close(fw.input)
	fw.wg.Wait()

	return nil
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
			if err := fw.runTask(spanName, input, task); err != nil {
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
	for flin := range fw.input {
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
				flin.done <- err
				return nil, err
			} else {
				flin.done <- nil
				return task, fw.runTask("put input to task and run", flin.input, task)
			}
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
			flin.done <- nil
			_ = fw.SetStatus(fw.ctx, database.FlowStatusWaiting)
			return nil, nil
		}
		err = fmt.Errorf("failed to create task for flow %d: %w", fw.flowCtx.FlowID, err)
		flin.done <- err
		return nil, err
	}

	flin.done <- nil
	spanName := fmt.Sprintf("perform task %d: %s", task.GetTaskID(), task.GetTitle())
	return task, fw.execTask(ctx, spanName, flin.input, task)
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
