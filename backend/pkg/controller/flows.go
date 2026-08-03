package controller

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"sync"
	"time"

	"pentagi/pkg/config"
	"pentagi/pkg/database"
	"pentagi/pkg/docker"
	"pentagi/pkg/graph/subscriptions"
	"pentagi/pkg/providers"
	"pentagi/pkg/providers/provider"
	"pentagi/pkg/tools"

	"github.com/sirupsen/logrus"
)

var (
	ErrFlowNotFound       = fmt.Errorf("flow not found")
	ErrFlowAlreadyStopped = fmt.Errorf("flow already stopped")
)

type FlowController interface {
	CreateFlow(
		ctx context.Context,
		userID int64,
		input string,
		prvname provider.ProviderName,
		prvtype provider.ProviderType,
		functions *tools.Functions,
		resources []database.UserResource,
	) (FlowWorker, error)
	CreateAssistant(
		ctx context.Context,
		userID int64,
		flowID int64,
		input string,
		useAgents bool,
		prvname provider.ProviderName,
		prvtype provider.ProviderType,
		functions *tools.Functions,
		resources []database.UserResource,
	) (AssistantWorker, error)
	LoadFlows(ctx context.Context) error
	ListFlows(ctx context.Context) []FlowWorker
	GetFlow(ctx context.Context, flowID int64) (FlowWorker, error)
	StopFlow(ctx context.Context, flowID int64) error
	FinishFlow(ctx context.Context, flowID int64) error
	RenameFlow(ctx context.Context, flowID int64, title string) error
	RenameFlowsProvider(ctx context.Context, userID int64, oldName, newName provider.ProviderName) error
	ResetFlowsProviderToDefault(
		ctx context.Context,
		userID int64,
		oldName provider.ProviderName,
		prvtype provider.ProviderType,
	) error
}

// reassignProviderTimeout bounds the provider reference sweep. It is generous
// for two indexed UPDATEs and only exists so a stuck database cannot pin the
// goroutine forever once the sweep is detached from the request context.
const reassignProviderTimeout = 30 * time.Second

type flowController struct {
	db     database.Querier
	mx     *sync.Mutex
	cfg    *config.Config
	flows  map[int64]FlowWorker
	docker docker.DockerClient
	provs  providers.ProviderController
	subs   subscriptions.SubscriptionsController
	alc    AgentLogController
	mlc    MsgLogController
	aslc   AssistantLogController
	slc    SearchLogController
	tlc    TermLogController
	vslc   VectorStoreLogController
	tclc   ToolCallLogController
	sc     ScreenshotController
}

func NewFlowController(
	db database.Querier,
	cfg *config.Config,
	docker docker.DockerClient,
	provs providers.ProviderController,
	subs subscriptions.SubscriptionsController,
) FlowController {
	return &flowController{
		db:     db,
		mx:     &sync.Mutex{},
		cfg:    cfg,
		flows:  make(map[int64]FlowWorker),
		docker: docker,
		provs:  provs,
		subs:   subs,
		alc:    NewAgentLogController(db),
		mlc:    NewMsgLogController(db),
		aslc:   NewAssistantLogController(db),
		slc:    NewSearchLogController(db),
		tlc:    NewTermLogController(db),
		vslc:   NewVectorStoreLogController(db),
		tclc:   NewToolCallLogController(db),
		sc:     NewScreenshotController(db),
	}
}

func (fc *flowController) LoadFlows(ctx context.Context) error {
	flows, err := fc.db.GetFlows(ctx)
	if err != nil {
		return fmt.Errorf("failed to load flows: %w", err)
	}

	for _, flow := range flows {
		fw, err := LoadFlowWorker(ctx, flow, flowWorkerCtx{
			db:     fc.db,
			cfg:    fc.cfg,
			docker: fc.docker,
			provs:  fc.provs,
			subs:   fc.subs,
			flowProviderControllers: flowProviderControllers{
				mlc:  fc.mlc,
				aslc: fc.aslc,
				alc:  fc.alc,
				slc:  fc.slc,
				tlc:  fc.tlc,
				vslc: fc.vslc,
				tclc: fc.tclc,
				sc:   fc.sc,
			},
		})
		if err != nil {
			if errors.Is(err, ErrNothingToLoad) {
				continue
			}

			logrus.WithContext(ctx).WithError(err).Errorf("failed to load flow %d", flow.ID)
			continue
		}

		fc.flows[flow.ID] = fw
	}

	return nil
}

func (fc *flowController) CreateFlow(
	ctx context.Context,
	userID int64,
	input string,
	prvname provider.ProviderName,
	prvtype provider.ProviderType,
	functions *tools.Functions,
	resources []database.UserResource,
) (FlowWorker, error) {
	fc.mx.Lock()
	defer fc.mx.Unlock()

	fw, err := NewFlowWorker(ctx, newFlowWorkerCtx{
		userID:    userID,
		input:     input,
		prvname:   prvname,
		prvtype:   prvtype,
		functions: functions,
		resources: resources,
		flowWorkerCtx: flowWorkerCtx{
			db:     fc.db,
			cfg:    fc.cfg,
			docker: fc.docker,
			provs:  fc.provs,
			subs:   fc.subs,
			flowProviderControllers: flowProviderControllers{
				mlc:  fc.mlc,
				aslc: fc.aslc,
				alc:  fc.alc,
				slc:  fc.slc,
				tlc:  fc.tlc,
				vslc: fc.vslc,
				tclc: fc.tclc,
				sc:   fc.sc,
			},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create flow worker: %w", err)
	}

	fc.flows[fw.GetFlowID()] = fw

	return fw, nil
}

func (fc *flowController) CreateAssistant(
	ctx context.Context,
	userID int64,
	flowID int64,
	input string,
	useAgents bool,
	prvname provider.ProviderName,
	prvtype provider.ProviderType,
	functions *tools.Functions,
	resources []database.UserResource,
) (AssistantWorker, error) {
	fc.mx.Lock()
	defer fc.mx.Unlock()

	var (
		fw  FlowWorker
		ok  bool
		err error
	)

	flowWorkerCtx := flowWorkerCtx{
		db:     fc.db,
		cfg:    fc.cfg,
		docker: fc.docker,
		provs:  fc.provs,
		subs:   fc.subs,
		flowProviderControllers: flowProviderControllers{
			mlc:  fc.mlc,
			aslc: fc.aslc,
			alc:  fc.alc,
			slc:  fc.slc,
			tlc:  fc.tlc,
			vslc: fc.vslc,
			tclc: fc.tclc,
			sc:   fc.sc,
		},
	}

	newFlow := func() error {
		fw, err = NewFlowWorker(ctx, newFlowWorkerCtx{
			userID:        userID,
			input:         input,
			dryRun:        true,
			prvname:       prvname,
			prvtype:       prvtype,
			functions:     functions,
			flowWorkerCtx: flowWorkerCtx,
		})
		if err != nil {
			return fmt.Errorf("failed to create flow worker: %w", err)
		}

		fc.flows[fw.GetFlowID()] = fw
		flowID = fw.GetFlowID()
		fw.SetStatus(ctx, database.FlowStatusWaiting)

		return nil
	}

	loadFlow := func() error {
		flow, err := fc.db.UpdateFlowStatus(ctx, database.UpdateFlowStatusParams{
			ID:     flowID,
			Status: database.FlowStatusWaiting,
		})
		if err != nil {
			return fmt.Errorf("failed to renew flow %d status: %w", flowID, err)
		}

		fw, err = LoadFlowWorker(ctx, flow, flowWorkerCtx)
		if err != nil {
			return fmt.Errorf("failed to load flow %d: %w", flowID, err)
		}

		fc.flows[flowID] = fw

		return nil
	}

	if flowID == 0 {
		if err := newFlow(); err != nil {
			return nil, err
		}
	} else if fw, ok = fc.flows[flowID]; ok {
		status, err := fw.GetStatus(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to get flow %d status: %w", flowID, err)
		}

		switch status {
		case database.FlowStatusCreated:
			return nil, fmt.Errorf("flow %d is not completed", flowID)
		case database.FlowStatusFinished, database.FlowStatusFailed:
			if err := loadFlow(); err != nil {
				return nil, err
			}
		case database.FlowStatusRunning, database.FlowStatusWaiting:
			break
		default:
			return nil, fmt.Errorf("flow %d is in unknown status: %s", flowID, status)
		}
	} else {
		if err := loadFlow(); err != nil {
			return nil, err
		}
	}

	if fw == nil { // just double check, this should never happen
		return nil, fmt.Errorf("unexpected error: flow %d not found", flowID)
	}

	aw, err := NewAssistantWorker(ctx, newAssistantWorkerCtx{
		userID:        userID,
		flowID:        flowID,
		input:         input,
		prvname:       prvname,
		prvtype:       prvtype,
		useAgents:     useAgents,
		functions:     functions,
		resources:     resources,
		fw:            fw,
		flowWorkerCtx: flowWorkerCtx,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create assistant: %w", err)
	}

	if err = fw.AddAssistant(ctx, aw); err != nil {
		return nil, fmt.Errorf("failed to add assistant to flow: %w", err)
	}

	return aw, nil
}

func (fc *flowController) ListFlows(ctx context.Context) []FlowWorker {
	fc.mx.Lock()
	defer fc.mx.Unlock()

	flows := make([]FlowWorker, 0)
	for _, flow := range fc.flows {
		flows = append(flows, flow)
	}

	sort.Slice(flows, func(i, j int) bool {
		return flows[i].GetFlowID() < flows[j].GetFlowID()
	})

	return flows
}

func (fc *flowController) GetFlow(ctx context.Context, flowID int64) (FlowWorker, error) {
	fc.mx.Lock()
	defer fc.mx.Unlock()

	flow, ok := fc.flows[flowID]
	if !ok {
		return nil, ErrFlowNotFound
	}

	return flow, nil
}

func (fc *flowController) StopFlow(ctx context.Context, flowID int64) error {
	fc.mx.Lock()
	defer fc.mx.Unlock()

	flow, ok := fc.flows[flowID]
	if !ok {
		return ErrFlowNotFound
	}

	err := flow.Stop(ctx)
	if err != nil {
		return fmt.Errorf("failed to stop flow %d: %w", flowID, err)
	}

	return nil
}

func (fc *flowController) FinishFlow(ctx context.Context, flowID int64) error {
	fc.mx.Lock()
	defer fc.mx.Unlock()

	flow, ok := fc.flows[flowID]
	if !ok {
		return ErrFlowNotFound
	}

	err := flow.Finish(ctx)
	if err != nil {
		return fmt.Errorf("failed to finish flow %d: %w", flowID, err)
	}

	delete(fc.flows, flowID)

	return nil
}

func (fc *flowController) RenameFlow(ctx context.Context, flowID int64, title string) error {
	fc.mx.Lock()
	defer fc.mx.Unlock()

	flow, ok := fc.flows[flowID]
	if !ok {
		return ErrFlowNotFound
	}

	return flow.Rename(ctx, title)
}

// RenameFlowsProvider repoints every flow and assistant of userID that still
// refers to oldName at newName, after the user renamed a custom LLM provider.
func (fc *flowController) RenameFlowsProvider(
	ctx context.Context,
	userID int64,
	oldName, newName provider.ProviderName,
) error {
	return fc.reassignFlowsProvider(ctx, userID, oldName, newName)
}

// ResetFlowsProviderToDefault repoints every flow and assistant of userID that
// referred to a just-deleted custom LLM provider at the built-in name for its
// type, which is literally the type string ("qwen", "openai", ...) — see
// provider.DefaultProviderName*. That name always resolves, so the flow stays
// loadable instead of failing with "provider not found by name".
func (fc *flowController) ResetFlowsProviderToDefault(
	ctx context.Context,
	userID int64,
	oldName provider.ProviderName,
	prvtype provider.ProviderType,
) error {
	return fc.reassignFlowsProvider(ctx, userID, oldName, provider.ProviderName(prvtype))
}

// reassignFlowsProvider rewrites the provider reference stored on a user's flow
// and assistant rows. It deliberately does *not* touch loaded workers:
//
//   - Nothing here blocks on an LLM. Building a provider instance probes the
//     upstream API to resolve a tool call ID template, so switching loaded
//     workers inline would tie a "rename provider" click to LLM latency and give
//     the caller time to cancel the request mid-cascade.
//   - Nothing here takes fc.mx or reaches into a worker, so the cascade cannot
//     deadlock against, or stall, any other flow operation.
//
// A running flow picks the change up on the user's next input (which already
// re-resolves the provider by name and calls flowWorker.switchProvider) or on
// the next backend start (which rebuilds the provider from the DB row). Both
// paths compare the provider's raw configuration, so they also catch the case
// where the name did not change but the configuration behind it did.
//
// The two sweeps only match rows still bearing oldName, which makes the whole
// operation idempotent and safe to retry. They are issued independently and
// their errors are joined, so a failure on one table never silently skips the
// other.
func (fc *flowController) reassignFlowsProvider(
	ctx context.Context,
	userID int64,
	oldName, newName provider.ProviderName,
) error {
	logger := logrus.WithContext(ctx).WithFields(logrus.Fields{
		"user_id":  userID,
		"old_name": oldName.String(),
		"new_name": newName.String(),
	})

	if oldName == newName {
		logger.Debug("provider name unchanged, nothing to reassign")
		return nil
	}

	// Only references that would otherwise dangle get rewritten. oldName can
	// still resolve after the provider is gone when it named an override of a
	// built-in — an intentional feature — in which case the built-in answers to
	// that name again and the stored value is already correct. Rewriting it
	// anyway would repoint rows that predate the override, and (when the
	// override's type differed from the built-in it was named after) would send
	// them to the wrong default entirely.
	if _, err := fc.provs.GetProvider(ctx, oldName, userID); err == nil {
		logger.Debug("old provider name still resolves, nothing to reassign")
		return nil
	}

	// Detached from the caller's request context: these are two short statements
	// and the reference must not be left half-rewritten because a browser tab
	// was closed. The timeout keeps a stuck DB from pinning the goroutine.
	ctx, cancel := context.WithTimeout(context.WithoutCancel(ctx), reassignProviderTimeout)
	defer cancel()

	flows, flowsErr := fc.db.UpdateFlowsProviderNameByOldName(ctx, database.UpdateFlowsProviderNameByOldNameParams{
		NewName: newName.String(),
		UserID:  userID,
		OldName: oldName.String(),
	})
	if flowsErr != nil {
		logger.WithError(flowsErr).Error("failed to bulk-update flows provider name")
		flowsErr = fmt.Errorf("failed to bulk-update flows provider name: %w", flowsErr)
	}

	assistants, asstErr := fc.db.UpdateAssistantsProviderNameByOldName(
		ctx, database.UpdateAssistantsProviderNameByOldNameParams{
			NewName: newName.String(),
			UserID:  userID,
			OldName: oldName.String(),
		})
	if asstErr != nil {
		logger.WithError(asstErr).Error("failed to bulk-update assistants provider name")
		asstErr = fmt.Errorf("failed to bulk-update assistants provider name: %w", asstErr)
	}

	// Publishing happens only after both writes are done. A subscriber that is
	// not draining its channel makes each publish cost up to the subscription
	// send timeout, so doing it in between would let a wedged websocket client
	// eat the deadline and starve the second UPDATE.
	for _, flow := range flows {
		// Skipped rather than published with no containers: FlowUpdated carries
		// the full terminal list and the client replaces its cached value with
		// whatever arrives, so an empty list would wipe the flow's terminals in
		// the UI. Same handling as flowWorker.switchProvider.
		containers, err := fc.db.GetFlowContainers(ctx, flow.ID)
		if err != nil {
			logger.WithError(err).Warnf("failed to get containers for flow %d, skipping its update event", flow.ID)
			continue
		}
		fc.subs.NewFlowPublisher(userID, flow.ID).FlowUpdated(ctx, flow, containers)
	}

	for _, assistant := range assistants {
		fc.subs.NewFlowPublisher(userID, assistant.FlowID).AssistantUpdated(ctx, assistant)
	}

	logger.WithFields(logrus.Fields{
		"flows_updated":      len(flows),
		"assistants_updated": len(assistants),
	}).Info("provider reference reassigned")

	return errors.Join(flowsErr, asstErr)
}
