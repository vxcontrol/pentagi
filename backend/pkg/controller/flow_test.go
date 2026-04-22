package controller

import (
	"context"
	"errors"
	"io"
	"sync"
	"testing"
	"time"

	"pentagi/pkg/database"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
)

const testFlowInputTimeout = 20 * time.Millisecond

type flowWorkerTestTaskController struct {
	tasks []TaskWorker
}

func (tc *flowWorkerTestTaskController) CreateTask(ctx context.Context, input string, updater FlowUpdater) (TaskWorker, error) {
	return nil, errors.New("unexpected CreateTask call")
}

func (tc *flowWorkerTestTaskController) LoadTasks(ctx context.Context, flowID int64, updater FlowUpdater) error {
	return nil
}

func (tc *flowWorkerTestTaskController) ListTasks(ctx context.Context) []TaskWorker {
	return append([]TaskWorker(nil), tc.tasks...)
}

func (tc *flowWorkerTestTaskController) GetTask(ctx context.Context, taskID int64) (TaskWorker, error) {
	for _, task := range tc.tasks {
		if task.GetTaskID() == taskID {
			return task, nil
		}
	}

	return nil, errors.New("task not found")
}

type flowWorkerTestTask struct {
	id              int64
	flowID          int64
	userID          int64
	title           string
	mu              sync.Mutex
	waiting         bool
	completed       bool
	inputs          []string
	runStarted      chan struct{}
	releaseRun      chan struct{}
	putInputStarted chan struct{}
	releasePutInput chan struct{}
	ignoreStop      bool
	runOnce         sync.Once
	putInputOnce    sync.Once
}

func (task *flowWorkerTestTask) GetTaskID() int64 {
	return task.id
}

func (task *flowWorkerTestTask) GetFlowID() int64 {
	return task.flowID
}

func (task *flowWorkerTestTask) GetUserID() int64 {
	return task.userID
}

func (task *flowWorkerTestTask) GetTitle() string {
	return task.title
}

func (task *flowWorkerTestTask) IsCompleted() bool {
	task.mu.Lock()
	defer task.mu.Unlock()

	return task.completed
}

func (task *flowWorkerTestTask) IsWaiting() bool {
	task.mu.Lock()
	defer task.mu.Unlock()

	return task.waiting
}

func (task *flowWorkerTestTask) GetStatus(ctx context.Context) (database.TaskStatus, error) {
	task.mu.Lock()
	defer task.mu.Unlock()

	switch {
	case task.completed:
		return database.TaskStatusFinished, nil
	case task.waiting:
		return database.TaskStatusWaiting, nil
	default:
		return database.TaskStatusRunning, nil
	}
}

func (task *flowWorkerTestTask) SetStatus(ctx context.Context, status database.TaskStatus) error {
	task.mu.Lock()
	defer task.mu.Unlock()

	switch status {
	case database.TaskStatusFinished, database.TaskStatusFailed:
		task.completed = true
		task.waiting = false
	case database.TaskStatusWaiting:
		task.completed = false
		task.waiting = true
	case database.TaskStatusRunning:
		task.completed = false
		task.waiting = false
	}

	return nil
}

func (task *flowWorkerTestTask) GetResult(ctx context.Context) (string, error) {
	return "", nil
}

func (task *flowWorkerTestTask) SetResult(ctx context.Context, result string) error {
	return nil
}

func (task *flowWorkerTestTask) PutInput(ctx context.Context, input string) error {
	task.putInputOnce.Do(func() {
		if task.putInputStarted != nil {
			close(task.putInputStarted)
		}
	})

	if task.releasePutInput != nil {
		<-task.releasePutInput
	}

	task.mu.Lock()
	defer task.mu.Unlock()

	task.inputs = append(task.inputs, input)
	task.waiting = false
	return nil
}

func (task *flowWorkerTestTask) Run(ctx context.Context) error {
	task.runOnce.Do(func() {
		if task.runStarted != nil {
			close(task.runStarted)
		}
	})

	if task.ignoreStop {
		if task.releaseRun != nil {
			<-task.releaseRun
		}
		task.mu.Lock()
		task.completed = true
		task.waiting = false
		task.mu.Unlock()
		return nil
	}

	<-ctx.Done()
	return ctx.Err()
}

func (task *flowWorkerTestTask) Finish(ctx context.Context) error {
	return nil
}

func (task *flowWorkerTestTask) Inputs() []string {
	task.mu.Lock()
	defer task.mu.Unlock()

	return append([]string(nil), task.inputs...)
}

func newFlowWorkerForInputTest(buffer int) (*flowWorker, context.CancelFunc) {
	ctx, cancel := context.WithCancel(context.Background())

	return &flowWorker{
		ctx:     ctx,
		cancel:  cancel,
		input:   make(chan flowInput, buffer),
		taskMX:  &sync.Mutex{},
		taskST:  func() {},
		taskWG:  &sync.WaitGroup{},
		inputWG: &sync.WaitGroup{},
		flowCtx: &FlowContext{
			FlowID: 42,
		},
		inputTO: testFlowInputTimeout,
	}, cancel
}

func newRunningFlowWorkerForInputTest(buffer int, tc TaskController) (*flowWorker, context.CancelFunc) {
	ctx, cancel := context.WithCancel(context.Background())
	logger := logrus.New()
	logger.SetOutput(io.Discard)

	fw := &flowWorker{
		tc:      tc,
		wg:      &sync.WaitGroup{},
		aws:     map[int64]AssistantWorker{},
		awsMX:   &sync.Mutex{},
		ctx:     ctx,
		cancel:  cancel,
		taskMX:  &sync.Mutex{},
		taskST:  func() {},
		taskWG:  &sync.WaitGroup{},
		inputWG: &sync.WaitGroup{},
		input:   make(chan flowInput, buffer),
		inputTO: testFlowInputTimeout,
		flowCtx: &FlowContext{
			FlowID: 42,
		},
		logger: logrus.NewEntry(logger),
	}

	fw.wg.Add(1)
	go fw.worker()

	return fw, cancel
}

func TestFlowWorkerPutInputEnqueuesWithoutImmediateConsumer(t *testing.T) {
	t.Parallel()

	fw, cancel := newFlowWorkerForInputTest(1)
	defer cancel()

	resultCh := make(chan error, 1)

	go func() {
		resultCh <- fw.PutInput(context.Background(), "queued input", nil)
	}()

	select {
	case err := <-resultCh:
		require.NoError(t, err)
	case <-time.After(fw.flowInputTimeout() + 250*time.Millisecond):
		t.Fatal("PutInput did not return after the queue handshake timeout")
	}

	require.Len(t, fw.input, 1)

	flin := <-fw.input
	require.Equal(t, "queued input", flin.input)
	require.Nil(t, flin.prv)
}

func TestFlowWorkerPutInputReturnsErrorWhenQueueIsFull(t *testing.T) {
	t.Parallel()

	fw, cancel := newFlowWorkerForInputTest(1)
	defer cancel()

	fw.input <- flowInput{input: "existing", done: make(chan error, 1)}

	err := fw.PutInput(context.Background(), "queued input", nil)
	require.ErrorContains(t, err, "input queue is full")
}

func TestFlowWorkerPutInputPropagatesEarlyError(t *testing.T) {
	t.Parallel()

	fw, cancel := newFlowWorkerForInputTest(1)
	defer cancel()

	go func() {
		flin := <-fw.input
		flin.done <- errors.New("queue rejected")
	}()

	err := fw.PutInput(context.Background(), "queued input", nil)
	require.ErrorContains(t, err, "queue rejected")
}

func TestFlowWorkerPutInputReturnsStoppedErrorWhenFlowCanceled(t *testing.T) {
	t.Parallel()

	fw, cancel := newFlowWorkerForInputTest(1)
	cancel()

	err := fw.PutInput(context.Background(), "queued input", nil)
	require.ErrorContains(t, err, "stopped")
	require.Len(t, fw.input, 0)
}

func TestFlowWorkerPutInputReturnsCanceledErrorWhenContextCanceled(t *testing.T) {
	t.Parallel()

	fw, cancelFlow := newFlowWorkerForInputTest(1)
	defer cancelFlow()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := fw.PutInput(ctx, "queued input", nil)
	require.ErrorContains(t, err, "input processing canceled")
	require.Len(t, fw.input, 0)
}

func TestFlowWorkerStopDropsQueuedInput(t *testing.T) {
	t.Parallel()

	fw, cancel := newFlowWorkerForInputTest(2)
	defer cancel()

	flin := flowInput{input: "queued input", done: make(chan error, 1)}
	fw.input <- flin

	err := fw.Stop(context.Background())
	require.NoError(t, err)
	require.Len(t, fw.input, 0)

	select {
	case stopErr := <-flin.done:
		require.ErrorContains(t, stopErr, "stopped before queued input was processed")
	case <-time.After(100 * time.Millisecond):
		t.Fatal("queued input was not drained on stop")
	}
}

func TestFlowWorkerStopTimeoutDropsQueuedInput(t *testing.T) {
	t.Parallel()

	fw, cancel := newFlowWorkerForInputTest(2)
	defer cancel()

	fw.taskWG.Add(1)
	defer fw.taskWG.Done()

	flin := flowInput{input: "queued input", done: make(chan error, 1)}
	fw.input <- flin

	err := fw.Stop(context.Background())
	require.ErrorContains(t, err, "task stop timeout")
	require.Len(t, fw.input, 0)

	select {
	case stopErr := <-flin.done:
		require.ErrorContains(t, stopErr, "stopped before queued input was processed")
	case <-time.After(100 * time.Millisecond):
		t.Fatal("queued input was not drained on stop timeout")
	}
}

func TestFlowWorkerStopDrainsQueuedInputWhileWorkerIsRunning(t *testing.T) {
	t.Parallel()

	task := &flowWorkerTestTask{
		id:         1,
		flowID:     42,
		userID:     7,
		title:      "waiting task",
		waiting:    true,
		runStarted: make(chan struct{}),
	}
	tc := &flowWorkerTestTaskController{
		tasks: []TaskWorker{task},
	}

	fw, cancel := newRunningFlowWorkerForInputTest(2, tc)
	defer func() {
		cancel()
		fw.wg.Wait()
	}()

	err := fw.PutInput(context.Background(), "first queued instruction", nil)
	require.NoError(t, err)

	select {
	case <-task.runStarted:
	case <-time.After(time.Second):
		t.Fatal("worker did not start running the active task")
	}

	drainedInput := flowInput{input: "second queued instruction", done: make(chan error, 1)}
	fw.input <- drainedInput

	err = fw.Stop(context.Background())
	require.NoError(t, err)
	require.Len(t, fw.input, 0)
	require.Equal(t, []string{"first queued instruction"}, task.Inputs())

	select {
	case stopErr := <-drainedInput.done:
		require.ErrorContains(t, stopErr, "stopped before queued input was processed")
	case <-time.After(100 * time.Millisecond):
		t.Fatal("queued input was not drained while the worker was stopping")
	}
}

func TestFlowWorkerStopWaitsForDequeuedInputBeforeTaskRegistration(t *testing.T) {
	t.Parallel()

	task := &flowWorkerTestTask{
		id:              1,
		flowID:          42,
		userID:          7,
		title:           "waiting task",
		waiting:         true,
		putInputStarted: make(chan struct{}),
		releasePutInput: make(chan struct{}),
		runStarted:      make(chan struct{}),
	}
	tc := &flowWorkerTestTaskController{
		tasks: []TaskWorker{task},
	}

	fw, cancel := newRunningFlowWorkerForInputTest(2, tc)
	defer func() {
		cancel()
		fw.wg.Wait()
	}()

	resultCh := make(chan error, 1)
	go func() {
		resultCh <- fw.PutInput(context.Background(), "queued instruction", nil)
	}()

	select {
	case <-task.putInputStarted:
	case <-time.After(time.Second):
		t.Fatal("worker did not start processing the dequeued input")
	}

	stopResultCh := make(chan error, 1)
	go func() {
		stopResultCh <- fw.Stop(context.Background())
	}()

	select {
	case err := <-stopResultCh:
		t.Fatalf("stop returned too early: %v", err)
	case <-time.After(100 * time.Millisecond):
	}

	close(task.releasePutInput)

	select {
	case err := <-stopResultCh:
		require.NoError(t, err)
	case <-time.After(time.Second):
		t.Fatal("stop did not wait for the dequeued input to finish unwinding")
	}

	select {
	case err := <-resultCh:
		require.NoError(t, err)
	case <-time.After(fw.flowInputTimeout() + time.Second):
		t.Fatal("PutInput did not finish after the stop sequence completed")
	}

	select {
	case <-task.runStarted:
		t.Fatal("task started running after stop completed")
	case <-time.After(100 * time.Millisecond):
	}

	require.Equal(t, []string{"queued instruction"}, task.Inputs())
}
