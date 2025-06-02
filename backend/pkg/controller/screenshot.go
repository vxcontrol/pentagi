package controller

import (
	"context"
	"fmt"
	"sync"

	"pentagi/pkg/database"
	"pentagi/pkg/graph/subscriptions"
)

type FlowScreenshotWorker interface {
	PutScreenshot(ctx context.Context, name, url string) (int64, error)
	GetScreenshot(ctx context.Context, screenshotID int64) (database.Screenshot, error)
}

type flowScreenshotWorker struct {
	db         database.Querier
	mx         *sync.Mutex
	flowID     int64
	containers map[int64]struct{}
	pub        subscriptions.FlowPublisher
}

func NewFlowScreenshotWorker(db database.Querier, flowID int64, pub subscriptions.FlowPublisher) FlowScreenshotWorker {
	return &flowScreenshotWorker{
		db:         db,
		mx:         &sync.Mutex{},
		flowID:     flowID,
		containers: make(map[int64]struct{}),
		pub:        pub,
	}
}

func (sw *flowScreenshotWorker) PutScreenshot(ctx context.Context, name, url string) (int64, error) {
	sw.mx.Lock()
	defer sw.mx.Unlock()

	screenshot, err := sw.db.CreateScreenshot(ctx, database.CreateScreenshotParams{
		Name:   sanitizeUTF8(name),
		Url:    sanitizeUTF8(url),
		FlowID: sw.flowID,
	})
	if err != nil {
		return 0, fmt.Errorf("failed to create screenshot: %w", err)
	}

	sw.pub.ScreenshotAdded(ctx, screenshot)

	return screenshot.ID, nil
}

func (sw *flowScreenshotWorker) GetScreenshot(ctx context.Context, screenshotID int64) (database.Screenshot, error) {
	screenshot, err := sw.db.GetScreenshot(ctx, screenshotID)
	if err != nil {
		return database.Screenshot{}, fmt.Errorf("failed to get screenshot: %w", err)
	}

	return screenshot, nil
}
