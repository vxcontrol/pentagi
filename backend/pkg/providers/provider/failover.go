package provider

import (
	"context"
	"errors"

	"pentagi/pkg/providers/pconfig"
	"pentagi/pkg/templates"

	"github.com/vxcontrol/langchaingo/llms"
	"github.com/vxcontrol/langchaingo/llms/streaming"
)

// SwitchNotifier is told that a call moved from one provider to another. It runs
// on the calling goroutine, so an implementation must not block.
type SwitchNotifier func(ctx context.Context, from, to ProviderName, err error)

type failoverProvider struct {
	Provider

	backup   Provider
	notifier SwitchNotifier
}

// WithFailover returns a provider that repeats a failed call on backup.
// Everything the interface exposes beyond the calling methods and the tool
// call ID template keeps answering for primary: a flow is still described by
// the provider it was created with.
func WithFailover(primary, backup Provider, notifier SwitchNotifier) Provider {
	if backup == nil || primary == nil || primary.Name() == backup.Name() {
		return primary
	}

	return &failoverProvider{Provider: primary, backup: backup, notifier: notifier}
}

func switchable(ctx context.Context, err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return false
	}

	return ctx.Err() == nil
}

func (f *failoverProvider) switched(ctx context.Context, err error) {
	if f.notifier != nil {
		f.notifier(ctx, f.Name(), f.backup.Name(), err)
	}
}

func (f *failoverProvider) Call(
	ctx context.Context, opt pconfig.ProviderOptionsType, prompt string,
) (string, error) {
	result, err := f.Provider.Call(ctx, opt, prompt)
	if !switchable(ctx, err) {
		return result, err
	}

	f.switched(ctx, err)

	return f.backup.Call(ctx, opt, prompt)
}

func (f *failoverProvider) CallEx(
	ctx context.Context,
	opt pconfig.ProviderOptionsType,
	chain []llms.MessageContent,
	streamCb streaming.Callback,
) (*llms.ContentResponse, error) {
	resp, err := f.Provider.CallEx(ctx, opt, chain, streamCb)
	if !switchable(ctx, err) {
		return resp, err
	}

	f.switched(ctx, err)

	return f.backup.CallEx(ctx, opt, chain, streamCb)
}

func (f *failoverProvider) CallWithTools(
	ctx context.Context,
	opt pconfig.ProviderOptionsType,
	chain []llms.MessageContent,
	tools []llms.Tool,
	streamCb streaming.Callback,
) (*llms.ContentResponse, error) {
	resp, err := f.Provider.CallWithTools(ctx, opt, chain, tools, streamCb)
	if !switchable(ctx, err) {
		return resp, err
	}

	f.switched(ctx, err)

	return f.backup.CallWithTools(ctx, opt, chain, tools, streamCb)
}

func (f *failoverProvider) CallWithExtraOptions(
	ctx context.Context,
	opt pconfig.ProviderOptionsType,
	chain []llms.MessageContent,
	tools []llms.Tool,
	streamCb streaming.Callback,
	extra ...llms.CallOption,
) (*llms.ContentResponse, error) {
	resp, err := f.Provider.CallWithExtraOptions(ctx, opt, chain, tools, streamCb, extra...)
	if !switchable(ctx, err) {
		return resp, err
	}

	f.switched(ctx, err)

	return f.backup.CallWithExtraOptions(ctx, opt, chain, tools, streamCb, extra...)
}

// GetToolCallIDTemplate is detected by calling the model when a flow or an
// assistant is created, so without a switch here a primary that is down
// fails the creation before any agent call could fall over.
func (f *failoverProvider) GetToolCallIDTemplate(ctx context.Context, prompter templates.Prompter) (string, error) {
	template, err := f.Provider.GetToolCallIDTemplate(ctx, prompter)
	if !switchable(ctx, err) {
		return template, err
	}

	f.switched(ctx, err)

	return f.backup.GetToolCallIDTemplate(ctx, prompter)
}
