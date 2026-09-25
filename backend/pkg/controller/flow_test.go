package controller

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"pentagi/pkg/config"
	"pentagi/pkg/database"
	"pentagi/pkg/docker"
	obs "pentagi/pkg/observability"
	"pentagi/pkg/providers"
	"pentagi/pkg/providers/provider"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type scriptedTasks struct {
	TaskController

	tasks       []TaskWorker
	createErr   error
	createDelay time.Duration
	createGate  chan struct{}
}

func (c *scriptedTasks) ListTasks(context.Context) []TaskWorker { return c.tasks }

func (c *scriptedTasks) CreateTask(context.Context, string, FlowUpdater) (TaskWorker, error) {
	if c.createGate != nil {
		<-c.createGate
	}
	time.Sleep(c.createDelay)

	return nil, c.createErr
}

// scriptedTask is task 11: its write-up blocks until cancelled unless answer is set, and Finish reports the flow waiting.
type scriptedTask struct {
	TaskWorker

	isWaiting bool
	inputErr  error
	runErr    error
	answer    error
	updater   FlowUpdater
	running   chan struct{}
	started   chan struct{}

	runOnce    sync.Once
	reportOnce sync.Once
	mx         sync.Mutex
	err        error
}

func (t *scriptedTask) GetTaskID() int64                          { return 11 }
func (t *scriptedTask) GetTitle() string                          { return "scan the host" }
func (t *scriptedTask) IsCompleted() bool                         { return false }
func (t *scriptedTask) IsWaiting() bool                           { return t.isWaiting }
func (t *scriptedTask) PutInput(context.Context, string) error    { return t.inputErr }
func (t *scriptedTask) GetResult(context.Context) (string, error) { return "", nil }
func (t *scriptedTask) GetStatus(context.Context) (database.TaskStatus, error) {
	return database.TaskStatusWaiting, nil
}

func (t *scriptedTask) Finish(ctx context.Context) error {
	return t.updater.SetStatus(ctx, database.FlowStatusWaiting)
}

func (t *scriptedTask) Run(context.Context) error {
	if t.running != nil {
		t.runOnce.Do(func() { close(t.running) })
	}

	return t.runErr
}

func (t *scriptedTask) Report(ctx context.Context) error {
	t.reportOnce.Do(func() { close(t.started) })
	if t.answer != nil {
		return t.answer
	}

	<-ctx.Done()

	t.mx.Lock()
	defer t.mx.Unlock()
	t.err = ctx.Err()

	return t.err
}

func (t *scriptedTask) reportErr() error {
	t.mx.Lock()
	defer t.mx.Unlock()

	return t.err
}

func startScriptedFlowWorker(t *testing.T, tc TaskController) (*flowWorker, *createFakeQuerier) {
	t.Helper()

	q := &createFakeQuerier{flow: database.Flow{ID: reservedFlowID, Status: database.FlowStatusWaiting}}
	pub := &cascadeFakePublisher{}
	ctx, cancel := context.WithCancel(context.Background())
	flowCtx := &FlowContext{
		DB:        q,
		FlowID:    reservedFlowID,
		Publisher: pub,
		MsgLog:    NewFlowMsgLogWorker(q, reservedFlowID, pub),
	}
	fw := &flowWorker{
		tc:      tc,
		wg:      &sync.WaitGroup{},
		ctx:     ctx,
		cancel:  cancel,
		taskMX:  &sync.Mutex{},
		taskST:  func() {},
		taskWG:  &sync.WaitGroup{},
		taskCCH: make(chan struct{}),
		input:   make(chan flowInput),
		flowCtx: flowCtx,
		logger:  logrus.WithField("flow_id", reservedFlowID),
	}
	fw.wg.Add(1)
	go fw.worker()
	t.Cleanup(func() {
		cancel()
		stopped := make(chan struct{})
		go func() {
			fw.wg.Wait()
			close(stopped)
		}()
		controllerReceive(t, stopped, "the flow worker to stop")
	})

	return fw, q
}

type report struct {
	message string
	result  string
	taskID  sql.NullInt64
}

func reports(q *createFakeQuerier) []report {
	var found []report
	for _, msg := range q.recordedMessages() {
		if msg.Type == database.MsglogTypeReport {
			found = append(found, report{message: msg.Message, taskID: msg.TaskID})
		}
	}
	for _, msg := range q.recordedResults() {
		if msg.Type == database.MsglogTypeReport {
			found = append(found, report{message: msg.Message, result: msg.Result, taskID: msg.TaskID})
		}
	}

	return found
}

type lifecycleFakeAssistant struct {
	AssistantWorker

	id int64
}

func (a *lifecycleFakeAssistant) GetAssistantID() int64 { return a.id }

func TestFlow_AddAssistant_RefusesAFinishedFlow(t *testing.T) {
	t.Parallel()

	fw := newFinishableWorker(
		&finishFakeQuerier{flow: database.Flow{ID: 42, Status: database.FlowStatusWaiting}},
		&callerFakeExecutor{},
	)
	require.NoError(t, fw.Finish(context.Background()))

	assert.ErrorIs(t, fw.AddAssistant(context.Background(), &lifecycleFakeAssistant{id: 5}), ErrFlowAlreadyStopped)
	assert.Empty(t, fw.ListAssistants(context.Background()),
		"nothing would ever finish an assistant added to a finished flow")
}

func TestFlow_Finish_LeavesTheFlowRecordedFinished(t *testing.T) {
	t.Parallel()

	sandboxErr := errors.New("docker is down")

	for _, tc := range []struct {
		name         string
		releaseErr   error
		openTask     bool
		wantStatuses []database.FlowStatus
	}{
		{
			name:         "a sandbox that cannot be released",
			releaseErr:   sandboxErr,
			wantStatuses: []database.FlowStatus{database.FlowStatusFinished},
		},
		{
			name:         "a last task that reports the flow waiting",
			openTask:     true,
			wantStatuses: []database.FlowStatus{database.FlowStatusWaiting, database.FlowStatusFinished},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			q := &finishFakeQuerier{flow: database.Flow{ID: 42, Status: database.FlowStatusRunning}}
			ex := &callerFakeExecutor{err: tc.releaseErr}
			fw := newFinishableWorker(q, ex)
			if tc.openTask {
				fw.tc = &scriptedTasks{tasks: []TaskWorker{&scriptedTask{updater: fw}}}
			}

			err := fw.Finish(context.Background())

			if tc.releaseErr != nil {
				assert.ErrorIs(t, err, tc.releaseErr, "the caller still learns the sandbox was left behind")
			} else {
				assert.NoError(t, err)
			}
			assert.True(t, ex.released)
			statuses := make([]database.FlowStatus, 0, len(q.flowStatus))
			for _, update := range q.flowStatus {
				statuses = append(statuses, update.Status)
			}
			assert.Equal(t, tc.wantStatuses, statuses,
				"a flow closed before its tasks would be reopened by their back propagation")
		})
	}
}

type lateInputFakeProvider struct {
	provider.Provider
}

func (p *lateInputFakeProvider) Name() provider.ProviderName { return provider.ProviderName("late") }
func (p *lateInputFakeProvider) Type() provider.ProviderType { return provider.ProviderOpenAI }

type lateInputFakeFlowProvider struct {
	providers.FlowProvider

	switches atomic.Int64
}

func (p *lateInputFakeFlowProvider) SetProvider(context.Context, provider.Provider) (bool, string, error) {
	p.switches.Add(1)

	return false, "", nil
}

func TestFlow_PutInput_RefusesInputAfterTheFlowFinished(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	flowProvider := &lateInputFakeFlowProvider{}
	flowCtx := &FlowContext{FlowID: 7, Provider: flowProvider}
	fw := &flowWorker{
		wg:      &sync.WaitGroup{},
		ctx:     ctx,
		cancel:  cancel,
		input:   make(chan flowInput),
		tc:      NewTaskController(flowCtx),
		flowCtx: flowCtx,
		logger:  logrus.WithField("flow_id", flowCtx.FlowID),
	}
	fw.wg.Add(1)
	go fw.worker()

	finished := make(chan error, 1)
	go func() { finished <- fw.finish() }()
	require.NoError(t, controllerReceive(t, finished, "finish to stop the worker"))

	for range 200 {
		require.NotPanics(t, func() {
			err := fw.PutInput(context.Background(), "late input", &lateInputFakeProvider{}, nil)
			assert.ErrorIs(t, err, context.Canceled)
		}, "an assistant's submit_flow_input can still hold the worker the user just finished")
	}
	assert.Zero(t, flowProvider.switches.Load(), "a finished flow does not have its provider switched by a late input")
}

func TestFlow_ReportFailure_TellsExactlyOneOfCallerAndMessageLog(t *testing.T) {
	notFound := errors.New("API returned unexpected status code: 404: The model `grok-4.7-latest` does not exist")
	onTask := sql.NullInt64{Int64: 11, Valid: true}

	for _, tc := range []struct {
		name       string
		tasks      *scriptedTasks
		input      string
		wantErr    string
		wantReport string
		wantTaskID sql.NullInt64
		wantResult string
	}{
		{
			name:       "a task that cannot start after the input was taken",
			tasks:      &scriptedTasks{createErr: notFound, createGate: make(chan struct{})},
			input:      "scan 10.0.0.5",
			wantReport: "does not exist",
			wantResult: "scan 10.0.0.5",
		},
		{
			name:       "a task that fails while running",
			tasks:      &scriptedTasks{tasks: []TaskWorker{&scriptedTask{isWaiting: true, runErr: errors.New("agent chain: 404")}}},
			input:      "go on",
			wantReport: "404",
			wantTaskID: onTask,
		},
		{
			name:       "a task that cannot resume after loading",
			tasks:      &scriptedTasks{tasks: []TaskWorker{&scriptedTask{runErr: errors.New("provider is gone")}}},
			wantReport: "provider is gone",
			wantTaskID: onTask,
		},
		{
			name:  "a task creation the user stopped",
			tasks: &scriptedTasks{createErr: context.Canceled},
			input: "scan 10.0.0.5",
		},
		{
			name: "an input the waiting task refuses",
			tasks: &scriptedTasks{tasks: []TaskWorker{
				&scriptedTask{isWaiting: true, inputErr: errors.New("subtask input rejected")},
			}},
			input:   "go on",
			wantErr: "subtask input rejected",
		},
		{
			name:    "a task that cannot start at once",
			tasks:   &scriptedTasks{createErr: notFound, createDelay: 50 * time.Millisecond},
			input:   "scan 10.0.0.5",
			wantErr: "does not exist",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			fw, q := startScriptedFlowWorker(t, tc.tasks)

			if tc.input != "" {
				err := fw.PutInput(t.Context(), tc.input, nil, nil)
				if tc.wantErr == "" {
					require.NoError(t, err)
				} else {
					require.ErrorContains(t, err, tc.wantErr, "the caller shows this one itself")
				}
				if tc.tasks.createGate != nil {
					close(tc.tasks.createGate)
				}
			}

			require.Eventually(t, func() bool {
				statuses := q.recordedStatuses()
				return len(statuses) > 0 && statuses[len(statuses)-1] == database.FlowStatusWaiting
			}, 5*time.Second, 10*time.Millisecond, "the flow goes back to waiting for input")

			if tc.wantReport == "" {
				assert.Never(t, func() bool { return len(reports(q)) > 0 }, 200*time.Millisecond, 10*time.Millisecond)

				return
			}
			require.Eventually(t, func() bool { return len(reports(q)) == 1 }, 5*time.Second, 10*time.Millisecond)
			found := reports(q)[0]
			assert.Contains(t, found.message, tc.wantReport)
			assert.Equal(t, tc.wantTaskID, found.taskID, "the report hangs on the task row when there is one")
			assert.Equal(t, tc.wantResult, found.result, "without a task row the input is kept for the user")
		})
	}
}

// The write-up reaches the flow through the status its task writes.
func TestFlow_SetStatus_KeepsAFinishedFlowFinished(t *testing.T) {
	fw, q := startScriptedFlowWorker(t, &scriptedTasks{})

	require.NoError(t, fw.SetStatus(t.Context(), database.FlowStatusFinished))

	tw := &taskWorker{
		mx:      &sync.RWMutex{},
		updater: fw,
		taskCtx: &TaskContext{
			TaskID:      11,
			FlowContext: FlowContext{DB: q, FlowID: reservedFlowID, Publisher: &cascadeFakePublisher{}},
		},
	}
	require.NoError(t, tw.SetStatus(t.Context(), database.TaskStatusFinished))

	assert.Equal(t, []database.FlowStatus{database.FlowStatusFinished}, q.recordedStatuses(),
		"a finished flow reading waiting lists as open while every input to it is refused")
}

func TestFlow_Stop_EndsAWriteUpInFlight(t *testing.T) {
	obs.InitObserver(context.Background(), nil, nil, nil)

	task := &scriptedTask{isWaiting: true, started: make(chan struct{})}
	fw, _ := startScriptedFlowWorker(t, &scriptedTasks{tasks: []TaskWorker{task}})

	require.NoError(t, fw.Report(t.Context(), 0))
	controllerReceive(t, task.started, "the write-up to start")

	stopped := make(chan error, 1)
	go func() { stopped <- fw.Stop(context.Background()) }()

	select {
	case err := <-stopped:
		require.NoError(t, err)
	case <-time.After(15 * time.Second):
		t.Fatal("Stop waited out its task timeout on a write-up it cannot interrupt")
	}
	assert.ErrorIs(t, task.reportErr(), context.Canceled)
}

// A flow accepts input during a write-up, and that input starts a task of its own.
func TestFlow_RunReport_KeepsAWriteUpInFlightWhenInputArrives(t *testing.T) {
	obs.InitObserver(context.Background(), nil, nil, nil)

	task := &scriptedTask{isWaiting: true, started: make(chan struct{}), running: make(chan struct{})}
	fw, _ := startScriptedFlowWorker(t, &scriptedTasks{tasks: []TaskWorker{task}})

	require.NoError(t, fw.Report(t.Context(), 0))
	controllerReceive(t, task.started, "the write-up to start")

	require.NoError(t, fw.PutInput(t.Context(), "keep going", nil, nil))
	controllerReceive(t, task.running, "the input to reach the task")

	assert.Never(t, func() bool { return task.reportErr() != nil }, time.Second, 20*time.Millisecond,
		"the write-up cannot be re-derived once the flow is finished")
}

func TestFlow_RunReport_TellsTheMessageLogWhyAWriteUpFailed(t *testing.T) {
	obs.InitObserver(context.Background(), nil, nil, nil)

	for _, tc := range []struct {
		name       string
		answer     error
		budget     time.Duration
		wantReport string
	}{
		{name: "a write-up that outlives its allowance", budget: 50 * time.Millisecond, wantReport: "deadline exceeded"},
		{
			name:       "a write-up whose retry outlives its allowance",
			answer:     errors.New("chain: 529 overloaded"),
			budget:     200 * time.Millisecond,
			wantReport: "529 overloaded",
		},
		{name: "a write-up the task refuses as already done", answer: fmt.Errorf("%w: task 11", ErrTaskAlreadyCompleted)},
	} {
		t.Run(tc.name, func(t *testing.T) {
			task := &scriptedTask{isWaiting: true, started: make(chan struct{}), answer: tc.answer}
			fw, q := startScriptedFlowWorker(t, &scriptedTasks{tasks: []TaskWorker{task}})

			require.NoError(t, fw.Report(t.Context(), tc.budget))
			controllerReceive(t, task.started, "the write-up to start")

			if tc.wantReport == "" {
				assert.Never(t, func() bool { return len(reports(q)) > 0 }, time.Second, 20*time.Millisecond,
					"the door answers a refusal to the caller, so it is no fault of the flow")

				return
			}
			require.Eventually(t, func() bool { return len(reports(q)) == 1 }, 10*time.Second, 10*time.Millisecond)
			assert.Contains(t, reports(q)[0].message, tc.wantReport)
			assert.Equal(t, int64(11), reports(q)[0].taskID.Int64, "the message belongs to the task it was asked about")
		})
	}
}

type completedTask struct {
	TaskWorker

	id   int64
	done bool
}

func (t *completedTask) GetTaskID() int64  { return t.id }
func (t *completedTask) IsCompleted() bool { return t.done }

func TestFlow_ReportableTask_PicksTheLastOpenTask(t *testing.T) {
	t.Parallel()

	open := func(id int64) TaskWorker { return &completedTask{id: id} }
	done := func(id int64) TaskWorker { return &completedTask{id: id, done: true} }

	for _, tc := range []struct {
		name  string
		tasks []TaskWorker
		want  int64
	}{
		{name: "the last of the open tasks", tasks: []TaskWorker{open(1), done(2), open(3), done(4)}, want: 3},
		{name: "no tasks at all"},
		{name: "only finished tasks", tasks: []TaskWorker{done(1), done(2)}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got := reportableTask(tc.tasks)
			if tc.want == 0 {
				assert.Nil(t, got)

				return
			}
			require.NotNil(t, got)
			assert.Equal(t, tc.want, got.GetTaskID(), "the one an operator means is the one they were just watching")
		})
	}
}

func TestFlow_ReportAllowance_FallsBackToTheDefaultWithoutABudget(t *testing.T) {
	t.Parallel()

	for budget, want := range map[time.Duration]time.Duration{
		0:                30 * time.Minute,
		-1:               30 * time.Minute,
		90 * time.Second: 90 * time.Second,
	} {
		assert.Equal(t, want, reportAllowance(budget), "budget %v", budget)
	}
}

func TestFlow_Report_RefusesAFlowWithNothingOpen(t *testing.T) {
	obs.InitObserver(context.Background(), nil, nil, nil)

	fw, _ := startScriptedFlowWorker(t, &scriptedTasks{})

	assert.ErrorIs(t, fw.Report(context.Background(), 0), ErrNothingToReport)
}

// sandboxFakeDocker ignores the context because the real sweep detaches from its caller itself.
type sandboxFakeDocker struct {
	docker.DockerClient

	mx      sync.Mutex
	killErr error
	killed  []string
	onSweep func()
}

func (d *sandboxFakeDocker) KillFlowCommands(_ context.Context, containerID string) error {
	d.mx.Lock()
	d.killed = append(d.killed, containerID)
	onSweep, killErr := d.onSweep, d.killErr
	d.mx.Unlock()

	if onSweep != nil {
		onSweep()
	}

	return killErr
}

func (d *sandboxFakeDocker) sweeps() []string {
	d.mx.Lock()
	defer d.mx.Unlock()

	return append([]string(nil), d.killed...)
}

func newStoppableWorker(dkr docker.DockerClient) *flowWorker {
	return &flowWorker{
		cfg:     &config.Config{},
		docker:  dkr,
		taskMX:  &sync.Mutex{},
		taskST:  func() {},
		taskWG:  &sync.WaitGroup{},
		flowCtx: &FlowContext{FlowID: 42},
		logger:  logrus.WithField("component", "test"),
	}
}

func TestFlow_Stop_EndsTheCommandsLeftRunningInTheSandbox(t *testing.T) {
	for _, tc := range []struct {
		name       string
		callerGone bool
		killErr    error
	}{
		{name: "with the caller waiting"},
		{name: "after the caller has gone", callerGone: true},
		{name: "when the sandbox cannot be swept", killErr: errors.New("docker is down")},
	} {
		t.Run(tc.name, func(t *testing.T) {
			dkr := &sandboxFakeDocker{killErr: tc.killErr}
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			if tc.callerGone {
				cancel()
			}

			require.NoError(t, newStoppableWorker(dkr).Stop(ctx), "a sandbox that cannot be swept does not keep the flow running")
			assert.Equal(t, []string{"pentagi-terminal-42", "pentagi-terminal-42"}, dkr.sweeps(),
				"cancelling only detaches this side of a container exec: the command inside keeps running")
		})
	}
}

func TestFlow_Stop_SweepsTheSandboxAgainOnceItsTasksHaveStopped(t *testing.T) {
	dkr := &sandboxFakeDocker{}
	fw := newStoppableWorker(dkr)

	var (
		tasksStopped atomic.Bool
		afterTasks   []bool
		release      sync.Once
	)
	released := make(chan struct{})
	dkr.onSweep = func() {
		afterTasks = append(afterTasks, tasksStopped.Load())
		release.Do(func() { close(released) })
	}

	fw.taskWG.Add(1)
	go func() {
		<-released
		tasksStopped.Store(true)
		fw.taskWG.Done()
	}()

	require.NoError(t, fw.Stop(context.Background()))

	assert.Equal(t, []bool{false, true}, afterTasks,
		"a command the flow launched while it was stopping outlived the stop")
}

func TestFlow_PutInput_SaysWhetherTheInputWasAccepted(t *testing.T) {
	// A caller that cannot tell "never started" from "still running" either
	// duplicates the action by retrying or drops it by not retrying.
	newWorker := func(ctx context.Context, cancel context.CancelFunc) *flowWorker {
		flowCtx := &FlowContext{FlowID: 7, Provider: &lateInputFakeFlowProvider{}}

		return &flowWorker{
			wg:      &sync.WaitGroup{},
			ctx:     ctx,
			cancel:  cancel,
			input:   make(chan flowInput),
			tc:      NewTaskController(flowCtx),
			flowCtx: flowCtx,
			logger:  logrus.WithField("flow_id", flowCtx.FlowID),
		}
	}

	t.Run("nobody takes the input", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		fw := newWorker(ctx, cancel)

		callerCtx, callerCancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
		defer callerCancel()

		err := fw.PutInput(callerCtx, "scan the host", &lateInputFakeProvider{}, nil)
		assert.ErrorIs(t, err, ErrInputNotAccepted, "the worker never took it, so retrying is safe")
		assert.NotErrorIs(t, err, ErrInputAccepted)
	})

	t.Run("the worker takes it and the caller gives up first", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		fw := newWorker(ctx, cancel)

		taken := make(chan struct{})
		go func() {
			<-fw.input
			close(taken)
		}()

		callerCtx, callerCancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
		defer callerCancel()

		err := fw.PutInput(callerCtx, "scan the host", &lateInputFakeProvider{}, nil)
		controllerReceive(t, taken, "the worker to take the input")
		assert.ErrorIs(t, err, ErrInputAccepted, "the work is running, so retrying would duplicate it")
		assert.NotErrorIs(t, err, ErrInputNotAccepted)
	})

	t.Run("the worker takes it and is simply slower than flowInputTimeout", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		fw := newWorker(ctx, cancel)

		taken := make(chan struct{})
		go func() {
			<-fw.input
			close(taken)
		}()

		err := fw.PutInput(context.Background(), "scan the host", &lateInputFakeProvider{}, nil)
		controllerReceive(t, taken, "the worker to take the input")
		assert.NoError(t, err, "a slow worker is the ordinary case and has always been quiet")
	})
}
