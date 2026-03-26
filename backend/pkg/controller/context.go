package controller

import (
	"context"
	"errors"
	"fmt"

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

	UserID int64
	FlowID int64

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

func wrapErrorEndSpan(ctx context.Context, span langfuse.Span, msg string, err error) error {
	logrus.WithContext(ctx).WithError(err).Error(msg)
	err = fmt.Errorf("%s: %w", msg, err)
	span.End(
		langfuse.WithSpanStatus(err.Error()),
		langfuse.WithSpanLevel(langfuse.ObservationLevelError),
	)
	return err
}
