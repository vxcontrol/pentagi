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

	errFlowAbandoned = errors.New("flow was closed while it was preparing")
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
	) (int64, error)
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
	) (int64, error)
	LoadFlows(ctx context.Context) error
	ListFlows(ctx context.Context) []FlowWorker
	GetFlow(ctx context.Context, flowID int64) (FlowWorker, error)
	StopFlow(ctx context.Context, flowID int64) error
	FinishFlow(ctx context.Context, flowID int64) error
	RenameFlow(ctx context.Context, flowID int64, title string) error
	ReportFlow(ctx context.Context, flowID int64, budget time.Duration) error
	RenameFlowsProvider(ctx context.Context, userID int64, oldName, newName provider.ProviderName) error
	ResetFlowsProviderToDefault(
		ctx context.Context,
		userID int64,
		oldName provider.ProviderName,
		prvtype provider.ProviderType,
	) error
}

// flowPrepareTimeout bounds the background preparation of a new flow. It is
// generous because it covers the provider's setup calls, and only exists so a
// stuck provider cannot pin the goroutine forever.
const flowPrepareTimeout = 30 * time.Minute

// reassignProviderTimeout bounds the provider reference sweep. It is generous
// for two indexed UPDATEs and only exists so a stuck database cannot pin the
// goroutine forever once the sweep is detached from the request context.
const reassignProviderTimeout = 30 * time.Second

// flowFinishTimeout bounds closing a flow once it is detached from the request
// context. It is generous because it covers stopping and removing the sandbox
// containers, and only exists so a stuck docker daemon cannot pin the goroutine
// forever.
const flowFinishTimeout = 5 * time.Minute

// flowLeftoversTimeout bounds closing what the worker did not hold in memory.
// It is separate from flowFinishTimeout because those writes must still land
// after a stuck sandbox has spent the whole finish budget.
const flowLeftoversTimeout = 30 * time.Second

// flowEntry is one flow's slot in the registry. fc.mx guards its fields; op is
// held across a lifecycle call and must never be taken while fc.mx is held.
type flowEntry struct {
	op        chan struct{}
	worker    FlowWorker
	preparing bool
	closing   bool
	committed bool
	abort     context.CancelFunc

	// done is closed once the teardown of an unloaded flow has ended and its
	// slot is gone; it is nil on a slot that holds a worker or a preparation.
	done chan struct{}
}

func newFlowEntry() *flowEntry {
	return &flowEntry{op: make(chan struct{}, 1)}
}

func (e *flowEntry) acquire(ctx context.Context) bool {
	select {
	case e.op <- struct{}{}:
		return true
	case <-ctx.Done():
		return false
	}
}

func (e *flowEntry) lock() {
	e.op <- struct{}{}
}

func (e *flowEntry) release() {
	<-e.op
}

func (e *flowEntry) available() bool {
	return e.worker != nil && !e.preparing && !e.closing
}

type flowController struct {
	db     database.Querier
	mx     *sync.Mutex
	cfg    *config.Config
	flows  map[int64]*flowEntry
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

	build func(ctx context.Context, flow database.Flow, fwc newFlowWorkerCtx, commit func() error) (FlowWorker, error)
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
		flows:  make(map[int64]*flowEntry),
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
		build:  buildFlowWorker,
	}
}

func (fc *flowController) register(flowID int64, worker FlowWorker) *flowEntry {
	fc.mx.Lock()
	defer fc.mx.Unlock()

	entry := newFlowEntry()
	entry.worker = worker
	fc.flows[flowID] = entry

	return entry
}

func (fc *flowController) unregister(flowID int64, entry *flowEntry) {
	fc.mx.Lock()
	defer fc.mx.Unlock()

	if fc.flows[flowID] == entry {
		delete(fc.flows, flowID)
	}
}

// lockSettled takes fc.mx once no teardown of an unloaded flow holds flowID's
// slot, so a delete or a new assistant acts on the flow the teardown left and
// not on containers still being removed. It returns with fc.mx held, or with
// ctx's error and fc.mx released.
func (fc *flowController) lockSettled(ctx context.Context, flowID int64) (*flowEntry, bool, error) {
	for {
		fc.mx.Lock()
		entry, ok := fc.flows[flowID]
		if !ok || entry.done == nil {
			return entry, ok, nil
		}
		done := entry.done
		fc.mx.Unlock()

		select {
		case <-done:
		case <-ctx.Done():
			return nil, false, ctx.Err()
		}
	}
}

func (fc *flowController) entryOf(flowID int64) (*flowEntry, bool) {
	fc.mx.Lock()
	defer fc.mx.Unlock()

	entry, ok := fc.flows[flowID]
	if !ok || !entry.available() {
		return nil, false
	}

	return entry, true
}

// claim hands out a flow with its lifecycle lock already held: the caller must
// release it with defer entry.release().
func (fc *flowController) claim(ctx context.Context, flowID int64) (*flowEntry, FlowWorker, bool) {
	entry, ok := fc.entryOf(flowID)
	if !ok {
		return nil, nil, false
	}

	if !entry.acquire(ctx) {
		return nil, nil, false
	}

	fc.mx.Lock()
	still := fc.flows[flowID] == entry && entry.available()
	worker := entry.worker
	fc.mx.Unlock()

	if !still {
		entry.release()
		return nil, nil, false
	}

	return entry, worker, true
}

func (fc *flowController) LoadFlows(ctx context.Context) error {
	flows, err := fc.db.GetFlows(ctx)
	if err != nil {
		return fmt.Errorf("failed to load flows: %w", err)
	}

	for _, flow := range flows {
		fw, err := LoadFlowWorker(ctx, flow, fc.flowWorkerCtx())
		if err != nil {
			if errors.Is(err, ErrNothingToLoad) {
				continue
			}

			logrus.WithContext(ctx).WithError(err).Errorf("failed to load flow %d", flow.ID)
			continue
		}

		fc.register(flow.ID, fw)
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
) (int64, error) {
	fwc := newFlowWorkerCtx{
		userID:        userID,
		input:         input,
		prvname:       prvname,
		prvtype:       prvtype,
		functions:     functions,
		resources:     resources,
		flowWorkerCtx: fc.flowWorkerCtx(),
	}

	flow, err := reserveFlow(ctx, fwc)
	if err != nil {
		return 0, fmt.Errorf("failed to create flow: %w", err)
	}

	entry, _ := fc.reserveEntry(flow.ID)
	fc.subs.NewFlowPublisher(userID, flow.ID).FlowCreated(ctx, flow, nil)

	go fc.prepareFlow(ctx, entry, flow, func(c context.Context, commit func() error) (FlowWorker, error) {
		return fc.build(c, flow, fwc, commit)
	})

	return flow.ID, nil
}

func (fc *flowController) prepareFlow(
	ctx context.Context,
	entry *flowEntry,
	flow database.Flow,
	build func(context.Context, func() error) (FlowWorker, error),
) {
	ctx, cancel := context.WithTimeout(context.WithoutCancel(ctx), flowPrepareTimeout)
	defer cancel()

	fc.mx.Lock()
	entry.abort = cancel
	fc.mx.Unlock()

	entry.lock()
	defer entry.release()

	worker, err := build(ctx, func() error { return fc.commitPreparation(entry) })
	logger := logrus.WithContext(ctx).WithField("flow_id", flow.ID)

	fc.mx.Lock()
	entry.preparing = false
	entry.worker = worker
	abandoned := !entry.committed && (entry.closing || fc.flows[flow.ID] != entry)
	if err != nil || abandoned {
		if fc.flows[flow.ID] == entry {
			delete(fc.flows, flow.ID)
		}
	}
	fc.mx.Unlock()

	switch {
	case abandoned && worker != nil:
		finishCtx, cancelFinish := context.WithTimeout(context.WithoutCancel(ctx), flowFinishTimeout)
		defer cancelFinish()

		if err := worker.Finish(finishCtx); err != nil {
			logger.WithError(err).Error("failed to release a flow that was closed while it was preparing")
		}
	case abandoned:
		logger.WithError(err).Warn("a flow closed while preparing failed to prepare")
	case err != nil:
		logger.WithError(err).Error("failed to prepare a flow")
		fc.failFlow(ctx, flow, err)
	}
}

func (fc *flowController) failFlow(ctx context.Context, flow database.Flow, cause error) {
	logger := logrus.WithContext(ctx).WithField("flow_id", flow.ID)
	pub := fc.subs.NewFlowPublisher(flow.UserID, flow.ID)

	if mlw, err := fc.mlc.NewFlowMsgLog(ctx, flow.ID, pub); err != nil {
		logger.WithError(err).Warn("failed to open the message log of a flow that could not start")
	} else if _, err := mlw.PutFlowMsg(
		ctx, database.MsglogTypeReport, "", fmt.Sprintf("This flow could not be started: %s", cause),
	); err != nil {
		logger.WithError(err).Warn("failed to record why a flow could not start")
	}

	failed, err := fc.db.UpdateFlowStatus(ctx, database.UpdateFlowStatusParams{
		ID:     flow.ID,
		Status: database.FlowStatusFailed,
	})
	if err != nil {
		logger.WithError(err).Error("failed to mark a flow failed after its preparation broke")
		return
	}

	containers, err := fc.db.GetFlowContainers(ctx, flow.ID)
	if err != nil {
		logger.WithError(err).Warn("failed to read containers of a failed flow, publishing it without them")
	}

	pub.FlowUpdated(ctx, failed, containers)
}

func (fc *flowController) flowWorkerCtx() flowWorkerCtx {
	return flowWorkerCtx{
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
) (int64, error) {
	fwc := newFlowWorkerCtx{
		userID:        userID,
		input:         input,
		dryRun:        true,
		prvname:       prvname,
		prvtype:       prvtype,
		functions:     functions,
		flowWorkerCtx: fc.flowWorkerCtx(),
	}

	entry, flowID, prepare, err := fc.flowForAssistant(ctx, flowID, fwc)
	if err != nil {
		return 0, err
	}

	awc := newAssistantWorkerCtx{
		userID:        userID,
		flowID:        flowID,
		input:         input,
		prvname:       prvname,
		prvtype:       prvtype,
		useAgents:     useAgents,
		functions:     functions,
		resources:     resources,
		flowWorkerCtx: fwc.flowWorkerCtx,
	}

	assistant, err := reserveAssistant(ctx, awc)
	if err != nil {
		go prepare(ctx)

		return 0, fmt.Errorf("failed to create assistant: %w", err)
	}

	fc.subs.NewFlowPublisher(userID, flowID).AssistantCreated(ctx, assistant)

	go func() {
		prepare(ctx)
		fc.prepareAssistant(ctx, entry, assistant, awc)
	}()

	return assistant.ID, nil
}

func (fc *flowController) flowForAssistant(
	ctx context.Context, flowID int64, fwc newFlowWorkerCtx,
) (*flowEntry, int64, func(context.Context), error) {
	if flowID == 0 {
		flow, err := reserveFlow(ctx, fwc)
		if err != nil {
			return nil, 0, nil, fmt.Errorf("failed to create flow: %w", err)
		}

		entry, _ := fc.reserveEntry(flow.ID)
		fc.subs.NewFlowPublisher(fwc.userID, flow.ID).FlowCreated(ctx, flow, nil)

		return entry, flow.ID, func(c context.Context) {
			fc.prepareFlow(c, entry, flow, func(c context.Context, commit func() error) (FlowWorker, error) {
				fw, err := fc.build(c, flow, fwc, commit)
				if err != nil {
					return nil, err
				}

				if err := fw.SetStatus(c, database.FlowStatusWaiting); err != nil {
					return nil, fmt.Errorf("failed to set flow %d status: %w", flow.ID, err)
				}

				return fw, nil
			})
		}, nil
	}

	entry, loaded, err := fc.lockSettled(ctx, flowID)
	if err != nil {
		return nil, 0, nil, err
	}
	closing := loaded && entry.closing
	preparing := loaded && entry.preparing
	fc.mx.Unlock()

	switch {
	case closing:
		return nil, 0, nil, ErrFlowNotFound
	case preparing:
		return nil, 0, nil, fmt.Errorf("flow %d is not ready yet", flowID)
	}

	flow, err := fc.db.GetFlow(ctx, flowID)
	if err != nil {
		return nil, 0, nil, fmt.Errorf("failed to get flow %d: %w", flowID, err)
	}

	switch flow.Status {
	case database.FlowStatusCreated:
		return nil, 0, nil, fmt.Errorf("flow %d is not completed", flowID)
	case database.FlowStatusRunning, database.FlowStatusWaiting:
		if loaded {
			return entry, flowID, func(context.Context) {}, nil
		}
	case database.FlowStatusFinished, database.FlowStatusFailed:
	default:
		return nil, 0, nil, fmt.Errorf("flow %d is in unknown status: %s", flowID, flow.Status)
	}

	reloaded, installed := fc.reserveEntry(flowID)
	if !installed {
		serving, ok := fc.entryOf(flowID)
		if !ok {
			return nil, 0, nil, fmt.Errorf("flow %d is not ready yet", flowID)
		}

		return serving, flowID, func(context.Context) {}, nil
	}

	return reloaded, flowID, func(c context.Context) {
		fc.prepareFlow(c, reloaded, flow, func(c context.Context, _ func() error) (FlowWorker, error) {
			renewed, err := fc.db.UpdateFlowStatus(c, database.UpdateFlowStatusParams{
				ID:     flowID,
				Status: database.FlowStatusWaiting,
			})
			if err != nil {
				return nil, fmt.Errorf("failed to renew flow %d status: %w", flowID, err)
			}

			return LoadFlowWorker(c, renewed, fc.flowWorkerCtx())
		})
	}, nil
}

func (fc *flowController) prepareAssistant(
	ctx context.Context,
	entry *flowEntry,
	assistant database.Assistant,
	awc newAssistantWorkerCtx,
) {
	ctx, cancel := context.WithTimeout(context.WithoutCancel(ctx), flowPrepareTimeout)
	defer cancel()

	fc.mx.Lock()
	ready := entry.available()
	fw := entry.worker
	fc.mx.Unlock()

	logger := logrus.WithContext(ctx).WithFields(logrus.Fields{
		"flow_id":      awc.flowID,
		"assistant_id": assistant.ID,
	})

	if !ready {
		logger.Error("an assistant was asked for on a flow that could not be started")
		fc.failAssistant(ctx, awc.userID, assistant, fmt.Errorf("flow %d could not be started", awc.flowID))

		return
	}

	awc.fw = fw

	if _, err := buildAssistantWorker(ctx, assistant, awc); err != nil {
		if errors.Is(err, ErrFlowAlreadyStopped) {
			logger.WithError(err).Warn("the flow finished while its assistant was being prepared")
			return
		}

		logger.WithError(err).Error("failed to prepare an assistant")
	}
}

func (fc *flowController) failAssistant(
	ctx context.Context, userID int64, assistant database.Assistant, cause error,
) {
	failed, err := fc.db.UpdateAssistantStatus(ctx, database.UpdateAssistantStatusParams{
		ID:     assistant.ID,
		Status: database.AssistantStatusFailed,
	})
	if err != nil {
		logrus.WithContext(ctx).WithError(err).
			Errorf("failed to mark assistant %d failed after %s", assistant.ID, cause)

		return
	}

	fc.subs.NewFlowPublisher(userID, assistant.FlowID).AssistantUpdated(ctx, failed)
}

// reserveEntry installs a preparing slot for flowID and reports whether this
// caller is the one that installed it; it never replaces a slot already there.
func (fc *flowController) reserveEntry(flowID int64) (*flowEntry, bool) {
	fc.mx.Lock()
	defer fc.mx.Unlock()

	if existing, taken := fc.flows[flowID]; taken {
		return existing, false
	}

	entry := newFlowEntry()
	entry.preparing = true
	fc.flows[flowID] = entry

	return entry, true
}

func (fc *flowController) commitPreparation(entry *flowEntry) error {
	fc.mx.Lock()
	defer fc.mx.Unlock()

	if entry.closing {
		return errFlowAbandoned
	}
	entry.committed = true

	return nil
}

func (fc *flowController) ListFlows(ctx context.Context) []FlowWorker {
	fc.mx.Lock()
	flows := make([]FlowWorker, 0, len(fc.flows))
	for _, entry := range fc.flows {
		if entry.available() {
			flows = append(flows, entry.worker)
		}
	}
	fc.mx.Unlock()

	sort.Slice(flows, func(i, j int) bool {
		return flows[i].GetFlowID() < flows[j].GetFlowID()
	})

	return flows
}

func (fc *flowController) GetFlow(ctx context.Context, flowID int64) (FlowWorker, error) {
	entry, ok := fc.entryOf(flowID)
	if !ok {
		return nil, ErrFlowNotFound
	}

	return entry.worker, nil
}

func (fc *flowController) StopFlow(ctx context.Context, flowID int64) error {
	entry, flow, ok := fc.claim(ctx, flowID)
	if !ok {
		return fc.missingFlowError(ctx, flowID)
	}
	defer entry.release()

	if err := flow.Stop(ctx); err != nil {
		return fmt.Errorf("failed to stop flow %d: %w", flowID, err)
	}

	return nil
}

func (fc *flowController) FinishFlow(ctx context.Context, flowID int64) error {
	return fc.finishFlowWithin(ctx, flowID, flowFinishTimeout)
}

func (fc *flowController) finishFlowWithin(ctx context.Context, flowID int64, budget time.Duration) error {
	entry, ok, err := fc.lockSettled(ctx, flowID)
	if err != nil {
		return err
	}
	if !ok {
		// A closing slot rather than fc.mx guards the teardown: other flows stay
		// served, and no assistant reloads this one while its containers go.
		entry = newFlowEntry()
		entry.closing = true
		entry.done = make(chan struct{})
		fc.flows[flowID] = entry
		fc.mx.Unlock()
		defer func() {
			fc.unregister(flowID, entry)
			close(entry.done)
		}()

		unloadedCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), budget)
		defer cancel()

		return fc.finishUnloadedFlow(unloadedCtx, flowID)
	}

	if entry.closing {
		fc.mx.Unlock()
		return nil
	}

	entry.closing = true
	abandon := entry.preparing && !entry.committed
	if abandon && entry.abort != nil {
		entry.abort()
	}
	fc.mx.Unlock()

	if abandon {
		preparingCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), budget)
		defer cancel()

		return fc.finishUnloadedFlow(preparingCtx, flowID)
	}

	entry.lock()
	defer entry.release()

	fc.mx.Lock()
	flow := entry.worker
	fc.mx.Unlock()

	ctx, cancel := context.WithTimeout(context.WithoutCancel(ctx), budget)
	defer cancel()

	defer fc.unregister(flowID, entry)

	if flow == nil {
		return fc.finishUnloadedFlow(ctx, flowID)
	}

	var finishErr error
	if err := flow.Finish(ctx); err != nil {
		finishErr = fmt.Errorf("failed to finish flow %d: %w", flowID, err)
	}

	leftoversCtx, cancelLeftovers := context.WithTimeout(context.WithoutCancel(ctx), flowLeftoversTimeout)
	defer cancelLeftovers()

	return errors.Join(finishErr, fc.finishFlowLeftovers(leftoversCtx, flowID))
}

// ReportFlow asks the flow's open task for its write-up. It returns once the
// report has been accepted, not once it has been written: see flowWorker.Report.
func (fc *flowController) ReportFlow(ctx context.Context, flowID int64, budget time.Duration) error {
	entry, flow, ok := fc.claim(ctx, flowID)
	if !ok {
		return fc.missingFlowError(ctx, flowID)
	}
	defer entry.release()

	return flow.Report(ctx, budget)
}

func (fc *flowController) RenameFlow(ctx context.Context, flowID int64, title string) error {
	entry, flow, ok := fc.claim(ctx, flowID)
	if !ok {
		return fc.missingFlowError(ctx, flowID)
	}
	defer entry.release()

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
