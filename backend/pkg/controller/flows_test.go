package controller

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"slices"
	"sync"
	"testing"
	"time"

	"pentagi/pkg/database"
	"pentagi/pkg/docker"
	obs "pentagi/pkg/observability"
	"pentagi/pkg/providers"
	"pentagi/pkg/providers/provider"
	"pentagi/pkg/providers/tester/mock"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// cascadeFakeQuerier fails its calls once the context is done, the way pgx does.
type cascadeFakeQuerier struct {
	database.Querier

	flowsCalls       []database.UpdateFlowsProviderNameByOldNameParams
	flowsResult      []database.Flow
	flowsErr         error
	assistantsCalls  []database.UpdateAssistantsProviderNameByOldNameParams
	assistantsResult []database.Assistant
	assistantsErr    error
	containersResult []database.Container
	containersErr    error
}

func (f *cascadeFakeQuerier) UpdateFlowsProviderNameByOldName(
	ctx context.Context, arg database.UpdateFlowsProviderNameByOldNameParams,
) ([]database.Flow, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	f.flowsCalls = append(f.flowsCalls, arg)
	if f.flowsErr != nil {
		return nil, f.flowsErr
	}

	return f.flowsResult, nil
}

func (f *cascadeFakeQuerier) UpdateAssistantsProviderNameByOldName(
	ctx context.Context, arg database.UpdateAssistantsProviderNameByOldNameParams,
) ([]database.Assistant, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	f.assistantsCalls = append(f.assistantsCalls, arg)
	if f.assistantsErr != nil {
		return nil, f.assistantsErr
	}

	return f.assistantsResult, nil
}

func (f *cascadeFakeQuerier) GetFlowContainers(ctx context.Context, _ int64) ([]database.Container, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if f.containersErr != nil {
		return nil, f.containersErr
	}

	return f.containersResult, nil
}

// cascadeFakeProviders resolves only the names in resolvable.
type cascadeFakeProviders struct {
	providers.ProviderController

	resolvable map[provider.ProviderName]bool
}

func (p *cascadeFakeProviders) GetProvider(
	_ context.Context, prvname provider.ProviderName, _ int64,
) (provider.Provider, error) {
	if p.resolvable[prvname] {
		return mock.NewProvider(provider.ProviderQwen, prvname, "model"), nil
	}

	return nil, fmt.Errorf("provider not found by name '%s'", prvname)
}

func newTestFlowController(
	q *cascadeFakeQuerier, resolvable ...provider.ProviderName,
) (*flowController, *cascadeFakePublisher) {
	pub := &cascadeFakePublisher{}
	names := make(map[provider.ProviderName]bool, len(resolvable))
	for _, name := range resolvable {
		names[name] = true
	}

	return &flowController{
		db:    q,
		mx:    &sync.Mutex{},
		flows: map[int64]*flowEntry{},
		subs:  &cascadeFakeSubscriptions{pub: pub},
		provs: &cascadeFakeProviders{resolvable: names},
	}, pub
}

func TestFlows_ReassignFlowsProvider_RenameSweepsBothTablesPastACancelledCaller(t *testing.T) {
	const userID = int64(1)

	q := &cascadeFakeQuerier{
		flowsResult:      []database.Flow{{ID: 10}, {ID: 11}},
		assistantsResult: []database.Assistant{{ID: 100, FlowID: 10}},
	}
	fc, pub := newTestFlowController(q)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	require.NoError(t, fc.RenameFlowsProvider(ctx, userID, "my-qwen", "my-qwen-renamed"))

	assert.Equal(t, []database.UpdateFlowsProviderNameByOldNameParams{
		{NewName: "my-qwen-renamed", UserID: userID, OldName: "my-qwen"},
	}, q.flowsCalls, "the sweep is scoped to the acting user")
	assert.Equal(t, []database.UpdateAssistantsProviderNameByOldNameParams{
		{NewName: "my-qwen-renamed", UserID: userID, OldName: "my-qwen"},
	}, q.assistantsCalls)
	assert.Len(t, pub.flowUpdated, 2, "every rewritten flow row is published")
	assert.Len(t, pub.assistantUpdated, 1, "every rewritten assistant row is published")
}

func TestFlows_ReassignFlowsProvider_RewritesOnlyADanglingName(t *testing.T) {
	for _, tc := range []struct {
		name       string
		oldName    provider.ProviderName
		prvtype    provider.ProviderType
		resolvable []provider.ProviderName
		wantNew    string
	}{
		{name: "a name already equal to its type's default", oldName: "qwen", prvtype: provider.ProviderQwen},
		{
			name: "a name a built-in of another type still answers to", oldName: "openai",
			prvtype: provider.ProviderCustom, resolvable: []provider.ProviderName{"openai"},
		},
		{name: "a name nothing answers to any more", oldName: "openai", prvtype: provider.ProviderCustom, wantNew: "custom"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			q := &cascadeFakeQuerier{flowsResult: []database.Flow{{ID: 10}}}
			fc, pub := newTestFlowController(q, tc.resolvable...)

			require.NoError(t, fc.ResetFlowsProviderToDefault(context.Background(), 1, tc.oldName, tc.prvtype))

			if tc.wantNew == "" {
				assert.Empty(t, q.flowsCalls, "a name that still resolves is never rewritten")
				assert.Empty(t, q.assistantsCalls)
				assert.Empty(t, pub.flowUpdated, "nor is every row of that provider republished")

				return
			}
			require.Len(t, q.flowsCalls, 1)
			assert.Equal(t, string(tc.oldName), q.flowsCalls[0].OldName)
			assert.Equal(t, tc.wantNew, q.flowsCalls[0].NewName, "the default name is the type string")
		})
	}
}

func TestFlows_ReassignFlowsProvider_ReportsEachSweepFailureWithoutSkippingTheOther(t *testing.T) {
	flowsErr, assistantsErr := errors.New("flows update exploded"), errors.New("assistants update exploded")

	for _, tc := range []struct {
		name          string
		assistantsErr error
		wantPublished int
	}{
		{name: "the flows sweep fails", wantPublished: 1},
		{name: "both sweeps fail", assistantsErr: assistantsErr},
	} {
		t.Run(tc.name, func(t *testing.T) {
			q := &cascadeFakeQuerier{
				flowsErr:         flowsErr,
				assistantsErr:    tc.assistantsErr,
				assistantsResult: []database.Assistant{{ID: 100, FlowID: 10}},
			}
			fc, pub := newTestFlowController(q)

			err := fc.RenameFlowsProvider(context.Background(), 1, "old", "new")

			assert.ErrorIs(t, err, flowsErr)
			if tc.assistantsErr != nil {
				assert.ErrorIs(t, err, tc.assistantsErr)
			}
			assert.Len(t, q.assistantsCalls, 1, "a failure on one table does not skip the other")
			assert.Len(t, pub.assistantUpdated, tc.wantPublished)
		})
	}
}

// An event with no containers would wipe the terminals an open tab shows.
func TestFlows_ReassignFlowsProvider_SkipsThePublishOfAFlowWhoseContainersCannotBeRead(t *testing.T) {
	q := &cascadeFakeQuerier{
		flowsResult:   []database.Flow{{ID: 10}},
		containersErr: errors.New("containers unavailable"),
	}
	fc, pub := newTestFlowController(q)

	require.NoError(t, fc.RenameFlowsProvider(context.Background(), 1, "old", "new"),
		"a container lookup failure is not a cascade failure")

	assert.Empty(t, pub.flowUpdated)
	assert.Len(t, q.flowsCalls, 1, "the rewrite itself still happened")
}

type creatingController struct {
	fc      *flowController
	q       *createFakeQuerier
	pub     *cascadeFakePublisher
	worker  *noopFlowWorker
	release chan struct{}
	entered chan struct{}

	commitsBeforeRelease bool
	commitsAfterRelease  bool
	ignoresCancel        bool
	holdsAfterCancel     bool
	committed            chan error
	cancelled            chan struct{}
	finishing            *finishRecordingWorker
}

func newCreatingController(t *testing.T, buildErr error) *creatingController {
	t.Helper()

	q := &createFakeQuerier{}
	pub := &cascadeFakePublisher{}
	built := &noopFlowWorker{flowID: reservedFlowID}

	c := &creatingController{
		q:         q,
		pub:       pub,
		worker:    built,
		release:   make(chan struct{}),
		entered:   make(chan struct{}, 1),
		committed: make(chan error, 1),
		cancelled: make(chan struct{}),
	}

	c.fc = &flowController{
		db:     q,
		mx:     &sync.Mutex{},
		flows:  map[int64]*flowEntry{},
		subs:   &cascadeFakeSubscriptions{pub: pub},
		docker: &finishFakeDocker{},
		mlc:    NewMsgLogController(q),
		build: func(ctx context.Context, _ database.Flow, _ newFlowWorkerCtx, commit func() error) (FlowWorker, error) {
			if c.commitsBeforeRelease {
				if err := commit(); err != nil {
					return nil, err
				}
			}
			c.entered <- struct{}{}

			switch {
			case c.ignoresCancel:
				<-c.release
			case c.commitsAfterRelease:
				<-c.release
				err := commit()
				c.committed <- err
				if err != nil {
					return nil, err
				}
			default:
				select {
				case <-c.release:
				case <-ctx.Done():
					close(c.cancelled)
					if c.holdsAfterCancel {
						<-c.release
					}
					return nil, ctx.Err()
				}
			}

			if buildErr != nil {
				return nil, buildErr
			}
			if c.finishing != nil {
				return c.finishing, nil
			}

			return built, nil
		},
	}

	return c
}

func (c *creatingController) waitForPreparation(t *testing.T) {
	t.Helper()

	controllerReceive(t, c.entered, "preparing the flow to start")
}

func (c *creatingController) waitUntilSettled(t *testing.T) {
	t.Helper()

	require.Eventually(t, func() bool {
		c.fc.mx.Lock()
		defer c.fc.mx.Unlock()

		entry, ok := c.fc.flows[reservedFlowID]

		return !ok || !entry.preparing
	}, 5*time.Second, time.Millisecond, "the flow never finished preparing")
}

// finishWhilePreparing fails, rather than hangs, when FinishFlow waits for a build still holding the flow.
func (c *creatingController) finishWhilePreparing(t *testing.T) {
	t.Helper()

	finished := make(chan error, 1)
	go func() { finished <- c.fc.FinishFlow(context.Background(), reservedFlowID) }()

	require.NoError(t, controllerReceive(t, finished, "finishing a flow whose sandbox is still being prepared"))
}

func (c *creatingController) registered() bool {
	c.fc.mx.Lock()
	defer c.fc.mx.Unlock()

	_, ok := c.fc.flows[reservedFlowID]

	return ok
}

func (c *creatingController) create(t *testing.T) int64 {
	t.Helper()

	flowID, err := c.fc.CreateFlow(context.Background(), 1, "scan it", "anthropic", "anthropic", nil, nil)
	require.NoError(t, err)

	return flowID
}

func TestFlows_CreateFlow_AnswersAndAnnouncesTheFlowBeforePreparingIt(t *testing.T) {
	c := newCreatingController(t, nil)

	answered := make(chan int64, 1)
	go func() {
		flowID, _ := c.fc.CreateFlow(context.Background(), 1, "scan it", "anthropic", "anthropic", nil, nil)
		answered <- flowID
	}()

	assert.Equal(t, reservedFlowID, controllerReceive(t, answered, "creating a flow while its sandbox is prepared"))
	created := c.pub.created()
	require.Len(t, created, 1, "the owner sees the flow before its sandbox is ready")
	assert.Equal(t, database.FlowStatusCreated, created[0].Status)

	c.waitForPreparation(t)
	close(c.release)
	c.waitUntilSettled(t)
}

func TestFlows_GetFlow_ServesAFlowOnlyOnceItIsPrepared(t *testing.T) {
	c := newCreatingController(t, nil)

	flowID := c.create(t)
	c.waitForPreparation(t)

	flow, err := c.fc.GetFlow(context.Background(), flowID)
	assert.Nil(t, flow, "a flow with no sandbox yet cannot serve anything")
	assert.ErrorIs(t, err, ErrFlowNotFound)
	assert.Empty(t, c.fc.ListFlows(context.Background()))

	close(c.release)
	c.waitUntilSettled(t)

	ready, err := c.fc.GetFlow(context.Background(), flowID)
	require.NoError(t, err, "a prepared flow is served")
	assert.Same(t, c.worker, ready)
}

func TestFlows_PrepareFlow_OutlivesTheRequestThatAskedForIt(t *testing.T) {
	c := newCreatingController(t, nil)

	ctx, cancel := context.WithCancel(context.Background())
	flowID, err := c.fc.CreateFlow(ctx, 1, "scan it", "anthropic", "anthropic", nil, nil)
	require.NoError(t, err)

	c.waitForPreparation(t)
	cancel()
	close(c.release)
	c.waitUntilSettled(t)

	ready, err := c.fc.GetFlow(context.Background(), flowID)
	require.NoError(t, err)
	assert.Same(t, c.worker, ready)
}

func TestFlows_FailFlow_RecordsWhyAndMarksTheFlowFailed(t *testing.T) {
	c := newCreatingController(t, errors.New("sandbox image is unreachable"))

	flowID := c.create(t)
	c.waitForPreparation(t)
	close(c.release)
	c.waitUntilSettled(t)

	require.Eventually(t, func() bool {
		return slices.Contains(c.q.recordedStatuses(), database.FlowStatusFailed)
	}, 5*time.Second, 10*time.Millisecond, "a flow the owner already saw must not vanish when its preparation breaks")

	messages := c.q.recordedMessages()
	require.Len(t, messages, 1, "the owner is told why the flow could not start")
	assert.Equal(t, database.MsglogTypeReport, messages[0].Type)
	assert.Contains(t, messages[0].Message, "sandbox image is unreachable")
	assert.Equal(t, flowID, messages[0].FlowID)
}

func TestFlows_FinishFlow_CancelsAPreparationWithoutWaitingForIt(t *testing.T) {
	c := newCreatingController(t, nil)
	c.holdsAfterCancel = true

	c.create(t)
	c.waitForPreparation(t)
	c.finishWhilePreparing(t)

	controllerReceive(t, c.cancelled, "the preparation of a closed flow to stop calling its provider")
	close(c.release)
	c.waitUntilSettled(t)
	assert.False(t, c.registered(), "a flow finished while preparing does not stay in the registry")
}

func TestFlows_CreateAssistant_AnswersFirstAndFailsTheAssistantOfAFlowThatCannotStart(t *testing.T) {
	c := newCreatingController(t, errors.New("sandbox image is unreachable"))

	answered := make(chan int64, 1)
	go func() {
		assistantID, _ := c.fc.CreateAssistant(
			context.Background(), 1, 0, "hello", false, "anthropic", "anthropic", nil, nil,
		)
		answered <- assistantID
	}()

	assert.Equal(t, reservedAssistantID,
		controllerReceive(t, answered, "creating an assistant while its flow's sandbox is prepared"))
	created, _ := c.pub.assistants()
	require.Len(t, created, 1, "the owner sees the assistant before its flow is ready")

	c.waitForPreparation(t)
	close(c.release)
	c.waitUntilSettled(t)

	assert.Eventually(t, func() bool {
		return slices.Contains(c.q.recordedAssistantStatuses(), database.AssistantStatusFailed)
	}, 5*time.Second, 5*time.Millisecond,
		"an assistant whose flow could not start stops looking like it is about to answer")
}

func TestFlows_CreateAssistant_StillPreparesTheFlowWhenTheAssistantCannotBeReserved(t *testing.T) {
	c := newCreatingController(t, nil)
	reserveErr := errors.New("assistants table is unavailable")
	c.q.mx.Lock()
	c.q.assistantErr = reserveErr
	c.q.mx.Unlock()

	_, err := c.fc.CreateAssistant(context.Background(), 1, 0, "hello", false, "anthropic", "anthropic", nil, nil)
	require.ErrorIs(t, err, reserveErr)

	c.waitForPreparation(t)
	close(c.release)
	c.waitUntilSettled(t)

	flow, err := c.fc.GetFlow(context.Background(), reservedFlowID)
	require.NoError(t, err, "a flow nobody finishes preparing answers not ready for the life of the process")
	assert.Same(t, c.worker, flow)
}

func TestFlows_FlowForAssistant_GivesTwoConcurrentLoadersOneSlot(t *testing.T) {
	c := newCreatingController(t, nil)

	const flowID = int64(77)
	c.q.mx.Lock()
	c.q.flow = database.Flow{ID: flowID, UserID: 1, Status: database.FlowStatusFinished}
	c.q.readDelay = 20 * time.Millisecond
	c.q.mx.Unlock()

	slots := make(chan *flowEntry, 2)
	start := make(chan struct{})
	for range 2 {
		go func() {
			<-start
			entry, _, _, err := c.fc.flowForAssistant(context.Background(), flowID, newFlowWorkerCtx{
				userID:        1,
				flowWorkerCtx: c.fc.flowWorkerCtx(),
			})
			if err != nil {
				entry = nil
			}
			slots <- entry
		}()
	}

	close(start)
	first := controllerReceive(t, slots, "the first loader")
	second := controllerReceive(t, slots, "the second loader")

	c.fc.mx.Lock()
	registered := c.fc.flows[flowID]
	c.fc.mx.Unlock()

	require.NotNil(t, registered, "one of the two callers takes the flow")
	// Compared as bare pointers: formatting a flowEntry reads fields a preparation may write.
	for _, slot := range []*flowEntry{first, second} {
		if slot != nil {
			assert.True(t, slot == registered,
				"a second slot means a second worker, and the loser's teardown releases the winner's sandbox")
		}
	}
}

func TestFlows_CommitPreparation_RefusesAFlowClosedBeforeTheCommit(t *testing.T) {
	c := newCreatingController(t, nil)
	c.commitsAfterRelease = true

	c.create(t)
	c.waitForPreparation(t)
	c.finishWhilePreparing(t)
	close(c.release)

	assert.ErrorIs(t, controllerReceive(t, c.committed, "the build to reach its commit point"), errFlowAbandoned,
		"a build not told the flow is gone starts the sandbox and runs its input")
	c.waitUntilSettled(t)
	assert.False(t, c.registered())
	assert.NotContains(t, c.q.recordedStatuses(), database.FlowStatusFailed, "closing a flow is not a failure")
}

func TestFlows_FinishFlow_WaitsForACommittedPreparation(t *testing.T) {
	for _, tc := range []struct {
		name     string
		buildErr error
	}{
		{name: "a preparation that completes has its worker finished"},
		{name: "a preparation that fails leaves nothing to finish", buildErr: errors.New("the sandbox refused the input")},
	} {
		t.Run(tc.name, func(t *testing.T) {
			c := newCreatingController(t, tc.buildErr)
			c.commitsBeforeRelease = true
			c.finishing = &finishRecordingWorker{noopFlowWorker: noopFlowWorker{flowID: reservedFlowID}}

			c.create(t)
			c.waitForPreparation(t)

			finished := make(chan error, 1)
			go func() { finished <- c.fc.FinishFlow(context.Background(), reservedFlowID) }()

			select {
			case <-finished:
				t.Fatal("finishing a flow whose sandbox is starting answered before the preparation ended")
			case <-time.After(100 * time.Millisecond):
			}
			close(c.release)

			require.NoError(t, controllerReceive(t, finished, "finishing once the preparation ended"))
			assert.Equal(t, tc.buildErr == nil, c.finishing.finished.Load(),
				"the worker the preparation built is not left running")
			assert.False(t, c.registered())
		})
	}
}

func TestFlows_PrepareFlow_FinishesTheWorkerOfAnAbandonedPreparationOnALiveContext(t *testing.T) {
	c := newCreatingController(t, nil)
	c.ignoresCancel = true
	c.finishing = &finishRecordingWorker{
		noopFlowWorker: noopFlowWorker{flowID: reservedFlowID},
		seen:           make(chan error, 1),
	}

	c.create(t)
	c.waitForPreparation(t)
	c.finishWhilePreparing(t)
	close(c.release)

	assert.NoError(t, controllerReceive(t, c.finishing.seen, "the worker of the abandoned preparation to be finished"),
		"a worker finished on a cancelled context keeps its sandbox")
	c.waitUntilSettled(t)
}

type gatedFlowWorker struct {
	noopFlowWorker

	entered chan string
	release chan struct{}

	mx       sync.Mutex
	inside   int
	overlaps bool
}

func (w *gatedFlowWorker) gate(name string) {
	w.mx.Lock()
	w.inside++
	if w.inside > 1 {
		w.overlaps = true
	}
	w.mx.Unlock()

	w.entered <- name
	<-w.release

	w.mx.Lock()
	w.inside--
	w.mx.Unlock()
}

func (w *gatedFlowWorker) sawOverlap() bool {
	w.mx.Lock()
	defer w.mx.Unlock()

	return w.overlaps
}

func (w *gatedFlowWorker) Stop(context.Context) error           { w.gate("stop"); return nil }
func (w *gatedFlowWorker) Finish(context.Context) error         { w.gate("finish"); return nil }
func (w *gatedFlowWorker) Rename(context.Context, string) error { w.gate("rename"); return nil }

const (
	gatedFlowID = int64(7)
	otherFlowID = int64(8)
)

// gatedDocker holds the container removal of an unloaded flow in the same gate.
type gatedDocker struct {
	docker.DockerClient

	gated *gatedFlowWorker
}

func (d *gatedDocker) RemoveContainer(ctx context.Context, _ string, _ int64) error {
	d.gated.gate("remove container")

	return ctx.Err()
}

func newGatedController(t *testing.T) (*flowController, *gatedFlowWorker) {
	t.Helper()

	fc, _, _ := newFinishController(&finishFakeQuerier{
		flow: database.Flow{ID: gatedFlowID, Status: database.FlowStatusRunning},
		containers: []database.Container{{
			ID:      1,
			FlowID:  gatedFlowID,
			Status:  database.ContainerStatusRunning,
			LocalID: sql.NullString{String: "cid-1", Valid: true},
		}},
	})

	gated := &gatedFlowWorker{
		noopFlowWorker: noopFlowWorker{flowID: gatedFlowID},
		entered:        make(chan string, 4),
		release:        make(chan struct{}),
	}
	fc.docker = &gatedDocker{gated: gated}
	fc.register(gatedFlowID, gated)
	fc.register(otherFlowID, &noopFlowWorker{flowID: otherFlowID})

	return fc, gated
}

// blockedOn starts call and returns once it is held inside the gated worker.
func blockedOn(t *testing.T, worker *gatedFlowWorker, call func() error) chan error {
	t.Helper()

	done := make(chan error, 1)
	go func() { done <- call() }()
	controllerReceive(t, worker.entered, "the flow to enter the blocking call")

	return done
}

func answersWhile(t *testing.T, worker *gatedFlowWorker, what string, probe func()) {
	t.Helper()

	answered := make(chan struct{})
	go func() {
		probe()
		close(answered)
	}()

	select {
	case <-answered:
	case <-time.After(2 * time.Second):
		close(worker.release)
		t.Fatalf("%s waited for another lifecycle call: one flow freezes the whole controller", what)
	}
}

func waitUntilClosing(t *testing.T, fc *flowController, flowID int64) {
	t.Helper()

	require.Eventually(t, func() bool {
		fc.mx.Lock()
		defer fc.mx.Unlock()

		entry, ok := fc.flows[flowID]

		return ok && entry.closing
	}, 5*time.Second, time.Millisecond, "the flow never entered its closing state")
}

// Subtests are keyed by the lifecycle call held open on one flow.
func TestFlows_OneFlowsLifecycleCallDoesNotFreezeTheController(t *testing.T) {
	for _, tc := range []struct {
		name     string
		unloaded bool
		call     func(fc *flowController) error
	}{
		{name: "a stop in progress", call: func(fc *flowController) error {
			return fc.StopFlow(context.Background(), gatedFlowID)
		}},
		{name: "a rename in progress", call: func(fc *flowController) error {
			return fc.RenameFlow(context.Background(), gatedFlowID, "slow")
		}},
		{name: "a finish in progress", call: func(fc *flowController) error {
			return fc.FinishFlow(context.Background(), gatedFlowID)
		}},
		{name: "a finish of an unloaded flow in progress", unloaded: true, call: func(fc *flowController) error {
			return fc.FinishFlow(context.Background(), gatedFlowID)
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			fc, gated := newGatedController(t)
			if tc.unloaded {
				delete(fc.flows, gatedFlowID)
			}
			held := blockedOn(t, gated, func() error { return tc.call(fc) })

			answersWhile(t, gated, "listing flows", func() { fc.ListFlows(context.Background()) })
			answersWhile(t, gated, "looking another flow up", func() {
				flow, err := fc.GetFlow(context.Background(), otherFlowID)
				assert.NoError(t, err)
				if assert.NotNil(t, flow) {
					assert.Equal(t, otherFlowID, flow.GetFlowID())
				}
			})
			answersWhile(t, gated, "renaming another flow", func() {
				assert.NoError(t, fc.RenameFlow(context.Background(), otherFlowID, "untouched"))
			})

			close(gated.release)
			require.NoError(t, controllerReceive(t, held, "the held call to return"))
		})
	}
}

func TestFlows_FinishFlow_TakesAClosingFlowOutOfService(t *testing.T) {
	obs.InitObserver(context.Background(), nil, nil, nil)

	for _, tc := range []struct {
		name     string
		unloaded bool
	}{
		{name: "a loaded flow"},
		{name: "an unloaded flow", unloaded: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			fc, gated := newGatedController(t)
			if tc.unloaded {
				delete(fc.flows, gatedFlowID)
			}
			finished := blockedOn(t, gated, func() error { return fc.FinishFlow(context.Background(), gatedFlowID) })

			answersWhile(t, gated, "looking the closing flow up", func() {
				flow, err := fc.GetFlow(context.Background(), gatedFlowID)
				assert.Nil(t, flow, "a flow whose containers are being stopped is not handed to a new caller")
				assert.ErrorIs(t, err, ErrFlowNotFound)
			})
			// An assistant on an unloaded flow waits for the teardown instead.
			if !tc.unloaded {
				answersWhile(t, gated, "creating an assistant on the closing flow", func() {
					assistantID, err := fc.CreateAssistant(
						context.Background(), 1, gatedFlowID, "hello", false, "anthropic", "anthropic", nil, nil,
					)
					assert.Zero(t, assistantID)
					assert.ErrorIs(t, err, ErrFlowNotFound, "a second worker for a flow that is shutting down would fight the first")
				})
			}

			fc.mx.Lock()
			entry, present := fc.flows[gatedFlowID]
			fc.mx.Unlock()
			require.True(t, present, "the closing flow keeps its slot until it is closed")
			if tc.unloaded {
				assert.Nil(t, entry.worker, "a worker was loaded for a flow still releasing its containers")
			} else {
				assert.Same(t, gated, entry.worker, "a second worker was loaded for a flow still stopping its containers")
			}

			close(gated.release)
			require.NoError(t, controllerReceive(t, finished, "finishing to return"))
		})
	}
}

// newTearingDownController serves the unloaded flow gatedFlowID, whose one container the database lists as running.
func newTearingDownController(t *testing.T) (*creatingController, *gatedFlowWorker) {
	t.Helper()

	c := newCreatingController(t, nil)
	c.q.mx.Lock()
	c.q.flow = database.Flow{ID: gatedFlowID, UserID: 1, Status: database.FlowStatusRunning}
	c.q.containers = []database.Container{{
		ID:      1,
		FlowID:  gatedFlowID,
		Status:  database.ContainerStatusRunning,
		LocalID: sql.NullString{String: "cid-1", Valid: true},
	}}
	c.q.mx.Unlock()

	gated := &gatedFlowWorker{
		noopFlowWorker: noopFlowWorker{flowID: gatedFlowID},
		entered:        make(chan string, 4),
		release:        make(chan struct{}),
	}
	c.fc.docker = &gatedDocker{gated: gated}

	return c, gated
}

// waitingBehind fails unless the caller answering on answered waits on ctx while gated holds a teardown.
func waitingBehind[T any](t *testing.T, gated *gatedFlowWorker, ctx *flowsWaitingCtx, answered <-chan T, what string) {
	t.Helper()

	select {
	case got := <-answered:
		close(gated.release)
		t.Fatalf("%s answered %+v while the flow's containers were still being removed", what, got)
	case <-ctx.waiting:
	case <-time.After(10 * time.Second):
		close(gated.release)
		t.Fatalf("%s never waited for the flow's teardown", what)
	}
}

func TestFlows_FinishFlow_WaitsForTheTeardownOfAnUnloadedFlowInProgress(t *testing.T) {
	for _, tc := range []struct {
		name    string
		givesUp bool
	}{
		{name: "it answers after the first finish and releases what that one left"},
		{name: "it answers its caller's cancellation", givesUp: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			fc, gated := newGatedController(t)
			delete(fc.flows, gatedFlowID)
			first := blockedOn(t, gated, func() error { return fc.FinishFlow(context.Background(), gatedFlowID) })

			parent, cancel := context.WithCancel(context.Background())
			defer cancel()
			ctx := &flowsWaitingCtx{Context: parent, waiting: make(chan struct{})}
			second := make(chan error, 1)
			go func() { second <- fc.FinishFlow(ctx, gatedFlowID) }()
			waitingBehind(t, gated, ctx, second, "a second finish")

			if tc.givesUp {
				cancel()
				assert.ErrorIs(t, controllerReceive(t, second, "the second finish to give up"), context.Canceled)
				close(gated.release)
				require.NoError(t, controllerReceive(t, first, "the first finish"))

				return
			}

			close(gated.release)
			require.NoError(t, controllerReceive(t, first, "the first finish"))
			require.NoError(t, controllerReceive(t, second, "the second finish"))
			select {
			case name := <-gated.entered:
				assert.Equal(t, "remove container", name)
			default:
				t.Error("the second finish trusted the first with a container the database still lists as running")
			}
			assert.NotContains(t, fc.flows, gatedFlowID)
		})
	}
}

type flowsCallerKey struct{}

// flowsCallerCtx names its caller to steppedDocker and sends on waits every time Done is read, which is every time
// a queued caller starts to wait.
type flowsCallerCtx struct {
	context.Context

	waits chan struct{}
}

func newFlowsCallerCtx(name string) *flowsCallerCtx {
	return &flowsCallerCtx{
		Context: context.WithValue(context.Background(), flowsCallerKey{}, name),
		waits:   make(chan struct{}, 8),
	}
}

func (c *flowsCallerCtx) Done() <-chan struct{} {
	select {
	case c.waits <- struct{}{}:
	default:
	}

	return c.Context.Done()
}

// steppedDocker holds each container removal until the release of the caller that asked for it.
type steppedDocker struct {
	docker.DockerClient

	entered  chan string
	releases map[string]chan struct{}
}

func (d *steppedDocker) RemoveContainer(ctx context.Context, _ string, _ int64) error {
	name, _ := ctx.Value(flowsCallerKey{}).(string)
	d.entered <- name
	<-d.releases[name]

	return ctx.Err()
}

func TestFlows_FinishFlow_RunsTheQueuedTeardownsOfAnUnloadedFlowOneAtATime(t *testing.T) {
	fc, _ := newGatedController(t)
	delete(fc.flows, gatedFlowID)
	d := &steppedDocker{entered: make(chan string, 4), releases: map[string]chan struct{}{}}
	for _, name := range []string{"a", "b", "c"} {
		d.releases[name] = make(chan struct{})
	}
	fc.docker = d
	released := map[string]bool{}
	release := func(name string) {
		released[name] = true
		close(d.releases[name])
	}
	t.Cleanup(func() {
		for name, ch := range d.releases {
			if !released[name] {
				close(ch)
			}
		}
	})

	first := make(chan error, 1)
	go func() { first <- fc.FinishFlow(newFlowsCallerCtx("a"), gatedFlowID) }()
	require.Equal(t, "a", controllerReceive(t, d.entered, "the first finish to begin its teardown"))

	callers := map[string]*flowsCallerCtx{"b": newFlowsCallerCtx("b"), "c": newFlowsCallerCtx("c")}
	answers := map[string]chan error{}
	for name, ctx := range callers {
		answered := make(chan error, 1)
		answers[name] = answered
		go func() { answered <- fc.FinishFlow(ctx, gatedFlowID) }()
		controllerReceive(t, ctx.waits, "finish "+name+" to wait for the first teardown")
	}

	release("a")
	require.NoError(t, controllerReceive(t, first, "the first finish"))
	winner := controllerReceive(t, d.entered, "a queued finish to begin its teardown")
	loser := map[string]string{"b": "c", "c": "b"}[winner]
	require.NotEmpty(t, loser, "the teardown was begun by %q, which is not a queued finish", winner)

	select {
	case err := <-answers[loser]:
		t.Fatalf("finish %s answered %v while the teardown of finish %s was still held", loser, err, winner)
	case name := <-d.entered:
		t.Fatalf("finish %s began a teardown while the one of finish %s was still held", name, winner)
	case <-callers[loser].waits:
	case <-time.After(10 * time.Second):
		t.Fatalf("finish %s never waited for the teardown of finish %s", loser, winner)
	}

	release(winner)
	require.NoError(t, controllerReceive(t, answers[winner], "finish "+winner))
	require.Equal(t, loser, controllerReceive(t, d.entered, "finish "+loser+" to begin its teardown"))
	release(loser)
	require.NoError(t, controllerReceive(t, answers[loser], "finish "+loser))

	fc.mx.Lock()
	defer fc.mx.Unlock()
	assert.NotContains(t, fc.flows, gatedFlowID)
}

func TestFlows_CreateAssistant_WaitsForTheTeardownOfAnUnloadedFlowInProgress(t *testing.T) {
	obs.InitObserver(context.Background(), nil, nil, nil)

	type answer struct {
		assistantID int64
		err         error
	}

	for _, tc := range []struct {
		name    string
		givesUp bool
	}{
		{name: "it reloads the flow once the teardown ends"},
		{name: "it answers its caller's cancellation", givesUp: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			c, gated := newTearingDownController(t)
			finished := blockedOn(t, gated, func() error { return c.fc.FinishFlow(context.Background(), gatedFlowID) })

			parent, cancel := context.WithCancel(context.Background())
			defer cancel()
			ctx := &flowsWaitingCtx{Context: parent, waiting: make(chan struct{})}
			answered := make(chan answer, 1)
			go func() {
				assistantID, err := c.fc.CreateAssistant(
					ctx, 1, gatedFlowID, "hello", false, "anthropic", "anthropic", nil, nil,
				)
				answered <- answer{assistantID: assistantID, err: err}
			}()
			waitingBehind(t, gated, ctx, answered, "creating an assistant")

			if tc.givesUp {
				cancel()
				got := controllerReceive(t, answered, "creating an assistant to give up")
				assert.Zero(t, got.assistantID)
				assert.ErrorIs(t, got.err, context.Canceled)
				close(gated.release)
				require.NoError(t, controllerReceive(t, finished, "the finish"))
				created, _ := c.pub.assistants()
				assert.Empty(t, created)

				return
			}

			close(gated.release)
			require.NoError(t, controllerReceive(t, finished, "the finish"))
			got := controllerReceive(t, answered, "creating an assistant")
			require.NoError(t, got.err)
			assert.Equal(t, reservedAssistantID, got.assistantID)

			require.Eventually(t, func() bool {
				return slices.Contains(c.q.recordedAssistantStatuses(), database.AssistantStatusFailed)
			}, 5*time.Second, 5*time.Millisecond, "the reload never ended")
			statuses := c.q.recordedStatuses()
			require.GreaterOrEqual(t, len(statuses), 2)
			assert.Equal(t, []database.FlowStatus{database.FlowStatusFinished, database.FlowStatusWaiting}, statuses[:2],
				"the flow is renewed only after its teardown has finished it")
		})
	}
}

// The window is narrow, so the scenario runs many times.
func TestFlows_Claim_RefusesAFlowThatFinishedWhileTheCallerQueued(t *testing.T) {
	for range 200 {
		fc, gated := newGatedController(t)

		renamed := blockedOn(t, gated, func() error {
			return fc.RenameFlow(context.Background(), gatedFlowID, "slow")
		})

		stopped := make(chan error, 1)
		go func() { stopped <- fc.StopFlow(context.Background(), gatedFlowID) }()

		finished := make(chan error, 1)
		go func() { finished <- fc.FinishFlow(context.Background(), gatedFlowID) }()

		waitUntilClosing(t, fc, gatedFlowID)
		close(gated.release)

		require.NoError(t, controllerReceive(t, renamed, "renaming to return"))
		require.NoError(t, controllerReceive(t, finished, "finishing to return"))
		require.ErrorIs(t, controllerReceive(t, stopped, "stopping to answer"), ErrFlowNotLoaded,
			"a flow that finished while the caller queued takes no other lifecycle call")
		require.False(t, gated.sawOverlap(), "two lifecycle calls were inside the same flow at once")
	}
}

func TestFlows_FinishFlow_WaitsForTheLifecycleCallInProgress(t *testing.T) {
	fc, gated := newGatedController(t)

	stopped := blockedOn(t, gated, func() error { return fc.StopFlow(context.Background(), gatedFlowID) })

	finished := make(chan error, 1)
	go func() { finished <- fc.FinishFlow(context.Background(), gatedFlowID) }()

	select {
	case name := <-gated.entered:
		close(gated.release)
		t.Fatalf("%q ran on a flow that was already stopping", name)
	case <-time.After(500 * time.Millisecond):
	}

	close(gated.release)
	require.NoError(t, controllerReceive(t, stopped, "stopping to return"))
	require.NoError(t, controllerReceive(t, finished, "finishing to return"))
	assert.False(t, gated.sawOverlap(), "two lifecycle calls were inside the same flow at once")
}

// flowsWaitingCtx closes waiting the first time Done is read, which is when a queued caller starts to wait.
type flowsWaitingCtx struct {
	context.Context

	once    sync.Once
	waiting chan struct{}
}

func (c *flowsWaitingCtx) Done() <-chan struct{} {
	c.once.Do(func() { close(c.waiting) })

	return c.Context.Done()
}

func TestFlows_StopFlow_ReleasesAQueuedCallerThatCancels(t *testing.T) {
	fc, gated := newGatedController(t)

	renamed := blockedOn(t, gated, func() error {
		return fc.RenameFlow(context.Background(), gatedFlowID, "slow")
	})

	parent, cancel := context.WithCancel(context.Background())
	ctx := &flowsWaitingCtx{Context: parent, waiting: make(chan struct{})}
	stopped := make(chan error, 1)
	go func() { stopped <- fc.StopFlow(ctx, gatedFlowID) }()

	controllerReceive(t, ctx.waiting, "stopping to wait on its context")
	cancel()

	select {
	case err := <-stopped:
		assert.ErrorIs(t, err, context.Canceled, "a flow that exists is not answered as missing")
	case <-time.After(2 * time.Second):
		close(gated.release)
		t.Fatal("stopping queued behind work that can run for minutes with no way out")
	}

	close(gated.release)
	require.NoError(t, controllerReceive(t, renamed, "renaming to return"))
}

func TestFlows_FinishFlow_OutlivesTheRequestThatAskedForIt(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name   string
		loaded bool
	}{
		{name: "a loaded flow", loaded: true},
		{name: "an unloaded flow"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			const flowID = int64(7)

			q := &finishFakeQuerier{flow: database.Flow{ID: flowID, Status: database.FlowStatusWaiting}}
			fc, _, _ := newFinishController(q)
			worker := &finishRecordingWorker{noopFlowWorker: noopFlowWorker{flowID: flowID}, seen: make(chan error, 1)}
			if tc.loaded {
				fc.register(flowID, worker)
			}

			ctx, cancel := context.WithCancel(context.Background())
			cancel()

			require.NoError(t, fc.FinishFlow(ctx, flowID), "a client that hung up does not leave the flow half closed")

			if tc.loaded {
				assert.NoError(t, controllerReceive(t, worker.seen, "the worker to be finished"),
					"the worker was handed the context the client had already cancelled")

				return
			}
			require.Len(t, q.flowStatus, 1)
			assert.Equal(t, database.FlowStatusFinished, q.flowStatus[0].Status)
		})
	}
}

type budgetEatingWorker struct {
	noopFlowWorker
}

func (w *budgetEatingWorker) Finish(ctx context.Context) error {
	<-ctx.Done()

	return ctx.Err()
}

func TestFlows_FinishFlow_ClosesTheLeftoversOfAWorkerThatDidNotFinish(t *testing.T) {
	t.Parallel()

	const flowID = int64(7)
	dockerDown := errors.New("docker is down")

	for _, tc := range []struct {
		name    string
		worker  FlowWorker
		budget  time.Duration
		wantErr error
	}{
		{
			name:    "a worker that fails to finish",
			worker:  &finishRecordingWorker{noopFlowWorker: noopFlowWorker{flowID: flowID}, err: dockerDown},
			budget:  time.Minute,
			wantErr: dockerDown,
		},
		{
			name:    "a worker that spends the whole budget",
			worker:  &budgetEatingWorker{noopFlowWorker: noopFlowWorker{flowID: flowID}},
			budget:  20 * time.Millisecond,
			wantErr: context.DeadlineExceeded,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			q := &finishFakeQuerier{
				flow:     database.Flow{ID: flowID, Status: database.FlowStatusWaiting},
				tasks:    []database.Task{{ID: 11, Status: database.TaskStatusRunning}},
				subtasks: map[int64][]database.Subtask{11: {{ID: 21, Status: database.SubtaskStatusRunning}}},
			}
			fc, _, _ := newFinishController(q)
			fc.register(flowID, tc.worker)

			require.ErrorIs(t, fc.finishFlowWithin(context.Background(), flowID, tc.budget), tc.wantErr,
				"the caller still learns the sandbox was left behind")

			require.Len(t, q.taskStatus, 1, "a task only the database knows about is closed by nobody else")
			assert.Equal(t, int64(11), q.taskStatus[0].ID)
			require.Len(t, q.subtaskStatus, 1)
			assert.Equal(t, int64(21), q.subtaskStatus[0].ID)
		})
	}
}

func TestFlows_FinishFlow_StartsItsBudgetOnceItsTurnComes(t *testing.T) {
	t.Parallel()

	const flowID = int64(7)

	q := &finishFakeQuerier{flow: database.Flow{ID: flowID, Status: database.FlowStatusWaiting}}
	fc, _, _ := newFinishController(q)
	entry := fc.register(flowID, newFinishableWorker(q, &callerFakeExecutor{}))

	entry.lock()
	released := make(chan struct{})
	go func() {
		time.Sleep(60 * time.Millisecond)
		entry.release()
		close(released)
	}()

	require.NoError(t, fc.finishFlowWithin(context.Background(), flowID, 20*time.Millisecond),
		"a flow waiting its turn still has time left to be recorded")
	controllerReceive(t, released, "the other lifecycle call to end")

	require.Len(t, q.flowStatus, 1)
	assert.Equal(t, database.FlowStatusFinished, q.flowStatus[0].Status)
}

func TestFlows_FinishFlow_LeavesNoSlotSoARetryClosesWhatItLeftOpen(t *testing.T) {
	t.Parallel()

	const flowID = int64(36)
	tasksErr := errors.New("connection reset")

	for _, tc := range []struct {
		name           string
		loaded         bool
		wantFlowStatus []database.UpdateFlowStatusParams
	}{
		{name: "a loaded flow whose worker finished the row", loaded: true},
		{
			name:           "an unloaded flow",
			wantFlowStatus: []database.UpdateFlowStatusParams{{Status: database.FlowStatusFinished, ID: flowID}},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			q := &finishFakeQuerier{
				flow:     database.Flow{ID: flowID, UserID: 5, Status: database.FlowStatusWaiting},
				tasks:    []database.Task{{ID: 5, Status: database.TaskStatusRunning}},
				tasksErr: tasksErr,
				subtasks: map[int64][]database.Subtask{5: {{ID: 11, Status: database.SubtaskStatusRunning}}},
			}
			fc, pub, _ := newFinishController(q)
			worker := &finishRecordingWorker{noopFlowWorker: noopFlowWorker{flowID: flowID}}
			if tc.loaded {
				fc.register(flowID, worker)
			}

			require.ErrorIs(t, fc.FinishFlow(context.Background(), flowID), tasksErr)
			assert.Equal(t, tc.loaded, worker.finished.Load())
			assert.NotContains(t, fc.flows, flowID, "a slot left behind turns every later finish into a no-op")

			q.tasksErr = nil
			if tc.loaded {
				q.flow.Status = database.FlowStatusFinished
			}

			require.NoError(t, fc.FinishFlow(context.Background(), flowID))
			require.Len(t, q.taskStatus, 1, "the retry closes the task the first attempt left open")
			require.Len(t, q.subtaskStatus, 1)
			assert.Len(t, pub.taskUpdated, 1)
			assert.Equal(t, tc.wantFlowStatus, q.flowStatus)
			assert.NotContains(t, fc.flows, flowID)
		})
	}
}

func TestFlows_PrepareAssistant_DoesNotHoldTheFlowAgainstStopOrRename(t *testing.T) {
	obs.InitObserver(context.Background(), nil, nil, nil)

	const flowID = int64(42)

	fc, _, _ := newFinishController(&finishFakeQuerier{
		flow: database.Flow{ID: flowID, Status: database.FlowStatusRunning},
	})
	entry := fc.register(flowID, &noopFlowWorker{flowID: flowID})

	provs := &lifecycleProviders{entered: make(chan struct{}), release: make(chan struct{})}
	var releaseOnce sync.Once
	release := func() { releaseOnce.Do(func() { close(provs.release) }) }
	t.Cleanup(release)

	awc := lifecycleAssistantCtx(&lifecycleQuerier{}, provs)
	assistant, err := reserveAssistant(context.Background(), awc)
	require.NoError(t, err)

	prepared := make(chan struct{})
	go func() {
		fc.prepareAssistant(context.Background(), entry, assistant, awc)
		close(prepared)
	}()
	controllerReceive(t, provs.entered, "the assistant's provider to be asked for")

	for name, call := range map[string]func() error{
		"stop":   func() error { return fc.StopFlow(context.Background(), flowID) },
		"rename": func() error { return fc.RenameFlow(context.Background(), flowID, "renamed") },
	} {
		answered := make(chan error, 1)
		go func() { answered <- call() }()

		select {
		case err := <-answered:
			require.NoError(t, err, name)
		case <-time.After(2 * time.Second):
			t.Fatalf("%s waited for the assistant being built on the flow", name)
		}
	}

	release()
	controllerReceive(t, prepared, "the assistant's preparation to end")
}
