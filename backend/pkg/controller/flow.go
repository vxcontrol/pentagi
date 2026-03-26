package controller

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"slices"
	"sync"
	"time"

	"pentagi/pkg/cast"
	"pentagi/pkg/config"
	"pentagi/pkg/database"
	"pentagi/pkg/docker"
	"pentagi/pkg/graph/subscriptions"
	obs "pentagi/pkg/observability"
	"pentagi/pkg/observability/langfuse"
	"pentagi/pkg/providers"
	"pentagi/pkg/providers/pconfig"
	"pentagi/pkg/providers/provider"
	"pentagi/pkg/templates"
	"pentagi/pkg/tools"

	"github.com/sirupsen/logrus"
)

const stopTaskTimeout = 5 * time.Second

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
	PutInput(ctx context.Context, input string) error
	Finish(ctx context.Context) error
	Stop(ctx context.Context) error
	Rename(ctx context.Context, title string) error
}

type flowWorker struct {
	tc      TaskController
	wg      *sync.WaitGroup
	aws     map[int64]AssistantWorker
	awsMX   *sync.Mutex
	ctx     context.Context
	cancel  context.CancelFunc
	taskMX  *sync.Mutex
	taskST  context.CancelFunc
	taskWG  *sync.WaitGroup
	input   chan flowInput
	flowCtx *FlowContext
	logger  *logrus.Entry
}

type newFlowWorkerCtx struct {
	userID    int64
	input     string
	dryRun    bool
	prvname   provider.ProviderName
	prvtype   provider.ProviderType
	functions *tools.Functions

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
	sc   ScreenshotController
}

type flowProviderWorkers struct {
	mlw  FlowMsgLogWorker
	alw  FlowAgentLogWorker
	slw  FlowSearchLogWorker
	tlw  FlowTermLogWorker
	vslw FlowVectorStoreLogWorker
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
) (FlowWorker, error) {
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
		logrus.WithError(err).Error("failed to create flow in DB")
		return nil, fmt.Errorf("failed to create flow in DB: %w", err)
	}

	logger := logrus.WithContext(ctx).WithFields(logrus.Fields{
		"flow_id":       flow.ID,
		"user_id":       fwc.userID,
		"provider_name": fwc.prvname.String(),
		"provider_type": fwc.prvtype.String(),
	})
	logger.Info("flow created in DB")

	user, err := fwc.db.GetUser(ctx, fwc.userID)
	if err != nil {
		logger.WithError(err).Error("failed to get user")
		return nil, fmt.Errorf("failed to get user %d: %w", fwc.userID, err)
	}

	ctx, observation := obs.Observer.NewObservation(ctx,
		langfuse.WithObservationTraceContext(
			langfuse.WithTraceName(fmt.Sprintf("%d flow worker", flow.ID)),
			langfuse.WithTraceUserID(user.Mail),
			langfuse.WithTraceTags([]string{"controller", "flow"}),
			langfuse.WithTraceInput(fwc.input),
			langfuse.WithTraceSessionID(fmt.Sprintf("flow-%d", flow.ID)),
			langfuse.WithTraceMetadata(langfuse.Metadata{
				"flow_id":       flow.ID,
				"user_id":       fwc.userID,
				"user_email":    user.Mail,
				"user_name":     user.Name,
				"user_hash":     user.Hash,
				"user_role":     user.RoleName,
				"provider_name": fwc.prvname.String(),
				"provider_type": fwc.prvtype.String(),
			}),
		),
	)
	flowSpan := observation.Span(langfuse.WithSpanName("prepare flow worker"))
	ctx, _ = flowSpan.Observation(ctx)

	prompter := templates.NewDefaultPrompter() // TODO: change to flow prompter by userID from DB
	executor, err := tools.NewFlowToolsExecutor(fwc.db, fwc.cfg, fwc.docker, fwc.functions, flow.ID)
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
	executor.SetGraphitiClient(fwc.provs.GraphitiClient())

	flowCtx := &FlowContext{
		DB:         fwc.db,
		UserID:     fwc.userID,
		FlowID:     flow.ID,
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
		ctx:     ctx,
		cancel:  cancel,
		taskMX:  &sync.Mutex{},
		taskST:  func() {},
		taskWG:  &sync.WaitGroup{},
		input:   make(chan flowInput),
		flowCtx: flowCtx,
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

	fw.wg.Add(1)
	go fw.worker()

	if !fwc.dryRun {
		if err := fw.PutInput(ctx, fwc.input); err != nil {
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
			langfuse.WithTraceName(fmt.Sprintf("%d flow worker", flow.ID)),
			langfuse.WithTraceUserID(user.Mail),
			langfuse.WithTraceTags([]string{"controller", "flow"}),
			langfuse.WithTraceSessionID(fmt.Sprintf("flow-%d", flow.ID)),
			langfuse.WithTraceMetadata(langfuse.Metadata{
				"flow_id":       flow.ID,
				"user_id":       flow.UserID,
				"user_email":    user.Mail,
				"user_name":     user.Name,
				"user_hash":     user.Hash,
				"user_role":     user.RoleName,
				"provider_name": flow.ModelProviderName,
				"provider_type": flow.ModelProviderType,
			}),
		),
	)
	flowSpan := observation.Span(langfuse.WithSpanName("prepare flow worker"))
	ctx, _ = flowSpan.Observation(ctx)

	functions := &tools.Functions{}
	if err := json.Unmarshal(flow.Functions, functions); err != nil {
		return nil, wrapErrorEndSpan(ctx, flowSpan, "failed to unmarshal functions", err)
	}

	prompter := templates.NewDefaultPrompter() // TODO: change to flow prompter by userID from DB
	executor, err := tools.NewFlowToolsExecutor(fwc.db, fwc.cfg, fwc.docker, functions, flow.ID)
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
	executor.SetGraphitiClient(fwc.provs.GraphitiClient())

	flowCtx := &FlowContext{
		DB:         fwc.db,
		UserID:     flow.UserID,
		FlowID:     flow.ID,
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
		ctx:     ctx,
		cancel:  cancel,
		taskMX:  &sync.Mutex{},
		taskST:  func() {},
		taskWG:  &sync.WaitGroup{},
		input:   make(chan flowInput),
		flowCtx: flowCtx,
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
		flowWorkerCtx: fwc,
	}
	for _, assistant := range assistants {
		aw, err := LoadAssistantWorker(ctx, assistant, awc)
		if err != nil {
			if errors.Is(err, ErrNothingToLoad) {
				continue
			}
			return nil, wrapErrorEndSpan(ctx, flowSpan, "failed to load assistant worker", err)
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

func (fw *flowWorker) PutInput(ctx context.Context, input string) error {
	ctx, span := obs.Observer.NewSpan(ctx, obs.SpanKindInternal, "controller.flowWorker.PutInput")
	defer span.End()

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
			return err // nil or error
		case <-timer.C:
			return nil // no early error
		case <-fw.ctx.Done():
			return fmt.Errorf("flow %d stopped: %w", fw.flowCtx.FlowID, fw.ctx.Err())
		case <-ctx.Done():
			return fmt.Errorf("flow %d input processing timeout: %w", fw.flowCtx.FlowID, ctx.Err())
		}
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

	if task, err := fw.tc.CreateTask(fw.ctx, flin.input, fw); err != nil {
		err = fmt.Errorf("failed to create task for flow %d: %w", fw.flowCtx.FlowID, err)
		flin.done <- err
		return nil, err
	} else {
		flin.done <- nil
		spanName := fmt.Sprintf("perform task %d: %s", task.GetTaskID(), task.GetTitle())
		return task, fw.runTask(spanName, flin.input, task)
	}
}

func (fw *flowWorker) runTask(spanName, input string, task TaskWorker) error {
	_, observation := obs.Observer.NewObservation(fw.ctx)
	span := observation.Span(
		langfuse.WithSpanName(spanName),
		langfuse.WithSpanInput(input),
		langfuse.WithSpanMetadata(langfuse.Metadata{
			"task_id": task.GetTaskID(),
		}),
	)

	fw.taskMX.Lock()
	fw.taskST()
	ctx, taskST := context.WithCancel(fw.ctx)
	fw.taskST = taskST
	fw.taskMX.Unlock()

	ctx, _ = span.Observation(ctx)
	defer taskST()

	fw.taskWG.Add(1)
	defer fw.taskWG.Done()

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
		sw:   sw,
	}, nil
}
