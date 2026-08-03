package controller

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"testing"

	"pentagi/pkg/database"
	"pentagi/pkg/graph/subscriptions"
	"pentagi/pkg/providers"
	"pentagi/pkg/providers/provider"
	"pentagi/pkg/providers/tester/mock"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---- fakes -----------------------------------------------------------------

// cascadeFakeQuerier embeds a nil database.Querier so any method the code under
// test calls without a stub here panics loudly (signalling a test gap) instead
// of silently returning zero values.
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
	f.flowsCalls = append(f.flowsCalls, arg)
	if f.flowsErr != nil {
		return nil, f.flowsErr
	}
	return f.flowsResult, nil
}

func (f *cascadeFakeQuerier) UpdateAssistantsProviderNameByOldName(
	ctx context.Context, arg database.UpdateAssistantsProviderNameByOldNameParams,
) ([]database.Assistant, error) {
	f.assistantsCalls = append(f.assistantsCalls, arg)
	if f.assistantsErr != nil {
		return nil, f.assistantsErr
	}
	return f.assistantsResult, nil
}

func (f *cascadeFakeQuerier) GetFlowContainers(ctx context.Context, flowID int64) ([]database.Container, error) {
	if f.containersErr != nil {
		return nil, f.containersErr
	}
	return f.containersResult, nil
}

// cascadeFakePublisher records every FlowUpdated/AssistantUpdated call.
type cascadeFakePublisher struct {
	subscriptions.FlowPublisher

	mx               sync.Mutex
	flowUpdated      []database.Flow
	assistantUpdated []database.Assistant
}

func (p *cascadeFakePublisher) FlowUpdated(ctx context.Context, flow database.Flow, terms []database.Container) {
	p.mx.Lock()
	defer p.mx.Unlock()
	p.flowUpdated = append(p.flowUpdated, flow)
}

func (p *cascadeFakePublisher) AssistantUpdated(ctx context.Context, assistant database.Assistant) {
	p.mx.Lock()
	defer p.mx.Unlock()
	p.assistantUpdated = append(p.assistantUpdated, assistant)
}

// cascadeFakeSubscriptions always hands out the same publisher regardless of
// userID/flowID, so a test can assert on everything published in one place.
type cascadeFakeSubscriptions struct {
	subscriptions.SubscriptionsController

	pub *cascadeFakePublisher
}

func (s *cascadeFakeSubscriptions) NewFlowPublisher(userID, flowID int64) subscriptions.FlowPublisher {
	return s.pub
}

// cascadeFakeProviders answers the "does the old name still resolve?" probe.
// resolvable lists the names that do; anything else fails, like a custom
// provider name that no built-in answers to.
type cascadeFakeProviders struct {
	providers.ProviderController

	resolvable map[provider.ProviderName]bool
}

func (p *cascadeFakeProviders) GetProvider(
	ctx context.Context, prvname provider.ProviderName, userID int64,
) (provider.Provider, error) {
	if p.resolvable[prvname] {
		return mock.NewProvider(provider.ProviderQwen, prvname, "model"), nil
	}
	return nil, fmt.Errorf("provider not found by name '%s'", prvname)
}

// newTestFlowController builds a flowController with fake dependencies,
// bypassing NewFlowController (which needs a real docker client the provider
// reassignment does not touch). flows stays empty on purpose: the reassignment
// is DB-only and must never reach into a worker. resolvable names the providers
// that still answer after the change, which is what decides whether a rewrite
// is needed at all.
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
		flows: map[int64]FlowWorker{},
		subs:  &cascadeFakeSubscriptions{pub: pub},
		provs: &cascadeFakeProviders{resolvable: names},
	}, pub
}

// ---- tests ------------------------------------------------------------------

func TestReassignFlowsProvider_RenameSweepsFlowsAndAssistants(t *testing.T) {
	const userID = int64(1)

	q := &cascadeFakeQuerier{
		flowsResult:      []database.Flow{{ID: 10}, {ID: 11}},
		assistantsResult: []database.Assistant{{ID: 100, FlowID: 10}},
	}
	fc, pub := newTestFlowController(q)

	err := fc.RenameFlowsProvider(context.Background(), userID, "my-qwen", "my-qwen-renamed")
	require.NoError(t, err)

	require.Len(t, q.flowsCalls, 1)
	assert.Equal(t, userID, q.flowsCalls[0].UserID, "the sweep must be scoped to the acting user")
	assert.Equal(t, "my-qwen", q.flowsCalls[0].OldName)
	assert.Equal(t, "my-qwen-renamed", q.flowsCalls[0].NewName)

	require.Len(t, q.assistantsCalls, 1)
	assert.Equal(t, userID, q.assistantsCalls[0].UserID)
	assert.Equal(t, "my-qwen", q.assistantsCalls[0].OldName)
	assert.Equal(t, "my-qwen-renamed", q.assistantsCalls[0].NewName)

	assert.Len(t, pub.flowUpdated, 2, "every rewritten flow row must be published")
	assert.Len(t, pub.assistantUpdated, 1, "every rewritten assistant row must be published")
}

func TestReassignFlowsProvider_ResetUsesProviderTypeAsDefaultName(t *testing.T) {
	q := &cascadeFakeQuerier{}
	fc, _ := newTestFlowController(q)

	err := fc.ResetFlowsProviderToDefault(context.Background(), 1, "my-custom-qwen", provider.ProviderQwen)
	require.NoError(t, err)

	require.Len(t, q.flowsCalls, 1)
	assert.Equal(t, "my-custom-qwen", q.flowsCalls[0].OldName)
	assert.Equal(t, string(provider.ProviderQwen), q.flowsCalls[0].NewName,
		"the default provider name is literally the provider type string")
}

// Deleting a user provider named exactly like the built-in of its type (an
// intentional way to override the product default) leaves the stored name
// already resolving to that built-in, so there is nothing to rewrite. Running
// flows drop the deleted configuration on their next input or on the next
// start — see flowProvider.SetProvider, which compares the raw config.
func TestReassignFlowsProvider_ShadowedDefaultNameIsNoOp(t *testing.T) {
	q := &cascadeFakeQuerier{}
	fc, pub := newTestFlowController(q, "qwen")

	err := fc.ResetFlowsProviderToDefault(context.Background(), 1, "qwen", provider.ProviderQwen)
	require.NoError(t, err)

	assert.Empty(t, q.flowsCalls, "a no-op rename must not issue a self-matching UPDATE")
	assert.Empty(t, q.assistantsCalls)
	assert.Empty(t, pub.flowUpdated, "and must not republish every row of that provider")
}

// The override may also be named after a built-in of a DIFFERENT type (a
// "custom"-typed provider called "openai"). Deleting it must not drag those rows
// onto the "custom" built-in: the name "openai" resolves again on its own, and
// rewriting it would repoint rows that predate the override.
func TestReassignFlowsProvider_OldNameStillResolvesIsNoOp(t *testing.T) {
	q := &cascadeFakeQuerier{}
	fc, pub := newTestFlowController(q, "openai")

	err := fc.ResetFlowsProviderToDefault(context.Background(), 1, "openai", provider.ProviderCustom)
	require.NoError(t, err)

	assert.Empty(t, q.flowsCalls, "a still-resolving name must never be rewritten")
	assert.Empty(t, q.assistantsCalls)
	assert.Empty(t, pub.flowUpdated)
}

// ...but if that built-in is not enabled, the name really would dangle, so the
// reset must go through and point the rows at the deleted provider's own type.
func TestReassignFlowsProvider_UnresolvableOldNameIsRewritten(t *testing.T) {
	q := &cascadeFakeQuerier{}
	fc, _ := newTestFlowController(q) // nothing resolves

	err := fc.ResetFlowsProviderToDefault(context.Background(), 1, "openai", provider.ProviderCustom)
	require.NoError(t, err)

	require.Len(t, q.flowsCalls, 1)
	assert.Equal(t, "openai", q.flowsCalls[0].OldName)
	assert.Equal(t, string(provider.ProviderCustom), q.flowsCalls[0].NewName)
}

func TestReassignFlowsProvider_FlowsSweepErrorStillRunsAssistantsSweep(t *testing.T) {
	flowsErr := errors.New("flows update exploded")
	q := &cascadeFakeQuerier{
		flowsErr:         flowsErr,
		assistantsResult: []database.Assistant{{ID: 100, FlowID: 10}},
	}
	fc, pub := newTestFlowController(q)

	err := fc.RenameFlowsProvider(context.Background(), 1, "old", "new")

	require.Error(t, err)
	assert.ErrorIs(t, err, flowsErr, "the failure must be reported, not swallowed")
	require.Len(t, q.assistantsCalls, 1,
		"the tables are independent: a failure on one must not skip the other")
	assert.Len(t, pub.assistantUpdated, 1)
}

func TestReassignFlowsProvider_BothSweepErrorsAreReported(t *testing.T) {
	flowsErr, assistantsErr := errors.New("flows boom"), errors.New("assistants boom")
	q := &cascadeFakeQuerier{flowsErr: flowsErr, assistantsErr: assistantsErr}
	fc, _ := newTestFlowController(q)

	err := fc.RenameFlowsProvider(context.Background(), 1, "old", "new")

	require.Error(t, err)
	assert.ErrorIs(t, err, flowsErr)
	assert.ErrorIs(t, err, assistantsErr)
}

// FlowUpdated carries the flow's full terminal list and the client replaces its
// cached value wholesale, so publishing with no containers would blank the
// terminals in the UI. Better to skip the event than to corrupt the view.
func TestReassignFlowsProvider_SkipsPublishWhenContainerLookupFails(t *testing.T) {
	q := &cascadeFakeQuerier{
		flowsResult:   []database.Flow{{ID: 10}},
		containersErr: errors.New("containers unavailable"),
	}
	fc, pub := newTestFlowController(q)

	err := fc.RenameFlowsProvider(context.Background(), 1, "old", "new")
	require.NoError(t, err, "a container lookup failure is not a cascade failure")

	assert.Empty(t, pub.flowUpdated, "publishing an empty terminal list would wipe the client's cache")
	require.Len(t, q.flowsCalls, 1, "the rewrite itself must still have happened")
}

// The sweep is detached from the caller's context on purpose: the provider row
// is already committed by the time it runs, so a disconnecting HTTP client must
// not leave half of the references pointing at a name that no longer exists.
func TestReassignFlowsProvider_RunsDespiteCancelledCallerContext(t *testing.T) {
	q := &cascadeFakeQuerier{flowsResult: []database.Flow{{ID: 10}}}
	fc, _ := newTestFlowController(q)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := fc.RenameFlowsProvider(ctx, 1, "old", "new")
	require.NoError(t, err)

	assert.Len(t, q.flowsCalls, 1)
	assert.Len(t, q.assistantsCalls, 1)
}
