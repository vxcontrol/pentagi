package langfuse

import (
	"context"
	"fmt"
	"sync"
	"time"

	"pentagi/pkg/observability/langfuse/api"

	"github.com/sirupsen/logrus"
)

const (
	defaultQueueSize    = 100
	defaultSendInterval = 10 * time.Second
	defaultTimeout      = 20 * time.Second
)

type Observer interface {
	NewObservation(
		ctx context.Context,
		opts ...ObservationContextOption,
	) (context.Context, Observation)
	Shutdown(ctx context.Context) error
	ForceFlush(ctx context.Context) error
}

type enqueue interface {
	enqueue(event *api.IngestionEvent)
}

type observer struct {
	mx        *sync.Mutex
	wg        *sync.WaitGroup
	ctx       context.Context
	cancel    context.CancelFunc
	client    *Client
	project   string
	release   string
	interval  time.Duration
	timeout   time.Duration
	queueSize int
	queue     chan *api.IngestionEvent
	flusher   chan chan error
}

func NewObserver(client *Client, opts ...ObserverOption) Observer {
	ctx, cancel := context.WithCancel(context.Background())
	o := &observer{
		mx:      &sync.Mutex{},
		wg:      &sync.WaitGroup{},
		ctx:     ctx,
		cancel:  cancel,
		client:  client,
		flusher: make(chan chan error),
	}
	for _, opt := range opts {
		opt(o)
	}
	if o.interval <= 0 || o.interval > 10*time.Minute {
		o.interval = defaultSendInterval
	}
	if o.timeout <= 0 || o.timeout > 2*time.Minute {
		o.timeout = defaultTimeout
	}
	if o.queueSize <= 0 || o.queueSize > 10000 {
		o.queueSize = defaultQueueSize
	}
	o.queue = make(chan *api.IngestionEvent, o.queueSize)

	o.wg.Add(1)
	go func() {
		defer o.wg.Done()
		o.sender()
	}()

	return o
}

func (o *observer) NewObservation(
	ctx context.Context,
	opts ...ObservationContextOption,
) (context.Context, Observation) {
	var obsCtx ObservationContext
	for _, opt := range opts {
		opt(&obsCtx)
	}

	parentObsCtx, parentObsCtxFound := getObservationContext(ctx)

	if obsCtx.TraceID == "" { // wants to use parent trace id in general or create new one
		if parentObsCtxFound && parentObsCtx.TraceID != "" {
			obsCtx.TraceID = parentObsCtx.TraceID
			if obsCtx.ObservationID == "" { // wants to use parent observation id in general
				obsCtx.ObservationID = parentObsCtx.ObservationID
			}
		} else {
			obsCtx.TraceID = newID()
		}
	}

	if obsCtx.TraceCtx != nil {
		o.putTraceInfo(obsCtx)
	}

	obs := &observation{
		obsCtx: observationContext{
			TraceID:       obsCtx.TraceID,
			ObservationID: obsCtx.ObservationID,
		},
		observer: o,
	}

	return putObservationContext(ctx, obs.obsCtx), obs
}

func (o *observer) Shutdown(ctx context.Context) error {
	o.mx.Lock()
	defer o.mx.Unlock()

	if err := o.ctx.Err(); err != nil {
		return err
	}

	o.cancel()
	o.wg.Wait()
	close(o.flusher)

	return nil
}

func (o *observer) ForceFlush(ctx context.Context) error {
	o.mx.Lock()
	defer o.mx.Unlock()

	if err := o.ctx.Err(); err != nil {
		return err
	}

	ch := make(chan error)
	o.flusher <- ch

	return <-ch
}

func (o *observer) enqueue(event *api.IngestionEvent) {
	o.mx.Lock()
	defer o.mx.Unlock()

	if err := o.ctx.Err(); err != nil {
		return
	}

	o.queue <- event
}

func (o *observer) flush(ctx context.Context, batch []*api.IngestionEvent) error {
	ctx, cancel := context.WithTimeout(ctx, o.timeout)
	defer cancel()

	if len(batch) == 0 {
		return nil
	}

	metadata := map[string]any{
		"batch_size":      len(batch),
		"app_release":     o.release,
		"sdk_integration": "langchain",
		"sdk_name":        "golang",
		"sdk_version":     InstrumentationVersion,
		"public_key":      o.client.PublicKey(),
		"project_id":      o.client.ProjectID(),
	}

	resp, err := o.client.Ingestion.Batch(ctx, &api.IngestionBatchRequest{
		Batch:    batch,
		Metadata: metadata,
	})
	if err != nil {
		return err
	}

	if len(resp.Errors) > 0 {
		for _, event := range resp.Errors {
			logrus.WithContext(ctx).WithFields(logrus.Fields{
				"event_id": event.Id,
				"status":   event.Status,
				"message":  event.Message,
				"error":    event.Error,
			}).Errorf("failed to send event to Langfuse")
		}

		return fmt.Errorf("failed to send %d events", len(resp.Errors))
	}

	return nil
}

func (o *observer) sender() {
	batch := make([]*api.IngestionEvent, 0, o.queueSize)
	ticker := time.NewTicker(o.interval)
	defer ticker.Stop()

	for {
		select {
		case <-o.ctx.Done():
			return
		case ch := <-o.flusher:
			ch <- o.flush(o.ctx, batch)
			batch = batch[:0]
			ticker.Reset(o.interval)
		case <-ticker.C:
			if err := o.flush(o.ctx, batch); err != nil {
				logrus.WithContext(o.ctx).WithError(err).Error("failed to flush events by interval")
			}
			batch = batch[:0]
			ticker.Reset(o.interval)
		case event := <-o.queue:
			batch = append(batch, event)
			if len(batch) >= o.queueSize {
				if err := o.flush(o.ctx, batch); err != nil {
					logrus.WithContext(o.ctx).WithError(err).Error("failed to flush events by queue size")
				}
				batch = batch[:0]
				ticker.Reset(o.interval)
			}
		}
	}
}

func (o *observer) putTraceInfo(obsCtx ObservationContext) {
	traceCreate := api.NewIngestionEventFromIngestionEventZero(&api.IngestionEventZero{
		Id:        newID(),
		Timestamp: getCurrentTimeString(),
		Type:      ingestionCreateTrace,
		Body: &api.TraceBody{
			Id:        getStringRef(obsCtx.TraceID),
			Timestamp: obsCtx.TraceCtx.Timestamp,
			Name:      obsCtx.TraceCtx.Name,
			UserId:    obsCtx.TraceCtx.UserId,
			Input:     obsCtx.TraceCtx.Input,
			Output:    obsCtx.TraceCtx.Output,
			SessionId: obsCtx.TraceCtx.SessionId,
			Release:   getStringRef(o.release),
			Version:   obsCtx.TraceCtx.Version,
			Metadata:  obsCtx.TraceCtx.Metadata,
			Tags:      obsCtx.TraceCtx.Tags,
			Public:    obsCtx.TraceCtx.Public,
		},
	})

	o.enqueue(traceCreate)
}
