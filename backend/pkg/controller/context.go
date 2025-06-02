package controller

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"unicode/utf8"

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

	UserID    int64
	FlowID    int64
	FlowTitle string

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
		langfuse.WithEndSpanStatus(err.Error()),
		langfuse.WithEndSpanLevel(langfuse.ObservationLevelError),
	)
	return err
}

func sanitizeUTF8(msg string) string {
	if msg == "" {
		return ""
	}

	var builder strings.Builder
	builder.Grow(len(msg)) // Pre-allocate for efficiency

	for i := 0; i < len(msg); {
		// Explicitly skip null bytes
		if msg[i] == '\x00' {
			i++
			continue
		}
		// Decode rune and check for errors
		r, size := utf8.DecodeRuneInString(msg[i:])
		if r == utf8.RuneError && size == 1 {
			// Invalid UTF-8 byte, replace with Unicode replacement character
			builder.WriteRune(utf8.RuneError)
			i += size
		} else {
			builder.WriteRune(r)
			i += size
		}
	}

	return builder.String()
}
