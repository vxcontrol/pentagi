package controller

import (
	"context"
	"errors"
	"fmt"
	"sync"

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
		prvtype provider.ProviderType,
		functions *tools.Functions,
	) (FlowWorker, error)
	LoadFlows(ctx context.Context) error
	ListFlows(ctx context.Context) []FlowWorker
	GetFlow(ctx context.Context, flowID int64) (FlowWorker, error)
	FinishFlow(ctx context.Context, flowID int64) error
	StopFlow(ctx context.Context, flowID int64) error
}

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
	slc    SearchLogController
	tlc    TermLogController
	vslc   VectorStoreLogController
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
		slc:    NewSearchLogController(db),
		tlc:    NewTermLogController(db),
		vslc:   NewVectorStoreLogController(db),
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
				alc:  fc.alc,
				slc:  fc.slc,
				tlc:  fc.tlc,
				vslc: fc.vslc,
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
	prvtype provider.ProviderType,
	functions *tools.Functions,
) (FlowWorker, error) {
	fc.mx.Lock()
	defer fc.mx.Unlock()

	fw, err := NewFlowWorker(ctx, newFlowWorkerCtx{
		userID:    userID,
		input:     input,
		prvtype:   prvtype,
		functions: functions,
		flowWorkerCtx: flowWorkerCtx{
			db:     fc.db,
			cfg:    fc.cfg,
			docker: fc.docker,
			provs:  fc.provs,
			subs:   fc.subs,
			flowProviderControllers: flowProviderControllers{
				mlc:  fc.mlc,
				alc:  fc.alc,
				slc:  fc.slc,
				tlc:  fc.tlc,
				vslc: fc.vslc,
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

func (fc *flowController) ListFlows(ctx context.Context) []FlowWorker {
	fc.mx.Lock()
	defer fc.mx.Unlock()

	flows := make([]FlowWorker, 0)
	for _, flow := range fc.flows {
		flows = append(flows, flow)
	}

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

	delete(fc.flows, flowID)

	return nil
}
