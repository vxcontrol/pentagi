package controller

import (
	"context"
	"errors"
	"fmt"
	"time"

	"pentagi/pkg/database"
	"pentagi/pkg/graph/subscriptions"
	"pentagi/pkg/observability/langfuse"
	"pentagi/pkg/providers"
	"pentagi/pkg/tools"

	"github.com/sirupsen/logrus"
)

var ErrNothingToLoad = errors.New("nothing to load")

type FlowContext struct {
	DB database.Querier

	UserID  int64
	FlowID  int64
	TraceID string

	Executor  tools.FlowToolsExecutor
	Provider  providers.FlowProvider
	Publisher subscriptions.FlowPublisher

	TermLog    FlowTermLogWorker
	MsgLog     FlowMsgLogWorker
	Screenshot FlowScreenshotWorker
}

type TaskContext struct {
	TaskID    int64
	TaskTitle string
	TaskInput string

	FlowContext
}

type SubtaskContext struct {
	MsgChainID         int64
	SubtaskID          int64
	SubtaskTitle       string
	SubtaskDescription string

	TaskContext
}

// markFlowFailed records a flow as failed after its resources couldn't be
// prepared. It uses a detached context so the write still lands when the
// caller's context was the thing that got canceled.
func markFlowFailed(db database.Querier, flowID int64) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if _, err := db.UpdateFlowStatus(ctx, database.UpdateFlowStatusParams{
		Status: database.FlowStatusFailed,
		ID:     flowID,
	}); err != nil {
		logrus.WithError(err).WithField("flow_id", flowID).
			Error("failed to mark flow as failed after resource preparation error")
	}
}

func wrapErrorEndSpan(ctx context.Context, span langfuse.Span, msg string, err error) error {
	logrus.WithContext(ctx).WithError(err).Error(msg)
	err = fmt.Errorf("%s: %w", msg, err)
	span.End(
		langfuse.WithSpanStatus(err.Error()),
		langfuse.WithSpanLevel(langfuse.ObservationLevelError),
	)
	return err
}
