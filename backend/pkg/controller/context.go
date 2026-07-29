package controller

import (
	"context"
	"errors"
	"fmt"

	"pentagi/pkg/config"
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

func wrapErrorEndSpan(ctx context.Context, span langfuse.Span, msg string, err error) error {
	logrus.WithContext(ctx).WithError(err).Error(msg)
	err = fmt.Errorf("%s: %w", msg, err)
	span.End(
		langfuse.WithSpanStatus(err.Error()),
		langfuse.WithSpanLevel(langfuse.ObservationLevelError),
	)
	return err
}

// tenantMeta adds a tenant_id key to trace metadata when a tenant is configured.
// The key is absent otherwise, so the emitted payload is byte-identical to what
// a single-instance deployment sends today.
func tenantMeta(cfg *config.Config, md langfuse.Metadata) langfuse.Metadata {
	if cfg == nil || !cfg.HasTenant() {
		return md
	}

	if md == nil {
		md = make(langfuse.Metadata)
	}

	md["tenant_id"] = cfg.TenantID

	return md
}

// tenantUserID namespaces the Langfuse "user" dimension. Instances share the
// seeded admin@pentagi.com account, so without this every tenant's traces would
// collapse onto one Langfuse user. Returns the address unchanged in
// single-instance mode.
func tenantUserID(cfg *config.Config, mail string) string {
	return cfg.TenantUserID(mail)
}

// tenantTags appends a "tenant:<id>" tag so traces from several instances
// sharing one Langfuse project stay filterable. Returns exactly the base tags in
// single-instance mode, so the emitted payload is unchanged.
func tenantTags(cfg *config.Config, base ...string) []string {
	return cfg.TenantTags(base...)
}
