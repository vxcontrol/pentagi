package queue

import (
	"context"
	"errors"
	"reflect"
	"sync"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

const defaultWorkersAmount = 10

var (
	ErrAlreadyRunning = errors.New("already running")
	ErrAlreadyStopped = errors.New("already stopped")
)

type Queue interface {
	Instance() uuid.UUID
	Running() bool
	Start() error
	Stop() error
}

type message[I any] struct {
	value   I
	doneCtx context.Context
	cancel  context.CancelFunc
}

type queue[I any, O any] struct {
	mx *sync.Mutex
	wg *sync.WaitGroup

	ctx    context.Context
	cancel context.CancelFunc

	instance uuid.UUID
	workers  int
	queue    chan *message[I]
	input    <-chan I
	output   chan O
	process  func(I) (O, error)
}

func NewQueue[I, O any](input <-chan I, output chan O, workers int, process func(I) (O, error)) Queue {
	mx, wg := &sync.Mutex{}, &sync.WaitGroup{}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	if workers <= 0 {
		workers = defaultWorkersAmount
	}

	return &queue[I, O]{
		mx: mx,
		wg: wg,

		ctx:    ctx,
		cancel: cancel,

		instance: uuid.New(),
		workers:  workers,
		input:    input,
		output:   output,
		process:  process,
	}
}

func (q *queue[I, O]) Instance() uuid.UUID {
	return q.instance
}

func (q *queue[I, O]) Running() bool {
	q.mx.Lock()
	defer q.mx.Unlock()

	return q.ctx.Err() == nil
}

func (q *queue[I, O]) Start() error {
	q.mx.Lock()
	defer q.mx.Unlock()

	if q.ctx.Err() == nil {
		return ErrAlreadyRunning
	}

	q.ctx, q.cancel = context.WithCancel(context.Background())
	q.queue = make(chan *message[I], q.workers*2)

	q.wg.Add(q.workers)
	for idx := 0; idx < q.workers; idx++ {
		go q.worker(idx)
	}

	// We make a buffered signal-only channel to raise possibility
	// q.reader() is running before Start() returns.
	ch := make(chan struct{}, 1)
	q.wg.Add(1)
	go func() {
		ch <- struct{}{}
		q.reader()
	}()
	<-ch

	return nil
}

func (q *queue[I, O]) Stop() error {
	q.mx.Lock()

	if q.ctx.Err() != nil {
		q.mx.Unlock()
		return ErrAlreadyStopped
	}

	q.cancel()
	q.mx.Unlock()
	q.wg.Wait()

	return nil
}

func (q *queue[I, O]) inputType() string {
	return reflect.Zero(reflect.TypeOf(new(I)).Elem()).Type().String()
}

func (q *queue[I, O]) outputType() string {
	return reflect.Zero(reflect.TypeOf(new(O)).Elem()).Type().String()
}

func (q *queue[I, O]) worker(wid int) {
	defer q.wg.Done()
	logger := logrus.WithFields(logrus.Fields{
		"component":   "queue_processor",
		"input_type":  q.inputType(),
		"output_type": q.outputType(),
		"instance":    q.instance,
		"worker":      wid,
	})
	logger.Debug("worker started")
	defer logger.Debug("worker exited")

	for msg := range q.queue {
		if q.process == nil {
			logger.Error("no processing function provided")
		} else if result, err := q.process(msg.value); err != nil {
			logger.WithError(err).Error("failed to process message")
		} else {
			// wait until the previous message is sent to the output channel
			<-msg.doneCtx.Done()

			// send the converted events to the output channel
			q.output <- result
		}

		// close the context to mark this operation as complete
		msg.cancel()
	}
}

func (q *queue[I, O]) reader() {
	defer q.wg.Done()
	defer close(q.queue)

	logger := logrus.WithFields(logrus.Fields{
		"component":   "queue_reader",
		"input_type":  q.inputType(),
		"output_type": q.outputType(),
		"instance":    q.instance,
	})
	logger.Debug("worker started")
	defer logger.Debug("worker exited")

	// create a root context as the initial doneCtx
	lastDoneCtx, cancel := context.WithCancel(context.Background())
	// cancel a root context because "previous" message was processed
	cancel()

	for {
		select {
		case <-q.ctx.Done():
			return
		case value, ok := <-q.input:
			// check if the input channel is closed and exit if so
			if !ok {
				q.mx.Lock()
				q.cancel()
				q.mx.Unlock()
				return
			}

			// create a new context for each message
			newCtx, cancel := context.WithCancel(context.Background())

			q.queue <- &message[I]{
				value:   value,
				doneCtx: lastDoneCtx,
				cancel:  cancel,
			}

			// update lastDoneCtx for next message
			lastDoneCtx = newCtx
		}
	}
}
