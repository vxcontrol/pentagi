package controller

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"

	"pentagi/pkg/cast"
	"pentagi/pkg/database"
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

const stopAssistantTimeout = 5 * time.Second

type AssistantWorker interface {
	GetAssistantID() int64
	GetUserID() int64
	GetFlowID() int64
	GetTitle() string
	GetStatus(ctx context.Context) (database.AssistantStatus, error)
	SetStatus(ctx context.Context, status database.AssistantStatus) error
	PutInput(ctx context.Context, input string, useAgents bool) error
	Finish(ctx context.Context) error
	Stop(ctx context.Context) error
}

type assistantWorker struct {
	id      int64
	flowID  int64
	userID  int64
	chainID int64
	aslw    FlowAssistantLogWorker
	ap      providers.AssistantProvider
	db      database.Querier
	wg      *sync.WaitGroup
	pub     subscriptions.FlowPublisher
	ctx     context.Context
	cancel  context.CancelFunc
	runMX   *sync.Mutex
	runST   context.CancelFunc
	runWG   *sync.WaitGroup
	input   chan assistantInput
	logger  *logrus.Entry
}

type newAssistantWorkerCtx struct {
	userID    int64
	flowID    int64
	input     string
	useAgents bool
	prvname   provider.ProviderName
	prvtype   provider.ProviderType
	functions *tools.Functions

	flowWorkerCtx
}

type assistantWorkerCtx struct {
	userID int64
	flowID int64

	flowWorkerCtx
}

const assistantInputTimeout = 2 * time.Second

type assistantInput struct {
	input     string
	useAgents bool
	done      chan error
}

func NewAssistantWorker(ctx context.Context, awc newAssistantWorkerCtx) (AssistantWorker, error) {
	ctx, span := obs.Observer.NewSpan(ctx, obs.SpanKindInternal, "controller.NewAssistantWorker")
	defer span.End()

	logger := logrus.WithContext(ctx).WithFields(logrus.Fields{
		"flow_id":       awc.flowID,
		"user_id":       awc.userID,
		"provider_name": awc.prvname.String(),
		"provider_type": awc.prvtype.String(),
	})

	user, err := awc.db.GetUser(ctx, awc.userID)
	if err != nil {
		logger.WithError(err).Error("failed to get user")
		return nil, fmt.Errorf("failed to get user %d: %w", awc.userID, err)
	}

	container, err := awc.db.GetFlowPrimaryContainer(ctx, awc.flowID)
	if err != nil {
		logger.WithError(err).Error("failed to get flow primary container")
		return nil, fmt.Errorf("failed to get flow primary container: %w", err)
	}

	assistant, err := awc.db.CreateAssistant(ctx, database.CreateAssistantParams{
		Title:              "untitled",
		Status:             database.AssistantStatusCreated,
		Model:              "unknown",
		ModelProviderName:  string(awc.prvname),
		ModelProviderType:  database.ProviderType(awc.prvtype),
		Language:           "English",
		ToolCallIDTemplate: cast.ToolCallIDTemplate,
		Functions:          []byte("{}"),
		FlowID:             awc.flowID,
		UseAgents:          awc.useAgents,
	})
	if err != nil {
		logger.WithError(err).Error("failed to create assistant in DB")
		return nil, fmt.Errorf("failed to create assistant in DB: %w", err)
	}

	logger = logger.WithField("assistant_id", assistant.ID)
	logger.Info("assistant created in DB")

	ctx, observation := obs.Observer.NewObservation(ctx,
		langfuse.WithObservationTraceContext(
			langfuse.WithTraceName(fmt.Sprintf("%d flow %d assistant worker", awc.flowID, assistant.ID)),
			langfuse.WithTraceUserID(user.Mail),
			langfuse.WithTraceTags([]string{"controller", "assistant"}),
			langfuse.WithTraceInput(awc.input),
			langfuse.WithTraceSessionID(fmt.Sprintf("assistant-%d-flow-%d", assistant.ID, awc.flowID)),
			langfuse.WithTraceMetadata(langfuse.Metadata{
				"assistant_id":  assistant.ID,
				"flow_id":       awc.flowID,
				"user_id":       awc.userID,
				"user_email":    user.Mail,
				"user_name":     user.Name,
				"user_hash":     user.Hash,
				"user_role":     user.RoleName,
				"provider_name": awc.prvname.String(),
				"provider_type": awc.prvtype.String(),
			}),
		),
	)
	assistantSpan := observation.Span(langfuse.WithSpanName("prepare assistant worker"))
	ctx, _ = assistantSpan.Observation(ctx)

	pub := awc.subs.NewFlowPublisher(awc.userID, awc.flowID)
	aslw, err := awc.aslc.NewFlowAssistantLog(ctx, awc.flowID, assistant.ID, pub)
	if err != nil {
		return nil, wrapErrorEndSpan(ctx, assistantSpan, "failed to create flow assistant log worker", err)
	}

	prompter := templates.NewDefaultPrompter() // TODO: change to flow prompter by userID from DB
	executor, err := tools.NewFlowToolsExecutor(awc.db, awc.cfg, awc.docker, awc.functions, awc.flowID)
	if err != nil {
		return nil, wrapErrorEndSpan(ctx, assistantSpan, "failed to create flow tools executor", err)
	}
	assistantProvider, err := awc.provs.NewAssistantProvider(ctx, awc.prvname, prompter, executor,
		assistant.ID, awc.flowID, awc.userID, container.Image, awc.input, aslw.StreamFlowAssistantMsg)
	if err != nil {
		return nil, wrapErrorEndSpan(ctx, assistantSpan, "failed to get assistant provider", err)
	}

	msgChainID, err := assistantProvider.PrepareAgentChain(ctx)
	if err != nil {
		return nil, wrapErrorEndSpan(ctx, assistantSpan, "failed to prepare assistant chain", err)
	}

	functionsBlob, err := json.Marshal(awc.functions)
	if err != nil {
		return nil, wrapErrorEndSpan(ctx, assistantSpan, "failed to marshal functions", err)
	}

	logger = logger.WithField("msg_chain_id", msgChainID)
	logger.Info("assistant provider prepared")

	assistant, err = awc.db.UpdateAssistant(ctx, database.UpdateAssistantParams{
		Title:              assistantProvider.Title(),
		Model:              assistantProvider.Model(pconfig.OptionsTypePrimaryAgent),
		Language:           assistantProvider.Language(),
		ToolCallIDTemplate: assistantProvider.ToolCallIDTemplate(),
		Functions:          functionsBlob,
		TraceID:            database.StringToNullString(observation.TraceID()),
		MsgchainID:         database.Int64ToNullInt64(&msgChainID),
		ID:                 assistant.ID,
	})
	if err != nil {
		logger.WithError(err).Error("failed to create assistant in DB")
		return nil, fmt.Errorf("failed to create assistant in DB: %w", err)
	}

	workers, err := getFlowProviderWorkers(ctx, awc.flowID, &awc.flowProviderControllers)
	if err != nil {
		return nil, wrapErrorEndSpan(ctx, assistantSpan, "failed to get flow provider workers", err)
	}

	assistantProvider.SetAgentLogProvider(workers.alw)
	assistantProvider.SetMsgLogProvider(aslw)

	executor.SetImage(container.Image)
	executor.SetEmbedder(assistantProvider.Embedder())
	executor.SetScreenshotProvider(workers.sw)
	executor.SetAgentLogProvider(workers.alw)
	executor.SetMsgLogProvider(aslw)
	executor.SetSearchLogProvider(workers.slw)
	executor.SetTermLogProvider(workers.tlw)
	executor.SetVectorStoreLogProvider(workers.vslw)
	executor.SetGraphitiClient(awc.provs.GraphitiClient())

	ctx, cancel := context.WithCancel(context.Background())
	ctx, _ = obs.Observer.NewObservation(ctx, langfuse.WithObservationTraceID(observation.TraceID()))
	aw := &assistantWorker{
		id:      assistant.ID,
		flowID:  awc.flowID,
		userID:  awc.userID,
		chainID: msgChainID,
		aslw:    aslw,
		ap:      assistantProvider,
		db:      awc.db,
		wg:      &sync.WaitGroup{},
		pub:     pub,
		ctx:     ctx,
		cancel:  cancel,
		runMX:   &sync.Mutex{},
		runST:   func() {},
		runWG:   &sync.WaitGroup{},
		input:   make(chan assistantInput),
		logger: logrus.WithFields(logrus.Fields{
			"msg_chain_id": msgChainID,
			"assistant_id": assistant.ID,
			"flow_id":      awc.flowID,
			"user_id":      awc.userID,
			"trace_id":     observation.TraceID(),
			"component":    "assistant",
		}),
	}

	pub.AssistantCreated(ctx, assistant)

	aw.wg.Add(1)
	go aw.worker()

	if err := aw.PutInput(ctx, awc.input, awc.useAgents); err != nil {
		return nil, wrapErrorEndSpan(ctx, assistantSpan, "failed to run assistant worker", err)
	}

	assistantSpan.End(langfuse.WithSpanStatus("assistant worker started"))

	return aw, nil
}

func LoadAssistantWorker(
	ctx context.Context, assistant database.Assistant, awc assistantWorkerCtx,
) (AssistantWorker, error) {
	ctx, span := obs.Observer.NewSpan(ctx, obs.SpanKindInternal, "controller.LoadAssistantWorker")
	defer span.End()

	switch assistant.Status {
	case database.AssistantStatusRunning, database.AssistantStatusWaiting:
	default:
		return nil, fmt.Errorf("assistant %d has status %s: loading aborted: %w", assistant.ID, assistant.Status, ErrNothingToLoad)
	}

	logger := logrus.WithContext(ctx).WithFields(logrus.Fields{
		"assistant_id":  assistant.ID,
		"flow_id":       awc.flowID,
		"user_id":       awc.userID,
		"msg_chain_id":  assistant.MsgchainID,
		"provider_name": assistant.ModelProviderName,
		"provider_type": assistant.ModelProviderType,
	})

	user, err := awc.db.GetUser(ctx, awc.userID)
	if err != nil {
		logger.WithError(err).Error("failed to get user")
		return nil, fmt.Errorf("failed to get user %d: %w", awc.userID, err)
	}

	container, err := awc.db.GetFlowPrimaryContainer(ctx, awc.flowID)
	if err != nil {
		logger.WithError(err).Error("failed to get flow primary container")
		return nil, fmt.Errorf("failed to get flow primary container: %w", err)
	}

	ctx, observation := obs.Observer.NewObservation(ctx,
		langfuse.WithObservationTraceContext(
			langfuse.WithTraceName(fmt.Sprintf("%d flow %d assistant worker", awc.flowID, assistant.ID)),
			langfuse.WithTraceUserID(user.Mail),
			langfuse.WithTraceTags([]string{"controller", "assistant"}),
			langfuse.WithTraceSessionID(fmt.Sprintf("assistant-%d-flow-%d", assistant.ID, awc.flowID)),
			langfuse.WithTraceMetadata(langfuse.Metadata{
				"assistant_id":  assistant.ID,
				"flow_id":       awc.flowID,
				"user_id":       awc.userID,
				"user_email":    user.Mail,
				"user_name":     user.Name,
				"user_hash":     user.Hash,
				"user_role":     user.RoleName,
				"provider_name": assistant.ModelProviderName,
				"provider_type": assistant.ModelProviderType,
			}),
		),
	)
	assistantSpan := observation.Span(langfuse.WithSpanName("prepare assistant worker"))
	ctx, _ = assistantSpan.Observation(ctx)

	functions := &tools.Functions{}
	if err := json.Unmarshal(assistant.Functions, functions); err != nil {
		return nil, wrapErrorEndSpan(ctx, assistantSpan, "failed to unmarshal functions", err)
	}

	pub := awc.subs.NewFlowPublisher(awc.userID, awc.flowID)
	aslw, err := awc.aslc.NewFlowAssistantLog(ctx, awc.flowID, assistant.ID, pub)
	if err != nil {
		return nil, wrapErrorEndSpan(ctx, assistantSpan, "failed to create flow assistant log worker", err)
	}

	prompter := templates.NewDefaultPrompter() // TODO: change to flow prompter by userID from DB
	executor, err := tools.NewFlowToolsExecutor(awc.db, awc.cfg, awc.docker, functions, awc.flowID)
	if err != nil {
		return nil, wrapErrorEndSpan(ctx, assistantSpan, "failed to create flow tools executor", err)
	}
	assistantProvider, err := awc.provs.LoadAssistantProvider(ctx, provider.ProviderName(assistant.ModelProviderName),
		prompter, executor, assistant.ID, awc.flowID, awc.userID, container.Image, assistant.Language, assistant.Title,
		assistant.ToolCallIDTemplate, aslw.StreamFlowAssistantMsg)
	if err != nil {
		return nil, wrapErrorEndSpan(ctx, assistantSpan, "failed to get assistant provider", err)
	}

	workers, err := getFlowProviderWorkers(ctx, awc.flowID, &awc.flowProviderControllers)
	if err != nil {
		return nil, wrapErrorEndSpan(ctx, assistantSpan, "failed to get flow provider workers", err)
	}

	assistantProvider.SetAgentLogProvider(workers.alw)
	assistantProvider.SetMsgLogProvider(aslw)

	executor.SetImage(container.Image)
	executor.SetEmbedder(assistantProvider.Embedder())
	executor.SetScreenshotProvider(workers.sw)
	executor.SetAgentLogProvider(workers.alw)
	executor.SetMsgLogProvider(aslw)
	executor.SetSearchLogProvider(workers.slw)
	executor.SetTermLogProvider(workers.tlw)
	executor.SetVectorStoreLogProvider(workers.vslw)

	var msgChainID int64
	pmsgChainID := database.NullInt64ToInt64(assistant.MsgchainID)
	if pmsgChainID != nil {
		msgChainID = *pmsgChainID
		assistantProvider.SetMsgChainID(msgChainID)
	} else {
		return nil, fmt.Errorf("assistant %d has no msgchain id", assistant.ID)
	}

	ctx, cancel := context.WithCancel(context.Background())
	ctx, _ = obs.Observer.NewObservation(ctx, langfuse.WithObservationTraceID(observation.TraceID()))
	aw := &assistantWorker{
		id:      assistant.ID,
		flowID:  awc.flowID,
		userID:  awc.userID,
		chainID: msgChainID,
		aslw:    aslw,
		ap:      assistantProvider,
		db:      awc.db,
		wg:      &sync.WaitGroup{},
		pub:     pub,
		ctx:     ctx,
		cancel:  cancel,
		runMX:   &sync.Mutex{},
		runST:   func() {},
		runWG:   &sync.WaitGroup{},
		input:   make(chan assistantInput),
		logger: logrus.WithFields(logrus.Fields{
			"msg_chain_id": msgChainID,
			"assistant_id": assistant.ID,
			"flow_id":      awc.flowID,
			"user_id":      awc.userID,
			"trace_id":     observation.TraceID(),
			"component":    "assistant",
		}),
	}

	assistant, err = awc.db.UpdateAssistantStatus(ctx, database.UpdateAssistantStatusParams{
		Status: database.AssistantStatusWaiting,
		ID:     assistant.ID,
	})
	if err != nil {
		return nil, wrapErrorEndSpan(ctx, assistantSpan, "failed to update assistant status", err)
	}

	pub.AssistantUpdated(ctx, assistant)

	aw.wg.Add(1)
	go aw.worker()

	assistantSpan.End(langfuse.WithSpanStatus("assistant worker started"))

	return aw, nil
}

func (aw *assistantWorker) worker() {
	defer aw.wg.Done()

	perform := func(ctx context.Context, input string, useAgents bool) error {
		aw.runWG.Add(1)
		defer aw.runWG.Done()

		_, err := aw.db.UpdateAssistantUseAgents(ctx, database.UpdateAssistantUseAgentsParams{
			UseAgents: useAgents,
			ID:        aw.id,
		})
		if err != nil {
			return fmt.Errorf("failed to update assistant use agents: %w", err)
		}

		if err := aw.SetStatus(ctx, database.AssistantStatusRunning); err != nil {
			aw.logger.WithError(err).Error("failed to set assistant status to waiting")
		}

		defer func() {
			if err := aw.SetStatus(ctx, database.AssistantStatusWaiting); err != nil {
				aw.logger.WithError(err).Error("failed to set assistant status to waiting")
			}
		}()

		_, err = aw.aslw.PutFlowAssistantMsg(ctx, database.MsglogTypeInput, "", input)
		if err != nil {
			return fmt.Errorf("failed to put input to flow assistant log: %w", err)
		}

		aw.runMX.Lock()
		ctx, aw.runST = context.WithCancel(aw.ctx)
		aw.runMX.Unlock()

		if err := aw.ap.PutInputToAgentChain(ctx, input); err != nil {
			return fmt.Errorf("failed to put input to agent chain: %w", err)
		}

		if err := aw.ap.PerformAgentChain(ctx); err != nil {
			if errors.Is(err, context.Canceled) {
				ctx = context.Background()
			}
			errChainConsistency := aw.ap.EnsureChainConsistency(ctx)
			if errChainConsistency != nil {
				err = errors.Join(err, errChainConsistency)
			}
			return fmt.Errorf("failed to perform agent chain: %w", err)
		}

		return nil
	}

	for {
		select {
		case <-aw.ctx.Done():
			return
		case ain := <-aw.input:
			err := perform(aw.ctx, ain.input, ain.useAgents)
			if err != nil {
				aw.logger.WithError(err).Error("failed to perform assistant chain")
			}
			ain.done <- err
		}
	}
}

func (aw *assistantWorker) GetAssistantID() int64 {
	return aw.id
}

func (aw *assistantWorker) GetUserID() int64 {
	return aw.userID
}

func (aw *assistantWorker) GetFlowID() int64 {
	return aw.flowID
}

func (aw *assistantWorker) GetTitle() string {
	return aw.ap.Title()
}

func (aw *assistantWorker) GetStatus(ctx context.Context) (database.AssistantStatus, error) {
	assistant, err := aw.db.GetAssistant(ctx, aw.id)
	if err != nil {
		return database.AssistantStatusFailed, err
	}

	return assistant.Status, nil
}

func (aw *assistantWorker) SetStatus(ctx context.Context, status database.AssistantStatus) error {
	assistant, err := aw.db.UpdateAssistantStatus(ctx, database.UpdateAssistantStatusParams{
		Status: status,
		ID:     aw.id,
	})
	if err != nil {
		return fmt.Errorf("failed to update assistant %d flow %d status: %w", aw.id, aw.flowID, err)
	}

	aw.pub.AssistantUpdated(ctx, assistant)

	return nil
}

func (aw *assistantWorker) PutInput(ctx context.Context, input string, useAgents bool) error {
	ctx, span := obs.Observer.NewSpan(ctx, obs.SpanKindInternal, "controller.assistantWorker.PutInput")
	defer span.End()

	ain := assistantInput{input: input, useAgents: useAgents, done: make(chan error, 1)}
	select {
	case <-aw.ctx.Done():
		close(ain.done)
		return fmt.Errorf("assistant %d flow %d stopped: %w", aw.id, aw.flowID, aw.ctx.Err())
	case <-ctx.Done():
		close(ain.done)
		return fmt.Errorf("assistant %d flow %d input processing timeout: %w", aw.id, aw.flowID, ctx.Err())
	case aw.input <- ain:
		timer := time.NewTimer(assistantInputTimeout)
		defer timer.Stop()

		select {
		case err := <-ain.done:
			return err // nil or error
		case <-timer.C:
			return nil // no early error
		case <-aw.ctx.Done():
			return fmt.Errorf("assistant %d flow %d stopped: %w", aw.id, aw.flowID, aw.ctx.Err())
		case <-ctx.Done():
			return fmt.Errorf("assistant %d flow %d input processing timeout: %w", aw.id, aw.flowID, ctx.Err())
		}
	}
}

func (aw *assistantWorker) Finish(ctx context.Context) error {
	ctx, span := obs.Observer.NewSpan(ctx, obs.SpanKindInternal, "controller.assistantWorker.Finish")
	defer span.End()

	if err := aw.ctx.Err(); err != nil {
		if errors.Is(err, context.Canceled) {
			return nil
		}
		return fmt.Errorf("assistant %d flow %d stop failed: %w", aw.id, aw.flowID, err)
	}

	aw.cancel()
	close(aw.input)
	aw.wg.Wait()

	if err := aw.SetStatus(ctx, database.AssistantStatusFinished); err != nil {
		aw.logger.WithError(err).Error("failed to set assistant status to finished")
	}

	return nil
}

func (aw *assistantWorker) Stop(ctx context.Context) error {
	ctx, span := obs.Observer.NewSpan(ctx, obs.SpanKindInternal, "controller.assistantWorker.Stop")
	defer span.End()

	aw.runST()
	done := make(chan struct{})
	timer := time.NewTimer(stopAssistantTimeout)
	defer timer.Stop()

	go func() {
		aw.runWG.Wait()
		close(done)
	}()

	select {
	case <-timer.C:
		return fmt.Errorf("assistant stop timeout")
	case <-done:
		return nil
	}
}
