package controller

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func newFlowWorkerForInputTest(buffer int) (*flowWorker, context.CancelFunc) {
	ctx, cancel := context.WithCancel(context.Background())

	return &flowWorker{
		ctx:    ctx,
		cancel: cancel,
		input:  make(chan flowInput, buffer),
		taskMX: &sync.Mutex{},
		taskST: func() {},
		taskWG: &sync.WaitGroup{},
		flowCtx: &FlowContext{
			FlowID: 42,
		},
	}, cancel
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
	case <-time.After(flowInputTimeout + 500*time.Millisecond):
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
