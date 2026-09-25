package controller

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"

	"pentagi/pkg/database"
	obs "pentagi/pkg/observability"
	"pentagi/pkg/providers"
	"pentagi/pkg/providers/embeddings"
	"pentagi/pkg/providers/pconfig"
	"pentagi/pkg/templates"
	"pentagi/pkg/tools"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type unavailableEmbedder struct {
	embeddings.Embedder
}

func (unavailableEmbedder) IsAvailable() bool { return false }

// lifecycleChain's chain blocks until cancelled when started is set, and otherwise answers err after delay.
type lifecycleChain struct {
	providers.AssistantProvider

	started   chan struct{}
	cancelled chan struct{}
	delay     time.Duration
	err       error
}

func (*lifecycleChain) Title() string                                      { return "assistant" }
func (*lifecycleChain) Language() string                                   { return "English" }
func (*lifecycleChain) ToolCallIDTemplate() string                         { return "call_%s" }
func (*lifecycleChain) Model(pconfig.ProviderOptionsType) string           { return "model" }
func (*lifecycleChain) PrepareAgentChain(context.Context) (int64, error)   { return 11, nil }
func (*lifecycleChain) Embedder() embeddings.Embedder                      { return unavailableEmbedder{} }
func (*lifecycleChain) SetAgentLogProvider(tools.AgentLogProvider)         {}
func (*lifecycleChain) SetMsgLogProvider(tools.MsgLogProvider)             {}
func (*lifecycleChain) SetFlowWorker(providers.FlowWorker)                 {}
func (*lifecycleChain) SetMsgChainID(int64)                                {}
func (*lifecycleChain) PutInputToAgentChain(context.Context, string) error { return nil }
func (*lifecycleChain) EnsureChainConsistency(context.Context) error       { return nil }

func (c *lifecycleChain) PerformAgentChain(ctx context.Context) error {
	if c.started == nil {
		time.Sleep(c.delay)

		return c.err
	}

	close(c.started)
	<-ctx.Done()
	close(c.cancelled)

	return ctx.Err()
}

// lifecycleFlowWorker records the assistant it is given, or refuses it the way a finished flow does.
type lifecycleFlowWorker struct {
	noopFlowWorker

	refuses bool
	mx      sync.Mutex
	added   AssistantWorker
}

func (w *lifecycleFlowWorker) AddAssistant(_ context.Context, aw AssistantWorker) error {
	if w.refuses {
		return fmt.Errorf("flow %d: %w", w.flowID, ErrFlowAlreadyStopped)
	}

	w.mx.Lock()
	defer w.mx.Unlock()
	w.added = aw

	return nil
}

func (w *lifecycleFlowWorker) registered() AssistantWorker {
	w.mx.Lock()
	defer w.mx.Unlock()

	return w.added
}

func TestAssistant_BuildAssistantWorker_MarksTheRowFailedOnlyForAFault(t *testing.T) {
	obs.InitObserver(context.Background(), nil, nil, nil)

	updateErr := errors.New("update failed")

	for _, tc := range []struct {
		name       string
		q          *lifecycleQuerier
		fw         *lifecycleFlowWorker
		wantErr    error
		wantFailed bool
	}{
		{
			name:       "a build that breaks marks the row it was given failed",
			q:          &lifecycleQuerier{updateErr: updateErr},
			fw:         &lifecycleFlowWorker{},
			wantErr:    updateErr,
			wantFailed: true,
		},
		{
			name:    "a flow that finished meanwhile leaves the row to its finish",
			q:       &lifecycleQuerier{},
			fw:      &lifecycleFlowWorker{refuses: true},
			wantErr: ErrFlowAlreadyStopped,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			chain := &lifecycleChain{started: make(chan struct{}), cancelled: make(chan struct{})}
			awc := lifecycleAssistantCtx(tc.q, &lifecycleProviders{provider: chain})
			awc.fw = tc.fw
			pub := awc.subs.(*cascadeFakeSubscriptions).pub

			assistant, err := reserveAssistant(context.Background(), awc)
			require.NoError(t, err)

			_, err = buildAssistantWorker(context.Background(), assistant, awc)

			require.ErrorIs(t, err, tc.wantErr)
			select {
			case <-chain.started:
				t.Fatal("an assistant that was not built ran its chain")
			default:
			}
			created, updated := pub.assistants()
			assert.Empty(t, created)
			if !tc.wantFailed {
				assert.Empty(t, tc.q.recordedStatuses())

				return
			}
			assert.Equal(t, []database.UpdateAssistantStatusParams{
				{Status: database.AssistantStatusFailed, ID: lifecycleAssistantID},
			}, tc.q.statusCalls)
			require.Len(t, updated, 1, "the caller was already told the assistant exists, so it is told it failed")
			assert.Equal(t, database.AssistantStatusFailed, updated[0].Status)
		})
	}
}

func TestAssistant_BuildAssistantWorker_RegistersAWorkerItsOwnerCanEnd(t *testing.T) {
	obs.InitObserver(context.Background(), nil, nil, nil)

	for _, tc := range []struct {
		name string
		end  func(AssistantWorker) error
	}{
		{name: "finished while its first input runs", end: func(aw AssistantWorker) error {
			return aw.Finish(context.Background())
		}},
		{name: "stopped while its first input runs", end: func(aw AssistantWorker) error {
			return aw.Stop(context.Background())
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			chain := &lifecycleChain{started: make(chan struct{}), cancelled: make(chan struct{})}
			q := &lifecycleQuerier{}
			fw := &lifecycleFlowWorker{noopFlowWorker: noopFlowWorker{flowID: 42}}
			awc := lifecycleAssistantCtx(q, &lifecycleProviders{provider: chain})
			awc.fw = fw

			assistant, err := reserveAssistant(context.Background(), awc)
			require.NoError(t, err)

			built := make(chan error, 1)
			go func() {
				_, err := buildAssistantWorker(context.Background(), assistant, awc)
				built <- err
			}()
			controllerReceive(t, chain.started, "the assistant chain to start")

			registered := fw.registered()
			require.NotNil(t, registered,
				"a running chain the flow does not know about cannot be stopped by deleting its assistant")
			require.NoError(t, tc.end(registered))
			controllerReceive(t, chain.cancelled, "the chain of the ended worker to stop")
			require.NoError(t, controllerReceive(t, built, "the build to return"), "an owner's stop is not a failed build")
			assert.NotContains(t, q.recordedStatuses(), database.AssistantStatusFailed)
			require.NoError(t, registered.Finish(context.Background()))
		})
	}
}

func loadRestoredAssistant(
	t *testing.T, q database.Querier, status database.AssistantStatus,
	log *lifecycleAssistantLog, chain providers.AssistantProvider,
) (AssistantWorker, error) {
	t.Helper()
	obs.InitObserver(context.Background(), nil, nil, nil)

	awc := assistantWorkerCtx{
		userID:        7,
		flowID:        42,
		fw:            &noopFlowWorker{flowID: 42},
		prompter:      templates.NewDefaultPrompter(),
		flowWorkerCtx: lifecycleAssistantCtx(q, &lifecycleProviders{provider: chain}).flowWorkerCtx,
	}
	awc.aslc = &lifecycleLogController{log: log}

	aw, err := LoadAssistantWorker(context.Background(), database.Assistant{
		ID:                5,
		FlowID:            42,
		Status:            status,
		Functions:         []byte("{}"),
		MsgchainID:        sql.NullInt64{Int64: 11, Valid: true},
		ModelProviderName: "custom",
	}, awc)
	if aw != nil {
		t.Cleanup(func() { _ = aw.Finish(context.Background()) })
	}

	return aw, err
}

func reportNotes(log *lifecycleAssistantLog) []lifecycleNote {
	var found []lifecycleNote
	for _, note := range log.recorded() {
		if note.msgType == database.MsglogTypeReport {
			found = append(found, note)
		}
	}

	return found
}

func TestAssistant_LoadAssistantWorker_NotesARunTheRestartCutShort(t *testing.T) {
	connReset := errors.New("connection reset")
	waiting := []database.AssistantStatus{database.AssistantStatusWaiting}

	for _, tc := range []struct {
		name         string
		status       database.AssistantStatus
		q            *lifecycleQuerier
		logErr       error
		wantErr      error
		wantStatuses []database.AssistantStatus
		wantNotes    []lifecycleNote
	}{
		{
			name:         "an assistant restored mid-run says its reply was cut short",
			status:       database.AssistantStatusRunning,
			q:            &lifecycleQuerier{},
			wantStatuses: waiting,
			wantNotes:    []lifecycleNote{{msgType: database.MsglogTypeReport, msg: interruptedRunNote}},
		},
		{
			name:         "an assistant restored while waiting is not marked",
			status:       database.AssistantStatusWaiting,
			q:            &lifecycleQuerier{},
			wantStatuses: waiting,
		},
		{
			name:         "a note that cannot be stored does not keep the assistant from loading",
			status:       database.AssistantStatusRunning,
			q:            &lifecycleQuerier{},
			logErr:       connReset,
			wantStatuses: waiting,
		},
		{
			name:    "no note is left when the assistant could not be set waiting",
			status:  database.AssistantStatusRunning,
			q:       &lifecycleQuerier{statusErr: connReset},
			wantErr: connReset,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			log := &lifecycleAssistantLog{err: tc.logErr}

			aw, err := loadRestoredAssistant(t, tc.q, tc.status, log, &lifecycleChain{})

			if tc.wantErr != nil {
				require.ErrorIs(t, err, tc.wantErr)
				assert.Empty(t, log.recorded(), "a note written before a failed load repeats on every restart")

				return
			}
			require.NoError(t, err)
			assert.NotNil(t, aw)
			assert.Equal(t, tc.wantStatuses, tc.q.recordedStatuses())
			assert.Equal(t, tc.wantNotes, log.recorded())
		})
	}
}

func TestAssistant_NoteInterruptedRun_WritesTheOwedNoteBeforeTheNextMessage(t *testing.T) {
	log := &lifecycleAssistantLog{err: errors.New("connection reset")}

	aw, err := loadRestoredAssistant(t, &lifecycleQuerier{}, database.AssistantStatusRunning, log, &lifecycleChain{})
	require.NoError(t, err)
	require.Empty(t, log.recorded())

	log.recovers()
	require.NoError(t, aw.PutInput(context.Background(), "go on", false, nil))

	require.Eventually(t, func() bool { return len(reportNotes(log)) == 1 }, 5*time.Second, 10*time.Millisecond)
	assert.Equal(t, lifecycleNote{msgType: database.MsglogTypeReport, msg: interruptedRunNote}, log.recorded()[0],
		"the note belongs above the message it explains")

	require.NoError(t, aw.PutInput(context.Background(), "and again", false, nil))
	assert.Never(t, func() bool { return len(reportNotes(log)) > 1 }, time.Second, 20*time.Millisecond)
}

func TestAssistant_ReportFailure_ReachesExactlyOneOfCallerAndLog(t *testing.T) {
	overloaded := errors.New("529: overloaded")

	for _, tc := range []struct {
		name     string
		delay    time.Duration
		toCaller bool
	}{
		{name: "a failure after the input was taken goes to the log", delay: assistantInputTimeout + 300*time.Millisecond},
		{name: "a failure at once goes to the caller only", delay: 50 * time.Millisecond, toCaller: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			log := &lifecycleAssistantLog{}
			aw, err := loadRestoredAssistant(t, &lifecycleQuerier{}, database.AssistantStatusWaiting, log,
				&lifecycleChain{delay: tc.delay, err: overloaded})
			require.NoError(t, err)

			err = aw.PutInput(context.Background(), "scan the host", false, nil)

			if tc.toCaller {
				require.ErrorIs(t, err, overloaded)
				assert.Never(t, func() bool { return len(reportNotes(log)) > 0 }, 200*time.Millisecond, 10*time.Millisecond)

				return
			}
			require.NoError(t, err)
			require.Eventually(t, func() bool { return len(reportNotes(log)) == 1 }, 5*time.Second, 10*time.Millisecond)
			assert.Contains(t, reportNotes(log)[0].msg, "529: overloaded")
		})
	}
}

func TestAssistant_PutInput_RefusesAnInputWhoseAssistantIsFinishedMeanwhile(t *testing.T) {
	log := &lifecycleAssistantLog{}
	aw, err := loadRestoredAssistant(t, &lifecycleQuerier{}, database.AssistantStatusWaiting, log,
		&lifecycleChain{delay: time.Second})
	require.NoError(t, err)

	answered := make(chan error, 1)
	go func() { answered <- aw.PutInput(context.Background(), "scan the host", false, nil) }()
	require.Eventually(t, func() bool { return len(log.recorded()) > 0 }, 5*time.Second, time.Millisecond,
		"the input never reached the assistant")

	finished := make(chan error, 1)
	go func() { finished <- aw.Finish(context.Background()) }()

	assert.ErrorIs(t, controllerReceive(t, answered, "the input to be answered"), context.Canceled,
		"a caller whose assistant was finished is told it stopped instead of waiting for the chain")
	require.NoError(t, controllerReceive(t, finished, "the assistant to finish"))
}
